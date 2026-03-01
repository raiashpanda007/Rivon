package controllers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/raiashpanda007/rivon/internals/services"
	"github.com/raiashpanda007/rivon/internals/services/auth"
	"github.com/raiashpanda007/rivon/internals/services/markets"
	"github.com/raiashpanda007/rivon/internals/types"
	"github.com/raiashpanda007/rivon/internals/utils"
)

type MarketController interface {
	GetMarkets(res http.ResponseWriter, req *http.Request)
	PlaceOrder(res http.ResponseWriter, req *http.Request)
}

type marketControllerUtils struct {
	svc markets.MarketServices
}

func InitMarketControllers(pgDb *pgxpool.Pool, orderRedis *redis.Client) MarketController {
	svc := services.InitMarketServices(pgDb, orderRedis)
	return &marketControllerUtils{
		svc: svc,
	}
}

func (r *marketControllerUtils) GetMarkets(res http.ResponseWriter, req *http.Request) {
	marketIdStr := req.URL.Query().Get("marketId")
	teamDetailsStr := req.URL.Query().Get("teamDetails")
	teamDetails := teamDetailsStr == "true"

	if marketIdStr == "" {
		markets, errType, err := r.svc.GetAllMarkets(req.Context(), teamDetails)
		if err != nil {
			slog.Error("Get all markets error ", "error", err)
			utils.WriteJson(res, utils.ErrorMap[errType].StatusCode, utils.GenerateError(errType, err))
			return
		}

		utils.WriteJson(res, http.StatusOK, utils.Response[[]types.MarketTable]{
			Status:  201,
			Heading: "Market Details",
			Message: "All market details",
			Data:    markets,
		})
		return
	} else {
		_, err := uuid.Parse(marketIdStr)
		if err != nil {
			slog.Error("Invalid league Id", "error", err)
			utils.WriteJson(res, http.StatusBadRequest, utils.GenerateError(utils.ErrBadRequest, errors.New("Invalid market Id")))
			return
		}

		markets, errType, err := r.svc.GetMarket(req.Context(), marketIdStr, teamDetails)

		if err != nil {
			utils.WriteJson(res, utils.ErrorMap[errType].StatusCode, utils.GenerateError(errType, err))
			return
		}

		utils.WriteJson(res, http.StatusOK, utils.Response[types.MarketTable]{
			Status:  200,
			Message: "Market Details",
			Heading: "Market Details",
			Data:    markets,
		})

	}

}

func (r *marketControllerUtils) PlaceOrder(res http.ResponseWriter, req *http.Request) {
	slog.Info("PLACE ORDER CALLED ...")
	userCred, ok := req.Context().Value("USER").(*auth.User)

	if !ok {
		slog.Error("User credentials not found in context")
		utils.WriteJson(res, http.StatusForbidden, utils.GenerateError(utils.ErrUnauthorized, errors.New("You are not authorized to place an order")))
		return
	}
	if userCred.Verified == false {
		slog.Error("User email not verified")
		utils.WriteJson(res, http.StatusForbidden, utils.GenerateError(utils.ErrUnauthorized, errors.New("Please verify your email to place an order")))
		return
	}
	var order types.MarketOrder

	err := json.NewDecoder(req.Body).Decode(&order)

	if err != nil {
		slog.Error("Error parsing order details", "error", err)
		utils.WriteJson(res, http.StatusBadRequest, utils.GenerateError(utils.ErrBadRequest, errors.New("Invalid order details")))
		return
	}

	if order.Quantity <= 0 || order.Price <= 0 || order.OrderType != "BUY" && order.OrderType != "SELL" {
		slog.Error("Invalid order parameters", "order", order)
		utils.WriteJson(res, http.StatusBadRequest, utils.GenerateError(utils.ErrBadRequest, errors.New("Invalid order parameters")))
		return
	}

	err = r.svc.PlaceOrder(req.Context(), userCred.Id, order.MarketId, order.Price, order.Quantity, order.OrderType)

	if err != nil {
		slog.Error("Error placing order", "error", err)
		utils.WriteJson(res, http.StatusInternalServerError, utils.GenerateError(utils.ErrInternal, errors.New("Failed to place order")))
		return
	}

	utils.WriteJson(res, http.StatusOK, utils.Response[string]{
		Status:  200,
		Message: "Order placed successfully",
		Heading: "Order Placed",
		Data:    "Order has been placed successfully",
	})

}
