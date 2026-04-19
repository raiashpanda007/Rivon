package pubsub

import (
	"context"

	"github.com/go-redis/redis/v8"
)

type PubSubService interface {
	Api() ApiPubSubServices
	WSIn() WSInPubSubServices
	WSOut() WSOutPubSubServices
}

type pubsubStruct struct {
	api   ApiPubSubServices
	wsIn  WSInPubSubServices
	wsOut WSOutPubSubServices
}

func InitPubSub(ctx context.Context, apiPubSubRedisClient *redis.Client, wsInPubSubRedisClient *redis.Client, wsOutPubSubRedisClient *redis.Client) PubSubService {
	apiPubSub := InitApiPubSub(ctx, apiPubSubRedisClient)
	wsInPubSub := InitWSInPubSub(ctx, wsInPubSubRedisClient)
	wsOutPubSub := InitWSOutPubSub(ctx, wsOutPubSubRedisClient)
	return &pubsubStruct{
		api:   apiPubSub,
		wsIn:  wsInPubSub,
		wsOut: wsOutPubSub,
	}
}

func (p *pubsubStruct) Api() ApiPubSubServices {
	return p.api
}

func (p *pubsubStruct) WSIn() WSInPubSubServices {
	return p.wsIn
}

func (p *pubsubStruct) WSOut() WSOutPubSubServices {
	return p.wsOut
}
