package wallet

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"time"

	"github.com/google/uuid"
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

type WalletRepo interface {
	GetWalletInfo(ctx context.Context, userID string) (*Wallet, utils.ErrorType, error)
}

func NewWalletRepo(pgDB *pgxpool.Pool) WalletRepo {
	return &walletRepoUtils{
		pgDb: pgDB,
	}

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
		if err == sql.ErrNoRows {
			slog.Error("Wallet not found for user", "userID", userID, "error", err)
			return nil, utils.ErrNotFound, errors.New("Invalid User id not found associated wallet")
		}
		slog.Error("Database error getting wallet info", "error", err)
		return nil, utils.ErrInternal, err
	}
	return &wallet, utils.NoError, nil
}
