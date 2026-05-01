package wallet

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/go-redis/redis/v8"
	"github.com/raiashpanda007/rivon/internals/utils"
)

type WalletServices interface {
	GetWalletState(ctx context.Context, userID string) (*Wallet, utils.ErrorType, error)
}

type walletServiceUtils struct {
	repo         WalletRepo
	userMapRedis *redis.Client
}

func NewWalletServices(walletRepo WalletRepo, userMapRedis *redis.Client) WalletServices {
	return &walletServiceUtils{repo: walletRepo, userMapRedis: userMapRedis}
}

type walletRedisData struct {
	Balance int64   `json:"balance"`
	Assets  []Asset `json:"assets"`
}

func (r *walletServiceUtils) pushToRedis(ctx context.Context, userID string, w *Wallet) error {
	assets, err := r.repo.GetUserAssets(ctx, userID)
	if err != nil {
		slog.Error("Unable to fetch assets for redis push", "userID", userID, "err", err)
		assets = []Asset{}
	}

	data, err := json.Marshal(walletRedisData{Balance: w.Balance, Assets: assets})
	if err != nil {
		return err
	}
	return r.userMapRedis.Set(ctx, userID, data, 0).Err()
}

func (r *walletServiceUtils) GetWalletState(ctx context.Context, userID string) (*Wallet, utils.ErrorType, error) {
	wallet, errType, err := r.repo.GetWalletInfo(ctx, userID)
	if err != nil {
		return nil, errType, err
	}

	if err := r.pushToRedis(ctx, userID, wallet); err != nil {
		slog.Error("Unable to save wallet to redis", "userID", userID, "err", err)
		return nil, utils.ErrInternal, err
	}

	return wallet, utils.NoError, nil
}
