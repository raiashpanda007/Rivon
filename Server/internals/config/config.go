package config

import (
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"os"
	"strings"
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
type FootballStaticCountry struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Code   string `json:"code"`
	Emblem string `json:"emblem"`
}

type FootballStaticLeague struct {
	ID      int                   `json:"id"`
	Name    string                `json:"name"`
	Code    string                `json:"code"`
	Emblem  string                `json:"emblem"`
	Country FootballStaticCountry `json:"country"`
}

type FootballStaticCompetitions struct {
	Competitions []FootballStaticLeague `json:"competitions"`
}

type Config struct {
	Auth               AuthConfig
	Server             HttpServer
	Db                 DataBase
	MailServerURL      string
	IsProduction       bool
	ClientBaseURL      string
	FootBallStaticData FootballStaticCompetitions
	FootbalOrgApiKey   []string
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

func ReadFromJSON() (*FootballStaticCompetitions, error) {
	data, err := os.ReadFile("./internals/config/football_org_static.json")
	if err != nil {
		log.Fatal("Not football static json file found please get it up ready first :: ", err)
		return nil, err
	}
	var competitions FootballStaticCompetitions
	err = json.Unmarshal(data, &competitions)
	return &competitions, err
}

func getAllFootBallOrgAPIKeys() []string {
	var keys []string

	for i := 1; ; i++ {
		key := os.Getenv(fmt.Sprintf("FOOTBALL_API_KEY_%d", i))
		if key == "" {
			break
		}
		keys = append(keys, key)
	}

	return keys
}
func MustLoad() *Config {
	var cfg Config
	log.Print("Loading Config ... ")
	err := godotenv.Load()
	if err != nil {
		log.Fatalln("ERROR ::  IN READING .env ", err.Error())
		return nil
	}

	CompetiionStaticMetaData, err := ReadFromJSON()
	if err != nil {
		log.Fatal("Not a valid Static json to be read from football static json file :: ", err)
	}
	cfg.FootBallStaticData = *CompetiionStaticMetaData
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
	cfg.FootbalOrgApiKey = getAllFootBallOrgAPIKeys()
	fmt.Println(cfg)
	return &cfg
}
