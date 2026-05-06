package candles

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Candle struct {
	OpenTime time.Time `json:"openTime"`
	MarketId string    `json:"marketId"`
	Interval string    `json:"interval"`
	Open     int64     `json:"open"`
	High     int64     `json:"high"`
	Low      int64     `json:"low"`
	Close    int64     `json:"close"`
	Volume   int64     `json:"volume"`
}

// lookbacks maps interval name → (bucket param, lookback param) for SQL.
var lookbacks = map[string][2]string{
	"1m":  {"1 minute", "6 hours"},
	"5m":  {"5 minutes", "24 hours"},
	"15m": {"15 minutes", "3 days"},
	"1h":  {"1 hour", "30 days"},
	"1d":  {"1 day", "180 days"},
}

// tickerDurations maps interval name → how often to recompute.
var tickerDurations = map[string]time.Duration{
	"1m":  1 * time.Minute,
	"5m":  5 * time.Minute,
	"15m": 15 * time.Minute,
	"1h":  1 * time.Hour,
	"1d":  24 * time.Hour,
}

type CandleService struct {
	db    *pgxpool.Pool
	redis *redis.Client
}

func NewCandleService(db *pgxpool.Pool, redisClient *redis.Client) *CandleService {
	return &CandleService{db: db, redis: redisClient}
}

func (s *CandleService) GenerateCandles(ctx context.Context, marketId, interval string) ([]Candle, error) {
	params, ok := lookbacks[interval]
	if !ok {
		return nil, nil
	}
	bucket, lookback := params[0], params[1]

	rows, err := s.db.Query(ctx, `
		WITH prev_close AS (
		    SELECT COALESCE(last(price, time), 0) AS price
		    FROM trade_ticks
		    WHERE market_id = $2::uuid
		      AND time < time_bucket($1::interval, NOW() - $3::interval)
		),
		raw AS (
		    SELECT
		        time_bucket($1::interval, time) AS open_time,
		        max(price)                      AS high,
		        min(price)                      AS low,
		        last(price, time)               AS close,
		        sum(quantity)                   AS volume
		    FROM trade_ticks
		    WHERE market_id = $2::uuid
		      AND time >= time_bucket($1::interval, NOW() - $3::interval)
		    GROUP BY 1
		)
		SELECT
		    open_time,
		    COALESCE(LAG(close) OVER (ORDER BY open_time), pc.price) AS open,
		    high,
		    low,
		    close,
		    volume
		FROM raw, prev_close pc
		ORDER BY open_time`,
		bucket, marketId, lookback,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var candles []Candle
	for rows.Next() {
		var c Candle
		c.MarketId = marketId
		c.Interval = interval
		if err := rows.Scan(&c.OpenTime, &c.Open, &c.High, &c.Low, &c.Close, &c.Volume); err != nil {
			return nil, err
		}
		candles = append(candles, c)
	}
	return candles, rows.Err()
}

// StartAllPublishers fetches all market IDs from the DB and starts a publisher goroutine per market.
func (s *CandleService) StartAllPublishers(ctx context.Context) {
	rows, err := s.db.Query(ctx, "SELECT id::text FROM markets")
	if err != nil {
		slog.Error("CandleService: failed to fetch market IDs", "err", err)
		return
	}
	var marketIds []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err == nil {
			marketIds = append(marketIds, id)
		}
	}
	rows.Close()

	for _, id := range marketIds {
		go s.startPublisher(ctx, id)
	}
	slog.Info("CandleService: started publishers", "markets", len(marketIds))
}

func (s *CandleService) startPublisher(ctx context.Context, marketId string) {
	for interval := range tickerDurations {
		s.publishCandles(ctx, marketId, interval)
	}

	tickers := map[string]*time.Ticker{}
	for interval, dur := range tickerDurations {
		tickers[interval] = time.NewTicker(dur)
	}
	defer func() {
		for _, t := range tickers {
			t.Stop()
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case <-tickers["1m"].C:
			s.publishCandles(ctx, marketId, "1m")
		case <-tickers["5m"].C:
			s.publishCandles(ctx, marketId, "5m")
		case <-tickers["15m"].C:
			s.publishCandles(ctx, marketId, "15m")
		case <-tickers["1h"].C:
			s.publishCandles(ctx, marketId, "1h")
		case <-tickers["1d"].C:
			s.publishCandles(ctx, marketId, "1d")
		}
	}
}

func (s *CandleService) publishCandles(ctx context.Context, marketId, interval string) {
	candles, err := s.GenerateCandles(ctx, marketId, interval)
	if err != nil {
		slog.Error("CandleService: candle generation failed", "marketId", marketId, "interval", interval, "err", err)
		return
	}
	if len(candles) == 0 {
		return
	}
	data, err := json.Marshal(candles)
	if err != nil {
		slog.Error("CandleService: marshal failed", "err", err)
		return
	}
	channel := "CANDLES_" + marketId + "_" + interval
	if err := s.redis.Publish(ctx, channel, string(data)).Err(); err != nil {
		slog.Error("CandleService: publish failed", "channel", channel, "err", err)
	}
}
