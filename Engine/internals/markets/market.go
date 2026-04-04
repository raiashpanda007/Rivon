package markets

import (
	"context"
	"log/slog"
	"math/rand"
	"time"

	"github.com/go-redis/redis/v8"
	orderbooks "github.com/raiashpanda007/rivon/engine/internals/Orderbooks"
	pubsub "github.com/raiashpanda007/rivon/engine/internals/PubSub"
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

func StarMarketProcess(ctx context.Context, ch chan OrderMessages, tradeRedis *redis.Client, pubsubSvc pubsub.PubSubService, marketId string) {

	var lastTradeId = ""
	var lastOrderId = ""
	var bids map[int][]orderbooks.Order
	var asks map[int][]orderbooks.Order
	var askHeap heap.MinHeap
	var bidHeap heap.MaxHeap
	var currentPrice int = 0

	OrderBook := orderbooks.NewOrderBook(lastTradeId, lastOrderId, "", bids, asks, &askHeap, &bidHeap, currentPrice)

	// Snapshot helper

	baseInterval := 1 * time.Minute

	// first timer
	timer := time.NewTimer(baseInterval + time.Duration(rand.Intn(10))*time.Second)

	for {
		select {

		// ---------------- ORDER PROCESSING ----------------
		case order := <-ch:

			slog.Info("order received",
				"orderId", order.OrderId,
				"marketId", order.MarketId,
				"streamID", order.StreamId,
			)

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

			// ✅ IMPORTANT: update streamId AFTER success
			OrderBook.LastStreamId = order.StreamId

			// async trade publish
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

			// async pubsub
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

		// ---------------- SNAPSHOT ----------------
		case <-timer.C:

			slog.Info("Taking snapshot", "marketId", marketId)

			// safe: same goroutine

			// async save
			go snapshots.SaveSnapShot(marketId, OrderBook)

			// reset timer with jitter
			timer.Reset(baseInterval + time.Duration(rand.Intn(10))*time.Second)
		}
	}
}
