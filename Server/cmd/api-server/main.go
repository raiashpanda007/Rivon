package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/raiashpanda007/rivon/internals/config"
	"github.com/raiashpanda007/rivon/internals/database"
	"github.com/raiashpanda007/rivon/internals/http/routes"
	"github.com/raiashpanda007/rivon/internals/services/auth"
)

func main() {
	slog.Info("Starting api-server :: ")
	cfg := config.MustLoad()

	auth.NewOAuth(cfg.Auth.GoAuthSecret, cfg.Server.CookieSecure, cfg.Auth.GoogleClientID, cfg.Auth.GoogleClientSecret, cfg.Auth.GithubClientID, cfg.Auth.GithubClientSecret, "http://"+cfg.Server.ApiServerAddr+"/api/rivon")

	Db, err := database.Init_DB(cfg.Db.PgURL, cfg.Db.OTPRedisURL)
	if err != nil {
		panic("UNABLE TO CONNECT TO DB" + err.Error())
	}

	router := routes.InitRouters(cfg, Db.PgDB, Db.Redis)

	server := http.Server{
		Addr:    cfg.Server.ApiServerAddr,
		Handler: router,
	}

	slog.Info("API_SERVER IS RUNNING ON ", "URL :: ", cfg.Server.ApiServerAddr)

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {

		err := server.ListenAndServe()

		if err != nil {
			log.Fatalf("ERROR :: UNABLE TO START SERVER :: %s", err.Error())
		}
	}()

	<-done

	slog.Info("Shutting down the server ")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()
	err = server.Shutdown(ctx)

	if err != nil {
		slog.Error("ERROR :: UNABLE TO CLOSE THE SERVER GRACEFULLY  :: ", slog.String("error", err.Error()))
	}

	slog.Info("Server shut down gracefull ")

}
