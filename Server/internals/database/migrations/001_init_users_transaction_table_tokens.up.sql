BEGIN;

CREATE TYPE user_type AS ENUM ('admin', 'user');
CREATE TYPE auth_provider AS ENUM ('github', 'google', 'credentials');
CREATE TYPE txn_type AS ENUM ('credit', 'debit');

CREATE TABLE users (
  id UUID PRIMARY KEY,
  type user_type NOT NULL DEFAULT 'user',
  name VARCHAR(50) NOT NULL,
  email TEXT NOT NULL,
  provider auth_provider NOT NULL,
  password_hash TEXT,
  display_photo TEXT,
  verified BOOLEAN NOT NULL DEFAULT FALSE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT unique_email_provider UNIQUE (email, provider),
  CONSTRAINT provider_password_check CHECK (
    (provider = 'credentials' AND password_hash IS NOT NULL)
    OR
    (provider IN ('google', 'github') AND password_hash IS NULL)
  )
);

CREATE TABLE wallets (
  id UUID PRIMARY KEY,
  user_id UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
  balance BIGINT NOT NULL DEFAULT 0,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT wallet_balance_non_negative CHECK (balance >= 0)
);

CREATE TABLE tokens (
  id UUID PRIMARY KEY,
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  token_hash TEXT NOT NULL,
  provider_hash_token TEXT,
  revoked BOOLEAN NOT NULL DEFAULT FALSE,
  expires_at TIMESTAMPTZ NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT unique_user_token UNIQUE (user_id, token_hash)
);

CREATE TABLE transactions (
  id UUID PRIMARY KEY,
  wallet_id UUID NOT NULL REFERENCES wallets(id) ON DELETE CASCADE,
  type txn_type NOT NULL,
  amount BIGINT NOT NULL CHECK (amount > 0),
  balance_before BIGINT NOT NULL,
  balance_after BIGINT NOT NULL,
  reason TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT valid_balance_transition CHECK (
    (type = 'credit' AND balance_after = balance_before + amount)
    OR
    (type = 'debit' AND balance_after = balance_before - amount)
  ),
  CONSTRAINT non_negative_balance_after CHECK (balance_after >= 0)
);

CREATE INDEX idx_users_email_provider ON users(email, provider);
CREATE INDEX idx_wallets_user_id ON wallets(user_id);
CREATE INDEX idx_tokens_user_id ON tokens(user_id);
CREATE INDEX idx_transactions_wallet_id ON transactions(wallet_id);
CREATE INDEX idx_transactions_created_at ON transactions(created_at);

INSERT INTO users (
  id,
  type,
  name,
  email,
  provider,
  password_hash,
  verified
) VALUES (
  '00000000-0000-0000-0000-000000000001',
  'admin',
  'Ashwin Rai',
  'raiashwin005@gmail.com',
  'credentials',
  '12345678',
  TRUE
)
ON CONFLICT (email, provider) DO NOTHING;COMMIT;
