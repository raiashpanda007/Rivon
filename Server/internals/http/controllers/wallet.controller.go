package controllers

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/raiashpanda007/rivon/internals/services"
	"github.com/raiashpanda007/rivon/internals/services/auth"
	"github.com/raiashpanda007/rivon/internals/services/wallet"
	"github.com/raiashpanda007/rivon/internals/utils"
)

type WalletController interface {
	GetWallet(res http.ResponseWriter, req *http.Request)
	GetTransactions(res http.ResponseWriter, req *http.Request)
	GetAssets(res http.ResponseWriter, req *http.Request)
}

type walletControllerUtils struct {
	svc wallet.WalletServices
}

func InitWalletController(pgDb *pgxpool.Pool, userMapRedis *redis.Client) WalletController {
	walletSvc := services.InitWalletServices(pgDb, userMapRedis)
	return &walletControllerUtils{
		svc: *walletSvc,
	}
}

func (r *walletControllerUtils) GetTransactions(res http.ResponseWriter, req *http.Request) {
	user, ok := req.Context().Value("USER").(*auth.User)
	if !ok {
		utils.WriteJson(res, http.StatusUnauthorized, utils.GenerateError(utils.ErrUnauthorized, errors.New("please login again")))
		return
	}

	limit := 50
	offset := 0
	if l := req.URL.Query().Get("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 && v <= 200 {
			limit = v
		}
	}
	if o := req.URL.Query().Get("offset"); o != "" {
		if v, err := strconv.Atoi(o); err == nil && v >= 0 {
			offset = v
		}
	}

	txns, errType, err := r.svc.GetTransactions(req.Context(), user.Id.String(), limit, offset)
	if err != nil {
		slog.Error("GetTransactions error", "error", err)
		utils.WriteJson(res, utils.ErrorMap[errType].StatusCode, utils.GenerateError(errType, err))
		return
	}
	if txns == nil {
		txns = []wallet.Transaction{}
	}
	utils.WriteJson(res, http.StatusOK, utils.Response[[]wallet.Transaction]{
		Heading: "Status Ok",
		Message: "Fetched your transaction history",
		Data:    txns,
		Status:  http.StatusOK,
	})
}

func (r *walletControllerUtils) GetAssets(res http.ResponseWriter, req *http.Request) {
	user, ok := req.Context().Value("USER").(*auth.User)
	if !ok {
		utils.WriteJson(res, http.StatusUnauthorized, utils.GenerateError(utils.ErrUnauthorized, errors.New("please login again")))
		return
	}
	assets, errType, err := r.svc.GetAssets(req.Context(), user.Id.String())
	if err != nil {
		slog.Error("GetAssets error", "error", err)
		utils.WriteJson(res, utils.ErrorMap[errType].StatusCode, utils.GenerateError(errType, err))
		return
	}
	utils.WriteJson(res, http.StatusOK, utils.Response[[]wallet.AssetWithMarket]{
		Heading: "Status Ok",
		Message: "Fetched your portfolio",
		Data:    assets,
		Status:  http.StatusOK,
	})
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
