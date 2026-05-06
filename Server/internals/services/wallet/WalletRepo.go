package wallet

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/raiashpanda007/rivon/internals/utils"
)

type Wallet struct {
	Id        uuid.UUID `json:"id"`
	UserId    uuid.UUID `json:"userId"`
	Balance   int64     `json:"balance"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type walletRepoUtils struct {
	pgDb *pgxpool.Pool
}

type Asset struct {
	MarketID string `json:"marketId"`
	Quantity int64  `json:"quantity"`
}

type Transaction struct {
	Id            uuid.UUID  `json:"id"`
	WalletId      uuid.UUID  `json:"walletId"`
	Type          string     `json:"type"`
	Amount        int64      `json:"amount"`
	BalanceBefore int64      `json:"balanceBefore"`
	BalanceAfter  int64      `json:"balanceAfter"`
	OrderId       *uuid.UUID `json:"orderId"`
	TradeId       *uuid.UUID `json:"tradeId"`
	CreatedAt     time.Time  `json:"createdAt"`
	MarketName    *string    `json:"marketName"`
	MarketCode    *string    `json:"marketCode"`
}

type AssetWithMarket struct {
	MarketId     string `json:"marketId"`
	MarketName   string `json:"marketName"`
	MarketCode   string `json:"marketCode"`
	Emblem       string `json:"emblem"`
	Quantity     int64  `json:"quantity"`
	LockedQty    int64  `json:"lockedQty"`
	AvgCost      int64  `json:"avgCost"`
	CurrentPrice int64  `json:"currentPrice"`
}

type WalletRepo interface {
	GetWalletInfo(ctx context.Context, userID string) (*Wallet, utils.ErrorType, error)
	GetUserAssets(ctx context.Context, userID string) ([]Asset, error)
	GetTransactions(ctx context.Context, userID string, limit, offset int) ([]Transaction, error)
	GetUserAssetsWithMarket(ctx context.Context, userID string) ([]AssetWithMarket, error)
}

func NewWalletRepo(pgDB *pgxpool.Pool) WalletRepo {
	return &walletRepoUtils{
		pgDb: pgDB,
	}

}

func (r *walletRepoUtils) GetUserAssets(ctx context.Context, userID string) ([]Asset, error) {
	id, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("invalid user ID format")
	}
	rows, err := r.pgDb.Query(ctx, `SELECT market_id, quantity FROM assets WHERE user_id = $1 AND quantity > 0`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var assets []Asset
	for rows.Next() {
		var a Asset
		if err := rows.Scan(&a.MarketID, &a.Quantity); err != nil {
			return nil, err
		}
		assets = append(assets, a)
	}
	return assets, nil
}

func (r *walletRepoUtils) GetTransactions(ctx context.Context, userID string, limit, offset int) ([]Transaction, error) {
	id, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("invalid user ID format")
	}
	rows, err := r.pgDb.Query(ctx, `
		SELECT
			t.id, t.wallet_id, t.type, t.amount, t.balance_before, t.balance_after,
			t.order_id, t.trade_id, t.created_at,
			m.market_name, m.market_code
		FROM transactions t
		JOIN wallets w ON t.wallet_id = w.id
		LEFT JOIN orders o ON t.order_id = o.id
		LEFT JOIN markets m ON o.market_id = m.id
		WHERE w.user_id = $1
		ORDER BY t.created_at DESC
		LIMIT $2 OFFSET $3`,
		id, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var txns []Transaction
	for rows.Next() {
		var tx Transaction
		if err := rows.Scan(
			&tx.Id, &tx.WalletId, &tx.Type, &tx.Amount, &tx.BalanceBefore, &tx.BalanceAfter,
			&tx.OrderId, &tx.TradeId, &tx.CreatedAt,
			&tx.MarketName, &tx.MarketCode,
		); err != nil {
			return nil, err
		}
		txns = append(txns, tx)
	}
	return txns, rows.Err()
}

func (r *walletRepoUtils) GetUserAssetsWithMarket(ctx context.Context, userID string) ([]AssetWithMarket, error) {
	id, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("invalid user ID format")
	}
	rows, err := r.pgDb.Query(ctx, `
		SELECT
			a.market_id::text,
			m.market_name,
			m.market_code,
			t.emblem,
			a.quantity,
			a.locked_qty,
			a.avg_cost,
			COALESCE(live.last_price, m.last_price) AS current_price
		FROM assets a
		JOIN markets m ON a.market_id = m.id
		JOIN teams t   ON m.team_id   = t.id
		LEFT JOIN LATERAL (
			SELECT last(price, time) AS last_price
			FROM trade_ticks
			WHERE market_id = m.id
			  AND time >= NOW() - INTERVAL '24 hours'
		) live ON true
		WHERE a.user_id = $1
		  AND a.quantity > 0
		ORDER BY a.updated_at DESC`,
		id,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var assets []AssetWithMarket
	for rows.Next() {
		var a AssetWithMarket
		if err := rows.Scan(&a.MarketId, &a.MarketName, &a.MarketCode, &a.Emblem,
			&a.Quantity, &a.LockedQty, &a.AvgCost, &a.CurrentPrice); err != nil {
			return nil, err
		}
		assets = append(assets, a)
	}
	return assets, rows.Err()
}

func (r *walletRepoUtils) GetWalletInfo(ctx context.Context, userID string) (*Wallet, utils.ErrorType, error) {
	var wallet Wallet
	id, err := uuid.Parse(userID)
	if err != nil {
		slog.Error("Invalid user ID format", "userID", userID, "error", err)
		return nil, utils.ErrBadRequest, errors.New("Please provide valid uuid UserId")
	}

	query := `
	SELECT id, user_id, balance, created_at, updated_at from wallets
	WHERE user_id = $1 ;
	`
	err = r.pgDb.QueryRow(ctx, query, id).Scan(&wallet.Id, &wallet.UserId, &wallet.Balance, &wallet.CreatedAt, &wallet.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			slog.Error("Wallet not found for user", "userID", userID, "error", err)
			return nil, utils.ErrNotFound, errors.New("Invalid User id not found associated wallet")
		}
		slog.Error("Database error getting wallet info", "error", err)
		return nil, utils.ErrInternal, err
	}
	return &wallet, utils.NoError, nil
}
