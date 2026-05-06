package config

import (
	"log"
	"log/slog"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	ENVIROMENT      string
	TRADE_REDIS_URL string
	PG_DB_URL       string
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
	slog.Info("Loading Config var for Db writter")

	err := godotenv.Load()
	if err != nil {
		log.Fatal("ERROR :: IN READING .env ", err.Error())
		return nil
	}

	cfg.ENVIROMENT = mustEnv("ENVIROMENT")
	cfg.PG_DB_URL = mustEnv("PG_DB_URL")
	cfg.TRADE_REDIS_URL = mustEnv("TRADE_REDIS_URL")
	return &cfg
}
