package pubsub

import (
	"context"

	"github.com/go-redis/redis/v8"
)

type PubSubService interface {
	ApiPubSubServices
}

type pubsubStruct struct {
	ApiPubSubServices
}

func InitPubSub(ctx context.Context, apiPubSubRedisClient *redis.Client) PubSubService {
	apiPubSub := InitApiPubSub(ctx, apiPubSubRedisClient)

	return &pubsubStruct{
		ApiPubSubServices: apiPubSub,
	}
}
