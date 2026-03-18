# Rivon

Rivon is a high-performance sports trading and betting exchange that models football teams as tradeable assets and enables real-time trading and betting on live matches. It combines a modern React/Next.js frontend (Turborepo) with a scalable Go backend (microservices) and lightweight auxiliary services (mail, Redis).

## PLEASE NOTE
Pull requests are welcome. For major changes, open an issue first to discuss design and scope. For code changes, include tests where applicable and update docs.

## Features

- Real-time exchange UI and betting flows across multiple Next.js apps (`base`, `bet`, `exchange`).
- Shared monorepo packages: `@workspace/ui`, `@workspace/api-caller`, `@workspace/store`, `@workspace/types`.
- Smart Standings Engine supporting multiple league rules and season navigation.
- JWT-based authentication, Redis-backed OTP, and OAuth scaffolding.
- High-performance Engine components for markets and orderbooks with Redis persistence.

## Architecture & Tech Stack

- Frontend: Turborepo with Next.js, React, Tailwind CSS, Framer Motion.
- Backend: Go microservices (`chi` router, `pgx` Postgres pooling).
- Engine: Go-based order matching and market logic under `Engine/`.
- MailServer: Bun + Express + Nodemailer for transactional email.
- Infrastructure: Docker Compose for Postgres and Redis; `migrate` for migrations.

## Quickstart (development)

1. Start infrastructure (Postgres + Redis):

```bash
docker-compose up -d
```

2. Frontend (Client):

```bash
cd Client
bun install
bun dev
```

3. MailServer (optional):

```bash
cd MailServer
bun install
bun dev
```

4. Server services:

```bash
cd Server
cp .env.sample .env
go run cli/main.go start api-server
```

For migrations:

```bash
go run cli/main.go migrate up
```

## CLI & Dev Commands

- Add a service: `go run cli/main.go add <service-name>`
- Start services: `go run cli/main.go start [<service-name>]`
- Build services: `go run cli/main.go build`
- Run migrations: `go run cli/main.go migrate up`

## Project layout

```
Rivon/
├─ Client/
├─ Server/
├─ Engine/
├─ MailServer/
├─ docker-compose.yml
└─ README.md
```

## Contributing

- Open issues for design or spec discussions for major changes.
- Follow existing code style; include tests and update docs for changes.

## Get precise change timestamps

Use git to see when files changed:

```bash
git log --pretty=format:"%h %ad %an %s" --date=short -- README.md
```

---


## What's New

**Last updated:** 2026-03-15

- **Frontend — Exchange & UI:** Dynamic League Hub with glassmorphism visuals; season navigation; smart standings engine supporting Champions League logic (Top 8 qualify; 9–24 play-offs; bottom elimination); color-coded zone indicators and Framer Motion animations for polished UX.
- **Frontend — Apps & Packages:** Shared UI package (`@workspace/ui`), API client (`@workspace/api-caller`), centralized state via Redux Toolkit across `base`, `bet`, and `exchange`; Tailwind CSS and custom fonts applied consistently.
- **Backend — API Server & Services:** `api-server` using `chi` router; JWT auth with HTTP-only cookies; Redis-backed OTP verification; OAuth scaffolding (Google/GitHub); PostgreSQL pooling via `pgx` and CLI-integrated migration tooling.
- **Engine:** Performance-focused market and orderbook components, optimized heap utilities, and Redis-backed order persistence.
- **Mail Server:** Standalone Bun service with Nodemailer and HTML templates for OTPs and transactional emails.
- **CLI & Tooling:** Custom CLI for scaffolding, building, starting, and migrating services; Docker Compose for Postgres and Redis; Turborepo monorepo structure for frontend apps.
- **Infrastructure & Reliability:** Environment validation at startup, graceful shutdown support, structured logging and middleware, and clear separation of internals (`config`, `database`, `http`, `services`, `types`).

Notes:

- For exact timestamps per change, inspect the git history (e.g., `git log -- README.md`).
- If you prefer, I can extract commit-based notes and generate a `CHANGELOG.md`.

```
