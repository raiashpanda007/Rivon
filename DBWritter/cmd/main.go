package main

import (
	"context"
	"log/slog"
	"os/signal"
	"syscall"

	config "github.com/raiashpanda007/rivon/dbwritter/internals/Config"
	db "github.com/raiashpanda007/rivon/dbwritter/internals/Db"
	candles "github.com/raiashpanda007/rivon/dbwritter/internals/Candles"
	repowriter "github.com/raiashpanda007/rivon/dbwritter/internals/RepoWriter"
	tradestreamreader "github.com/raiashpanda007/rivon/dbwritter/internals/TradeStreamReader"
	tsdb "github.com/raiashpanda007/rivon/dbwritter/internals/TSDB"
	redisconnection "github.com/raiashpanda007/rivon/dbwritter/internals/redisConnection"
	"os"
)

func main() {
	slog.Info("Started Db writter.")

	cfg := config.MustLoad()
	if cfg == nil {
		slog.Error("Unable to read the config from env")
		os.Exit(1)
	}

	pool, err := db.ConnectToData(cfg.PG_DB_URL)
	if err != nil {
		panic(err)
	}

	redisClient, err := redisconnection.ConnectToTradeStream(cfg.TRADE_REDIS_URL)
	if err != nil {
		panic(err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	writer := repowriter.NewRepoWriter(pool)
	tsdbWriter := tsdb.NewTSDBWriter(pool)
	candleSvc := candles.NewCandleService(pool, redisClient)

	go tsdbWriter.StartFlushLoop(ctx)
	go candleSvc.StartAllPublishers(ctx)

	tradestreamreader.InitTradeConsumer(ctx, redisClient, writer, tsdbWriter)
}
