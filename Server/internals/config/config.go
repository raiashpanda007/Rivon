package config

import (
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type DataBase struct {
	PgURL       string
	OTPRedisURL string
}
type AuthConfig struct {
	AuthSecret         string
	GoogleClientID     string
	GoogleClientSecret string
	GithubClientID     string
	GithubClientSecret string
	GoAuthSecret       string
}
type HttpServer struct {
	ApiServerAddr string
	CookieSecure  bool
}
type Config struct {
	Auth          AuthConfig
	Server        HttpServer
	Db            DataBase
	MailServerURL string
	IsProduction  bool
	ClientBaseURL string
}

func mustEnv(key string) string {
	val := strings.TrimSpace(os.Getenv(key))
	if val == "" {
		log.Fatalf("ERROR :: MISSING OR EMPTY env VARS  :: %s", key)
	}
	return val
}
func stringTobool(str string) bool {
	if str == "TRUE" || str == "true" || str == "True" {
		return true
	}
	return false
}
func MustLoad() *Config {
	var cfg Config
	log.Print("Loading Config ... ")
	err := godotenv.Load()
	if err != nil {
		log.Fatalln("ERROR ::  IN READING .env ", err.Error())
		return nil
	}
	var authCfg = AuthConfig{
		AuthSecret:         mustEnv("AUTH_SECRET"),
		GoogleClientID:     mustEnv("GOOGLE_AUTH_CLIENT_ID"),
		GoogleClientSecret: mustEnv("GOOGLE_AUTH_CLIENT_SECRET"),
		GithubClientID:     mustEnv("GITHUB_AUTH_CLIENT_ID"),
		GithubClientSecret: mustEnv("GITHUB_AUTH_CLIENT_SECRET"),
		GoAuthSecret:       mustEnv("GO_AUTH_COOKIE_SECRET"),
	}
	var httpCfg = HttpServer{
		ApiServerAddr: mustEnv("API_SERVER_URL"),
		CookieSecure:  stringTobool(mustEnv("COOKIE_SECURE")),
	}
	var dbCfg = DataBase{
		PgURL:       mustEnv("DATABASE_POSTGRES_URL"),
		OTPRedisURL: mustEnv("OTP_REDIS_URL"),
	}

	cfg.Auth = authCfg
	cfg.Server = httpCfg
	cfg.Db = dbCfg
	cfg.MailServerURL = mustEnv("MAIL_SERVER_URL")
	cfg.IsProduction = stringTobool(mustEnv("PRODUCTION"))
	cfg.ClientBaseURL = mustEnv("CLIENT_BASE_URL")
	return &cfg
}
