package markets

import (
	"context"
	"log/slog"
	"math/rand"
	"time"

	"github.com/go-redis/redis/v8"
	orderbooks "github.com/raiashpanda007/rivon/engine/internals/Orderbooks"
	pubsub "github.com/raiashpanda007/rivon/engine/internals/PubSub"
	redisStream "github.com/raiashpanda007/rivon/engine/internals/Redis"
	snapshots "github.com/raiashpanda007/rivon/engine/internals/Snapshots"
	heap "github.com/raiashpanda007/rivon/engine/internals/utils"
	tradestream "github.com/raiashpanda007/rivon/engine/internals/utils/TradeStream"
	wsmessagestypes "github.com/raiashpanda007/rivon/engine/internals/utils/WsMessagesTypes"
)

// copyDepth returns a shallow copy of a depth map safe to hand off to a goroutine.
func copyDepth(depth map[int]int) map[int]int {
	cp := make(map[int]int, len(depth))
	for k, v := range depth {
		cp[k] = v
	}
	return cp
}

// toPublicFills strips user identifiers from fills before broadcasting.
func toPublicFills(fills []orderbooks.Fills) []wsmessagestypes.PublicFill {
	result := make([]wsmessagestypes.PublicFill, len(fills))
	for i, f := range fills {
		result[i] = wsmessagestypes.PublicFill{
			Price:    f.Price,
			Quantity: f.Quantity,
			TradeId:  f.TradeId,
		}
	}
	return result
}

// pushOrderbookUpdate sends current depth + fills to wsOutChannel without blocking the caller.
func pushOrderbookUpdate(wsOutChannel chan wsmessagestypes.WSOutMessageStruct, bidDepth, askDepth map[int]int, currentPrice int, fills []orderbooks.Fills) {
	publicFills := toPublicFills(fills)
	go func() {
		wsOutChannel <- wsmessagestypes.WSOutMessageStruct{
			MessageType: wsmessagestypes.ORDERBOOK_UPDATE,
			Payload: wsmessagestypes.OrderbookUpdatePayload{
				BidDepth:     bidDepth,
				AskDepth:     askDepth,
				CurrentPrice: currentPrice,
				Fills:        publicFills,
			},
		}
	}()
}

type OrderMessages struct {
	OrderId   string
	UserId    string
	MarketId  string
	Price     int
	Quantity  int
	OrderType string
	StreamId  string
}

