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
- **Current State**: The infrastructure is ready, but no specific microservices have been added to `cmd/` yet.

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
pnpm install
```

To start the development server for all apps:

```bash
pnpm dev
```

Or to run a specific app (e.g., `bet`):

```bash
pnpm --filter bet dev
```

### 2. Server Setup (Backend)

Navigate to the `Server` directory:

```bash
cd Server
```

The backend is managed via a custom CLI tool located in `cli/main.go`.

#### Managing Services

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
# go run cli/main.go start auth payment
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
│   ├── cmd/                 # Microservices entry points (currently empty)
│   ├── services.json        # Registry of available services
│   └── ...
└── README.md
```
