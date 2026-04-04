package redis

import (
	"context"
	"log/slog"

	"github.com/go-redis/redis/v8"
)

func Init_Order_Redis(orderRedisURL string) (*redis.Client, error) {

	ctx := context.Background()

	slog.Info("CONNECTING TO ORDER REDIS DB :: ")

	OrderRedis := redis.NewClient(&redis.Options{
		Addr: orderRedisURL,
	})

	_, err := OrderRedis.Ping(ctx).Result()

	if err != nil {
		slog.Error("REDIS Ping failed ", slog.Any("ERROR :: ", err))
		return nil, err
	}

	slog.Info("CONNECTED TO ORDER REDIS DB :: ")
	return OrderRedis, nil
}
