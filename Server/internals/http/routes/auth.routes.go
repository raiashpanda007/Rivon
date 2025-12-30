package routes

import (
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/markbates/goth/gothic"
	"github.com/raiashpanda007/rivon/internals/config"
	"github.com/raiashpanda007/rivon/internals/http/controllers"
	"github.com/raiashpanda007/rivon/internals/http/middlewares"
)

func NewAuthRouter(cfg *config.Config, pgDb *pgxpool.Pool, Controllers controllers.Controllers) chi.Router {
	router := chi.NewRouter()
	Middlewares := middlewares.NewMiddlewares(cfg, pgDb)
	router.Get("/{provider}", gothic.BeginAuthHandler)
	router.Get("/{provider}/callback", Controllers.OAuthLogin)
	router.Route("/credentials", func(r chi.Router) {
		r.Post("/signin", Controllers.CredentialSignIn)
		r.Post("/signup", Controllers.CredentialSignUp)
		r.Post("/refresh", Controllers.CredentialRefresh)
		r.With(Middlewares.AuthVerifyMiddleware).Delete("/signout", Controllers.CredentialSignOut)
	})
	router.With(Middlewares.AuthVerifyMiddleware).Get("/me", Controllers.Me)
	router.With(Middlewares.AuthVerifyMiddleware).Post("/verify/send_otp", Controllers.SendVerifyOTP)
	router.With(Middlewares.AuthVerifyMiddleware).Post("/verify/verify_otp", Controllers.VerifyOTP)
	return router
}
