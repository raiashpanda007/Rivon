package tradestream

import (
	"context"
	"encoding/json"
	"github.com/go-redis/redis/v8"
	orderbooks "github.com/raiashpanda007/rivon/engine/internals/Orderbooks"
	"log/slog"
)

type TradeStreamTypes string

const (
	ORDER_UPDATED   TradeStreamTypes = "order_updated"
	CANCELLED_ORDER TradeStreamTypes = "order_cancelled"
)

func TradeRedisStreamPublisher(ctx context.Context, tradeType TradeStreamTypes, orderId, marketId, lastOrderId, lastTradeId string, fills []orderbooks.Fills, executedQty int, price int, tradeRedisClient *redis.Client) {

	fillsJSON, err := json.Marshal(fills)
	if err != nil {
		slog.Error("Unable to marshal fills", "error", err)
		return
	}

	_, err = tradeRedisClient.XAdd(ctx, &redis.XAddArgs{
		Stream: "TRADES",
		Values: map[string]any{
			"tradeType":   string(tradeType),
			"marketId":    marketId,
			"fills":       string(fillsJSON),
			"executedQty": executedQty,
			"price":       price,
			"orderId":     orderId,
			"lastOrderId": lastOrderId,
			"lastTradeId": lastTradeId,
		},
	}).Result()

	if err != nil {
		slog.Error("Unable to save the update on the stream", "error :: ", err)
	}

	slog.Info("Trade successfully recorded")

}
