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

func (r *walletServiceUtils) GetWalletState(ctx context.Context, userID string) (*Wallet, utils.ErrorType, error) {
	wallet, errType, err := r.repo.GetWalletInfo(ctx, userID)
	if err != nil {
		return nil, errType, err
	}

	count, err := r.userMapRedis.Exists(ctx, userID).Result()
	if err != nil {
		slog.Error("Unable to check the user status in redis instance", slog.Any("Error :: ", err))
		return nil, utils.ErrInternal, err
	}

	if count > 0 {
		return wallet, utils.NoError, err
	} else {

		data, err := json.Marshal(wallet)

		if err != nil {
			slog.Error("Unable to marshal data", slog.Any("Error :: ", err))
			return nil, utils.ErrInternal, err
		}

		err = r.userMapRedis.Set(ctx, userID, data, 0).Err()

		if err != nil {
			slog.Error("Unable to save the user wallet with info in redis", slog.Any("ERROR :: ", err))
			return nil, utils.ErrInternal, err
		}

	}

	return wallet, utils.NoError, nil
}
