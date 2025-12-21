package routes

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/raiashpanda007/rivon/internals/config"
	"github.com/raiashpanda007/rivon/internals/http/controllers"
	"github.com/raiashpanda007/rivon/internals/http/middlewares"
)

func NewAuthRouter(cfg *config.Config, pgDb *pgxpool.Pool, OtpRedis *redis.Client) chi.Router {
	router := chi.NewRouter()
	Controllers := controllers.NewController(pgDb, OtpRedis, cfg.Auth.AuthSecret, cfg.MailServerURL, cfg.Server.CookieSecure)
	Middlewares := middlewares.NewMiddlewares(cfg, pgDb)
	router.Route("/credentials", func(r chi.Router) {
		r.Post("/signin", Controllers.CredentialSignIn)
		r.Post("/signup", Controllers.CredentialSignUp)
		r.Post("/refresh", Controllers.CredentialRefresh)
		r.With(Middlewares.AuthVerifyMiddleware).Delete("/signout", Controllers.CredentialSignOut)
	})
	router.With(Middlewares.AuthVerifyMiddleware).Post("/verify/send_otp", Controllers.SendVerifyOTP)
	router.With(Middlewares.AuthVerifyMiddleware).Post("/verify/verify_otp", Controllers.VerifyOTP)
	return router
}
