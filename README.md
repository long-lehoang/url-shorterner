# URL Shortener — High Performance (Go + Redis + Postgres)

## Final Requirements & Technical Specification

This document defines the finalized **functional**, **non-functional**, and **technical requirements** for a production-ready, high-performance URL Shortener service built in **Go**, using **Redis**, **Postgres**, **Bloom Filter**, and **Prometheus metrics**.

This README will serve directly as your project's `README.md` on GitHub.

---

# 1. Overview

A high-performance URL Shortener service optimized for real-world usage:

* Sub-ms latency on Redis cache hit
* Bloom filter reducing invalid DB lookups
* Fully public service with rate limiting
* Click analytics (IP, UA, timestamp)
* Custom alias support
* URL expiration
* Batch shortening
* Sliding Window rate limiting
* Prometheus metrics

The project demonstrates senior-level backend engineering capabilities: scalability, performance, reliability, clean architecture, and production-oriented design.

---

# 2. Architecture

```
                      ┌─────────────────────┐
                      │      Client / UI     │
                      └──────────┬───────────┘
                                 │
                         HTTPS / REST
                                 │
                      ┌──────────▼───────────┐
                      │     Go API (Gin)     │
                      │  - Shorten URL       │
                      │  - Redirect          │
                      │  - Rate Limiting     │
                      │  - Metrics           │
                      └──────────┬───────────┘
                                 │
               ┌─────────────────┼──────────────────┐
               │                 │                  │
     ┌─────────▼────────┐ ┌──────▼───────┐ ┌────────▼────────┐
     │ Redis Cache       │ │ Bloom Filter │ │ Postgres (RDS)   │
     │ short:code -> URL │ │ Fast reject  │ │ Persistent store │
     │ Rate limit data   │ │ for invalid   │ │ URL + Analytics │
     └───────────────────┘ │ codes          │ └─────────────────┘
                           └───────────────┘
```

---

# 3. Functional Requirements

## 3.1 Shorten URL

**Endpoint:** `POST /shorten`

**Request:**

```json
{
  "url": "https://example.com/...",
  "expires_in": 86400,
  "alias": "longle123" // optional
}
```

**Rules:**

* Validate URL format
* If alias provided → check for conflict
* Generates short code (base62 / uuid segment)
* Stores in Postgres
* Writes to Redis cache (TTL = expires_in)
* Adds alias/code to Bloom filter

**Response:**

```json
{
  "short_code": "aZ81kd02",
  "short_url": "https://domain/aZ81kd02",
  "expires_at": "2025-11-15T12:00:00Z"
}
```

---

## 3.2 Batch Shortening

**Endpoint:** `POST /shorten/batch`

**Request:**

```json
{
  "items": [
    {"url": "https://a.com", "expires_in": 3600},
    {"url": "https://b.com", "alias": "bb"}
  ]
}
```

**Response:**

```json
{
  "results": [
    {"url": "https://a.com", "short": "..."},
    {"url": "https://b.com", "short": "..."}
  ]
}
```

---

## 3.3 Redirect

**Endpoint:** `GET /:short_code`

**Flow:**

1. Bloom filter → if "definitely not" → 404
2. Redis → cache hit → redirect
3. Postgres → fetch
4. Cache warming
5. Store analytics asynchronously

---

## 3.4 Click Analytics (Full)

**Stored in Postgres:**

* short_code
* timestamp
* IP (hashed or anonymized)
* user-agent
* referer

**Endpoint:** (optional for admin)

```
GET /analytics/:code
```

Returns aggregated metrics.

---

## 3.5 Rate Limiting

* **Sliding Window** algorithm (Redis)
* Per-IP limit (default: 100 req/min)
* Blocks abusive traffic

---

# 4. Non-Functional Requirements

## 4.1 Performance goals

* Cache hit latency: **1–2 ms**
* DB hit latency: **10–20 ms**
* 95% cache hit ratio target

## 4.2 Scalability

* Stateless API → horizontally scalable
* Redis as central cache

## 4.3 Security

* URL validation
* Prevent SSRF-like payloads
* Public service → strict rate limiting

## 4.4 Metrics

Expose Prometheus metrics:

* http_request_total
* redirect_latency_seconds
* cache_hit_ratio
* rate_limit_blocked_total

Route: `/metrics`

## 4.5 Monitoring

* Prometheus scraping
* Grafana dashboards (optional)

---

# 5. Tech Stack

* **Go 1.25+**
* **Gin** (HTTP framework)
* **Redis** (cache + rate limiting)
* **Postgres** (persistent store)
* **willf/bloom** (Bloom filter)
* **Prometheus client_golang** (metrics)
* **Docker Compose** (local dev)

---

# 6. API Summary

| Method | Endpoint           | Description          |
| ------ | ------------------ | -------------------- |
| POST   | `/shorten`         | Create shortened URL |
| POST   | `/shorten/batch`   | Create multiple URLs |
| GET    | `/:code`           | Redirect             |
| GET    | `/analytics/:code` | Get analytics        |
| GET    | `/metrics`         | Prometheus metrics   |
| GET    | `/swagger/index.html` | Swagger API documentation |

