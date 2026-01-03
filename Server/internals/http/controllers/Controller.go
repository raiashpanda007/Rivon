package controllers

import (
	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Controllers struct {
	AuthController
	WalletController
}

func NewController(pgDb *pgxpool.Pool, otpRedis *redis.Client, jwtSecret, mailServerURL string, cookieSecure bool, clientBaseURL string) Controllers {
	auth := InitAuthController(pgDb, otpRedis, jwtSecret, mailServerURL, cookieSecure, clientBaseURL)
	walletController := InitWalletController(pgDb)
	return Controllers{
		AuthController:   auth,
		WalletController: walletController,
	}

}
