package repowriter

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Fill struct {
	Price        int    `json:"Price"`
	Quantity     int    `json:"Quantity"`
	TradeId      string `json:"TradeId"`
	OtherUserId  string `json:"OtherUserId"`
	OtherOrderId string `json:"OtherOrderId"`
	OrderId      string `json:"OrderId"`
}

type TradeMessage struct {
	ExecutedAt  time.Time
	TradeType   string
	MarketId    string
	OrderId     string
	UserId      string
	Side        string // "BUY" or "SELL"
	Quantity    int64
	ExecutedQty int64
	Price       int64
	Fills       []Fill
}

type RepoWriter struct {
	db *pgxpool.Pool
}

func NewRepoWriter(db *pgxpool.Pool) *RepoWriter {
	return &RepoWriter{db: db}
}

func ParseTradeMessage(values map[string]interface{}) (*TradeMessage, error) {
	getString := func(k string) string {
		if v, ok := values[k]; ok {
			return fmt.Sprintf("%v", v)
		}
		return ""
	}
	getInt64 := func(k string) int64 {
		n, _ := strconv.ParseInt(getString(k), 10, 64)
		return n
	}

	fillsJSON := getString("fills")
	var fills []Fill
	if fillsJSON != "" && fillsJSON != "null" {
		if err := json.Unmarshal([]byte(fillsJSON), &fills); err != nil {
			return nil, fmt.Errorf("parse fills JSON: %w", err)
		}
	}

	return &TradeMessage{
		TradeType:   getString("tradeType"),
		MarketId:    getString("marketId"),
		OrderId:     getString("orderId"),
		UserId:      getString("userId"),
		Side:        getString("side"),
		Quantity:    getInt64("quantity"),
		ExecutedQty: getInt64("executedQty"),
		Price:       getInt64("price"),
		Fills:       fills,
	}, nil
}

func (r *RepoWriter) ProcessTradeMessage(ctx context.Context, msg TradeMessage) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if msg.TradeType == "order_cancelled" {
		if msg.OrderId != "" {
			if err := r.cancelOrder(ctx, tx, msg.OrderId); err != nil {
				return fmt.Errorf("cancel order %s: %w", msg.OrderId, err)
			}
		}
		return tx.Commit(ctx)
	}
	if len(msg.Fills) == 0 {
		// Queued order with no immediate match — already written as 'pending', nothing to update.
		return tx.Commit(ctx)
	}

	takerStatus := "partial"
	if msg.ExecutedQty >= msg.Quantity && msg.Quantity > 0 {
		takerStatus = "filled"
	}

	if err := r.upsertOrder(ctx, tx, msg.OrderId, msg.UserId, msg.MarketId, msg.Side, msg.Price, msg.Quantity, msg.ExecutedQty, takerStatus); err != nil {
		return fmt.Errorf("upsert taker order %s: %w", msg.OrderId, err)
	}

	makerSide := "SELL"
	if msg.Side == "SELL" {
		makerSide = "BUY"
	}

	for _, fill := range msg.Fills {
		rowsAffected, err := r.writeTrade(ctx, tx, fill.TradeId, msg.MarketId, fill.OrderId, fill.OtherOrderId, fill.Price, fill.Quantity)
		if err != nil {
			return fmt.Errorf("write trade %s: %w", fill.TradeId, err)
		}
		if rowsAffected == 0 {
			continue // already processed — idempotency
		}

		if err := r.upsertMakerOrder(ctx, tx, fill.OtherOrderId, fill.OtherUserId, msg.MarketId, makerSide, fill.Price, fill.Quantity); err != nil {
			return fmt.Errorf("upsert maker order %s: %w", fill.OtherOrderId, err)
		}

		var buyerId, sellerId, takerOrderId, makerOrderId string
		if msg.Side == "BUY" {
			buyerId, sellerId = msg.UserId, fill.OtherUserId
			takerOrderId, makerOrderId = fill.OrderId, fill.OtherOrderId
		} else {
			buyerId, sellerId = fill.OtherUserId, msg.UserId
			takerOrderId, makerOrderId = fill.OtherOrderId, fill.OrderId
		}
		amount := int64(fill.Price) * int64(fill.Quantity)
		if err := r.settleWallets(ctx, tx, buyerId, sellerId, msg.MarketId, fill.TradeId, takerOrderId, makerOrderId, amount, int64(fill.Quantity), int64(fill.Price)); err != nil {
			return fmt.Errorf("settle wallets for trade %s: %w", fill.TradeId, err)
		}
	}

	return tx.Commit(ctx)
}

func (r *RepoWriter) cancelOrder(ctx context.Context, tx pgx.Tx, orderId string) error {
	_, err := tx.Exec(ctx, `
		UPDATE orders SET status = 'cancelled', updated_at = NOW()
		WHERE id = $1 AND status NOT IN ('filled', 'cancelled')`,
		orderId,
	)
	return err
}

