package main

import (
	"context"
	"log/slog"

	config "github.com/raiashpanda007/rivon/engine/internals/Config"
	database "github.com/raiashpanda007/rivon/engine/internals/Database"
	engine "github.com/raiashpanda007/rivon/engine/internals/Engine"
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

	engine.InitEngine(ctx, orderRedis, tradeRedis, db)
	slog.Info("Trade engine running... Press Ctrl+C to stop")

	select {}

}
