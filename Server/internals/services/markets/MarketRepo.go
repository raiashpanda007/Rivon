package markets

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/raiashpanda007/rivon/internals/types"
)

type MarketRepo interface {
	CreateMarket(ctx context.Context, teamId uuid.UUID, marketName, marketCode string, lastPrice, volume24H, totalVolume, openPrice24H int64) (types.MarketTable, error)
	GetAllMarkets(ctx context.Context, teamDetails bool) ([]types.MarketTable, error)
	GetMarket(ctx context.Context, marketID uuid.UUID, teamDetails bool) (types.MarketTable, error)
}

type marketRepo struct {
	db *pgxpool.Pool
}

func NewMarketRepoServices(db *pgxpool.Pool) MarketRepo {
	return &marketRepo{
		db: db,
	}
}

func (r *marketRepo) CreateMarket(ctx context.Context, teamId uuid.UUID, marketName, marketCode string, lastPrice, volume24H, totalVolume, openPrice24H int64) (types.MarketTable, error) {
	var createdMarket types.MarketTable
	query := `
INSERT INTO markets (
    id,
    team_id,
    market_name,
    market_code,
    last_price,
    volume_24h,
    total_volume,
    open_price_24h
) VALUES (
    $1,$2,$3,$4,$5,$6,$7,$8
)
ON CONFLICT (market_code) DO UPDATE SET updated_at = NOW()
RETURNING
    id,
    team_id,
    market_name,
    market_code,
    last_price,
    status,
    volume_24h,
    total_volume,
    open_price_24h,
    created_at,
    updated_at;
`

	err := r.db.QueryRow(ctx, query, uuid.New(), teamId, marketName, marketCode, lastPrice, volume24H, totalVolume, openPrice24H).Scan(&createdMarket.Id, &createdMarket.TeamID, &createdMarket.MarketName, &createdMarket.MarketCode, &createdMarket.LastPrice, &createdMarket.MarketStatus, &createdMarket.Volume24H, &createdMarket.TotalVolume, &createdMarket.OpenPrice24H, &createdMarket.CreatedAt, &createdMarket.UpdatedAt)

	if err != nil {
		slog.Error("Error in creating markets ", "error", err)
		return createdMarket, err
	}

	return createdMarket, nil
}

func (r *marketRepo) GetAllMarkets(ctx context.Context, teamDetails bool) ([]types.MarketTable, error) {
	var markets []types.MarketTable
	query := `
	SELECT
    m.id,
    m.team_id,
    m.market_name,
    m.market_code,
    m.last_price,
    m.status,
    m.volume_24h,
    m.total_volume,
    m.open_price_24h,
    m.created_at,
    m.updated_at
	`
	if teamDetails {
		query += `,
		t.id,
		t.name,
		t.short_name,
		t.code,
		t.tla,
		t.emblem,
		t.football_org_id
		FROM markets m
		JOIN teams t ON m.team_id = t.id;`
	} else {
		query += ` FROM markets m;`
	}

	rows, err := r.db.Query(ctx, query)

	if err != nil {
		slog.Error("Error in querying markets ", "error", err)
		return markets, err
	}

	for rows.Next() {
		var market types.MarketTable
		if teamDetails {
			var team types.TeamDetails
			err = rows.Scan(&market.Id, &market.TeamID, &market.MarketName, &market.MarketCode, &market.LastPrice, &market.MarketStatus, &market.Volume24H, &market.TotalVolume, &market.OpenPrice24H, &market.CreatedAt, &market.UpdatedAt, &team.ID, &team.Name, &team.ShortName, &team.Code, &team.TLA, &team.Emblem, &team.FootballOrgId)
			market.TeamDetails = &team
		} else {
			err = rows.Scan(&market.Id, &market.TeamID, &market.MarketName, &market.MarketCode, &market.LastPrice, &market.MarketStatus, &market.Volume24H, &market.TotalVolume, &market.OpenPrice24H, &market.CreatedAt, &market.UpdatedAt)
		}

		if err != nil {
			slog.Error("Error in scanning markets ", "error", err)
			return markets, err
		}

		markets = append(markets, market)

	}
	if err := rows.Err(); err != nil {
		slog.Error("Error iterating markets", "error", err)
		return markets, err
	}

	return markets, nil
}

func (r *marketRepo) GetMarket(ctx context.Context, marketID uuid.UUID, teamDetails bool) (types.MarketTable, error) {
	var market types.MarketTable
	query := `
	SELECT
    m.id,
    m.team_id,
    m.market_name,
    m.market_code,
    m.last_price,
    m.status,
    m.volume_24h,
    m.total_volume,
    m.open_price_24h,
    m.created_at,
    m.updated_at
	`

	if teamDetails {
		query += `,
		t.id,
		t.name,
		t.short_name,
		t.code,
		t.tla,
		t.emblem,
		t.football_org_id
		FROM markets m
		JOIN teams t ON m.team_id = t.id
		WHERE m.id = $1;`
	} else {
		query += ` FROM markets m WHERE m.id = $1;`
	}

	var err error
	if teamDetails {
		var team types.TeamDetails
		err = r.db.QueryRow(ctx, query, marketID).Scan(&market.Id, &market.TeamID, &market.MarketName, &market.MarketCode, &market.LastPrice, &market.MarketStatus, &market.Volume24H, &market.TotalVolume, &market.OpenPrice24H, &market.CreatedAt, &market.UpdatedAt, &team.ID, &team.Name, &team.ShortName, &team.Code, &team.TLA, &team.Emblem, &team.FootballOrgId)
		market.TeamDetails = &team
	} else {
		err = r.db.QueryRow(ctx, query, marketID).Scan(&market.Id, &market.TeamID, &market.MarketName, &market.MarketCode, &market.LastPrice, &market.MarketStatus, &market.Volume24H, &market.TotalVolume, &market.OpenPrice24H, &market.CreatedAt, &market.UpdatedAt)
	}

	if err != nil {
		slog.Error("Error in getting market details for a market id ", "error", err)
		return market, err
	}

	return market, nil
}
