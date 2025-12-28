package auth

import (
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/github"
	"github.com/markbates/goth/providers/google"
)

func NewOAuth(OAuthKey string, cookieSecure bool, GoogleClientId string, GoogleClientSecret string, GithubClientId string, GithubClientSecret string, BaseAppUrl string) {
	store := sessions.NewCookieStore([]byte(OAuthKey))
	store.MaxAge(86400 * 15)
	store.Options.Path = "/"
	store.Options.HttpOnly = true
	store.Options.Secure = cookieSecure
	store.Options.SameSite = http.SameSiteLaxMode

	gothic.Store = store

	goth.UseProviders(
		google.New(GoogleClientId, GoogleClientSecret, BaseAppUrl+"/auth/google/callback"),
		github.New(GithubClientId, GithubClientSecret, BaseAppUrl+"/auth/github/callback", "user:email"),
	)
}
