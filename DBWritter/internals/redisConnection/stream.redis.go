package redisconnection

import (
	"context"
	"log/slog"

	"github.com/go-redis/redis/v8"
)

func ConnectToTradeStream(tradeStreamURL string) (*redis.Client, error) {

	ctx := context.Background()
	TradeRedis := redis.NewClient(&redis.Options{
		Addr: tradeStreamURL,
	})
	_, err := TradeRedis.Ping(ctx).Result()
	if err != nil {
		slog.Error("REDIS Ping failed ", slog.Any("ERROR :: ", err))
		return nil, err

	}
	return TradeRedis, nil
}
