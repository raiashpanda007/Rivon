package redis

import (
	"context"
	"log/slog"

	"github.com/go-redis/redis/v8"
)

func Init_Api_PubSub_Redis(apiPubSubRedisURL string) (*redis.Client, error) {

	ctx := context.Background()

	slog.Info("CONNECTING TO ORDER REDIS DB :: ")

	ApiPubsSubRedis := redis.NewClient(&redis.Options{
		Addr: apiPubSubRedisURL,
	})

	_, err := ApiPubsSubRedis.Ping(ctx).Result()

	if err != nil {
		slog.Error("REDIS Ping failed ", slog.Any("ERROR :: ", err))
		return nil, err
	}

	slog.Info("CONNECTED TO API PUBSUB :: ")
	return ApiPubsSubRedis, nil
}
