package main

import (
	"context"
	"log/slog"

	"github.com/raiashpanda007/rivon/internals/config"
	"github.com/raiashpanda007/rivon/internals/database"
	"github.com/raiashpanda007/rivon/internals/jobs"
)

func main() {
	slog.Info("Starting Cron Jobs")
	cfg := config.MustLoad()
	db, err := database.Init_DB(cfg.Db.PgURL, cfg.Db.OTPRedisURL)
	if err != nil {
		panic("Unable to connect DB" + err.Error())
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	jobs.RunStartUpJobs(ctx, db.PgDB, cfg)

	if err := jobs.RunCronJobs(ctx, db.PgDB, cfg); err != nil {
		panic(err)
	}
}
