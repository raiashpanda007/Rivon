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
)

type OrderMessages struct {
	OrderId   string
	UserId    string
	MarketId  string
	Price     int
	Quantity  int
	OrderType string
	StreamId  string
}

func StarMarketProcess(ctx context.Context, ch chan OrderMessages, tradeRedis *redis.Client, pubsubSvc pubsub.PubSubService, marketId string, orderRedis *redis.Client) {

	// Restore orderbook from latest snapshot, or start fresh.
	var OrderBook orderbooks.OrderBook
	if snap, ok := snapshots.ReadLastSnapShotForMarket(marketId); ok {
		OrderBook = *snap
		slog.Info("Orderbook restored from snapshot",
			"marketId", marketId,
			"currentPrice", OrderBook.CurrentPrice,
			"lastStreamId", OrderBook.LastStreamId,
		)
	} else {
		OrderBook = orderbooks.NewOrderBook("", "", "", nil, nil, heap.NewMinHeap(), heap.NewMaxHeap(), 0)
		slog.Info("Started fresh orderbook", "marketId", marketId)
	}

	replayStartId := OrderBook.LastStreamId
	if replayStartId == "" {
		replayStartId = "0"
	}

	// --- Crash-recovery replay ---
	// Find the orderId of the last trade that was actually executed for this market.
	// This is our "pivot": everything up to it was already processed (suppress re-publish),
	// everything after it is new (publish trades + pubsub normally).
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

		slog.Info("Replay complete",
			"marketId", marketId,
			"silent", silentCount,
			"normal", normalCount,
			"lastStreamId", OrderBook.LastStreamId,
		)
	} else {
		slog.Info("No replay messages found", "marketId", marketId)
	}
	// --- end replay ---

	baseInterval := 1 * time.Minute
	timer := time.NewTimer(baseInterval + time.Duration(rand.Intn(10))*time.Second)

	for {
		select {

		case order := <-ch:

			slog.Info("order received",
				"orderId", order.OrderId,
				"marketId", order.MarketId,
				"streamID", order.StreamId,
				"orderType", order.OrderType,
			)

			OrderBook.LastStreamId = order.StreamId

			if order.OrderType == "CANCEL_ORDER" {
				cancelled := OrderBook.CancelOrder(order.OrderId, order.UserId)
				slog.Info("cancel order processed", "orderId", order.OrderId, "cancelled", cancelled)

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

			go pubsubSvc.Api().Publish(pubsub.PubSubOrderMessage{
				OrderId:          order.OrderId,
				Fills:            Fills,
				ExecutedQuantity: executedQty,
				MessageType:      pubsub.ORDER_UPDATE,
			})

			bids, asks := OrderBook.GetDepth()
			slog.Info("orderbook state",
				"bidLevels", len(bids),
				"askLevels", len(asks),
				"currentPrice", OrderBook.CurrentPrice,
				"lastOrderId", OrderBook.LastOrderId,
				"lastTradeId", OrderBook.LastTradeId,
				"lastStreamID", OrderBook.LastStreamId,
			)

		case <-timer.C:

			slog.Info("Taking snapshot", "marketId", marketId)

			snap := OrderBook.GetSnapshot()
			go snapshots.SaveSnapShot(marketId, snap)

			timer.Reset(baseInterval + time.Duration(rand.Intn(10))*time.Second)
		}
	}
}
