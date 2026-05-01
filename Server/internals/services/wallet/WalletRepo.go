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

type WalletRepo interface {
	GetWalletInfo(ctx context.Context, userID string) (*Wallet, utils.ErrorType, error)
	GetUserAssets(ctx context.Context, userID string) ([]Asset, error)
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
