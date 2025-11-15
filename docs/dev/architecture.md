# Architecture

## System Overview

The URL Shortener service follows a clean architecture pattern with clear separation of concerns.

## Layers

### 1. Transport Layer (`svc/api/*/transport`)
- HTTP handlers and routing
- Request/response DTOs
- API documentation annotations

### 2. Application Layer (`svc/*/app`)
- Business logic
- Domain services
- Use cases

### 3. Domain Layer (`svc/*/entity`)
- Domain entities
- Value objects
- Domain events

### 4. Infrastructure Layer (`internal/*`)
- Database access (DAO/Repository)
- Caching (Redis)
- External service clients

## Data Flow

```
HTTP Request
    ↓
Transport Layer (http.go, api.go)
    ↓
Application Service (app/service.go)
    ↓
Domain Store (store/dao.go, store/repository.go)
    ↓
Infrastructure (internal/storage)
    ↓
Database/Redis
```

## CQRS Pattern

- **Read Operations**: Use DAO with reader connection pool
- **Write Operations**: Use Repository with writer connection pool

