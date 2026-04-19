BEGIN;

DROP INDEX IF EXISTS idx_transactions_trade_id;
DROP INDEX IF EXISTS idx_transactions_order_id;
DROP INDEX IF EXISTS idx_assets_user_market;
DROP INDEX IF EXISTS idx_trades_order_id;
DROP INDEX IF EXISTS idx_trades_market_id;
DROP INDEX IF EXISTS idx_orders_status;
DROP INDEX IF EXISTS idx_orders_user_id;
DROP INDEX IF EXISTS idx_orders_market_id;

ALTER TABLE transactions DROP COLUMN IF EXISTS trade_id;
ALTER TABLE transactions DROP COLUMN IF EXISTS order_id;
ALTER TABLE transactions ADD COLUMN reason TEXT;

DROP TABLE IF EXISTS assets;
DROP TABLE IF EXISTS trades;
DROP TABLE IF EXISTS orders;

DROP TYPE IF EXISTS order_side;
DROP TYPE IF EXISTS order_status;

ALTER TABLE markets DROP COLUMN IF EXISTS total_supply;

-- Recreate txn_type without trading values
ALTER TABLE transactions ALTER COLUMN type TYPE TEXT;
DROP TYPE txn_type;
CREATE TYPE txn_type AS ENUM ('credit', 'debit');
ALTER TABLE transactions ALTER COLUMN type TYPE txn_type USING type::txn_type;

COMMIT;
