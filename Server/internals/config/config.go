package config

import (
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type DataBase struct {
	PgURL string
}
type AuthConfig struct {
	AuthSecret         string
	GoogleClientID     string
	GoogleClientSecret string
	GithubClientID     string
	GithubClientSecret string
}
type HttpServer struct {
	ApiServerAddr string
}
type Config struct {
	Auth   AuthConfig
	Server HttpServer
	Db     DataBase
}

func mustEnv(key string) string {
	val := strings.TrimSpace(os.Getenv(key))
	if val == "" {
		log.Fatalf("ERROR :: MISSING OR EMPTY env VARS  :: %s", key)
	}
	return val
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
	}
	var httpCfg = HttpServer{
		ApiServerAddr: mustEnv("API_SERVER_URL"),
	}
	var dbCfg = DataBase{
		PgURL: mustEnv("DATABASE_POSTGRES_URL"),
	}

	cfg.Auth = authCfg
	cfg.Server = httpCfg
	cfg.Db = dbCfg
	return &cfg
}
