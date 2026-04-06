package redis

import (
	"context"
	"log/slog"

	"github.com/go-redis/redis/v8"
)

func Init_WS_PubSub_Redis(wsPubSubRedisURL string) (*redis.Client, error) {

	ctx := context.Background()

	slog.Info("CONNECTING TO WS REDIS FOR PUBSUB :: ")

	WSPubSubRedis := redis.NewClient(&redis.Options{
		Addr: wsPubSubRedisURL,
	})

	_, err := WSPubSubRedis.Ping(ctx).Result()

	if err != nil {
		slog.Error("REDIS Ping failed ", slog.Any("ERROR :: ", err))
		return nil, err
	}

	slog.Info("CONNECTED TO WS PUBSUB :: ")
	return WSPubSubRedis, nil
}
