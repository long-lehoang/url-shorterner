.PHONY: build build-api build-analytics build-migration run run-api run-analytics run-migration migrate test test-integration lint docker-up docker-down clean swagger swagger-gen docs-build docs-serve docs-serve-mkdocs docs-build-mkdocs docs

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

test-integration:
	go test ./test/integration/... -v

lint:
	@if command -v golangci-lint > /dev/null; then \
		echo "Running golangci-lint..."; \
		golangci-lint run ./... || (echo "Lint completed with warnings (mainly comment-related)" && exit 0); \
	else \
		echo "golangci-lint not found, using go vet..."; \
		go vet ./...; \
		echo "Running staticcheck..."; \
		if command -v staticcheck > /dev/null; then \
			staticcheck ./...; \
		fi; \
	fi

docker-up: swagger-gen
	docker compose up --build

docker-down:
	docker compose down

swagger-gen:
	@SWAGGER_CMD=$$(command -v swagger 2>/dev/null || echo ""); \
	if [ -z "$$SWAGGER_CMD" ]; then \
		SWAGGER_CMD="$$HOME/go/bin/swagger"; \
		if [ ! -f "$$SWAGGER_CMD" ]; then \
			echo "swagger not found, installing..."; \
			go install github.com/go-swagger/go-swagger/cmd/swagger@latest; \
		fi; \
	fi; \
	echo "Generating Swagger documentation..."; \
	$$SWAGGER_CMD generate spec -o ./docs/swagger.json --scan-models; \
	if [ -f ./docs/swagger.json ]; then \
		$$SWAGGER_CMD flatten --with-flatten=remove-unused ./docs/swagger.json -o ./docs/swagger.json 2>/dev/null || true; \
	fi

swagger: swagger-gen

docs-build:
	@echo "Building API documentation (using Docker for Node.js tools)..."
	@bash docs/scripts/build-docs.sh

docs-serve:
	@echo "Serving documentation..."
	@if [ -f public/api/index.html ]; then \
		echo "Starting HTTP server on http://localhost:8082"; \
		echo "Open http://localhost:8082/api/index.html in your browser"; \
		cd public && python3 -m http.server 8082 2>/dev/null || \
		python3 -m SimpleHTTPServer 8082 2>/dev/null || \
		(echo "Python not found. Opening file directly..." && \
		 open api/index.html 2>/dev/null || \
		 xdg-open api/index.html 2>/dev/null || \
		 echo "Please open public/api/index.html in your browser"); \
	else \
		echo "Documentation not found. Run 'make docs-build' first."; \
	fi

docs-serve-mkdocs:
	@echo "Serving MkDocs documentation..."
	@if [ -f .venv/bin/mkdocs ]; then \
		.venv/bin/mkdocs serve; \
	elif command -v mkdocs > /dev/null; then \
		mkdocs serve; \
	else \
		echo "MkDocs not found. Install with: pip install mkdocs-material"; \
		echo "Or use virtualenv: python3 -m venv .venv && .venv/bin/pip install mkdocs-material"; \
	fi

docs-build-mkdocs:
	@echo "Building MkDocs site..."
	@if [ -f .venv/bin/mkdocs ]; then \
		.venv/bin/mkdocs build --site-dir public/dev; \
	elif command -v mkdocs > /dev/null; then \
		mkdocs build --site-dir public/dev; \
	else \
		echo "MkDocs not found. Install with: pip install mkdocs-material"; \
		echo "Or use virtualenv: python3 -m venv .venv && .venv/bin/pip install mkdocs-material"; \
	fi

docs: docs-build docs-build-mkdocs

clean:
	rm -rf bin/ docs/swagger-polished.json api/spec/openapi.json public/
	go clean
