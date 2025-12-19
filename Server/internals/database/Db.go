package database

import (
	"context"
	"log/slog"

	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DataBase struct {
	PgDB  *pgxpool.Pool
	Redis *redis.Client
}

func Init_DB(pgUrl string, rdUrl string) (*DataBase, error) {
	ctx := context.Background()
	db, err := pgxpool.New(ctx, pgUrl)
	if err != nil {
		slog.Error("UNABLE TO CONNECT TO THE DATABASE")
		return nil, err
	}
	err = db.Ping(ctx)
	if err != nil {
		slog.Error("ERROR :: IN PING OUR PG DB ", slog.Any("ERROR :: ", err))
		return nil, err
	}

	slog.Info("CONNECTED TO PG DB :: ")
	Redis := redis.NewClient(&redis.Options{
		Addr: rdUrl,
	})
	_, err = Redis.Ping(ctx).Result()
	if err != nil {
		slog.Error("REDIS Ping failed ", slog.Any("ERROR :: ", err))
		return nil, err
	}
	slog.Info("CONNECTED TO REDIS DB :: ")

	return &DataBase{
		PgDB:  db,
		Redis: Redis,
	}, nil

}
