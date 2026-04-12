package database

import (
	"context"
	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v5/pgxpool"
	"log/slog"
)

type DataBase struct {
	PgDB                  *pgxpool.Pool
	OtpRedis              *redis.Client
	UserMapRedis          *redis.Client
	OrderRedis            *redis.Client
	ApiEnginerPubSubRedis *redis.Client
}

func Init_DB(pgUrl string, otpRedisUrl string, orderRedisUrl string, apiEnginePubSubRedisURL string) (*DataBase, error) {
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
	OtpRedis := redis.NewClient(&redis.Options{
		Addr: otpRedisUrl,
	})
	_, err = OtpRedis.Ping(ctx).Result()
	if err != nil {
		slog.Error("REDIS Ping failed ", slog.Any("ERROR :: ", err))
		return nil, err
	}
	slog.Info("CONNECTED TO OTP REDIS DB :: ")

	OrderRedis := redis.NewClient(&redis.Options{
		Addr: orderRedisUrl,
	})
	_, err = OrderRedis.Ping(ctx).Result()
	if err != nil {
		slog.Error("REDIS Ping failed ", slog.Any("ERROR :: ", err))
		return nil, err
	}
	slog.Info("CONNECTED TO ORDER REDIS DB :: ")

	ApiEnginePubSubRedis := redis.NewClient(&redis.Options{
		Addr: apiEnginePubSubRedisURL,
	})
	_, err = ApiEnginePubSubRedis.Ping(ctx).Result()
	if err != nil {
		slog.Error("REDIS Ping failed ", slog.Any("ERROR :: ", err))
		return nil, err
	}

	UserMapRedis := redis.NewClient(&redis.Options{
		Addr: otpRedisUrl,
	})
	_, err = UserMapRedis.Ping(ctx).Result()
	if err != nil {
		slog.Error("REDIS PING failed ", slog.Any("ERROR :: ", err))
		return nil, err
	}

	return &DataBase{
		PgDB:                  db,
		OtpRedis:              OtpRedis,
		OrderRedis:            OrderRedis,
		ApiEnginerPubSubRedis: ApiEnginePubSubRedis,
		UserMapRedis:          UserMapRedis,
	}, nil

}
