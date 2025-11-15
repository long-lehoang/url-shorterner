# Documentation Pipeline

This project uses a comprehensive documentation pipeline for generating API documentation.

## Overview

The documentation pipeline consists of:

1. **MkDocs Material** - Development documentation (markdown files)
2. **Swagger/OpenAPI 3.0** - API specifications
3. **Redocly CLI** - Interactive HTML documentation generation

## Build Process

The documentation build process follows these steps (all Node.js tools run in Docker containers):

1. **Generate Swagger 2.0** from Go code annotations using `swaggo/swag`
2. **Polish Swagger JSON** - Remove Go-specific extensions (Docker: `node:20-alpine`)
3. **Convert to OpenAPI 3.0** - Migrate Swagger 2.0 → OpenAPI 3.0 (Docker: `mermade/swagger2openapi`)
4. **Apply Patches** - Apply JSON patch files for metadata (Docker: `node:20-alpine`)
5. **Sort Schema Properties** - Sort properties using x-order annotations (Docker: `node:20-alpine`)
6. **Lint OpenAPI** - Validate spec (Docker: `redocly/cli`)
7. **Generate HTML** - Create interactive HTML documentation (Docker: `redocly/cli`)

## Quick Start

### Prerequisites

The build script uses Docker containers for all Node.js tools, so no local npm installation is required!

**Required:**
- Docker (for running Node.js tools in containers)
- Go 1.25+ (for generating Swagger from annotations)

**Optional (for local development):**
- Node.js 18+ (if you want to run scripts locally instead of Docker)
- MkDocs Material (for dev docs): `pip install mkdocs-material`

### Build Documentation

```bash
# Build all documentation
make docs

# Build only API docs
make docs-build

# Build only MkDocs site
make docs-build-mkdocs

# Serve MkDocs locally
make docs-serve
```

## Directory Structure

```
docs/
├── api/
│   └── swagger.yml          # Base Swagger definitions
├── scripts/
│   ├── build-docs.sh        # Main build script
│   ├── polish-swagger.js    # Clean Swagger JSON
│   ├── polish-openapi.js    # Sort OpenAPI schemas
│   └── apply-patches.js    # Apply JSON patches
├── patches/
│   ├── info.patch.yaml      # API info metadata
│   └── meta.patch.yaml      # Tags, servers, etc.
└── dev/                     # MkDocs source files
    └── index.md

api/
└── spec/
    └── openapi.json         # Generated OpenAPI 3.0 spec

public/
├── api/
│   └── index.html           # Generated HTML docs
└── dev/                     # MkDocs site
```

## Configuration Files

### `redocly.yaml`

Redocly configuration for theme, linting rules, and API definitions.

### `mkdocs.yml`

MkDocs Material configuration for development documentation.

### Patch Files

- `docs/patches/info.patch.yaml` - API metadata (title, version, description, contact, license)
- `docs/patches/meta.patch.yaml` - Servers, tags, and other metadata

## Docker Build

You can also build documentation using Docker:

```bash
docker build -f Dockerfile.docs -t url-shortener-docs .
docker run -p 8080:80 url-shortener-docs
```

## Manual Steps

If you need to run steps manually (all using Docker):

```bash
# 1. Generate Swagger
make swagger-gen

# 2. Polish Swagger (Docker)
docker run --rm \
  -v $(pwd)/docs:/docs \
  -v $(pwd)/docs/scripts:/scripts \
  -w /docs \
  node:20-alpine \
  sh -c "node /scripts/polish-swagger.js /docs/swagger.json /docs/swagger-polished.json"

# 3. Convert to OpenAPI 3.0 (Docker)
docker run --rm \
  -v $(pwd)/docs:/docs \
  -v $(pwd)/api/spec:/output \
  mermade/swagger2openapi:latest \
  swagger2openapi /docs/swagger-polished.json -o /output/openapi.json

# 4. Apply patches (Docker)
docker run --rm \
  -v $(pwd)/api/spec:/spec \
  -v $(pwd)/docs/patches:/patches \
  -v $(pwd)/docs/scripts:/scripts \
  -w /spec \
  node:20-alpine \
  sh -c "npm install -g fast-json-patch js-yaml && node /scripts/apply-patches.js /spec/openapi.json /patches /spec/openapi.json"

# 5. Polish OpenAPI (Docker)
docker run --rm \
  -v $(pwd)/api/spec:/spec \
  -v $(pwd)/docs/scripts:/scripts \
  -w /spec \
  node:20-alpine \
  sh -c "node /scripts/polish-openapi.js /spec/openapi.json /spec/openapi.json"

# 6. Lint (Docker)
docker run --rm \
  -v $(pwd):/app \
  -w /app \
  redocly/cli:latest \
  redocly lint api/spec/openapi.json

# 7. Generate HTML (Docker)
docker run --rm \
  -v $(pwd):/app \
  -w /app \
  redocly/cli:latest \
  redocly build-docs api/spec/openapi.json --output public/api/index.html --config redocly.yaml
```

## Output

- **OpenAPI Spec**: `api/spec/openapi.json`
- **HTML Docs**: `public/api/index.html`
- **MkDocs Site**: `public/dev/`

## Troubleshooting

### Docker images not found

The build script automatically pulls required Docker images:
- `node:20-alpine` - For Node.js scripts
- `mermade/swagger2openapi:latest` - For Swagger to OpenAPI conversion
- `redocly/cli:latest` - For linting and HTML generation

If images are not found, Docker will automatically pull them on first run.

### OpenAPI conversion fails

The script uses `mermade/swagger2openapi` Docker image. If it fails, ensure Docker is running and has internet access to pull the image.

### MkDocs not found

Install MkDocs Material (optional, only for dev docs):
```bash
pip install mkdocs-material
```

Or use Docker:
```bash
docker run --rm -p 8000:8000 -v $(pwd):/docs squidfunk/mkdocs-material mkdocs serve -a 0.0.0.0:8000
```

