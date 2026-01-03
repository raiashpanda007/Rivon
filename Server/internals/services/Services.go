package services

import (
	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/raiashpanda007/rivon/internals/services/auth"
	"github.com/raiashpanda007/rivon/internals/services/wallet"
)

func InitAuthServices(pgDb *pgxpool.Pool, otpRedis *redis.Client, jwtSecret string, mailServerURL string) *auth.AuthServices {
	userRepo := auth.NewUserRepo(pgDb)
	tokenServices := auth.NewTokenServices(jwtSecret, pgDb)
	otpServices := auth.NewOTPServices(otpRedis, mailServerURL)

	authService := auth.NewAuthServices(userRepo, tokenServices, otpServices)
	return &authService

}

func InitWalletServices(pgDb *pgxpool.Pool) *wallet.WalletServices {
	walletRepo := wallet.NewWalletRepo(pgDb)
	walletServices := wallet.NewWalletServices(walletRepo)
	return &walletServices
}
