<div align="center">

# ⚡ Rivon

### Real-time Sports Trading Exchange

**Trade team positions like stocks. Prices move on demand. No house edge.**

*A continuous double-auction prediction market built for sports — powered by a custom order-matching engine, Redis Streams, and a Go backend designed for low-latency throughput.*

---

![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat-square&logo=go&logoColor=white)
![Next.js](https://img.shields.io/badge/Next.js-16-black?style=flat-square&logo=next.js)
![Redis](https://img.shields.io/badge/Redis-Streams-DC382D?style=flat-square&logo=redis&logoColor=white)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-pgx_v5-336791?style=flat-square&logo=postgresql&logoColor=white)
![Bun](https://img.shields.io/badge/Bun-1.3+-fbf0df?style=flat-square&logo=bun)
![CI](https://img.shields.io/badge/CI-GitHub_Actions-2088FF?style=flat-square&logo=github-actions&logoColor=white)

</div>

---

## Preview

<div align="center">
  <img src="docs/screenshots/Rivon.png" alt="Rivon Platform" width="100%" />
</div>

---

## What is Rivon?

Traditional sports betting platforms fix odds at booking time and rely on a house margin to stay profitable. Rivon replaces that model with a **continuous double auction** — the same mechanism used by financial exchanges. Every price is determined solely by the order book.

When more users buy a team's position, the price rises. When confidence drops, it falls. No fixed odds. No bookmaker margin. Just a market.

This creates a set of genuinely hard engineering problems:

- Low-latency order matching across hundreds of concurrent markets
- Real-time price propagation without bottlenecking on shared state
- Consistent fill settlement across a distributed service boundary
- Partial fills, order cancellation, and depth queries at exchange speed

All of these are solved here — not with off-the-shelf trading frameworks, but with a purpose-built engine in Go.

---

## Why Go (and not Rust or C++)

The honest answer is constraints. A proper exchange-grade matching engine in Rust or C++ would yield lower latency at the cost of significantly higher implementation complexity, longer iteration cycles, and a much steeper maintenance burden.

Go hits the right balance for this project: goroutines give true concurrency without OS-thread overhead, the GC pause profile is acceptable at these throughput levels, and the language is fast to write correctly. The result is an engine that's production-viable without requiring three engineers to maintain the allocator.

The same logic applies to the concurrency model inside the engine. The theoretically correct scaling approach is one container per market — fully isolated processes, no shared memory, horizontally scalable across a cluster. That architecture is absolutely achievable with the current design (the Redis Stream interface is already the boundary), but running N containers for N markets is an infrastructure cost that isn't justified at this stage. Go's goroutine scheduler gives a very similar isolation guarantee — each market owns its goroutine and its orderbook, with zero shared mutable state — at a fraction of the operational cost.

When scale demands it, the container-per-market path is open. Nothing in the current design forecloses it.

---

## Architecture

Four independent services connected through Redis Streams:

```
┌────────────────────────────────────────────────────────────────────────────┐
│                             Client  (Next.js 16)                           │
│                                                                            │
│        ┌──────────────┐   ┌─────────────────────┐   ┌──────────────────┐  │
│        │   base app   │   │  exchange terminal  │   │  bet interface   │  │
│        └──────────────┘   └─────────────────────┘   └──────────────────┘  │
└────────────────────────────────────┬───────────────────────────────────────┘
                                     │  REST API  /  WebSocket
┌────────────────────────────────────▼───────────────────────────────────────┐
│                             Server  (Go / Chi)                             │
│                                                                            │
│   Auth · Wallet · Markets · Football Metadata · Cron Jobs                 │
│                                                                            │
│   Publishes orders ──────────────────────────────────────────────────────► │
│                     ORDERS_<market_id>  (Redis Stream)                     │
└──────────────────────────┬──────────────────────────┬──────────────────────┘
                           │                          │ Trade Redis (fills)
┌──────────────────────────▼──────────────────────────▼──────────────────────┐
│                             Engine  (Go)                                   │
│                                                                            │
│   ┌──────────────────────────────────────────────────────────────────────┐ │
│   │  Redis Stream Consumer Groups                                        │ │
│   │  Batched: 20 markets per goroutine · XReadGroup block 5s · count 10 │ │
│   └─────────────────────────────┬────────────────────────────────────────┘ │
│                                 │  non-blocking channel send               │
│   ┌─────────────────────────────▼────────────────────────────────────────┐ │
│   │  Per-Market Goroutine  (1 per market, buffered channel cap=50)       │ │
│   └─────────────────────────────┬────────────────────────────────────────┘ │
│                                 │                                           │
│   ┌─────────────────────────────▼────────────────────────────────────────┐ │
│   │  Orderbook  —  Price-Time Priority Matching                          │ │
│   │  MaxHeap (bids) · MinHeap (asks) · O(log n) best-price lookup       │ │
│   │  Partial fills · Order cancellation · Depth + Snapshot queries      │ │
│   └─────────────────────────────┬────────────────────────────────────────┘ │
│                                 │  Fill events                             │
│   ┌─────────────────────────────▼────────────────────────────────────────┐ │
│   │  Trade Redis Stream  →  consumed by Server for settlement            │ │
│   └──────────────────────────────────────────────────────────────────────┘ │
└────────────────────────────────────────────────────────────────────────────┘
           │
┌──────────▼───────────────┐
│  MailServer (Bun/Express) │
│  OTP · Transactional mail │
└───────────────────────────┘
```

### Order Lifecycle

```
[1] User submits order via Client UI
[2] Server validates request, writes to ORDERS_<market_id> Redis Stream
[3] Engine consumer group picks up message (batch read, 5s block timeout)
[4] Message dispatched to per-market goroutine channel (non-blocking)
[5] Orderbook.AddOrder() — price-time priority match against resting orders
      ├── Full fill   → remove matched orders, emit fills
      ├── Partial fill → emit partial fills, queue remainder at limit price
      └── No match    → queue order at limit price, no fills
[6] Fill events published to Trade Redis Stream
[7] Server settlement worker reads fills, updates wallet + positions
[8] WebSocket broadcast → live orderbook depth and trade feed to clients
```

Every step from [3] onwards is fully asynchronous. The matching path itself is lock-free per market — one goroutine owns one orderbook, no mutexes required.

---

## Matching Engine — Deep Dive

The engine (`Engine/`) is a standalone Go binary. It has no HTTP interface. Its entire surface area is two Redis connections: one for reading order streams, one for writing fill events.

### Data Structures

```
OrderBook
├── Bids        map[int][]*Order    — price level → FIFO queue of resting buy orders
├── Asks        map[int][]*Order    — price level → FIFO queue of resting sell orders
├── BidHeap     MaxHeap             — O(log n) best bid price
├── AskHeap     MinHeap             — O(log n) best ask price
├── UserOrderMap map[string]map[string]*Order  — O(1) lookup by userId → orderId
└── CurrentPrice int                — last matched price (the "market price")
```

The heap gives O(log n) insert and O(1) peek for the best price at each side. The price-level slice gives O(1) FIFO access within a level. Total complexity for a match: **O(log n + k)** where k is the number of fills generated.

### Matching Algorithm

For a **BUY** order at price `p`:
1. Peek `bestAsk` from `AskHeap`
2. If `p < bestAsk` → no match possible, queue in `Bids[p]`
3. Otherwise, iterate the FIFO queue at `Asks[bestAsk]`, fill as many contracts as available
4. If the ask level is exhausted, pop from `AskHeap` and repeat with the next best ask
5. If the buy order still has remaining quantity after all matchable asks, queue the remainder at `p`

Partial fills on both sides are tracked via the `Filled` field on each `Order`. An order is considered fully filled when `Filled == Quantity`.

```go
// Simplified — bid side matching loop
for order.Filled < order.Quantity {
    if r.AskHeap.Size() == 0 {
        r.addOrderToBids(order, price)   // no asks — rest the order
        break
    }
    bestAsk := r.AskHeap.Peek()
    if price < bestAsk {
        r.addOrderToBids(order, price)   // price too low — rest the order
        break
    }
    // match against FIFO queue at Asks[bestAsk], generate fills
}
```

### Concurrency Without Containers

The engine uses a **goroutine-per-market** model. Each market runs its matching loop in an isolated goroutine — no locks, no channels shared between markets, no possibility of one market's load affecting another's latency.

Redis stream consumption is batched: one goroutine handles `XReadGroup` for up to 20 markets simultaneously, then fans messages out to individual market channels. This keeps the goroutine count at `O(markets / 20)` for I/O and `O(markets)` for processing — far cheaper than one OS thread per market.

```
Redis Consumer Goroutine  (1 per batch of 20 markets)
  ├── XReadGroup → reads up to 10 messages across 20 streams per call
  └── dispatches to marketMap[id] <- msg  (non-blocking, drops if channel full)

Market Goroutine  (1 per market)
  └── for msg := range channel { orderbook.AddOrder(...) }
```

The fully isolated container-per-market model remains the right long-term architecture for exchange-grade horizontal scaling. The goroutine model is a deliberate, cost-conscious choice that achieves the same isolation guarantee within a single process at near-zero infrastructure overhead.

---

## Tech Stack

| Layer | Technology | Notes |
|---|---|---|
| Frontend | Next.js 16, Turborepo, Bun | Three apps in one monorepo |
| State Management | Redux Toolkit | Shared `@workspace/store` package |
| API Server | Go, Chi router | Layered: routes → controllers → services |
| DB Driver | pgx v5 | Connection pool, prepared statements |
| Matching Engine | Go, custom heaps | Goroutine-per-market, zero shared state |
| Message Bus | Redis Streams | Consumer groups, at-least-once delivery |
| Cache | Redis × 3 | OTP / Orders / Trades — isolated by role |
| Database | PostgreSQL | 3 migrations via golang-migrate |
| Auth | JWT + OTP + OAuth | HTTP-only cookies, Google + GitHub |
| Email | Bun, Express v5, Nodemailer | Input validated with Zod |
| CI/CD | GitHub Actions | Build + migrate on every push to `dev`/`main` |
| Dev Tooling | Custom Go CLI | Scaffold, build, start, migrate — one tool |

### Three Redis Instances — Why

| Instance | Port | Role |
|---|---|---|
| OTP Redis | 6379 | TTL-based key/value, high write churn, no persistence needed |
| Order Redis | 6380 | Persistent streams, consumer group offsets, append-only workload |
| Trade Redis | 6381 | Fill delivery from engine to server, short-lived consumer |

Keeping these separate prevents a high-throughput stream append from adding latency to OTP lookups, and lets each instance carry only the persistence and memory configuration it actually needs.

---

## Services

### Client — `Client/`

Next.js 16 monorepo powered by Turborepo and Bun. Three independent apps share a common design system and state layer:

| App | Purpose |
|---|---|
| `base` | Auth, user dashboard, portfolio overview, league browser |
| `exchange` | Live trading terminal — orderbook depth, price chart, order entry |
| `bet` | Match predictions, position management, settlement history |

**Shared packages:**

- `@workspace/ui` — component library, design tokens
- `@workspace/store` — Redux Toolkit slices, shared across all apps
- `@workspace/api-caller` — typed HTTP client wrapping fetch
- `@workspace/types` — shared TypeScript types
- `@workspace/logger` — structured client-side logging

### Server — `Server/`

Go REST API. Stateless — all session and order state lives in Redis or PostgreSQL.

```
cmd/
  api-server/   ← HTTP server entry point
  jobs/         ← Cron job runner (football metadata sync)

internals/
  http/
    routes/       ← Route registration
    controllers/  ← Request/response handling
    middlewares/  ← Auth, logging, error handling
  services/
    auth/         ← JWT, OTP, OAuth
    markets/      ← Market CRUD, order publishing
    wallet/       ← Balance management, fill settlement
    footballMeta/ ← League, team, match data
  database/       ← pgx pool, migrations
  config/         ← Environment validation at startup
```

Football metadata is synced via cron jobs using three rotated API keys to stay within rate limits on football-api.org.

### Engine — `Engine/`

Standalone Go binary. No HTTP. Reads from Order Redis, writes to Trade Redis.

| File | Responsibility |
|---|---|
| `internals/Engine/engine.go` | Boot: load markets from DB, init streams, spawn goroutines |
| `internals/markets/market.go` | Per-market goroutine — owns its orderbook for its lifetime |
| `internals/Orderbooks/orderbook.go` | Matching logic, partial fills, cancellation, depth queries |
| `internals/utils/Heap.go` | MinHeap and MaxHeap — O(log n) insert/pop/peek |
| `internals/utils/TradeStream/` | Publishes fill events to Trade Redis Stream |

### MailServer — `MailServer/`

Express v5 microservice for transactional email. Handles OTP delivery and trade/account notifications. All inputs validated with Zod before processing.

---

## Getting Started

### Prerequisites

- Go 1.22+
- Bun 1.3+
- Docker + Docker Compose

### 1. Start Infrastructure

```bash
docker-compose up -d
```

This brings up:
- PostgreSQL on `:5432`
- OTP Redis on `:6379`
- Order Redis on `:6380`
- Trade Redis on `:6381`

### 2. Configure Environment

```bash
cp Server/.env.sample Server/.env
cp Engine/.env.sample Engine/.env
```

Key variables to fill in:

```env
DATABASE_POSTGRES_URL=postgres://...
OTP_REDIS_URL=redis://localhost:6379
ORDER_REDIS_URL=redis://localhost:6380
TRADE_REDIS_URL=redis://localhost:6381
AUTH_SECRET=<your-jwt-signing-key>
GOOGLE_AUTH_CLIENT_ID=...
GITHUB_AUTH_CLIENT_ID=...
FOOTBALL_API_KEY_1=...
```

### 3. Run Migrations and Build

```bash
cd cli

go run main.go migrate up    # Apply all DB migrations
go run main.go build         # Compile all services → bin/
go run main.go start         # Start all Go services
```

Or start services individually:

```bash
go run main.go start api-server
go run main.go start engine
go run main.go start jobs
```

### 4. Frontend

```bash
cd Client
bun install
bun dev      # Turborepo starts all three apps
```

| App | URL |
|---|---|
| base | http://localhost:3000 |
| exchange | http://localhost:3001 |
| bet | http://localhost:3002 |

### CLI Reference

```bash
go run main.go build [service]       # Build one or all services
go run main.go start [service]       # Start one or all services
go run main.go migrate up            # Apply migrations
go run main.go migrate down          # Roll back migrations
go run main.go add <name>            # Scaffold new Go service in Server/cmd/
```

---

## Scalability

### Current Constraints and Paths Forward

**Engine — goroutines today, containers tomorrow**

The goroutine-per-market model gives market-level isolation within a single process. When throughput demands it, the engine can be sharded across multiple container instances by market ID range — the Redis Stream interface is already the clean boundary. Consumer groups natively support competing consumers, so adding engine instances requires no code changes, only deployment configuration.

**Server — stateless and horizontally ready**

No local state. Session data is in Redis, order state is in PostgreSQL. Any number of server instances can run behind a load balancer without coordination.

**Redis Streams — tunable throughput**

The `Count` parameter in `XReadGroup` trades latency for throughput. At low volume, count=1 gives the fastest individual order processing. Under load, batching (count=10, current default) amortizes round-trip cost. Consumer group semantics guarantee at-least-once delivery and allow lag monitoring.

**Database — pooled and replica-ready**

pgx v5 connection pooling handles concurrent request load. Read-heavy queries (market depth snapshots, trade history) can be routed to read replicas with a connection string swap.

**Trade Settlement — decoupling opportunity**

Settlement currently happens synchronously after the server reads from Trade Redis. Extracting this into a dedicated settlement worker would remove it from the critical path entirely and allow independent scaling.

---

## Roadmap

- [ ] **WebSocket server** — live orderbook depth, trade tape, and price feed without polling
- [ ] **Persistent orderbook** — restore open orders from PostgreSQL on engine restart
- [ ] **Market resolution** — settle positions on match outcome (win / draw / loss)
- [ ] **In-play markets** — prices update as live match events occur (goals, red cards, penalties)
- [ ] **Portfolio analytics** — P&L tracking, realized/unrealized breakdown, position history
- [ ] **Engine observability** — Prometheus metrics: latency histograms, fill rate, queue depth
- [ ] **Market maker incentives** — rebate programs for tight-spread liquidity providers
- [ ] **Multi-sport support** — extend beyond football to basketball, cricket, tennis

---

## Contributing

Pull requests are welcome. For significant changes, open an issue first to align on design and scope. Include tests for new behavior and update relevant documentation.

---

## Vision

Prediction markets produce more accurate forecasts than fixed-odds bookmakers. A properly designed continuous double auction — one where price is set by participants, not a house — is a fundamentally better model for sports prediction.

Rivon is an attempt to build that. The domain is sports, but the engineering standards are the same ones you would hold a financial exchange to: deterministic matching, transparent pricing, no hidden margins, and latency low enough that trading feels real-time.

The project is also an exercise in building serious distributed systems infrastructure on a solo or small-team budget. Every architectural decision here — the goroutine model, the three Redis instances, the CLI-driven build system — reflects that constraint without compromising on correctness or scalability headroom.

---

<div align="center">

Built with Go, Next.js, and Redis Streams.

</div>
