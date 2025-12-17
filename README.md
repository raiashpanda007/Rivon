# Rivon

Rivon is a modern, full-stack monorepo application built with high-performance technologies. It features a scalable frontend architecture using Turborepo and Next.js, and a flexible backend infrastructure powered by Go.

## 🚀 Project Status

### Client-Side (Frontend)
The frontend is structured as a **Turborepo** monorepo containing:
- **Apps**:
  - `bet`: A Next.js 16 application for betting features.
  - `exchange`: A Next.js 16 application for exchange features.
- **Packages**:
  - Shared UI components (`@workspace/ui`).
  - Shared configurations (`eslint-config`, `typescript-config`).
- **Tech Stack**: Next.js 16, React 19, TailwindCSS, TypeScript.

### Server-Side (Backend)
The backend is initialized as a **Go module** with a custom CLI tool for managing microservices.
- **CLI Tool**: Located in `Server/cli`, this tool helps in scaffolding, building, and starting services.
- **Services**:
  - `api-server`: The main API server built with `chi` router. It includes middleware for logging, recovery, and timeouts, and supports graceful shutdown.
- **Shared Internals**:
  - `config`: Centralized configuration management using `godotenv` to load environment variables.
- **Current State**: The infrastructure is set up with the first microservice (`api-server`) and shared configuration logic.

---

## 🛠️ Setup Instructions

Follow these steps to set up the application locally.

### Prerequisites
- **Node.js** (Latest LTS recommended)
- **pnpm** (Package manager)
- **Go** (v1.24+)

### 1. Client Setup (Frontend)

Navigate to the `Client` directory and install dependencies:

```bash
cd Client
bun install
```

To start the development server for all apps:

```bash
bun dev
```

Or to run a specific app (e.g., `bet`):

```bash
bun --filter bet dev
```

### 2. Server Setup (Backend)

Navigate to the `Server` directory:

```bash
cd Server
```

#### Environment Configuration
Copy the sample environment file and configure the necessary variables:
```bash
cp .env.sample .env
```
Ensure the following variables are set in `.env`:
- `AUTH_SECRET`
- `GOOGLE_AUTH_CLIENT_ID`
- `GOOGLE_AUTH_CLIENT_SECRET`
- `GITHUB_AUTH_CLIENT_ID`
- `GITHUB_AUTH_CLIENT_SECRET`
- `API_SERVER_URL`

#### Managing Services

The backend is managed via a custom CLI tool located in `cli/main.go`.

**Add a new service:**
This will create a new service directory in `cmd/<service-name>`.
```bash
go run cli/main.go add <service-name>
# Example: go run cli/main.go add auth
```

**Start services:**
Starts all registered services (defined in `services.json`).
```bash
go run cli/main.go start
# Or start specific services:
# go run cli/main.go start api-server
```

**Build services:**
Builds binaries for the services into the `bin/` directory.
```bash
go run cli/main.go build
```

## 📂 Project Structure

```
Rivon/
├── Client/                  # Frontend Monorepo
│   ├── apps/
│   │   ├── bet/             # Betting Application
│   │   └── exchange/        # Exchange Application
│   ├── packages/            # Shared libraries (UI, configs)
│   └── ...
├── Server/                  # Backend Go Module
│   ├── cli/                 # Custom CLI for service management
│   ├── cmd/                 # Microservices entry points
│   │   └── api-server/      # Main API Server
│   ├── internals/           # Shared internal packages (config, etc.)
│   ├── services.json        # Registry of available services
│   └── ...
└── README.md
```
