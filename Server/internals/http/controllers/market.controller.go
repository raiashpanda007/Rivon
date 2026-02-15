package controllers

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/raiashpanda007/rivon/internals/services/markets"
	"github.com/raiashpanda007/rivon/internals/types"
	"github.com/raiashpanda007/rivon/internals/utils"
)

type MarketController interface {
	GetMarkets(res http.ResponseWriter, req *http.Request)
}

type marketControllerUtils struct {
	svc markets.MarketServices
}

func InitMarketControllers(pgDb *pgxpool.Pool) MarketController {
	svc := markets.NewMarketServices(pgDb)

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
