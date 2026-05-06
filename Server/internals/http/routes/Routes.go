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
	pubsub "github.com/raiashpanda007/rivon/internals/pub-sub"
	"github.com/raiashpanda007/rivon/internals/registry"
	"github.com/raiashpanda007/rivon/internals/utils"
)

func InitRouters(cfg *config.Config, PgDb *pgxpool.Pool, OtpRedis *redis.Client, OrderRedis *redis.Client, PubSubConn pubsub.Pubsub, reg *registry.Registry, UserMapRedis *redis.Client, TradeRedis *redis.Client) chi.Router {
	router := chi.NewRouter()
	Controllers := controllers.NewController(PgDb, OtpRedis, OrderRedis, cfg.Auth.AuthSecret, cfg.MailServerURL, cfg.Server.CookieSecure, cfg.ClientBaseURL, PubSubConn, reg, UserMapRedis, TradeRedis)
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

	router.Get("/api/rivon/health-check", func(w http.ResponseWriter, r *http.Request) {
		utils.WriteJson(w, http.StatusOK, utils.Response[string]{
			Status:  200,
			Data:    "Server is running fine ",
			Message: "Hi , this is api-server",
			Heading: "STATUS OK",
		})
	})

	// SSE endpoint — no timeout middleware; EventSource reconnects automatically.
	CandleRouter := NewCandleRoutes(Controllers)
	router.Mount("/api/rivon/candles", CandleRouter)

	// All other routes with a 60-second request timeout.
	router.Group(func(r chi.Router) {
		r.Use(middleware.Timeout(60 * time.Second))
		AuthRouter := NewAuthRouter(cfg, PgDb, Controllers)
		WalletRouter := NewWalletRouter(PgDb, cfg, Controllers)
		FootBallMetaRouter := NewFootBallMetaRoutes(cfg, PgDb, Controllers)
		MarketRouter := NewMarketRoutes(cfg, PgDb, Controllers)
		r.Mount("/api/rivon/auth", AuthRouter)
		r.Mount("/api/rivon/wallet", WalletRouter)
		r.Mount("/api/rivon/football-meta", FootBallMetaRouter)
		r.Mount("/api/rivon/markets", MarketRouter)
	})
	return router
}