## 6.1 API Documentation

The API includes interactive Swagger/OpenAPI documentation that can be accessed in two ways:

### Option 1: Embedded Swagger UI (Default)
**Swagger UI:** `http://localhost:8080/swagger/index.html`

The Swagger UI is embedded in the API server and automatically served at `/swagger/*` when the API server is running.

### Option 2: Redoc Documentation (Docker)
**Redoc UI:** `http://localhost:8081`

When using Docker Compose, a Redoc documentation service is available using `redocly/redoc`:
- Runs on port `8081`
- Serves beautiful, interactive API documentation from the generated `docs/swagger.yaml` file
- Automatically starts with `docker compose up`
- Provides a modern, responsive documentation interface optimized for readability
- Features include: code samples, request/response examples, and interactive API exploration

### Generating Documentation

To generate or update the API documentation:

```bash
make swagger-gen
# or
make swagger
```

This will:
- Parse Swagger annotations from the code
- Generate OpenAPI/Swagger JSON and YAML files in the `docs/` directory
- Create the documentation package for serving the Swagger UI

---

# 7. Environment Variables

```
PORT=8080
DATABASE_URL=postgres://postgres:password@postgres:5432/shortener?sslmode=disable
DATABASE_READER_URL=postgres://postgres:password@postgres-read:5432/shortener?sslmode=disable
REDIS_ADDR=redis:6379
REDIS_PASSWORD=
SHORT_CODE_LENGTH=8
RATE_LIMIT_MAX=100
RATE_LIMIT_WINDOW_SECONDS=60
BLOOM_N=1000000
BLOOM_P=0.001
DOMAIN=https://short.ly
```

**Database Configuration:**
- `DATABASE_URL` - Primary database (writer) - **Required**
- `DATABASE_READER_URL` - Read replica (reader) - **Optional**
  - If not set, defaults to `DATABASE_URL` (same database for local development)
  - In production, set to a separate read replica endpoint for better scalability

---

# 8. Docker Compose

* Postgres 15
* Redis 7
* API Service (Go)
* Analytics Service (Go)
* Migration Service (Go)

```bash
# Start all services
docker compose up --build

# Or use Makefile
make docker-up
```

Services:
- `migration` - Runs database migrations on startup
- `api` - Main HTTP API server (port 8080)
- `analytics` - Background analytics processing service

---

# 9. Local Development

## Prerequisites

* Go 1.25+
* Postgres 15+ (or use Docker Compose)
* Redis 7+ (or use Docker Compose)

## Setup

```bash
# Install dependencies
go mod tidy

# Run migrations
make migrate
# or
go run ./cmd/migration

# Run API server
make run-api
# or
go run ./cmd/api

# Run analytics service
make run-analytics
# or
go run ./cmd/analytics
```

## Available Make Commands

```bash
# Build all services
make build

# Build individual services
make build-api
make build-analytics
make build-migration

# Run services locally
make run-api
make run-analytics
make run-migration

# Run migrations
make migrate

# Run tests
make test

# Run linter
make lint

# Docker commands
make docker-up
make docker-down

# Clean build artifacts
make clean
```

## Testing the API

Once the API service is running (via `make run-api` or `docker compose up`), test endpoints:

```bash
# Shorten a URL
curl -X POST http://localhost:8080/shorten \
  -H "Content-Type: application/json" \
  -d '{"url": "https://example.com"}'

# Batch shorten
curl -X POST http://localhost:8080/shorten/batch \
  -H "Content-Type: application/json" \
  -d '{"items": [{"url": "https://google.com"}]}'

# Redirect (replace <code> with actual short code)
curl -v http://localhost:8080/<code>

# Get analytics
curl http://localhost:8080/analytics/<code>

# Prometheus metrics
curl http://localhost:8080/metrics
```

---

# 10. Project Structure

```
├── cmd/
│   ├── api/
│   │   ├── Dockerfile
│   │   └── main.go              # API server entry point
│   ├── analytics/
│   │   ├── Dockerfile
│   │   └── main.go              # Analytics service entry point
│   └── migration/
│       ├── Dockerfile
│       └── main.go              # Migration tool entry point
├── internal/                     # Common packages
│   ├── cache/                   # Redis cache implementation
│   ├── config/                  # Configuration management
│   ├── prometheus/              # Prometheus metrics
│   ├── rate/                    # Rate limiting
│   └── storage/                 # Database layer (DAO/Repo)
├── svc/                         # Business logic services
│   ├── api/                     # API handlers, middleware, routing
│   ├── analytics/              # Analytics service
│   └── shortener/              # URL shortening service
├── migrations/                  # Database migrations
│   ├── 001_create_tables.up.sql
│   └── 001_create_tables.down.sql
├── docker-compose.yml           # Docker Compose configuration
├── Makefile                     # Build and development commands
├── .golangci.yml               # Linter configuration
└── README.md
```

## Architecture Overview

- **cmd/**: Application entry points (separate binaries for API, analytics, migration)
- **internal/**: Shared infrastructure code (cache, config, storage, etc.)
- **svc/**: Business domain logic (API handlers, shortener service, analytics service)
- **migrations/**: Database schema migrations

---
