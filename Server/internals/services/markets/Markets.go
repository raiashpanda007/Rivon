package markets

import (
	"context"
	"errors"
	"log/slog"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/raiashpanda007/rivon/internals/types"
	"github.com/raiashpanda007/rivon/internals/utils"
)

type MarketServices interface {
	GetAllMarkets(ctx context.Context, teamDetails bool) ([]types.MarketTable, utils.ErrorType, error)
	GetMarket(ctx context.Context, marketID string, teamDetails bool) (types.MarketTable, utils.ErrorType, error)
	CreateMarket(ctx context.Context, teamID, marketName, marketCode string, lastPrice, volume24H, totalVolume, openPrice24H int64) (types.MarketTable, utils.ErrorType, error)
	PlaceOrder(ctx context.Context, userId, marketId uuid.UUID, price int64, quantity int64, orderType types.OrderTypes) error
}

type marketSvc struct {
	repo       MarketRepo
	orderRedis *redis.Client
}

func NewMarketServices(db *pgxpool.Pool, orderRedis *redis.Client) MarketServices {
	repo := NewMarketRepoServices(db)
	return &marketSvc{
		repo:       repo,
		orderRedis: orderRedis,
	}
}

func (r *marketSvc) GetAllMarkets(ctx context.Context, teamDetails bool) ([]types.MarketTable, utils.ErrorType, error) {
	markets, err := r.repo.GetAllMarkets(ctx, teamDetails)

	if err != nil {
		return markets, utils.ErrInternal, err
	}
	return markets, utils.NoError, nil
}

func (r *marketSvc) GetMarket(ctx context.Context, marketID string, teamDetails bool) (types.MarketTable, utils.ErrorType, error) {
	var market types.MarketTable
	marketIdUUID, err := uuid.Parse(marketID)
	if err != nil {
		slog.Error("Error in Get market Controller Invalid UUID :: ", "error ", err)
		return market, utils.ErrBadRequest, errors.New("Invalid market ID" + err.Error())
	}

	market, err = r.repo.GetMarket(ctx, marketIdUUID, teamDetails)

	if err != nil {
		slog.Error("Error in Get market controller in db query :: ", "error", err)
		if err == pgx.ErrNoRows {
			return market, utils.ErrNotFound, errors.New("No market exists of this ID")
		}

		return market, utils.ErrInternal, err

	}

	return market, utils.NoError, nil

}

func (r *marketSvc) CreateMarket(ctx context.Context, teamID, marketName, marketCode string, lastPrice, volume24H, totalVolume, openPrice24H int64) (types.MarketTable, utils.ErrorType, error) {
	var market types.MarketTable

	teamIDUuid, err := uuid.Parse(teamID)
	if err != nil {
		slog.Error("Error in create market Controller Invalid UUID :: ", "error ", err)
		return market, utils.ErrBadRequest, errors.New("Invalid team Id")
	}
	market, err = r.repo.CreateMarket(ctx, teamIDUuid, marketName, marketCode, lastPrice, volume24H, totalVolume, openPrice24H)

	if err != nil {
		slog.Error("Error in create market controller in db query :: ", "error", err)
		return market, utils.ErrInternal, err
	}

	return market, utils.NoError, nil

}

func (r *marketSvc) PlaceOrder(ctx context.Context, userId, marketId uuid.UUID, price int64, quantity int64, orderType types.OrderTypes) error {

	orderId := uuid.New()

	_, err := r.orderRedis.XAdd(ctx, &redis.XAddArgs{
		Stream: "ORDERS_" + marketId.String(),
		Values: map[string]interface{}{
			"orderId":   orderId.String(),
			"userId":    userId.String(),
			"marketId":  marketId.String(),
			"price":     price,
			"quantity":  int(quantity),
			"orderType": string(orderType),
		},
	}).Result()

	if err != nil {
		slog.Error("Unable to write on redis stream.")
		return err
	}
	slog.Info("Order placed successfully in redis stream for market ", "marketId", marketId, "userId", userId, "price", price, "quantity", quantity, "orderType", orderType)

	return nil
}
