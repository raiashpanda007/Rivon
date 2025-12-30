package routes

import (
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/raiashpanda007/rivon/internals/config"
	"github.com/raiashpanda007/rivon/internals/http/controllers"
	"github.com/raiashpanda007/rivon/internals/http/middlewares"
)

func NewWalletRouter(pgDb *pgxpool.Pool, cfg *config.Config, Controller controllers.Controllers) chi.Router {
	router := chi.NewRouter()
	Middlewares := middlewares.NewMiddlewares(cfg, pgDb)
	router.With(Middlewares.AuthVerifyMiddleware).Get("/me", Controller.GetWallet)
	return router
}
