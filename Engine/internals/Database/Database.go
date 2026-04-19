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

type AdminData struct {
	Balance       int
	LockedBalance int
	Assets        []AdminAsset
}

type AdminAsset struct {
	MarketID  string
	Quantity  int
	LockedQty int
}

func (r *Database) GetAdminData(adminID string) (*AdminData, error) {
	ctx := context.Background()

	var balance, lockedBalance int64
	err := r.pgdb.QueryRow(ctx,
		"SELECT balance, locked_balance FROM wallets WHERE user_id = $1", adminID,
	).Scan(&balance, &lockedBalance)
	if err != nil {
		return nil, err
	}

	rows, err := r.pgdb.Query(ctx,
		"SELECT market_id::text, quantity, locked_qty FROM assets WHERE user_id = $1", adminID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var assets []AdminAsset
	for rows.Next() {
		var marketID string
		var qty, lockedQty int64
		if err := rows.Scan(&marketID, &qty, &lockedQty); err != nil {
			return nil, err
		}
		assets = append(assets, AdminAsset{MarketID: marketID, Quantity: int(qty), LockedQty: int(lockedQty)})
	}

	return &AdminData{Balance: int(balance), LockedBalance: int(lockedBalance), Assets: assets}, nil
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
