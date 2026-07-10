# 🏭 Warehouse Management System

A fullstack warehouse item management system built as a technical test for Anteraja. The backend exposes a REST API for managing warehouse items, locations, and stock levels. The frontend is a React dashboard for browsing, searching, and managing items.

---

## Architecture Overview

```
your-repo/
├── backend/                # Go REST API (Gin + PostgreSQL + optional Redis)
│   ├── cmd/                # Application entry point
│   ├── internal/
│   │   ├── config/         # Environment variable loading
│   │   ├── database/       # PostgreSQL connection + migration runner
│   │   ├── cache/          # Optional Redis cache wrapper
│   │   ├── model/          # Domain structs (Item, Location, Stock)
│   │   ├── repository/     # Data access layer (pgx)
│   │   ├── service/        # Business logic & rules
│   │   ├── handler/        # HTTP handlers (Gin)
│   │   ├── middleware/      # Logger, recovery, response envelope
│   │   └── app/            # Composition root (dependency wiring)
│   ├── migrations/         # Plain SQL migration files
│   ├── docs/               # OpenAPI 3.0 spec + Swagger UI
│   ├── Dockerfile
│   └── .env.example
├── frontend/               # React 18 + TypeScript dashboard
│   └── src/
│       ├── components/     # UI components (Header, ItemTable, ItemForm, etc.)
│       ├── hooks/          # Custom React hooks (useItems, useStock, useDebounce)
│       ├── pages/          # Route-level page components
│       ├── services/       # API client functions
│       └── types/          # TypeScript interface definitions
├── docker-compose.yml      # Full stack orchestration
└── README.md
```

**Architectural pattern:** Layered (handler → service → repository) with a single composition root (`app.Container`) that wires all dependencies. Each layer only knows about the layer below it, never above.

---

## Prerequisites

| Tool | Minimum Version |
|------|----------------|
| Docker | 24.x |
| Docker Compose | v2.x |
| Go | 1.22 (dev only) |
| Node.js | 20.x (dev only) |

---

## 🚀 Quick Start (Docker)

```bash
# 1. Clone the repository
git clone https://github.com/YOUR_USERNAME/warehouse-app.git
cd warehouse-app

# 2. Copy and adjust environment variables (optional — defaults work out of the box)
cp .env.example .env

# 3. Build and start all services
docker compose up --build

# 4. Access the application
# Frontend dashboard → http://localhost:3000
# Backend API       → http://localhost:8080
# Swagger UI        → http://localhost:3000/docs/ or http://localhost:8080/docs
```

> **Note:** The backend automatically runs database migrations on startup, so no manual migration step is required.

---

## 💻 Local Development Setup

### Backend

```bash
cd backend

# Copy and configure environment
cp .env.example .env

# Start a local PostgreSQL (or point .env at an existing instance)
# docker run -d -p 5432:5432 -e POSTGRES_PASSWORD=postgres -e POSTGRES_DB=warehouse postgres:16-alpine

# Download dependencies
go mod download

# Run the server (migrations run automatically)
go run ./cmd
```

### Frontend

```bash
cd frontend

# Install dependencies
npm install

# Start the dev server (proxies /api to localhost:8080)
npm run dev

# Open http://localhost:5173
```

---

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/api/v1/items` | Create new item |
| `GET` | `/api/v1/items` | List items (filter by `category`, pagination: `page`, `limit`) |
| `GET` | `/api/v1/items/:id` | Get item detail |
| `PUT` | `/api/v1/items/:id` | Update item |
| `DELETE` | `/api/v1/items/:id` | Soft delete item |
| `POST` | `/api/v1/stock/receive` | Receive stock into a location |
| `GET` | `/api/v1/stock/:item_id` | Check stock per item (all locations) |
| `GET` | `/api/v1/locations` | List locations _(bonus)_ |
| `GET` | `/api/v1/items-categories` | List distinct categories _(frontend helper)_ |
| `GET` | `/health` | Health check |
| `GET` | `/docs` | Swagger UI |

### Response Envelope

All endpoints use a standard envelope format:

```json
// Success (list with pagination)
{
  "success": true,
  "message": "Items retrieved successfully",
  "data": [...],
  "meta": { "page": 1, "limit": 10, "total": 42 }
}

