package pubsub

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/go-redis/redis/v8"
	"github.com/raiashpanda007/rivon/internals/registry"
	"github.com/raiashpanda007/rivon/internals/types"
)

type Pubsub interface {
	Subscribe(ctx context.Context, stream string) error
}

type pubsub struct {
	registry    *registry.Registry
	redisClient *redis.Client
}

func InitPubSub(redisClient *redis.Client, reg *registry.Registry) Pubsub {
	return &pubsub{
		redisClient: redisClient,
		registry:    reg,
	}
}

func (r *pubsub) Subscribe(ctx context.Context, stream string) error {
	sub := r.redisClient.Subscribe(ctx, stream)
	if _, err := sub.Receive(ctx); err != nil {
		slog.Error("PubSub: failed to subscribe", "error", err)
		return err
	}
	slog.Info("PubSub: subscribed", "stream", stream)

	ch := sub.Channel()
	for {
		select {
		case msg, ok := <-ch:
			if !ok {
				slog.Info("PubSub: channel closed")
				return nil
			}
			var orderMsg types.PubSubOrderMessage
			if err := json.Unmarshal([]byte(msg.Payload), &orderMsg); err != nil {
				slog.Error("PubSub: unmarshal failed", "payload", msg.Payload, "error", err)
				continue
			}
			r.registry.Resolve(orderMsg.OrderId, types.FillResult{
				OrderId:          orderMsg.OrderId,
				ExecutedQuantity: orderMsg.ExecutedQuantity,
				Fills:            orderMsg.Fills,
			})
		case <-ctx.Done():
			slog.Info("PubSub: context cancelled, stopping subscriber")
			return ctx.Err()
		}
	}
}
