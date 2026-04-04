package engine

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	database "github.com/raiashpanda007/rivon/engine/internals/Database"
	pubsub "github.com/raiashpanda007/rivon/engine/internals/PubSub"
	"github.com/raiashpanda007/rivon/engine/internals/markets"
)

func parseOrderMessage(values map[string]interface{}, streamId string) (markets.OrderMessages, error) {

	priceStr := values["price"].(string)
	qtyStr := values["quantity"].(string)

	price, err := strconv.Atoi(priceStr)
	if err != nil {
		return markets.OrderMessages{}, err
	}

	qty, err := strconv.Atoi(qtyStr)
	if err != nil {
		return markets.OrderMessages{}, err
	}

	return markets.OrderMessages{
		OrderId:   values["orderId"].(string),
		UserId:    values["userId"].(string),
		MarketId:  values["marketId"].(string),
		Price:     price,
		Quantity:  qty,
		OrderType: values["orderType"].(string),
		StreamId:  streamId,
	}, nil
}

func redisStreamProducers(ctx context.Context, redisClient *redis.Client, markets []database.Market) error {
	group := "engine"
	var wg sync.WaitGroup
	errChan := make(chan error, len(markets))

	for _, market := range markets {
		wg.Add(1)
		go func(m database.Market) {
			defer wg.Done()
			stream := "ORDERS_" + m.Id

			err := redisClient.XGroupCreateMkStream(ctx, stream, group, "0").Err()
			if err != nil {
				// ignore if group already exists
				if !strings.Contains(err.Error(), "BUSYGROUP") {
					errChan <- err
					return
				}
			}
		}(market)
	}

	// Wait for all goroutines to complete
	wg.Wait()
	close(errChan)

	// Check if any errors occurred
	for err := range errChan {
		if err != nil {
			return err
		}
	}

	slog.Info("All streams initialized", "total_streams", len(markets))
	return nil
}

func ensureConsumerGroups(ctx context.Context, redisClient *redis.Client, marketBatch []database.Market) {
	group := "engine"
	for _, market := range marketBatch {
		stream := "ORDERS_" + market.Id
		err := redisClient.XGroupCreateMkStream(ctx, stream, group, "0").Err()
		if err != nil {
			if !strings.Contains(err.Error(), "BUSYGROUP") {
				slog.Error("Failed to ensure consumer group", "stream", stream, "error", err)
			}
		}
	}
}

const batchSize = 20

func redisStreamBatchConsumer(ctx context.Context, redisClient *redis.Client, marketMap map[string]chan markets.OrderMessages, batch []database.Market, batchId int) {
	// Build streams slice: [stream1, stream2, ..., >, >, ...]
	streams := make([]string, 0, len(batch)*2)
	for _, market := range batch {
		streams = append(streams, "ORDERS_"+market.Id)
	}
	for range batch {
		streams = append(streams, ">")
	}

	slog.Info("Batch consumer started", "batch_id", batchId, "stream_count", len(batch))

	for {
		res, err := redisClient.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    "engine",
			Consumer: fmt.Sprintf("engine-%d", batchId),
			Streams:  streams,
			Count:    10,
			Block:    5 * time.Second,
		}).Result()

		if err != nil {
			if err == redis.Nil {
				continue
			}
			if strings.Contains(err.Error(), "NOGROUP") {
				slog.Warn("Consumer group missing in batch, re-creating...", "batch_id", batchId)
				ensureConsumerGroups(ctx, redisClient, batch)
				time.Sleep(1 * time.Second)
				continue
			}
			slog.Error("Redis read error", "batch_id", batchId, "error", err)
			time.Sleep(1 * time.Second)
			continue
		}

		for _, stream := range res {
			for _, message := range stream.Messages {
				msg, err := parseOrderMessage(message.Values, message.ID)

				if err != nil {
					slog.Error("Error in parsing the stream message :: ", err)
					continue
				}
				ch, ok := marketMap[msg.MarketId]
				if !ok {
					continue
				}

				select {
				case ch <- msg:

				/*
						___________________________________
					|                                    |
					|Warning in case of message dropping |
					|____________________________________|

				*/

				default:
					slog.Warn("Channel full, dropping message", "marketId", msg.MarketId)
				}
			}
		}
	}
}

func startBatchedConsumers(ctx context.Context, redisClient *redis.Client, marketMap map[string]chan markets.OrderMessages, allMarkets []database.Market) {
	ensureConsumerGroups(ctx, redisClient, allMarkets)
	slog.Info("Consumer groups verified", "total_streams", len(allMarkets))

	batchCount := 0
	for i := 0; i < len(allMarkets); i += batchSize {
		end := i + batchSize
		if end > len(allMarkets) {
			end = len(allMarkets)
		}

		batch := allMarkets[i:end]
		go redisStreamBatchConsumer(ctx, redisClient, marketMap, batch, batchCount)
		batchCount++
	}

	slog.Info("All batch consumers started", "total_batches", batchCount, "total_streams", len(allMarkets))
}

func InitEngine(ctx context.Context, OrderRedis, TradeRedis *redis.Client, Db *database.Database, pubsubSvc pubsub.PubSubService) error {
	allMarkets, err := Db.GetAllMarkets()
	var marketChannelMap = make(map[string]chan markets.OrderMessages)

	if err != nil {
		slog.Error("ERROR :: IN GETTING ALL MARKETS :: ", slog.Any("ERROR :: ", err))
		return err
	}

	slog.Info("Creating market channels", "count", len(allMarkets))

	for _, market := range allMarkets {
		marketChannelMap[market.Id] = make(chan markets.OrderMessages, 50)
		go markets.StarMarketProcess(ctx, marketChannelMap[market.Id], TradeRedis, pubsubSvc, market.Id)
	}

	slog.Info("Initializing Redis streams...", "count", len(allMarkets))
	if err := redisStreamProducers(ctx, OrderRedis, allMarkets); err != nil {
		slog.Error("Failed to initialize streams", "error", err)
		return err
	}
	slog.Info("All streams ready, starting consumers...")
	startBatchedConsumers(ctx, OrderRedis, marketChannelMap, allMarkets)

	slog.Info("Engine initialized successfully", "market_count", len(allMarkets))
	return nil
}
