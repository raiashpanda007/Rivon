package pubsub

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/go-redis/redis/v8"
)

type Pubsub interface {
	Subscribe(ctx context.Context, stream string) error
}
type pubsub struct {
	OrderResponseMap map[string]http.ResponseWriter
	redisClient      *redis.Client
}

func InitPubSub(redisClient *redis.Client) Pubsub {

	return &pubsub{
		redisClient: redisClient,
	}

}

func (r *pubsub) Subscribe(ctx context.Context, stream string) error {
	sub := r.redisClient.Subscribe(ctx, stream)
	_, err := sub.Receive(ctx)
	if err != nil {
		slog.Error("Unable data read from subscriber :: ", err)
		return err
	}

	slog.Info("Successfully subscribed to read response for orders")
	ch := sub.Channel()

	for msg := range ch {
		slog.Debug("Subscriber recieved message :: ", msg)
	}

	return nil

}
