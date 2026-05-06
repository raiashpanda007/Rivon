package wallet

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/raiashpanda007/rivon/internals/utils"
)

type WalletServices interface {
	GetWalletState(ctx context.Context, userID string) (*Wallet, utils.ErrorType, error)
	GetTransactions(ctx context.Context, userID string, limit, offset int) ([]Transaction, utils.ErrorType, error)
	GetAssets(ctx context.Context, userID string) ([]AssetWithMarket, utils.ErrorType, error)
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

func (r *walletServiceUtils) GetAssets(ctx context.Context, userID string) ([]AssetWithMarket, utils.ErrorType, error) {
	assets, err := r.repo.GetUserAssetsWithMarket(ctx, userID)
	if err != nil {
		return nil, utils.ErrInternal, err
	}
	if assets == nil {
		assets = []AssetWithMarket{}
	}
	return assets, utils.NoError, nil
}

func (r *walletServiceUtils) GetTransactions(ctx context.Context, userID string, limit, offset int) ([]Transaction, utils.ErrorType, error) {
	txns, err := r.repo.GetTransactions(ctx, userID, limit, offset)
	if err != nil {
		return nil, utils.ErrInternal, err
	}
	return txns, utils.NoError, nil
}

func (r *walletServiceUtils) GetWalletState(ctx context.Context, userID string) (*Wallet, utils.ErrorType, error) {
	val, err := r.userMapRedis.Get(ctx, userID).Result()
	if err == nil {
		var cached walletRedisData
		if jsonErr := json.Unmarshal([]byte(val), &cached); jsonErr == nil {
			uid, _ := uuid.Parse(userID)
			return &Wallet{UserId: uid, Balance: cached.Balance}, utils.NoError, nil
		}
	}

	wallet, errType, dbErr := r.repo.GetWalletInfo(ctx, userID)
	if dbErr != nil {
		return nil, errType, dbErr
	}

	if err := r.pushToRedis(ctx, userID, wallet); err != nil {
		slog.Error("Unable to save wallet to redis", "userID", userID, "err", err)
	}

	return wallet, utils.NoError, nil
}
