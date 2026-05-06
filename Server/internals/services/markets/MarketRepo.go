package markets

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/raiashpanda007/rivon/internals/types"
)

type UserOrder struct {
	Id          uuid.UUID `json:"orderId"`
	Side        string    `json:"side"`
	Price       int64     `json:"price"`
	Quantity    int64     `json:"quantity"`
	ExecutedQty int64     `json:"filled"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"createdAt"`
}

type MarketRepo interface {
	CreateMarket(ctx context.Context, teamId uuid.UUID, marketName, marketCode string, lastPrice, volume24H, totalVolume, openPrice24H int64) (types.MarketTable, error)
	GetAllMarkets(ctx context.Context, teamDetails bool) ([]types.MarketTable, error)
	GetMarket(ctx context.Context, marketID uuid.UUID, teamDetails bool) (types.MarketTable, error)
	CreateOrder(ctx context.Context, orderId, userId, marketId uuid.UUID, side string, price, quantity int64) error
	UpdateOrderStatus(ctx context.Context, orderId uuid.UUID, status string, executedQty int64) error
	GetUserOpenOrders(ctx context.Context, userId, marketId uuid.UUID) ([]UserOrder, error)
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

func (r *marketRepo) CreateOrder(ctx context.Context, orderId, userId, marketId uuid.UUID, side string, price, quantity int64) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO orders (id, market_id, user_id, side, price, quantity, executed_qty, status)
		VALUES ($1, $2, $3, $4::order_side, $5, $6, 0, 'pending')
		ON CONFLICT (id) DO NOTHING`,
		orderId, marketId, userId, side, price, quantity,
	)
	return err
}

func (r *marketRepo) UpdateOrderStatus(ctx context.Context, orderId uuid.UUID, status string, executedQty int64) error {
	_, err := r.db.Exec(ctx, `
		UPDATE orders
		   SET status = $2::order_status,
		       executed_qty = $3,
		       updated_at = NOW()
		 WHERE id = $1
		   AND status NOT IN ('filled', 'cancelled')`,
		orderId, status, executedQty,
	)
	return err
}

// liveStatsJoins computes last_price, volume_24h (last 24 h), and today's open_price
// directly from trade_ticks so the markets table never needs manual updates.
const liveStatsJoins = `
LEFT JOIN LATERAL (
    SELECT
        last(price, time)  AS last_price,
        sum(quantity)      AS volume_24h
    FROM trade_ticks
    WHERE market_id = m.id
      AND time >= NOW() - INTERVAL '24 hours'
) day24h ON true
LEFT JOIN LATERAL (
    SELECT first(price, time) AS open_price
    FROM trade_ticks
    WHERE market_id = m.id
      AND time >= date_trunc('day', NOW())
) today ON true`

const liveStatsSelect = `
SELECT
    m.id,
    m.team_id,
    m.market_name,
    m.market_code,
    COALESCE(day24h.last_price, m.last_price) AS last_price,
    m.status,
    COALESCE(day24h.volume_24h, 0)            AS volume_24h,
    m.total_volume,
    COALESCE(today.open_price, m.open_price_24h) AS open_price_24h,
    m.created_at,
    m.updated_at`

func (r *marketRepo) GetAllMarkets(ctx context.Context, teamDetails bool) ([]types.MarketTable, error) {
	var markets []types.MarketTable
	var query string
	if teamDetails {
		query = liveStatsSelect + `,
		t.id, t.name, t.short_name, t.code, t.tla, t.emblem, t.football_org_id
		FROM markets m` + liveStatsJoins + `
		JOIN teams t ON m.team_id = t.id`
	} else {
		query = liveStatsSelect + `
		FROM markets m` + liveStatsJoins
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
	var query string
	if teamDetails {
		query = liveStatsSelect + `,
		t.id, t.name, t.short_name, t.code, t.tla, t.emblem, t.football_org_id
		FROM markets m` + liveStatsJoins + `
		JOIN teams t ON m.team_id = t.id
		WHERE m.id = $1`
	} else {
		query = liveStatsSelect + `
		FROM markets m` + liveStatsJoins + `
		WHERE m.id = $1`
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

func (r *marketRepo) GetUserOpenOrders(ctx context.Context, userId, marketId uuid.UUID) ([]UserOrder, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, side, price, quantity, executed_qty, status, created_at
		FROM orders
		WHERE user_id = $1
		  AND market_id = $2
		  AND status IN ('pending', 'partial')
		ORDER BY created_at DESC`,
		userId, marketId,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []UserOrder
	for rows.Next() {
		var o UserOrder
		if err := rows.Scan(&o.Id, &o.Side, &o.Price, &o.Quantity, &o.ExecutedQty, &o.Status, &o.CreatedAt); err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}
	return orders, rows.Err()
}
