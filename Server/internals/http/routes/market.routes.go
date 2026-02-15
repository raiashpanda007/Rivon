package routes

import (
	"github.com/go-chi/chi/v5"
	"github.com/raiashpanda007/rivon/internals/http/controllers"
)

func NewMarketRoutes(Controllers controllers.Controllers) chi.Router {
	router := chi.NewRouter()
	router.Get("/", Controllers.GetMarkets)
	return router
}
