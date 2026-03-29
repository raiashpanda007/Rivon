package pubsub

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/go-redis/redis/v8"
	orderbooks "github.com/raiashpanda007/rivon/engine/internals/Orderbooks"
)

type PubSubOrderMessageType string

const (
	ORDER_UPDATE PubSubOrderMessageType = "UPDATE_ORDER"
	ORDER_CANCEL PubSubOrderMessageType = "CANCEL_ORDER"
)

type PubSubOrderMessage struct {
	OrderId          string                 `json:"orderId"`
	Fills            []orderbooks.Fills     `json:"fills"`
	ExecutedQuantity int                    `json:"executedQty"`
	MessageType      PubSubOrderMessageType `json:"type"`
}

type ApiPubSubServices interface {
	Publish(message PubSubOrderMessage) error
	Subscribe() (any, error)
}

type apiPubSubStruct struct {
	ctx         context.Context
	redisClient *redis.Client
}

func InitApiPubSub(ctx context.Context, apiPubSubRedisClient *redis.Client) ApiPubSubServices {

	return &apiPubSubStruct{
		ctx:         ctx,
		redisClient: apiPubSubRedisClient,
	}

}

func (r *apiPubSubStruct) Publish(message PubSubOrderMessage) error {

	data, err := json.Marshal(message)
	if err != nil {
		return err
	}

	err = r.redisClient.Publish(r.ctx, "ORDERS", string(data)).Err()

	if err != nil {
		return err
	}

	return nil
}

func (r *apiPubSubStruct) Subscribe() (any, error) {
	sub := r.redisClient.Subscribe(r.ctx, "ORDERS")
	_, err := sub.Receive(r.ctx)
	if err != nil {
		slog.Error("Unable to subscribe to api-engine ", err)
		return nil, err
	}

	ch := sub.Channel()

	for msg := range ch {
		fmt.Println("Received:", msg.Payload)
	}

	return nil, nil
}
