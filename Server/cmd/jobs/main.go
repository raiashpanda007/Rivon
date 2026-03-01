package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/raiashpanda007/rivon/internals/config"
	"github.com/raiashpanda007/rivon/internals/database"
	"github.com/raiashpanda007/rivon/internals/jobs"
	"github.com/robfig/cron/v3"
)

func main() {
	slog.Info("Starting Cron Jobs")
	cfg := config.MustLoad()
	db, err := database.Init_DB(cfg.Db.PgURL, cfg.Db.OTPRedisURL, cfg.Db.OrderRedisURL)
	if err != nil {
		panic("Unable to connect DB" + err.Error())
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Run startup jobs only when the application starts
	if err := jobs.RunStartUpJobs(ctx, db.PgDB, cfg); err != nil {
		slog.Error("Failed to run startup jobs", "error", err)
	}

	// Run cron jobs immediately when the application starts
	if err := jobs.RunCronJobs(ctx, db.PgDB, db.OrderRedis, cfg); err != nil {
		slog.Error("Failed to run initial cron jobs", "error", err)
	}

	// Initialize Cron scheduler
	c := cron.New()

	// Schedule cron jobs to run every midnight
	_, err = c.AddFunc("@midnight", func() {
		slog.Info("Running scheduled cron jobs")
		// Use a fresh context for the scheduled job, or context.Background()
		if err := jobs.RunCronJobs(context.Background(), db.PgDB, db.OrderRedis, cfg); err != nil {
			slog.Error("Failed to run scheduled cron jobs", "error", err)
		}
	})
	if err != nil {
		panic("Failed to add cron job: " + err.Error())
	}

	c.Start()
	slog.Info("Cron scheduler started")

	// Block until a signal is received
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	slog.Info("Shutting down Cron Jobs")
	c.Stop()
}
