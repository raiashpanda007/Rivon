package tradestream

import (
	"context"
	"log/slog"

	"github.com/go-redis/redis/v8"
	orderbooks "github.com/raiashpanda007/rivon/engine/internals/Orderbooks"
)

func TradeRedisStreamPublisher(ctx context.Context, orderId, marketId string, fills []orderbooks.Fills, executedQty int, price int, tradeRedisClient *redis.Client) {

	_, err := tradeRedisClient.XAdd(ctx, &redis.XAddArgs{
		Stream: "TRADES",
		Values: map[string]any{
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
