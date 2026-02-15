BEGIN;
CREATE TYPE market_state AS ENUM('open',
'closed',
'suspended');
CREATE TABLE markets(
  id UUID PRIMARY KEY,
  team_id UUID UNIQUE NOT NULL REFERENCES teams(id),
  market_name TEXT UNIQUE NOT NULL,
  market_code TEXT UNIQUE NOT NULL,
  last_price BIGINT NOT NULL DEFAULT 0,
  status market_state NOT NULL DEFAULT 'open',
  volume_24h BIGINT NOT NULL DEFAULT 0,
  total_volume BIGINT DEFAULT 0,
  open_price_24h BIGINT NOT NULL DEFAULT 0,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMIT;
