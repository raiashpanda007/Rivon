# Rivon

Rivon is a modern sports trading and betting exchange, where you can but football teams as stocks and trade them in real-time and also you can use those stocks to bet on live matches.The main idea behind this project is to build robust and super system with as much as automation as possible and efficient scalling methods .It is built with high-performance technologies. It features a scalable frontend architecture using Turborepo and Next.js, and a flexible backend infrastructure powered by Go.

## PLEASE NOTE !!! 
Pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change. But please make sure try to completely fix the issue then raise a pr and also not forget to update the docs and readme files. If you are raising a pr for both client side and server side then please try to raise it together not forcing anything but this is for fullstack students so they will get better exposure . And please keep pr standards high and make sure to raise pr for small changes also.

## 🚀 Current Status & Recent Highlights

We have made significant progress in building the core **Exchange Application**, focusing on data visualization and user experience.

### **Latest Features:**
- **Dynamic League Hub**: A visually stunning dashboard in the Exchange app (`apps/exchange`) allows users to explore football leagues with a premium "glassmorphism" UI.
- **Smart Standings Engine**: 
  - **Champions League Logic**: The standings table now intelligently adapts handling for the UCL format (Top 8 Qualify, 9-24 Play-off, Bottom 8 Eliminated).
  - **Standard Leagues**: Automatically switches to standard promotion/relegation rules for leagues like Premier League or La Liga.
  - **Visual Indicators**: Color-coded zones (Green/Yellow/Red) and animations make understanding team positions instant.
- **Season Navigation**: Users can seamlessly switch between current and historical seasons to analyze team performance over time.
- **Premium UI/UX Polish**: 
  - Extensive use of **Framer Motion** for smooth page transitions and element entry.
  - **Scroll-Aware Header**: A smart header component that adapts its visibility and transparency based on user scroll behavior.
  - **Interactive Elements**: Hover effects, gradient text, and custom illustrations (e.g., specific messages for La Liga being a "Real Madrid bias" league 😉).

## ✨ Key Features & Achievements

### 🎨 Frontend (Client)
- **Modern UI/UX**:
  - **Landing Page**: A fully responsive, high-performance landing page with smooth animations and a "glassmorphism" aesthetic.
  - **Authentication**: Beautifully designed `Login` and `Register` cards with entry animations (Framer Motion).
  - **Interactive Elements**: Custom "Nice Toast" notifications, animated loading screens with blurred backgrounds, and dynamic buttons.
  - **Design System**: A centralized UI package (`@workspace/ui`) ensuring consistency across apps using Tailwind CSS and custom fonts (Outfit).
- **State Management**: Robust Redux implementation for managing user sessions and global state.
- **API Integration**: A dedicated `api-caller` package to streamline communication between client and server.

### ⚙️ Backend (Server)
- **Microservices Architecture**: Built with Go, featuring a custom CLI for easy service management (scaffolding, building, running).
- **Authentication System**:
  - **Secure Auth**: JWT-based authentication with HTTP-only cookies.
  - **OTP Verification**: Redis-backed One-Time Password system for secure email verification.
  - **OAuth**: Infrastructure for Google and GitHub login integration.
- **Database**: High-performance PostgreSQL connection pooling using `pgx`.
- **Mail Service**: A standalone Bun service for handling transactional emails with custom HTML templates.

### 🏗️ Infrastructure
- **Monorepo**: Efficiently managed using Turborepo for the frontend and Go modules for the backend.
- **Containerization**: Docker Compose setup for PostgreSQL and Redis dependencies.

## 🏗️ Architecture & Tech Stack

### Client-Side (Frontend)
The frontend is structured as a **Turborepo** monorepo containing:
- **Apps**:
  - `base`: The main landing and authentication application (Port 3000).
  - `bet`: A Next.js 16 application for betting features (Port 3001).
  - `exchange`: A Next.js 16 application for exchange features (Port 3002).
- **Packages**:
  - Shared UI components (`@workspace/ui`).
  - API Client (`@workspace/api-caller`).
  - State Management (`@workspace/store`).
  - Logging (`@workspace/logger`).
  - Shared Types (`@workspace/types`).
  - Shared configurations (`eslint-config`, `typescript-config`).
- **Tech Stack**: Next.js 16, React 19, TailwindCSS, TypeScript, Redux Toolkit.

