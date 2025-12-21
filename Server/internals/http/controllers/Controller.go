package controllers

import (
	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Controllers struct {
	AuthController
}

func NewController(pgDb *pgxpool.Pool, rDb *redis.Client, jwtSecret, mailServerURL string, cookieSecure bool) Controllers {
	auth := InitAuthController(pgDb, rDb, jwtSecret, mailServerURL, cookieSecure)
	return Controllers{
		AuthController: auth,
	}

}
