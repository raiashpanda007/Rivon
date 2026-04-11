package redis

import (
	"context"
	"log/slog"
	"strconv"

	"github.com/go-redis/redis/v8"
)

type ReplayOrderStreamMessage struct {
	OrderId   string
	UserId    string
	MarketId  string
	Price     int
	Quantity  int
	OrderType string
	StreamId  string
}

func getString(values map[string]interface{}, key string) string {
	if v, ok := values[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func getInt(values map[string]interface{}, key string) int {
	if v, ok := values[key]; ok {
		switch val := v.(type) {
		case int:
			return val
		case int64:
			return int(val)
		case float64:
			return int(val)
		case string:
			i, _ := strconv.Atoi(val)
			return i
		}
	}
	return 0
}
func Init_Order_Redis(orderRedisURL string) (*redis.Client, error) {

	ctx := context.Background()

	slog.Info("CONNECTING TO ORDER REDIS DB :: ")

	OrderRedis := redis.NewClient(&redis.Options{
		Addr: orderRedisURL,
	})

	_, err := OrderRedis.Ping(ctx).Result()

	if err != nil {
		slog.Error("REDIS Ping failed ", slog.Any("ERROR :: ", err))
		return nil, err
	}

	slog.Info("CONNECTED TO ORDER REDIS DB :: ")
	return OrderRedis, nil
}

// ReadLastTradeOrderIdForMarket scans the TRADES stream (newest-first) and
// returns the orderId of the most recent trade entry for the given marketId.
// Returns ("", false) if no matching entry is found.
func ReadLastTradeOrderIdForMarket(ctx context.Context, tradeRedis *redis.Client, marketId string) (string, bool) {
	msgs, err := tradeRedis.XRevRangeN(ctx, "TRADES", "+", "-", 500).Result()
	if err != nil {
		return "", false
	}
	for _, msg := range msgs {
		if getString(msg.Values, "marketId") == marketId {
			return getString(msg.Values, "orderId"), true
		}
	}
	return "", false
}

func ReplayOrderStream(ctx context.Context, orderRedis *redis.Client, streamName string, lastStreamID string) ([]ReplayOrderStreamMessage, error) {

	var messages []ReplayOrderStreamMessage

	for {
		// XRange reads existing messages without blocking, ideal for replay
		msgs, err := orderRedis.XRangeN(ctx, streamName, lastStreamID, "+", 100).Result()
		if err != nil {
			return nil, err
		}

		if len(msgs) == 0 {
			break
		}

		// skip the first message if it matches lastStreamID (XRange is inclusive)
		start := 0
		if msgs[0].ID == lastStreamID {
			start = 1
		}

		for _, msg := range msgs[start:] {
			order := ReplayOrderStreamMessage{
				OrderId:   getString(msg.Values, "orderId"),
				UserId:    getString(msg.Values, "userId"),
				MarketId:  getString(msg.Values, "marketId"),
				OrderType: getString(msg.Values, "orderType"),
				Price:     getInt(msg.Values, "price"),
				Quantity:  getInt(msg.Values, "quantity"),
				StreamId:  msg.ID,
			}
			messages = append(messages, order)
			lastStreamID = msg.ID
		}

		if len(msgs)-start < 100 {
			break // fetched fewer than requested → reached end of stream
		}
	}

	return messages, nil
}
