package db

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
)

func ConnectToData(pgDbURL string) (*pgxpool.Pool, error) {

	ctx := context.Background()
	db, err := pgxpool.New(ctx, pgDbURL)
	if err != nil {
		slog.Error("Unable to connect")
		return nil, err

	}

	err = db.Ping(ctx)
	if err != nil {
		slog.Error("ERROR :: IN PING OUR PG DB ", slog.Any("ERROR :: ", err))
		return nil, err
	}

	return db, nil
}
