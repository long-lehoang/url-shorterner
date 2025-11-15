# URL Shortener Development Documentation

Welcome to the URL Shortener development documentation.

## Overview

This service provides a high-performance URL shortening solution with:

- Sub-millisecond latency on cache hits
- Bloom filter for fast invalid lookups
- Click analytics with detailed tracking
- Rate limiting using sliding window algorithm
- Prometheus metrics for monitoring

## Quick Start

```bash
# Start all services
make docker-up

# Run API server locally
make run-api

# Run tests
make test

# Build documentation
make docs-build
```

## Project Structure

```
├── cmd/              # Application entry points
│   ├── api/         # HTTP API server
│   ├── analytics/   # Analytics service
│   └── migration/    # Database migrations
├── internal/         # Shared packages
│   ├── cache/        # Redis cache
│   ├── config/       # Configuration
│   ├── storage/      # Database layer
│   └── ...
├── svc/              # Business logic
│   ├── api/          # API handlers
│   ├── shortener/    # URL shortening domain
│   └── analytics/    # Analytics domain
└── docs/             # Documentation
```

## Development Workflow

1. Make changes to code
2. Run tests: `make test`
3. Run linter: `make lint`
4. Generate docs: `make docs-build`
5. Test locally: `make run-api`

