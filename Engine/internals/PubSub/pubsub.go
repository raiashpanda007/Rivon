package pubsub

import (
	"context"

	"github.com/go-redis/redis/v8"
)

type PubSubService interface {
	Api() ApiPubSubServices
}

type pubsubStruct struct {
	api ApiPubSubServices
}

func InitPubSub(ctx context.Context, apiPubSubRedisClient *redis.Client) PubSubService {
	apiPubSub := InitApiPubSub(ctx, apiPubSubRedisClient)

	return &pubsubStruct{
		api: apiPubSub,
	}
}

func (p *pubsubStruct) Api() ApiPubSubServices {
	return p.api
}
