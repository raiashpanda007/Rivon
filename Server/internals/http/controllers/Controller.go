package controllers

import (
	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v5/pgxpool"
	pubsub "github.com/raiashpanda007/rivon/internals/pub-sub"
)

type Controllers struct {
	AuthController
	WalletController
	FootballMetaController
	MarketController
}

func NewController(pgDb *pgxpool.Pool, otpRedis *redis.Client, orderRedis *redis.Client, jwtSecret, mailServerURL string, cookieSecure bool, clientBaseURL string, PubSubConn pubsub.Pubsub) Controllers {
	auth := InitAuthController(pgDb, otpRedis, jwtSecret, mailServerURL, cookieSecure, clientBaseURL)
	walletController := InitWalletController(pgDb)
	footballMetaController := InitFootballMetaController(pgDb)
	marketController := InitMarketControllers(pgDb, orderRedis, PubSubConn)
	return Controllers{
		AuthController:         auth,
		WalletController:       walletController,
		FootballMetaController: footballMetaController,
		MarketController:       marketController,
	}

}
