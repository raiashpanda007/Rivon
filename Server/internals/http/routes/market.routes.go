package routes

import (
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/raiashpanda007/rivon/internals/config"
	"github.com/raiashpanda007/rivon/internals/http/controllers"
	"github.com/raiashpanda007/rivon/internals/http/middlewares"
)

func NewMarketRoutes(cfg *config.Config, pgDb *pgxpool.Pool, Controllers controllers.Controllers) chi.Router {
	router := chi.NewRouter()
	Middlewares := middlewares.NewMiddlewares(cfg, pgDb)
	router.Get("/", Controllers.GetMarkets)
	router.With(Middlewares.AuthVerifyMiddleware).Post("/create-order", Controllers.PlaceOrder)
	return router
}
