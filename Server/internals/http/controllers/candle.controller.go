package controllers

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/raiashpanda007/rivon/internals/utils"
)

type CandleController interface {
	StreamCandles(res http.ResponseWriter, req *http.Request)
	GetCandleHistory(res http.ResponseWriter, req *http.Request)
}

type candleControllerUtils struct {
	tradeRedis *redis.Client
	db         *pgxpool.Pool
}

func InitCandleController(tradeRedis *redis.Client, db *pgxpool.Pool) CandleController {
	return &candleControllerUtils{tradeRedis: tradeRedis, db: db}
}

type candleHistory struct {
	OpenTime time.Time `json:"openTime"`
	Open     int64     `json:"open"`
	High     int64     `json:"high"`
	Low      int64     `json:"low"`
	Close    int64     `json:"close"`
	Volume   int64     `json:"volume"`
}

var candleIntervalBuckets = map[string]string{
	"1m":  "1 minute",
	"5m":  "5 minutes",
	"15m": "15 minutes",
	"1h":  "1 hour",
	"1d":  "1 day",
}

func (c *candleControllerUtils) StreamCandles(res http.ResponseWriter, req *http.Request) {
	marketId := req.URL.Query().Get("market")
	interval := req.URL.Query().Get("interval")
	if marketId == "" || interval == "" {
		http.Error(res, "missing market or interval query params", http.StatusBadRequest)
		return
	}

	if _, ok := candleIntervalBuckets[interval]; !ok {
		http.Error(res, "interval must be one of: 1m, 5m, 15m, 1h, 1d", http.StatusBadRequest)
		return
	}

	flusher, ok := res.(http.Flusher)
	if !ok {
		http.Error(res, "streaming not supported", http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", "text/event-stream")
	res.Header().Set("Cache-Control", "no-cache")
	res.Header().Set("Connection", "keep-alive")
	res.Header().Set("X-Accel-Buffering", "no")
	res.WriteHeader(http.StatusOK)
	flusher.Flush()

	channel := "CANDLES_" + marketId + "_" + interval
	ctx := req.Context()
	sub := c.tradeRedis.Subscribe(ctx, channel)
	defer sub.Close()

	if _, err := sub.Receive(ctx); err != nil {
		slog.Error("SSE candle subscribe failed", "channel", channel, "err", err)
		return
	}

	ch := sub.Channel()
	heartbeat := time.NewTicker(30 * time.Second)
	defer heartbeat.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-ch:
			if !ok {
				return
			}
			fmt.Fprintf(res, "event: candle\ndata: %s\n\n", msg.Payload)
			flusher.Flush()
		case <-heartbeat.C:
			fmt.Fprintf(res, ": heartbeat\n\n")
			flusher.Flush()
		}
	}
}

func (c *candleControllerUtils) GetCandleHistory(res http.ResponseWriter, req *http.Request) {
	marketId := req.URL.Query().Get("market")
	interval := req.URL.Query().Get("interval")
	fromStr := req.URL.Query().Get("from")
	toStr := req.URL.Query().Get("to")

	if marketId == "" || interval == "" || fromStr == "" || toStr == "" {
		utils.WriteJson(res, http.StatusBadRequest, utils.GeneralError(
			errors.New("missing market, interval, from, or to query params"),
			"Invalid request",
			http.StatusBadRequest,
			"Bad Request",
		))
		return
	}

	if _, err := uuid.Parse(marketId); err != nil {
		utils.WriteJson(res, http.StatusBadRequest, utils.GeneralError(
			err,
			"Invalid market id",
			http.StatusBadRequest,
			"Bad Request",
		))
		return
	}

	bucket, ok := candleIntervalBuckets[interval]
	if !ok {
		utils.WriteJson(res, http.StatusBadRequest, utils.GeneralError(
			errors.New("invalid interval"),
			"Interval must be one of: 1m, 5m, 15m, 1h, 1d",
			http.StatusBadRequest,
			"Bad Request",
		))
		return
	}

	fromSec, err := strconv.ParseInt(fromStr, 10, 64)
	if err != nil || fromSec <= 0 {
		utils.WriteJson(res, http.StatusBadRequest, utils.GeneralError(
			errors.New("invalid from"),
			"Invalid from value",
			http.StatusBadRequest,
			"Bad Request",
		))
		return
	}
	toSec, err := strconv.ParseInt(toStr, 10, 64)
	if err != nil || toSec <= 0 || toSec <= fromSec {
		utils.WriteJson(res, http.StatusBadRequest, utils.GeneralError(
			errors.New("invalid to"),
			"Invalid to value",
			http.StatusBadRequest,
			"Bad Request",
		))
		return
	}

	rows, err := c.db.Query(req.Context(), `
		WITH prev_close AS (
		    SELECT COALESCE(last(price, time), 0) AS price
		    FROM trade_ticks
		    WHERE market_id = $2::uuid
		      AND time < to_timestamp($3)
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
		      AND time >= time_bucket($1::interval, to_timestamp($3))
		      AND time <= to_timestamp($4)
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
		bucket,
		marketId,
		fromSec,
		toSec,
	)
	if err != nil {
		slog.Error("Candle history query failed", "err", err)
		utils.WriteJson(res, http.StatusInternalServerError, utils.GeneralError(
			err,
			"Unable to fetch candle history",
			http.StatusInternalServerError,
			"Internal Server Error",
		))
		return
	}
	defer rows.Close()

	var candles []candleHistory
	for rows.Next() {
		var cdl candleHistory
		if err := rows.Scan(&cdl.OpenTime, &cdl.Open, &cdl.High, &cdl.Low, &cdl.Close, &cdl.Volume); err != nil {
			slog.Error("Candle history scan failed", "err", err)
			utils.WriteJson(res, http.StatusInternalServerError, utils.GeneralError(
				err,
				"Unable to parse candle history",
				http.StatusInternalServerError,
				"Internal Server Error",
			))
			return
		}
		candles = append(candles, cdl)
	}

	if err := rows.Err(); err != nil {
		slog.Error("Candle history rows error", "err", err)
		utils.WriteJson(res, http.StatusInternalServerError, utils.GeneralError(
			err,
			"Unable to fetch candle history",
			http.StatusInternalServerError,
			"Internal Server Error",
		))
		return
	}

	utils.WriteJson(res, http.StatusOK, utils.Response[[]candleHistory]{
		Status:  http.StatusOK,
		Heading: "Candle History",
		Message: "Candle history fetched",
		Data:    candles,
	})
}
