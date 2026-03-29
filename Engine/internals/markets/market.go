package markets

import (
	"context"
	"log/slog"

	"github.com/go-redis/redis/v8"
	orderbooks "github.com/raiashpanda007/rivon/engine/internals/Orderbooks"
	pubsub "github.com/raiashpanda007/rivon/engine/internals/PubSub"
	heap "github.com/raiashpanda007/rivon/engine/internals/utils"
	tradestream "github.com/raiashpanda007/rivon/engine/internals/utils/TradeStream"
)

type OrderMessages struct {
	OrderId   string `json:"orderId"`
	UserId    string `json:"userId"`
	MarketId  string `json:"marketId"`
	Price     int    `json:"price"`
	Quantity  int    `json:"quantity"`
	OrderType string `json:"orderType"`
}

func StarMarketProcess(ctx context.Context, ch chan OrderMessages, tradeRedis *redis.Client, pubsubSvc pubsub.PubSubService) {
	var lastTradeId = ""
	var bids map[int][]orderbooks.Order
	var asks map[int][]orderbooks.Order
	var askHeap heap.MinHeap
	var bidHeap heap.MaxHeap
	var currentPrice int = 0
	OrderBook := orderbooks.NewOrderBook(lastTradeId, bids, asks, &askHeap, &bidHeap, currentPrice)

	for order := range ch {
		slog.Info("order received", "orderId", order.OrderId, "userId", order.UserId, "marketId", order.MarketId, "price", order.Price, "qty", order.Quantity, "side", order.OrderType)
		var inputOrder = orderbooks.Order{
			Id:       order.OrderId,
			Quantity: order.Quantity,
			Side:     orderbooks.OrderSide(order.OrderType),
			Price:    order.Price,
			UserId:   order.UserId,
			Filled:   0,
		}

		Fills, executedQty, err := OrderBook.AddOrder(inputOrder, order.Price)

		if err != nil {
			slog.Error("Error in Adding new order :: ", "err :: ", err)
		}

		tradestream.TradeRedisStreamPublisher(ctx, tradestream.ORDER_UPDATED, order.OrderId, order.MarketId, Fills, executedQty, order.Price, tradeRedis)

		pubsubOrderMessage := pubsub.PubSubOrderMessage{
			OrderId:          order.OrderId,
			Fills:            Fills,
			ExecutedQuantity: executedQty,
			MessageType:      pubsub.ORDER_UPDATE,
		}

		err = pubsubSvc.Api().Publish(pubsubOrderMessage)

		if err != nil {
			slog.Error("Unable to publish to api-server", err)
			panic("Unable publish to api-server :: " + err.Error())
		}

		// We have to cancel mechanism here too.

		bids, asks := OrderBook.GetDepth()
		snap := OrderBook.GetSnapshot()
		slog.Info("orderbook state",
			"bidLevels", len(bids),
			"askLevels", len(asks),
			"currentPrice", snap.CurrentPrice,
			"lastTradeId", snap.LastTradeId,
		)
	}

}
