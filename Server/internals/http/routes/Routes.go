package routes

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/raiashpanda007/rivon/internals/config"
	"github.com/raiashpanda007/rivon/internals/http/controllers"
	"github.com/raiashpanda007/rivon/internals/utils"
)

func InitRouters(cfg *config.Config, PgDb *pgxpool.Pool, OtpRedis *redis.Client) chi.Router {
	router := chi.NewRouter()
	Controllers := controllers.NewController(PgDb, OtpRedis, cfg.Auth.AuthSecret, cfg.MailServerURL, cfg.Server.CookieSecure, cfg.ClientBaseURL)
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{
			"http://localhost:3000",
			"http://localhost:3001",
			"http://localhost:3002",
			"http://localhost:5173",
		},
		AllowedMethods: []string{
			"GET",
			"POST",
			"PUT",
			"PATCH",
			"DELETE",
			"OPTIONS",
		},
		AllowedHeaders: []string{
			"Accept",
			"Authorization",
			"Content-Type",
			"X-CSRF-Token",
		},
		ExposedHeaders: []string{
			"Set-Cookie",
		},
		AllowCredentials: true,
		MaxAge:           300,
	}))

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

	AuthRouter := NewAuthRouter(cfg, PgDb, Controllers)
	WalletRouter := NewWalletRouter(PgDb, cfg, Controllers)
	router.Mount("/api/rivon/auth", AuthRouter)
	router.Mount("/api/rivon/wallet", WalletRouter)

	return router
}