### Server-Side (Backend)
The backend is initialized as a **Go module** with a custom CLI tool for managing microservices.
- **CLI Tool**: Located in `Server/cli`, this tool helps in scaffolding, building, and starting services.
- **Services**:
  - `api-server`: The main API server built with `chi` router. It includes middleware for logging, recovery, and timeouts, supports graceful shutdown, and initializes the database connection at startup.
- **Shared Internals** (`Server/internals`):
  - **`config`**:
    - **Purpose**: Centralized configuration management for the backend.
    - **Functionality**: Loads environment variables using `godotenv` and maps them to strongly-typed structs (`AuthConfig`, `HttpServer`, `DataBase`).
    - **Validation**: Enforces the presence of critical environment variables (e.g., `AUTH_SECRET`, `API_SERVER_URL`, `DATABASE_POSTGRES_URL`) at startup, preventing runtime errors due to missing configuration.
  - **`database`**:
    - **Purpose**: Manages database connections and migrations.
    - **Functionality**: Uses `pgx/v5/pgxpool` for efficient PostgreSQL connection pooling.
    - **Features**: Includes an `Init_DB` function to establish connections and verify them with a ping.
  - **`http`**:
    - **Purpose**: Handles HTTP requests, routing, and middleware.
    - **Components**:
      - `controllers`: Request handlers and input validation.
      - `routes`: URL routing and handler mapping (using `chi`).
      - `middlewares`: Request processing (e.g., logging, auth verification).
  - **`services`**:
    - **Purpose**: Encapsulates business logic and domain services.
    - **Components**:
      - `auth`: Handles authentication logic, OTP generation/verification, and token management.
  - **`types`**:
    - **Purpose**: Shared type definitions and data structures used across the application.
- **Current State**: The infrastructure is set up with the first microservice (`api-server`) and shared configuration logic.

### Mail Server
A standalone service for handling email communications.
- **Tech Stack**: Bun, Express, Nodemailer.
- **Functionality**: Sends emails (e.g., OTPs) using HTML templates.
- **Location**: `MailServer/` directory.

---

## 🛠️ Setup Instructions

Follow these steps to set up the application locally.

### Prerequisites
- **Node.js** (Latest LTS recommended)
- **bun** (Package manager)
- **Go** (v1.24+)
- **migrate** (CLI tool for database migrations)

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

Or to run a specific app (e.g., `base`, `bet`):

```bash
bun --filter base dev
# or
bun --filter bet dev
```



### 2. Infrastructure Setup (Database)

The project uses **Docker Compose** to spin up necessary infrastructure services (e.g., PostgreSQL).

Ensure you have Docker installed and running, then execute:

```bash
# From the project root
docker-compose up -d
```

This will start:
- A **PostgreSQL** instance on port `5432`.
- A **Redis** instance on port `6379` (used for OTP management).

The credentials and configurations match those in `.env.sample`.

### 3. Server Setup (Backend)

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
- `DATABASE_POSTGRES_URL`
- `OTP_REDIS_URL`
- `COOKIE_SECURE`
- `MAIL_SERVER_URL`

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

**Database Migrations:**
Runs database migrations using the `migrate` tool. Requires the `migrate` CLI to be installed.
```bash
go run cli/main.go migrate up
# Or to rollback:
go run cli/main.go migrate down
```

### 4. Mail Server Setup

Navigate to the `MailServer` directory and install dependencies:

```bash
cd MailServer
bun install
```

To start the mail server:

```bash
bun dev
```
The server runs on port `8001` (default) and listens for email sending requests.

## 📂 Project Structure

```
Rivon/
├── Client/                  # Frontend Monorepo
│   ├── apps/
│   │   ├── base/            # Landing & Auth Application
│   │   ├── bet/             # Betting Application
│   │   ├── exchange/        # Exchange Application
│   ├── packages/            # Shared libraries (UI, api-caller, store, etc.)
│   └── ...
├── MailServer/              # Email Service (Bun/Express)
├── Server/                  # Backend Go Module
│   ├── cli/                 # Custom CLI for service management
│   ├── cmd/                 # Microservices entry points
│   │   └── api-server/      # Main API Server
│   ├── internals/           # Shared internal packages (config, database, http, services, types)
│   ├── services.json        # Registry of available services
│   └── ...
└── README.md
```
