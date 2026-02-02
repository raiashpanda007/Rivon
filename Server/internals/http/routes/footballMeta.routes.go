package routes

import (
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/raiashpanda007/rivon/internals/config"
	"github.com/raiashpanda007/rivon/internals/http/controllers"
	"github.com/raiashpanda007/rivon/internals/http/middlewares"
)

func NewFootBallMetaRoutes(cfg *config.Config, pgDb *pgxpool.Pool, Controllers controllers.Controllers) chi.Router {
	router := chi.NewRouter()
	Middlewares := middlewares.NewMiddlewares(cfg, pgDb)
	router.With(Middlewares.AuthVerifyMiddleware).Get("/standings", Controllers.GetCompetitionTeamStandings)
	router.With(Middlewares.AuthVerifyMiddleware).Get("/competitions", Controllers.GetCompetitions)
	router.With(Middlewares.AuthVerifyMiddleware).Get("/seasons", Controllers.GetAllSeasons)
	return router
}
