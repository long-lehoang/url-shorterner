# Development Guide

## Local Setup

### Prerequisites

- Go 1.25+
- Docker & Docker Compose
- Node.js 18+ (for documentation)
- PostgreSQL 15+ (or use Docker)
- Redis 7+ (or use Docker)

### Getting Started

```bash
# Clone repository
git clone <repo-url>
cd url-shorterner

# Install dependencies
go mod download
npm install

# Start services
make docker-up

# Run migrations
make migrate

# Start API server
make run-api
```

## Code Structure

### Adding a New Endpoint

1. Define interface in `svc/api/{domain}/transport/http.go`
2. Add Swagger annotations
3. Implement handler in `svc/api/{domain}/transport/api.go`
4. Register route in `SetupRouter`
5. Generate docs: `make docs-build`

### Adding a New Domain

1. Create domain structure:
   ```
   svc/{domain}/
   ├── app/
   │   └── service.go
   ├── entity/
   │   └── {entity}.go
   └── store/
       ├── dao.go
       └── repository.go
   ```

2. Implement domain logic
3. Create API transport layer
4. Register in main.go

## Testing

```bash
# Run all tests
make test

# Run tests with coverage
go test -cover ./...

# Run specific package tests
go test ./svc/shortener/app/...
```

## Linting

```bash
# Run linter
make lint

# Fix issues automatically (if supported)
golangci-lint run --fix ./...
```

## Documentation

```bash
# Generate API documentation
make docs-build

# Serve MkDocs locally
make docs-serve

# Build all documentation
make docs
```

