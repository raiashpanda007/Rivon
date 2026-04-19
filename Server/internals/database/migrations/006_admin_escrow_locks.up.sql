BEGIN;
ALTER TABLE wallets ADD COLUMN locked_balance BIGINT NOT NULL DEFAULT 0 CHECK(locked_balance >= 0);
ALTER TABLE assets  ADD COLUMN locked_qty     BIGINT NOT NULL DEFAULT 0 CHECK(locked_qty >= 0);
COMMIT;
