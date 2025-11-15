#!/bin/bash

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

echo "Building API documentation..."

# Step 1: Generate Swagger 2.0 from Go annotations
echo "Step 1: Generating Swagger 2.0 spec..."
cd "$PROJECT_ROOT"
make swagger-gen

# Step 2: Polish Swagger JSON using Docker Node.js
echo "Step 2: Polishing Swagger JSON..."
docker run --rm \
  -v "$PROJECT_ROOT/docs:/docs" \
  -v "$PROJECT_ROOT/docs/scripts:/scripts" \
  -w /docs \
  node:20-alpine \
  sh -c "node /scripts/polish-swagger.js /docs/swagger.json /docs/swagger-polished.json"

# Step 3: Convert Swagger 2.0 to OpenAPI 3.0
echo "Step 3: Converting Swagger 2.0 to OpenAPI 3.0..."
mkdir -p "$PROJECT_ROOT/api/spec"

docker run --rm \
  -v "$PROJECT_ROOT/docs:/docs" \
  -v "$PROJECT_ROOT/api/spec:/output" \
  mermade/swagger2openapi:latest \
  swagger2openapi /docs/swagger-polished.json -o /output/openapi.json

# Step 4: Apply patches using Docker Node.js
echo "Step 4: Applying patches..."
docker run --rm \
  -v "$PROJECT_ROOT/api/spec:/spec" \
  -v "$PROJECT_ROOT/docs/patches:/patches" \
  -v "$PROJECT_ROOT/docs/scripts:/scripts" \
  -w /scripts \
  node:20-alpine \
  sh -c "npm install --no-save fast-json-patch js-yaml && node /scripts/apply-patches.js /spec/openapi.json /patches /spec/openapi.json"

# Step 5: Polish OpenAPI JSON using Docker Node.js
echo "Step 5: Polishing OpenAPI JSON..."
docker run --rm \
  -v "$PROJECT_ROOT/api/spec:/spec" \
  -v "$PROJECT_ROOT/docs/scripts:/scripts" \
  -w /spec \
  node:20-alpine \
  sh -c "node /scripts/polish-openapi.js /spec/openapi.json /spec/openapi.json"

# Step 6: Lint OpenAPI spec using Docker Redocly
echo "Step 6: Linting OpenAPI spec..."
docker run --rm \
  --entrypoint="" \
  -v "$PROJECT_ROOT:/app" \
  -w /app \
  redocly/cli:latest \
  sh -c "redocly lint api/spec/openapi.json" || {
    echo "Warning: Linting found issues, but continuing..."
  }

# Step 7: Generate HTML documentation using Docker Redocly
echo "Step 7: Generating HTML documentation..."
mkdir -p "$PROJECT_ROOT/public/api"

docker run --rm \
  --entrypoint="" \
  -v "$PROJECT_ROOT:/app" \
  -w /app \
  redocly/cli:latest \
  sh -c "redocly build-docs api/spec/openapi.json --output public/api/index.html --config redocly.yaml" || {
    echo "Warning: Redocly build failed, using fallback..."
    # Fallback: Use Redoc standalone HTML
    cat > "$PROJECT_ROOT/public/api/index.html" << 'EOF'
<!DOCTYPE html>
<html>
  <head>
    <title>URL Shortener API Documentation</title>
    <meta charset="utf-8"/>
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <link href="https://fonts.googleapis.com/css?family=Montserrat:300,400,700|Roboto:300,400,700" rel="stylesheet">
    <style>
      body { margin: 0; padding: 0; }
    </style>
  </head>
  <body>
    <redoc spec-url='api/spec/openapi.json'></redoc>
    <script src="https://cdn.redoc.ly/redoc/latest/bundles/redoc.standalone.js"></script>
  </body>
</html>
EOF
  }

echo "✓ Documentation build complete!"
echo "  OpenAPI spec: $PROJECT_ROOT/api/spec/openapi.json"
echo "  HTML docs: $PROJECT_ROOT/public/api/index.html"