// Error
{
  "success": false,
  "message": "SKU already exists",
  "errors": [{ "field": "sku", "reason": "Duplicate entry" }]
}
```

---

## Business Rules

- **SKU uniqueness**: Duplicate SKUs return HTTP 409. Enforced at the DB level (partial unique index on non-deleted rows) AND at the service layer.
- **Soft delete**: `DELETE /items/:id` sets `deleted_at` — the row is never removed from the database.
- **Non-negative stock**: `qty` must be a positive integer on receive. Validated at the service layer (before the DB) and also has a `CHECK` constraint in the schema as defense-in-depth.
- **Stock mutation logging**: Every `stock/receive` call is logged to stdout with `[STOCK-LOG]` prefix.
- **Input validation**: All request bodies are validated in the handler layer using gin's `binding` tags before reaching the service.

---

## Technical Decisions & Trade-offs

### 1. `pgx/v5` (pool) instead of `database/sql` + ORM

I chose `pgx` directly for PostgreSQL access rather than an ORM like GORM because:
- **Explicit SQL** makes the query plan visible and reviewable — no surprise N+1 queries hidden by magic `Preload` calls.
- The `pgxpool` provides connection pooling out of the box with fine-grained control over pool size and timeouts.
- **Trade-off**: More boilerplate scan code compared to an ORM. For a project of this scope the verbosity is acceptable; for a much larger schema, a query builder like `sqlc` would be a better middle ground.

### 2. Single-pass migration runner (no migration tool dependency)

Instead of adding `goose` or `migrate`, I wrote a lightweight runner that reads `.sql` files in lexical order and applies each in its own transaction. This keeps the binary self-contained (`docker compose up --build` is the only command needed), removes one external dependency, and is trivial to reason about.

**Trade-off**: It does not track applied migrations in a migrations table, so re-running the server re-runs idempotent SQL (`CREATE TABLE IF NOT EXISTS`, `CREATE INDEX IF NOT EXISTS`). For a production system with complex, destructive migrations, a proper tool like `goose` would be the right call.

### 3. Client-side search + server-side pagination

The backend list endpoint supports server-side category filtering and offset pagination. The frontend search input (debounced 300 ms) runs an in-memory filter over the current page rather than adding a `search` query parameter to the API.

**Rationale**: Implementing full-text search in PostgreSQL (`tsvector`, `ts_rank`, or `ILIKE`) would be straightforward but adds schema complexity. The current approach is fast enough for typical warehouse item counts per page and keeps the backend minimal.

**Trade-off**: Searching across all pages (not just the current one) requires a separate server-side search query. The code is written so that switching to server-side search later only requires adding a `search` param to `buildQuery()` in `services/items.ts` and a corresponding SQL `WHERE` clause.

### 4. Optional Redis caching (valor agregado)

The `cache.Cache` type is a thin wrapper around `go-redis`. It is _optional_ — when `REDIS_ENABLED=false` (the default for local development without Redis), all cache methods are no-ops, so the application functions identically without Redis. This design means:
- Zero complexity added to the service layer — the service just calls `cache.Get / Set / Delete` without caring whether Redis is actually present.
- In Docker Compose, Redis is enabled by default.
- **Trade-off**: The TTL-based invalidation strategy (versus event-driven) may serve stale data for up to 5 minutes. For write-heavy workloads a pub/sub or shorter TTL would be more appropriate.

---

## 🤖 AI Tools Declaration

As required, here is a transparent declaration of AI assistance:

| AI Tool | What it helped with |
|---------|-------------------|
| **Claude (Anthropic)** | Used as a code generation assistant for boilerplate scaffolding (middleware envelope, Dockerfile templates, Tailwind class composition), SQL schema review, and TypeScript interface generation. |

**Parts written / designed without AI assistance:**
- Architectural decisions (layered architecture with single composition root, why pgx over ORM, migration runner design)
- Business rule implementation and their placement across layers
- Error hierarchy and mapping (service sentinel errors → HTTP status codes)
- Query design (especially the partial unique index for soft-delete SKU uniqueness)
- The debounce + client-side search strategy

**All generated code was reviewed, understood, and in many cases modified** before being committed. The architecture and technical trade-offs described above reflect genuine design thinking, not AI suggestions.

---

## Environment Variables Reference

### Backend (`backend/.env.example`)

| Variable | Default | Description |
|----------|---------|-------------|
| `SERVER_PORT` | `8080` | HTTP listen port |
| `APP_ENV` | `development` | `development` or `production` |
| `DB_HOST` | `localhost` | PostgreSQL host |
| `DB_PORT` | `5432` | PostgreSQL port |
| `DB_USER` | `postgres` | Database user |
| `DB_PASSWORD` | `postgres` | Database password |
| `DB_NAME` | `warehouse` | Database name |
| `DB_SSLMODE` | `disable` | PostgreSQL SSL mode |
| `REDIS_ENABLED` | `false` | Enable Redis caching |
| `REDIS_HOST` | `localhost` | Redis host |
| `REDIS_PORT` | `6379` | Redis port |
| `REDIS_PASSWORD` | _(empty)_ | Redis password |
| `MIGRATIONS_DIR` | `./migrations` | Path to SQL migration files |

---

## Project Status

- [x] Backend REST API (all required endpoints)
- [x] Layered architecture (handler → service → repository)
- [x] Request logging, error handler, response envelope middleware
- [x] PostgreSQL + migrations
- [x] Soft delete with `deleted_at` flag
- [x] Stock mutation logging (`[STOCK-LOG]`)
- [x] Docker + Docker Compose (`docker compose up --build`)
- [x] OpenAPI 3.0 / Swagger UI (`/docs`)
- [x] `.env.example`
- [x] Frontend React 18 + TypeScript
- [x] TanStack Query for data fetching
- [x] React Router v6 (2+ routes)
- [x] Debounced search (300 ms)
- [x] Category filter (server-side)
- [x] Pagination
- [x] Loading skeletons + spinner
- [x] Toast notifications (react-hot-toast)
- [x] Responsive layout (desktop + tablet)
- [x] Custom hooks (`useItems`, `useStock`, `useDebounce`)
- [x] TypeScript strict mode
- [x] Redis caching _(bonus)_
- [x] Locations endpoint _(bonus)_
- [x] Error Boundary _(bonus)_
