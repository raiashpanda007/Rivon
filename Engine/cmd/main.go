package main

import (
	"context"
	"log/slog"

	config "github.com/raiashpanda007/rivon/engine/internals/Config"
	database "github.com/raiashpanda007/rivon/engine/internals/Database"
	engine "github.com/raiashpanda007/rivon/engine/internals/Engine"
	pubsub "github.com/raiashpanda007/rivon/engine/internals/PubSub"
	redis "github.com/raiashpanda007/rivon/engine/internals/Redis"
)

func main() {
	slog.Info("TRADE ENGINE STARTED :: ")
	cfg := config.MustLoad()
	db, err := database.InitDb(cfg.PG_URL)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err != nil {
		slog.Error("ERROR :: IN CONNECTING TO PG DB :: ", slog.Any("ERROR :: ", err))
	}

	orderRedis, err := redis.Init_Order_Redis(cfg.ORDER_REDIS_URL)
	if err != nil {
		slog.Error("ERROR :: IN CONNECTION TO ORDER REDIS :: ", slog.Any("ERROR :: ", err))
	}

	tradeRedis, err := redis.Init_Trade_Redis(cfg.TRADE_REDIS_URL)

	if err != nil {
		slog.Error("ERROR :: IN CONNECTION TO TRADE REDIS :: ", slog.Any("ERROR :: ", err))
	}
	apiPubSubRedisClient, err := redis.Init_Api_PubSub_Redis(cfg.API_PUB_SUB_REDIS_URL)

	if err != nil {
		slog.Error("ERROR :: IN CONNECTION TO API PUB SUB REDIS :: ", slog.Any("ERROR :: ", err))
	}

	wsInPubSubRedisClient, err := redis.Init_WS_PubSub_Redis(cfg.WS_PUB_SUB_REDIS_URL)

	if err != nil {
		slog.Error("ERROR :: IN CONNECTION TO WS IN PUB SUB REDIS :: ", slog.Any("ERROR :: ", err))
	}

	wsOutPubSubRedisClient, err := redis.Init_WS_PubSub_Redis(cfg.WS_PUB_SUB_REDIS_URL)
	if err != nil {
		slog.Error("ERROR :: IN CONNECTION TO WS OUT PUB SUB REDIS :: ", slog.Any("ERROR :: ", err))
	}

	userWalletMapRedis, err := redis.InitWalletMapRedisClient(ctx, cfg.WALLET_USER_MAP)
	if err != nil {
		slog.Error("ERROR :: IN CONNECTION TO WALLET MAP REDIS :: ", slog.Any("ERROR :: ", err))
	}

	pubsubSvc := pubsub.InitPubSub(ctx, apiPubSubRedisClient, wsInPubSubRedisClient, wsOutPubSubRedisClient)

	if err := engine.InitEngine(ctx, orderRedis, tradeRedis, db, pubsubSvc, userWalletMapRedis); err != nil {
		slog.Error("ERROR :: FAILED TO INITIALIZE ENGINE :: ", slog.Any("ERROR :: ", err))
		cancel()
		return
	}
	slog.Info("Trade engine running... Press Ctrl+C to stop")

	select {}

}
