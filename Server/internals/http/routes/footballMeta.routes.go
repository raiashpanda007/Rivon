package routes

import (
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/raiashpanda007/rivon/internals/config"
	"github.com/raiashpanda007/rivon/internals/http/controllers"
)

func NewFootBallMetaRoutes(cfg *config.Config, pgDb *pgxpool.Pool, Controllers controllers.Controllers) chi.Router {
	router := chi.NewRouter()
	router.Get("/standings", Controllers.GetCompetitionTeamStandings)
	router.Get("/competitions", Controllers.GetCompetitions)
	router.Get("/seasons", Controllers.GetAllSeasons)
	router.Get("/knockout", Controllers.GetKnockoutRounds)
	return router
}
