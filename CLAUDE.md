# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Rivon is a sports prediction/trading platform with four services:
- **Client** – Next.js 16 frontend monorepo (Turborepo + Bun)
- **Server** – Go REST API (Chi router, PostgreSQL, Redis)
- **Engine** – Go high-performance order-matching engine (Redis Streams)
- **MailServer** – Bun/Express transactional email service

All services are managed through a custom **CLI tool** (`cli/`) and registered in `services.json`.

---

## Infrastructure

```bash
docker-compose up -d   # PostgreSQL (5432), OTP Redis (6379), Order Redis (6380), Trade Redis (6381)
```

---

## CLI Tool (primary build/run interface)

```bash
cd cli

go run main.go build              # Build all services → bin/
go run main.go build api-server   # Build specific service
go run main.go start              # Start all services
go run main.go start engine       # Start specific service
go run main.go migrate up         # Apply DB migrations
go run main.go migrate down       # Rollback DB migrations
go run main.go add <name>         # Scaffold new Go service in Server/cmd/
```

---

## Frontend (Client/)

```bash
cd Client
bun install
bun dev           # Run all apps via Turborepo (base: 3000, exchange, bet)
bun run build
bun run lint
bun run format
```

Turborepo workspace packages: `@workspace/ui`, `@workspace/store` (Redux Toolkit), `@workspace/api-caller`, `@workspace/types`, `@workspace/logger`.

---

## Server (Go API)

```bash
cd Server
go mod tidy
go test ./...                         # Run all tests
go test ./internals/services/auth/... # Run specific package tests
```

Structure: `cmd/api-server/` and `cmd/jobs/` are entry points. Business logic lives in `internals/services/`. HTTP layer is in `internals/http/{routes,controllers,middlewares}`.

---

## Engine (Go Trading Engine)

```bash
cd Engine
go mod tidy
go test ./...
```

Structure:
- `internals/Engine/engine.go` – Orchestrates market consumer groups; batches 20 markets per goroutine batch
- `internals/Orderbooks/orderbook.go` – Price-time priority matching using custom MinHeap (asks) / MaxHeap (bids)
- `internals/markets/market.go` – Receives `OrderMessage` from Redis Streams and routes to orderbook
- `internals/utils/Heap.go` – Custom heap implementations for O(log n) order operations

**Order flow:**
```
Redis Stream (ORDERS_<market_id>) → Consumer Group → Market Channel (buf 50) → Orderbook → Fills
```

`Order` struct fields: `Id`, `UserId`, `Price`, `Quantity`, `Filled`, `Side` (BUY/SELL).
`Fills` struct fields: `Price`, `Quantity`, `TradeId`, `OtherUserId`, `OrderId`.

---

## MailServer

```bash
cd MailServer
bun install
bun dev
```

Express v5 + Nodemailer + Zod validation. Handles OTP delivery and transactional email.

---

## CI/CD

GitHub Actions (`.github/workflows/build.yml`) runs on PRs to `main` and pushes to `dev`/`main`:
1. Installs Bun 1.3.1 and Go (from `Server/go.mod`)
2. Installs `golang-migrate`
3. Runs `cli migrate up` against a test PostgreSQL instance
4. Runs `bun run build` (Client)
5. Runs `cli build` (all Go services)

---

## Environment

Copy `.env.sample` files in `Server/` and `Engine/` to `.env`. Key variables:
- `DATABASE_POSTGRES_URL`, `OTP_REDIS_URL`, `ORDER_REDIS_URL` – infrastructure connections
- `AUTH_SECRET` – JWT signing key
- `GOOGLE_AUTH_*`, `GITHUB_AUTH_*` – OAuth credentials
- `MAIL_SERVER_URL`, `CLIENT_BASE_URL`
- `FOOTBALL_API_KEY_1/2/3` – rotated keys for football-api.org

---

## Architecture Notes

- **Auth**: JWT in HTTP-only cookies + Redis-backed OTP + OAuth (Google/GitHub via Gorilla sessions)
- **Database**: pgx v5 connection pooling; 3 SQL migrations in `Server/internals/database/migrations/`
- **Redis**: Three separate instances—OTP, Order streaming, Trade data—each serving a distinct purpose
- **Football metadata**: Multiple API keys rotated to avoid rate limits; synced via cron jobs in `Server/cmd/jobs/`
- **Orderbook**: Partial fills are supported; cancellation via `CancelOrder()`; `GetDepth()` and `GetSnapshot()` for market data queries
