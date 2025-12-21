package services

import (
	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/raiashpanda007/rivon/internals/services/auth"
)

func InitAuthServices(pgDb *pgxpool.Pool, rDb *redis.Client, jwtSecret string, mailServerURL string) *auth.AuthServices {
	userRepo := auth.NewUserRepo(pgDb)
	tokenServices := auth.NewTokenServices(jwtSecret, pgDb)
	otpServices := auth.NewOTPServices(rDb, mailServerURL)

	authService := auth.NewAuthServices(userRepo, tokenServices, otpServices)
	return &authService

}
