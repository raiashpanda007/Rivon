package redis

import (
	"context"
	"log/slog"

	"github.com/go-redis/redis/v8"
)

func InitWalletMapRedisClient(ctx context.Context, userMapWalletRedisURL string) (*redis.Client, error) {

	walletMapRedis := redis.NewClient(&redis.Options{
		Addr: userMapWalletRedisURL,
	})

	_, err := walletMapRedis.Ping(ctx).Result()
	if err != nil {
		slog.Error("Unable to ping wallet map redis pool", slog.Any("Error :: ", err))
		return nil, err
	}

	return walletMapRedis, nil

}
