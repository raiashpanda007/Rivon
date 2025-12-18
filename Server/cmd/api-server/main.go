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

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/raiashpanda007/rivon/internals/config"
	"github.com/raiashpanda007/rivon/internals/database"
)

func main() {
	slog.Info("Starting api-server :: ")
	cfg := config.MustLoad()

	_, err := database.Init_DB(cfg.Db.PgURL)
	if err != nil {
		panic("UNABLE TO CONNECT TO DB" + err.Error())
	}

	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Timeout(60 * time.Second))

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
