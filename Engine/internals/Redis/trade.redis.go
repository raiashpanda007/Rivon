package redis

import (
	"context"
	"log/slog"

	"github.com/go-redis/redis/v8"
)

func Init_Trade_Redis(tradeOrderRedis string) (*redis.Client, error) {
	ctx := context.Background()

	slog.Info("CONNECTING TO ORDER REDIS DB :: ")

	TradeRedis := redis.NewClient(&redis.Options{
		Addr: tradeOrderRedis,
	})

	_, err := TradeRedis.Ping(ctx).Result()

	if err != nil {
		slog.Error("REDIS Ping failed ", slog.Any("ERROR :: ", err))
		return nil, err
	}

	slog.Info("CONNECTED TO ORDER REDIS DB :: ")
	return TradeRedis, nil

}