// upsertOrder uses GREATEST for executed_qty — idempotent across retries.
func (r *RepoWriter) upsertOrder(ctx context.Context, tx pgx.Tx, orderId, userId, marketId, side string, price, quantity, executedQty int64, status string) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO orders (id, market_id, user_id, side, price, quantity, executed_qty, status)
		VALUES ($1, $2, $3, $4::order_side, $5, $6, $7, $8::order_status)
		ON CONFLICT (id) DO UPDATE SET
		    executed_qty = GREATEST(orders.executed_qty, EXCLUDED.executed_qty),
		    status       = EXCLUDED.status,
		    updated_at   = NOW()
		WHERE orders.status NOT IN ('filled', 'cancelled')`,
		orderId, marketId, userId, side, price, quantity, executedQty, status,
	)
	return err
}

// upsertMakerOrder accumulates executed_qty — only called when behind the idempotency gate.
func (r *RepoWriter) upsertMakerOrder(ctx context.Context, tx pgx.Tx, orderId, userId, marketId, side string, price, quantity int) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO orders (id, market_id, user_id, side, price, quantity, executed_qty, status)
		VALUES ($1, $2, $3, $4::order_side, $5, $6, $6, 'partial'::order_status)
		ON CONFLICT (id) DO UPDATE SET
		    executed_qty = orders.executed_qty + $6,
		    status = CASE
		        WHEN orders.executed_qty + $6 >= orders.quantity THEN 'filled'::order_status
		        ELSE 'partial'::order_status
		    END,
		    updated_at = NOW()
		WHERE orders.status NOT IN ('filled', 'cancelled')`,
		orderId, marketId, userId, side, price, quantity,
	)
	return err
}

func (r *RepoWriter) writeTrade(ctx context.Context, tx pgx.Tx, tradeId, marketId, orderId, otherOrderId string, price, quantity int) (int64, error) {
	tag, err := tx.Exec(ctx, `
		INSERT INTO trades (id, market_id, order_id, other_order_id, price, quantity)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (id) DO NOTHING`,
		tradeId, marketId, orderId, otherOrderId, price, quantity,
	)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}

func (r *RepoWriter) settleWallets(ctx context.Context, tx pgx.Tx, buyerId, sellerId, marketId, tradeId, takerOrderId, makerOrderId string, amount, qty, price int64) error {
	// Credit seller — CTE reads pre-update balance to satisfy valid_balance_transition CHECK.
	_, err := tx.Exec(ctx, `
		WITH upd AS (
		    UPDATE wallets
		    SET balance = balance + $1, updated_at = NOW()
		    WHERE user_id = $2
		    RETURNING id AS wallet_id, balance - $1 AS balance_before, balance AS balance_after
		)
		INSERT INTO transactions (id, wallet_id, type, amount, balance_before, balance_after, order_id, trade_id)
		SELECT gen_random_uuid(), upd.wallet_id, 'credit', $1, upd.balance_before, upd.balance_after, $3::uuid, $4::uuid
		FROM upd`,
		amount, sellerId, makerOrderId, tradeId,
	)
	if err != nil {
		return fmt.Errorf("credit seller %s: %w", sellerId, err)
	}

	// Debit buyer — CTE reads pre-update balance.
	tag, err := tx.Exec(ctx, `
		WITH upd AS (
		    UPDATE wallets
		    SET balance = balance - $1, updated_at = NOW()
		    WHERE user_id = $2 AND balance >= $1
		    RETURNING id AS wallet_id, balance + $1 AS balance_before, balance AS balance_after
		)
		INSERT INTO transactions (id, wallet_id, type, amount, balance_before, balance_after, order_id, trade_id)
		SELECT gen_random_uuid(), upd.wallet_id, 'debit', $1, upd.balance_before, upd.balance_after, $3::uuid, $4::uuid
		FROM upd`,
		amount, buyerId, takerOrderId, tradeId,
	)
	if err != nil {
		return fmt.Errorf("debit buyer %s: %w", buyerId, err)
	}
	if tag.RowsAffected() == 0 {
		slog.Error("buyer debit skipped — insufficient balance in DB (Engine/DB wallet divergence)",
			"buyerId", buyerId, "tradeId", tradeId, "amount", amount)
	}

	// Buyer gains asset.
	_, err = tx.Exec(ctx, `
		INSERT INTO assets (id, user_id, market_id, quantity, avg_cost)
		VALUES (gen_random_uuid(), $1, $2, $3, $4)
		ON CONFLICT (user_id, market_id) DO UPDATE SET
		    avg_cost   = (assets.quantity * assets.avg_cost + EXCLUDED.quantity * EXCLUDED.avg_cost)
		                 / (assets.quantity + EXCLUDED.quantity),
		    quantity   = assets.quantity + EXCLUDED.quantity,
		    updated_at = NOW()`,
		buyerId, marketId, qty, price,
	)
	if err != nil {
		return fmt.Errorf("upsert buyer asset: %w", err)
	}

	// Seller reduces asset (locked_qty released + quantity reduced).
	_, err = tx.Exec(ctx, `
		UPDATE assets
		   SET locked_qty  = GREATEST(0, locked_qty - $1),
		       quantity    = quantity - $1,
		       updated_at  = NOW()
		 WHERE user_id = $2 AND market_id = $3`,
		qty, sellerId, marketId,
	)
	if err != nil {
		return fmt.Errorf("reduce seller asset: %w", err)
	}

	return nil
}
