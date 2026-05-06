package routes

import (
	"github.com/go-chi/chi/v5"
	"github.com/raiashpanda007/rivon/internals/http/controllers"
)

func NewCandleRoutes(Controllers controllers.Controllers) chi.Router {
	router := chi.NewRouter()
	router.Get("/stream", Controllers.StreamCandles)
	router.Get("/history", Controllers.GetCandleHistory)
	return router
}