func StarMarketProcess(ctx context.Context, ch chan OrderMessages, tradeRedis *redis.Client, pubsubSvc pubsub.PubSubService, marketId string, orderRedis *redis.Client, wsInChannel chan wsmessagestypes.WSInMessageStruct, wsOutChannel chan wsmessagestypes.WSOutMessageStruct) {

	// wsOut publisher — reads from wsOutChannel and publishes to Redis PubSub WS_OUT_<marketId>
	// in a continuous loop so the WS server always receives the latest orderbook state.
	go func() {
		for {
			select {
			case msg := <-wsOutChannel:
				if err := pubsubSvc.WSOut().Publish(marketId, msg); err != nil {
					slog.Error("wsOut publish failed", "marketId", marketId, "err", err)
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	// Restore orderbook from latest snapshot, or start fresh.
	var OrderBook orderbooks.OrderBook
	if snap, ok := snapshots.ReadLastSnapShotForMarket(marketId); ok {
		OrderBook = *snap
	} else {
		OrderBook = orderbooks.NewOrderBook("", "", "", nil, nil, heap.NewMinHeap(), heap.NewMaxHeap(), 0)
	}

	replayStartId := OrderBook.LastStreamId
	if replayStartId == "" {
		replayStartId = "0"
	}

	pivotOrderId, hasPivot := redisStream.ReadLastTradeOrderIdForMarket(ctx, tradeRedis, marketId)

	replayMsgs, replayErr := redisStream.ReplayOrderStream(ctx, orderRedis, "ORDERS_"+marketId, replayStartId)
	if replayErr != nil {
		slog.Error("Replay failed, starting from snapshot state", "marketId", marketId, "err", replayErr)
	} else if len(replayMsgs) > 0 {
		// Determine whether the pivot actually falls inside this replay window.
		pivotInReplay := false
		if hasPivot {
			for _, m := range replayMsgs {
				if m.OrderId == pivotOrderId {
					pivotInReplay = true
					break
				}
			}
		}

		silent := pivotInReplay // start silent only when the pivot is in-window
		silentCount, normalCount := 0, 0

		for _, msg := range replayMsgs {
			OrderBook.LastStreamId = msg.StreamId

			if msg.OrderType == "CANCEL_ORDER" {
				OrderBook.CancelOrder(msg.OrderId, msg.UserId)
				if !silent {
					normalCount++
					go tradestream.TradeRedisStreamPublisher(
						ctx,
						tradestream.CANCELLED_ORDER,
						msg.OrderId,
						marketId,
						OrderBook.LastOrderId,
						OrderBook.LastTradeId,
						nil,
						0,
						0,
						tradeRedis,
					)
					go pubsubSvc.Api().Publish(pubsub.PubSubOrderMessage{
						OrderId:     msg.OrderId,
						MessageType: pubsub.ORDER_CANCEL,
					})
				} else {
					silentCount++
				}
				if silent && msg.OrderId == pivotOrderId {
					silent = false
				}
				continue
			}

			inputOrder := orderbooks.Order{
				Id:       msg.OrderId,
				Quantity: msg.Quantity,
				Side:     orderbooks.OrderSide(msg.OrderType),
				Price:    msg.Price,
				UserId:   msg.UserId,
				Filled:   0,
				StreamId: msg.StreamId,
			}

			fills, executedQty, err := OrderBook.AddOrder(inputOrder, msg.Price)
			if err != nil {
				slog.Error("Replay AddOrder error", "orderId", msg.OrderId, "err", err)
				continue
			}

			if !silent {
				normalCount++
				go tradestream.TradeRedisStreamPublisher(
					ctx,
					tradestream.ORDER_UPDATED,
					msg.OrderId,
					marketId,
					OrderBook.LastOrderId,
					OrderBook.LastTradeId,
					fills,
					executedQty,
					msg.Price,
					tradeRedis,
				)
				go pubsubSvc.Api().Publish(pubsub.PubSubOrderMessage{
					OrderId:          msg.OrderId,
					Fills:            fills,
					ExecutedQuantity: executedQty,
					MessageType:      pubsub.ORDER_UPDATE,
				})
			} else {
				silentCount++
			}

			// After processing the pivot, switch to normal mode for all subsequent messages.
			if silent && msg.OrderId == pivotOrderId {
				silent = false
			}
		}

	} else {
	}
	// --- end replay ---

	baseInterval := 1 * time.Minute
	timer := time.NewTimer(baseInterval + time.Duration(rand.Intn(10))*time.Second)

	for {
		select {

		case order := <-ch:
			slog.Info("processor received", "orderId", order.OrderId)
			t0 := time.Now()
			OrderBook.LastStreamId = order.StreamId

			if order.OrderType == "CANCEL_ORDER" {
				_ = OrderBook.CancelOrder(order.OrderId, order.UserId)

				go tradestream.TradeRedisStreamPublisher(
					ctx,
					tradestream.CANCELLED_ORDER,
					order.OrderId,
					order.MarketId,
					OrderBook.LastOrderId,
					OrderBook.LastTradeId,
					nil,
					0,
					0,
					tradeRedis,
				)
				go pubsubSvc.Api().Publish(pubsub.PubSubOrderMessage{
					OrderId:     order.OrderId,
					MessageType: pubsub.ORDER_CANCEL,
				})
				pushOrderbookUpdate(wsOutChannel, copyDepth(OrderBook.BidDepth), copyDepth(OrderBook.AskDepth), OrderBook.CurrentPrice, nil)
				continue
			}

			inputOrder := orderbooks.Order{
				Id:       order.OrderId,
				Quantity: order.Quantity,
				Side:     orderbooks.OrderSide(order.OrderType),
				Price:    order.Price,
				UserId:   order.UserId,
				Filled:   0,
				StreamId: order.StreamId,
			}

			Fills, executedQty, err := OrderBook.AddOrder(inputOrder, order.Price)
			if err != nil {
				slog.Error("Error adding order", "err", err)
				continue
			}
			matchDur := time.Since(t0)

			pubsubStart := time.Now()
			orderId := order.OrderId
			go func() {
				pubsubSvc.Api().Publish(pubsub.PubSubOrderMessage{
					OrderId:          orderId,
					Fills:            Fills,
					ExecutedQuantity: executedQty,
					MessageType:      pubsub.ORDER_UPDATE,
				})
				slog.Info("latency",
					"orderId", orderId,
					"matchUs", matchDur.Microseconds(),
					"pubsubUs", time.Since(pubsubStart).Microseconds(),
					"totalUs", time.Since(t0).Microseconds(),
				)
			}()

			go tradestream.TradeRedisStreamPublisher(
				ctx,
				tradestream.ORDER_UPDATED,
				order.OrderId,
				order.MarketId,
				OrderBook.LastOrderId,
				OrderBook.LastTradeId,
				Fills,
				executedQty,
				order.Price,
				tradeRedis,
			)
			pushOrderbookUpdate(wsOutChannel, copyDepth(OrderBook.BidDepth), copyDepth(OrderBook.AskDepth), OrderBook.CurrentPrice, Fills)

		case <-timer.C:

			snap := OrderBook.GetSnapshot()
			go snapshots.SaveSnapShot(marketId, snap)

			timer.Reset(baseInterval + time.Duration(rand.Intn(10))*time.Second)

		case wsInMsg := <-wsInChannel:
			switch wsInMsg.MessageType {
			case wsmessagestypes.ORDERBOOK_SUBSCIRBE:
				bidDepth := copyDepth(OrderBook.BidDepth)
				askDepth := copyDepth(OrderBook.AskDepth)
				currentPrice := OrderBook.CurrentPrice
				userId := wsInMsg.UserId
				go func() {
					wsOutChannel <- wsmessagestypes.WSOutMessageStruct{
						MessageType: wsmessagestypes.ORDERBOOK_DATA,
						Payload: wsmessagestypes.OrderbookPayload{
							BidDepth:     bidDepth,
							AskDepth:     askDepth,
							CurrentPrice: currentPrice,
						},
						UserId: userId,
					}
				}()

			case wsmessagestypes.DEPTH_SUBSCRIBE:
				bidDepth := copyDepth(OrderBook.BidDepth)
				askDepth := copyDepth(OrderBook.AskDepth)
				userId := wsInMsg.UserId
				go func() {
					wsOutChannel <- wsmessagestypes.WSOutMessageStruct{
						MessageType: wsmessagestypes.DEPTH_DATA,
						Payload: wsmessagestypes.DepthPayload{
							BidDepth: bidDepth,
							AskDepth: askDepth,
						},
						UserId: userId,
					}
				}()
			}

		}

	}
}
