package tsdb

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	repowriter "github.com/raiashpanda007/rivon/dbwritter/internals/RepoWriter"
)

type tradeTick struct {
	at       time.Time
	marketId string
	price    int64
	quantity int64
	tradeId  string
}

type TSDBWriter struct {
	db      *pgxpool.Pool
	mu      sync.Mutex
	batch   []tradeTick
	flushCh chan struct{}
}

func NewTSDBWriter(db *pgxpool.Pool) *TSDBWriter {
	return &TSDBWriter{
		db:      db,
		flushCh: make(chan struct{}, 1),
	}
}

func (w *TSDBWriter) Enqueue(msg repowriter.TradeMessage) {
	w.mu.Lock()
	// Spread fills by microsecond per index so first(price, time)/last(price, time)
	// in candle aggregates have a deterministic ordering matching match order.
	for i, f := range msg.Fills {
		w.batch = append(w.batch, tradeTick{
			at:       msg.ExecutedAt.Add(time.Duration(i) * time.Microsecond),
			marketId: msg.MarketId,
			price:    int64(f.Price),
			quantity: int64(f.Quantity),
			tradeId:  f.TradeId,
		})
	}
	shouldFlush := len(w.batch) >= 100
	w.mu.Unlock()

	if shouldFlush {
		select {
		case w.flushCh <- struct{}{}:
		default:
		}
	}
}

func (w *TSDBWriter) StartFlushLoop(ctx context.Context) {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			w.flush(ctx)
		case <-w.flushCh:
			w.flush(ctx)
		case <-ctx.Done():
			w.flush(ctx)
			return
		}
	}
}

func (w *TSDBWriter) flush(ctx context.Context) {
	w.mu.Lock()
	if len(w.batch) == 0 {
		w.mu.Unlock()
		return
	}
	ticks := w.batch
	w.batch = nil
	w.mu.Unlock()

	_, err := w.db.CopyFrom(
		ctx,
		pgx.Identifier{"trade_ticks"},
		[]string{"time", "market_id", "price", "quantity", "trade_id"},
		pgx.CopyFromSlice(len(ticks), func(i int) ([]any, error) {
			t := ticks[i]
			return []any{t.at, t.marketId, t.price, t.quantity, t.tradeId}, nil
		}),
	)
	if err != nil {
		// CopyFrom fails entirely on UNIQUE conflict — fall back to individual inserts.
		slog.Warn("TSDB bulk copy failed (likely duplicate trade_id), falling back to individual inserts", "err", err)
		w.flushIndividual(ctx, ticks)
	}
}

func (w *TSDBWriter) flushIndividual(ctx context.Context, ticks []tradeTick) {
	for _, t := range ticks {
		_, err := w.db.Exec(ctx, `
			INSERT INTO trade_ticks (time, market_id, price, quantity, trade_id)
			VALUES ($1, $2, $3, $4, $5)`,
			t.at, t.marketId, t.price, t.quantity, t.tradeId,
		)
		if err != nil {
			slog.Error("TSDB individual insert failed", "tradeId", t.tradeId, "err", err)
		}
	}
}
