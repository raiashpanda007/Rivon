package database

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Database struct {
	pgdb *pgxpool.Pool
}

type Market struct {
	Id   string
	Name string
}

func InitDb(pgUrl string) (*Database, error) {
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

	return &Database{
		pgdb: db,
	}, nil
}

func (r *Database) GetAllMarkets() ([]Market, error) {
	ctx := context.Background()
	rows, err := r.pgdb.Query(ctx, "SELECT id, market_name FROM markets")
	if err != nil {
		slog.Error("ERROR :: IN GETTING ALL MARKETS ", slog.Any("ERROR :: ", err))
		return nil, err
	}
	defer rows.Close()

	var markets []Market
	for rows.Next() {
		var market Market
		err := rows.Scan(&market.Id, &market.Name)
		if err != nil {
			slog.Error("ERROR :: IN SCANNING MARKET ROW ", slog.Any("ERROR :: ", err))
			return nil, err
		}
		markets = append(markets, market)
	}

	return markets, nil

}
