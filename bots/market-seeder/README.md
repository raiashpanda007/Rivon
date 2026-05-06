# Rivon Market Seeder

Seeds dummy users via SQL, warms wallet data, subscribes to market streams, and places random BUY/SELL orders through the API so the flow is api -> engine -> db.

## Setup

```bash
cd bots/market-seeder
bun install
cp .env.sample .env
```

Edit `.env` as needed.

## Run

```bash
bun run seed
```

## Notes

- The script logs in each seeded user, calls `/api/rivon/wallet/me`, and opens a WebSocket subscription per market before placing orders.
- Orders use dollar inputs for `PRICE_MIN`/`PRICE_MAX` and are converted to cents for the API.
- `PAIR_RATIO` controls how many orders are sent as matched BUY/SELL pairs to guarantee trades.
- `WS_MODE=per-market` keeps one WS connection per market. Use `per-user` only if you need UI-level fidelity.

## Environment Variables

- `BASE_URL` API base URL (default: http://localhost:8000)
- `WS_URL` WebSocket base URL (default: ws://localhost:8003)
- `DATABASE_POSTGRES_URL` Postgres connection string
- `PASSWORD` password for all dummy users
- `USER_COUNT` number of dummy users to seed
- `WALLET_BALANCE` wallet balance in cents
- `ASSET_QTY` asset quantity per market per user
- `PRICE_MIN` min price in dollars
- `PRICE_MAX` max price in dollars
- `QTY_MIN` min quantity per order
- `QTY_MAX` max quantity per order
- `ORDERS_PER_MARKET_MIN` min orders per market
- `ORDERS_PER_MARKET_MAX` max orders per market
- `PAIR_RATIO` fraction of orders sent as matched pairs
- `ORDER_JITTER_MS` base jitter between orders (ms)
- `MARKET_LIMIT` optional cap on markets processed
- `WS_MODE` per-market or per-user
- `REPORT_DIR` output directory for reports
