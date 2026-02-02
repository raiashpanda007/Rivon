package controllers

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/raiashpanda007/rivon/internals/services"
	"github.com/raiashpanda007/rivon/internals/services/auth"
	"github.com/raiashpanda007/rivon/internals/services/wallet"
	"github.com/raiashpanda007/rivon/internals/utils"
)

type WalletController interface {
	GetWallet(res http.ResponseWriter, req *http.Request)
}

type walletControllerUtils struct {
	svc wallet.WalletServices
}

func InitWalletController(pgDb *pgxpool.Pool) WalletController {
	walletSvc := services.InitWalletServices(pgDb)
	return &walletControllerUtils{
		svc: *walletSvc,
	}
}

func (r *walletControllerUtils) GetWallet(res http.ResponseWriter, req *http.Request) {
	user, ok := req.Context().Value("USER").(*auth.User)
	if !ok {
		slog.Error("Failed to retrieve user from context")
		utils.WriteJson(res, http.StatusUnauthorized, utils.GenerateError(utils.ErrUnauthorized, errors.New("Please login again to get your wallet details")))
		return
	}
	userWalllet, errType, err := r.svc.GetWalletState(req.Context(), user.Id.String())

	if err != nil {
		slog.Error("GetWalletState service error", "error", err)
		utils.WriteJson(res, utils.ErrorMap[errType].StatusCode, utils.GenerateError(errType, err))
		return
	}

	utils.WriteJson(res, http.StatusOK, utils.Response[wallet.Wallet]{
		Heading: "Status Ok",
		Message: "Fetched your wallet details",
		Data:    *userWalllet,
		Status:  http.StatusOK,
	})

}
