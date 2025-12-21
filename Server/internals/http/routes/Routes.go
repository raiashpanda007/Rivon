package routes

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/raiashpanda007/rivon/internals/config"
	"github.com/raiashpanda007/rivon/internals/utils"
)

func InitRouters(cfg *config.Config, PgDb *pgxpool.Pool, OtpRedis *redis.Client) chi.Router {
	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Timeout(60 * time.Second))
	router.Get("/api/rivon/health-check", func(w http.ResponseWriter, r *http.Request) {
		utils.WriteJson(w, http.StatusOK, utils.Response[string]{
			Status:  200,
			Data:    "Server is running fine ",
			Message: "Hi , this is api-server",
			Heading: "STATUS OK",
		})
	})
	AuthRouter := NewAuthRouter(cfg, PgDb, OtpRedis)

	router.Mount("/api/rivon/auth", AuthRouter)
	return router
}
