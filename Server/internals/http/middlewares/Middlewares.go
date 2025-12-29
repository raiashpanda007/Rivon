package middlewares

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/raiashpanda007/rivon/internals/config"
	"github.com/raiashpanda007/rivon/internals/services/auth"
)

type Middlewares struct {
	AuthVerifyMiddleware func(http.Handler) http.Handler
}

func NewMiddlewares(cfg *config.Config, Db *pgxpool.Pool) Middlewares {
	tokenServices := auth.NewTokenServices(cfg.Auth.AuthSecret, Db)
	verifyMiddleware := VerifyMiddleware(tokenServices)
	return Middlewares{
		AuthVerifyMiddleware: verifyMiddleware,
	}
}
