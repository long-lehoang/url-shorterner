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

* **Go 1.21+**
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

---

# 7. Environment Variables

```
PORT=8080
DATABASE_URL=postgres://postgres:password@postgres:5432/shortener?sslmode=disable
REDIS_ADDR=redis:6379
REDIS_PASSWORD=
SHORT_CODE_LENGTH=8
RATE_LIMIT_MAX=100
RATE_LIMIT_WINDOW_SECONDS=60
BLOOM_N=1000000
BLOOM_P=0.001
DOMAIN=https://short.ly
```

---

# 8. Docker Compose

* Postgres 15
* Redis 7
* App (Go)

```bash
docker compose up --build
```

---

# 9. Local Development

```bash
go mod tidy
go run cmd/server/main.go
```

---

# 10. Project Structure

```
├── cmd/server/main.go
├── internal/
│   ├── api/
│   ├── cache/
│   ├── rate/
│   ├── shortener/
│   ├── analytics/
│   ├── prometheus/
│   └── storage/
├── migrations/
├── Dockerfile
├── docker-compose.yml
└── README.md
```

---
