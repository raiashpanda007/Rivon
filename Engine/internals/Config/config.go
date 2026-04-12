package config

import (
	"log"
	"log/slog"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	ENVIRONMENT           string
	ORDER_REDIS_URL       string
	TRADE_REDIS_URL       string
	API_PUB_SUB_REDIS_URL string
	PG_URL                string
	WS_PUB_SUB_REDIS_URL  string
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
	slog.Info("Loading Config for env .... ")

	err := godotenv.Load()
	if err != nil {
		log.Fatal("ERROR :: IN READING .env", err.Error())
		return nil
	}

	cfg.ENVIRONMENT = mustEnv("ENVIRONMENT")

	cfg.ORDER_REDIS_URL = mustEnv("ORDER_REDIS_URL")

	cfg.TRADE_REDIS_URL = mustEnv("TRADE_REDIS_URL")

	cfg.API_PUB_SUB_REDIS_URL = mustEnv("API_PUB_SUB_REDIS_URL")

	cfg.PG_URL = mustEnv("PG_URL")

	cfg.WS_PUB_SUB_REDIS_URL = mustEnv("WS_PUB_SUB_REDIS_URL")
	return &cfg
}
