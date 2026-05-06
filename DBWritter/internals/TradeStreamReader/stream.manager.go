package tradestreamreader

import (
	"context"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	repowriter "github.com/raiashpanda007/rivon/dbwritter/internals/RepoWriter"
	tsdb "github.com/raiashpanda007/rivon/dbwritter/internals/TSDB"
)

func createGroupForTradeStream(redisClient *redis.Client, ctx context.Context) {
	err := redisClient.XGroupCreateMkStream(ctx, "TRADES", "DB_WRITTER", "$").Err()
	if err != nil && !strings.Contains(err.Error(), "BUSYGROUP") {
		panic(err)
	}
}

func InitTradeConsumer(ctx context.Context, tradeRedisStreamClient *redis.Client, writer *repowriter.RepoWriter, tsdbWriter *tsdb.TSDBWriter) {
	slog.Info("Starting up trade consumer readers")

	createGroupForTradeStream(tradeRedisStreamClient, ctx)

	retries := map[string]int{}

	for {
		streams, err := tradeRedisStreamClient.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    "DB_WRITTER",
			Consumer: "DB_WRITTER_1",
			Streams:  []string{"TRADES", ">"},
			Count:    10,
			Block:    0,
		}).Result()

		if err != nil {
			if err == redis.Nil {
				continue
			}
			slog.Error("Redis XReadGroup error", "err", err)
			time.Sleep(time.Second)
			continue
		}

		for _, stream := range streams {
			for _, msg := range stream.Messages {
				tradeMsg, err := repowriter.ParseTradeMessage(msg.Values)
				if err != nil {
					slog.Error("Failed to parse trade message", "id", msg.ID, "err", err)
					// Unparseable message — ack immediately so it doesn't block the stream.
					tradeRedisStreamClient.XAck(ctx, "TRADES", "DB_WRITTER", msg.ID)
					continue
				}
				// Stream entry IDs are "<ms-since-epoch>-<seq>"; use this as the
				// trade execution time so candle buckets reflect when the trade happened.
				if parts := strings.SplitN(msg.ID, "-", 2); len(parts) == 2 {
					if ms, convErr := strconv.ParseInt(parts[0], 10, 64); convErr == nil {
						tradeMsg.ExecutedAt = time.UnixMilli(ms).UTC()
					}
				}
				if tradeMsg.ExecutedAt.IsZero() {
					tradeMsg.ExecutedAt = time.Now().UTC()
				}

				if err := writer.ProcessTradeMessage(ctx, *tradeMsg); err != nil {
					retries[msg.ID]++
					slog.Error("ProcessTradeMessage failed", "id", msg.ID, "attempt", retries[msg.ID], "err", err)
					if retries[msg.ID] >= 5 {
						slog.Error("Dead-lettering message after 5 failed attempts", "id", msg.ID)
						tradeRedisStreamClient.XAck(ctx, "TRADES", "DB_WRITTER", msg.ID)
						delete(retries, msg.ID)
					}
					continue
				}

				tsdbWriter.Enqueue(*tradeMsg)
				delete(retries, msg.ID)
				tradeRedisStreamClient.XAck(ctx, "TRADES", "DB_WRITTER", msg.ID)
			}
		}
	}
}
