.PHONY: build build-api build-analytics build-migration run run-api run-analytics run-migration migrate test lint docker-up docker-down clean

build: build-api build-analytics build-migration

build-api:
	go build -o bin/api ./cmd/api

build-analytics:
	go build -o bin/analytics ./cmd/analytics

build-migration:
	go build -o bin/migration ./cmd/migration

run-api:
	go run ./cmd/api

run-analytics:
	go run ./cmd/analytics

run-migration:
	go run ./cmd/migration

migrate: run-migration

test:
	go test ./...

lint:
	@if command -v golangci-lint > /dev/null; then \
		echo "Running golangci-lint..."; \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not found, using go vet..."; \
		go vet ./...; \
		echo "Running staticcheck..."; \
		if command -v staticcheck > /dev/null; then \
			staticcheck ./...; \
		fi; \
	fi

docker-up:
	docker compose up --build

docker-down:
	docker compose down

clean:
	rm -rf bin/
	go clean
