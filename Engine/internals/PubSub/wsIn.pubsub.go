package pubsub

import (
	"context"
	"encoding/json"
	"log"

	"github.com/go-redis/redis/v8"
	wsmessagestypes "github.com/raiashpanda007/rivon/engine/internals/utils/WsMessagesTypes"
)

type WSInPubSubServices interface {
	Subscribe(marketID string, wsInChan chan wsmessagestypes.WSInMessageStruct) error
}

type wsInPubSubStruct struct {
	ctx         context.Context
	redisClient *redis.Client
}

func InitWSInPubSub(ctx context.Context, wsInPubSubRedisClient *redis.Client) WSInPubSubServices {

	return &wsInPubSubStruct{
		ctx:         ctx,
		redisClient: wsInPubSubRedisClient,
	}
}

func (r *wsInPubSubStruct) Subscribe(marketID string, wsInChan chan wsmessagestypes.WSInMessageStruct) error {
	pubsub := r.redisClient.Subscribe(r.ctx, "MARKET_"+marketID)

	_, err := pubsub.Receive(r.ctx)
	if err != nil {
		return err
	}

	go func() {
		defer pubsub.Close()

		ch := pubsub.Channel()

		for msg := range ch {
			var parsed wsmessagestypes.WSInMessageStruct

			err := json.Unmarshal([]byte(msg.Payload), &parsed)
			if err != nil {
				log.Println("failed to unmarshal:", err)
				continue
			}

			wsInChan <- parsed
		}
	}()

	return err
}
