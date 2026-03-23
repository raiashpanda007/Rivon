package tradestream

import (
	"context"
	"github.com/go-redis/redis/v8"
	orderbooks "github.com/raiashpanda007/rivon/engine/internals/Orderbooks"
	"log/slog"
)

type TradeStreamTypes string

const (
	ORDER_UPDATED   TradeStreamTypes = "order_updated"
	CANCELLED_ORDER TradeStreamTypes = "order_cancelled"
)

func TradeRedisStreamPublisher(ctx context.Context, tradeType TradeStreamTypes, orderId, marketId string, fills []orderbooks.Fills, executedQty int, price int, tradeRedisClient *redis.Client) {

	_, err := tradeRedisClient.XAdd(ctx, &redis.XAddArgs{
		Stream: "TRADES",
		Values: map[string]any{
			"tradeType":   tradeType,
			"marketId":    marketId,
			"fills":       fills,
			"executedQty": executedQty,
			"price":       price,
			"orderId":     orderId,
		},
	}).Result()

	if err != nil {
		slog.Error("Unable to save the update on the stream", "error :: ", err)
	}

	slog.Info("Trade successfully recorded")

}
