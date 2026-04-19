-- ALTER TYPE ... ADD VALUE cannot run inside a transaction block
ALTER TYPE txn_type ADD VALUE IF NOT EXISTS 'reserve';
ALTER TYPE txn_type ADD VALUE IF NOT EXISTS 'refund';
ALTER TYPE txn_type ADD VALUE IF NOT EXISTS 'settle';
BEGIN;
ALTER TABLE markets ADD COLUMN total_supply BIGINT NOT NULL DEFAULT 1000000000;
CREATE TYPE order_status AS ENUM(
  'pending',
  'partial',
  'filled',
  'cancelled'
);
CREATE TYPE order_side AS ENUM(
  'BUY',
  'SELL'
);
CREATE TABLE orders(
  id UUID PRIMARY KEY,
  market_id UUID NOT NULL REFERENCES markets(id),
  user_id UUID NOT NULL REFERENCES users(id),
  side order_side NOT NULL,
  price BIGINT NOT NULL CHECK(price > 0),
  quantity BIGINT NOT NULL CHECK(quantity > 0),
  executed_qty BIGINT NOT NULL DEFAULT 0 CHECK(executed_qty >= 0),
  status order_status NOT NULL DEFAULT 'pending',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT executed_lte_quantity CHECK(executed_qty <= quantity)
);
CREATE TABLE trades(
  id UUID PRIMARY KEY,
  market_id UUID NOT NULL REFERENCES markets(id),
  order_id UUID NOT NULL REFERENCES orders(id),
  other_order_id UUID NOT NULL REFERENCES orders(id),
  price BIGINT NOT NULL CHECK(price > 0),
  quantity BIGINT NOT NULL CHECK(quantity > 0),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE TABLE assets(
  id UUID PRIMARY KEY,
  user_id UUID NOT NULL REFERENCES users(id),
  market_id UUID NOT NULL REFERENCES markets(id),
  quantity BIGINT NOT NULL DEFAULT 0 CHECK(quantity >= 0),
  avg_cost BIGINT NOT NULL DEFAULT 0 CHECK(avg_cost >= 0),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(user_id, market_id)
);
ALTER TABLE transactions DROP COLUMN reason;
ALTER TABLE transactions ADD COLUMN order_id UUID REFERENCES orders(id);
ALTER TABLE transactions ADD COLUMN trade_id UUID REFERENCES trades(id);
-- Seed admin's position for any markets that already exist
INSERT INTO assets(
  id,
  user_id,
  market_id,
  quantity,
  avg_cost
) SELECT
  gen_random_uuid(),
  '00000000-0000-0000-0000-000000000001',
  id,
  1000000000,
  0
FROM
  markets
ON CONFLICT (
    user_id,
    market_id
  ) DO NOTHING;
CREATE INDEX idx_orders_market_id
ON orders(market_id);
CREATE INDEX idx_orders_user_id
ON orders(user_id);
CREATE INDEX idx_orders_status
ON orders(status);
CREATE INDEX idx_trades_market_id
ON trades(market_id);
CREATE INDEX idx_trades_order_id
ON trades(order_id);
CREATE INDEX idx_assets_user_market
ON assets(
  user_id,
  market_id
);
CREATE INDEX idx_transactions_order_id
ON transactions(order_id);
CREATE INDEX idx_transactions_trade_id
ON transactions(trade_id);
COMMIT;
