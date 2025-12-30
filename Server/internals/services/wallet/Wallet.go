package wallet

import (
	"context"

	"github.com/raiashpanda007/rivon/internals/utils"
)

type WalletServices interface {
	GetWalletState(ctx context.Context, userID string) (*Wallet, utils.ErrorType, error)
}

type walletServiceUtils struct {
	repo WalletRepo
}

func NewWalletServices(walletRepo WalletRepo) WalletServices {
	return &walletServiceUtils{repo: walletRepo}
}

func (r *walletServiceUtils) GetWalletState(ctx context.Context, userID string) (*Wallet, utils.ErrorType, error) {
	return r.repo.GetWalletInfo(ctx, userID)
}
