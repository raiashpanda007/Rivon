CREATE EXTENSION IF NOT EXISTS timescaledb CASCADE;

CREATE TABLE IF NOT EXISTS trade_ticks (
    time      TIMESTAMPTZ NOT NULL,
    market_id UUID        NOT NULL,
    price     BIGINT      NOT NULL,
    quantity  BIGINT      NOT NULL,
    trade_id  UUID        NOT NULL
);

SELECT create_hypertable('trade_ticks', 'time', if_not_exists => TRUE);

CREATE INDEX IF NOT EXISTS idx_trade_ticks_market_time ON trade_ticks (market_id, time DESC);
