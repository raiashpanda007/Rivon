package pubsub

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/go-redis/redis/v8"
	wsmessagestypes "github.com/raiashpanda007/rivon/engine/internals/utils/WsMessagesTypes"
)

type WSOutPubSubServices interface {
	Publish(marketID string, msg wsmessagestypes.WSOutMessageStruct) error
}

type wsOutPubSubStruct struct {
	ctx         context.Context
	redisClient *redis.Client
}

func InitWSOutPubSub(ctx context.Context, wsOutPubSubRedisClient *redis.Client) WSOutPubSubServices {
	return &wsOutPubSubStruct{
		ctx:         ctx,
		redisClient: wsOutPubSubRedisClient,
	}
}

func (r *wsOutPubSubStruct) Publish(marketID string, msg wsmessagestypes.WSOutMessageStruct) error {
	data, err := json.Marshal(msg)
	if err != nil {
		slog.Error("failed to marshal wsOut message", "marketId", marketID, "err", err)
		return err
	}
	return r.redisClient.Publish(r.ctx, "WS_OUT_"+marketID, string(data)).Err()
}
