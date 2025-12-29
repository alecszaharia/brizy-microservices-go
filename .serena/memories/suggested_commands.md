# Suggested Commands

## Initial Setup

Install required development tools:

```bash
# From repository root
make init
```

This installs:
- buf, protoc plugins (protoc-gen-go, protoc-gen-go-grpc, protoc-gen-go-http, protoc-gen-openapi, protoc-gen-validate)
- wire (dependency injection)
- kratos CLI

## Root-Level Commands (Contracts/Proto)

Execute from repository root:

```bash
# Full workflow: format, lint, and generate
make contracts-all

# Generate code from proto files (gRPC, Connect RPC, OpenAPI)
make contracts-generate

# Format proto files
make contracts-format

# Lint protobuf files
make contracts-lint

# Check for breaking API changes against main branch
make contracts-breaking

# Clean generated files
make contracts-clean

# View all available make targets
make help
```

**Important**: Always run from repository root. Generated code goes to `contracts/{service}/v1/`.

## Service-Level Commands

Execute from `services/{service-name}/` directory:

```bash
cd services/symbols  # Or cd services/{service-name}

# Generate Wire dependency injection code
make generate
# Note: Uses GOWORK=off to avoid workspace conflicts

# Build the service binary
make build
# Output: bin/{service-name}

# Run unit tests with race detection
make test

# Generate coverage report (opens in browser)
make coverage

# Generate config protobuf code
make config

# Full service workflow: config + generate
make all

# View service-specific make targets
make help
```

## Testing Commands

```bash
# Run all tests for a service
cd services/symbols && go test ./internal/...

# Run specific test by name
cd services/symbols && go test -run TestSymbolUseCase_Create ./internal/biz/

# Run tests with verbose output
cd services/symbols && go test -v ./internal/...

# Run tests with coverage
cd services/symbols && go test -coverprofile=coverage.out ./internal/...

# View coverage in browser
cd services/symbols && go tool cover -html=coverage.out

# Run tests with race detection (recommended)
cd services/symbols && go test -race ./internal/...
```

## Running the Service

```bash
# Build and run locally
cd services/symbols
make build
./bin/symbols -conf configs/config.yaml

# Run with Docker Compose
docker-compose up symbols

# Run in detached mode
docker-compose up -d symbols

# View service logs
docker-compose logs -f symbols
```

## Common Development Workflows

### Workflow 1: Modifying Proto Files

```bash
# 1. Edit proto files in api/{service}/v1/
# 2. Generate contracts from repository root
make contracts-all

# 3. If service code needs updating, go to service directory
cd services/symbols

# 4. Update service implementation
# 5. Run tests
make test

# 6. Build
make build
```

### Workflow 2: Adding New Business Logic

```bash
cd services/symbols

# 1. Add/modify code in internal/biz/ or internal/service/
# 2. If dependencies changed, update wire.go and regenerate
make generate

# 3. Run tests
make test

# 4. Build
make build
```

### Workflow 3: Full Clean Build

```bash
# From repository root
make contracts-all

cd services/symbols
make generate
make test
make build
```

## Go Module Commands

```bash
# Tidy dependencies
go mod tidy

# Download dependencies
go mod download

# Verify dependencies
go mod verify

# List all modules
go list -m all
```

## Git Operations

```bash
# Standard git commands work as expected
git status
git add .
git commit -m "descriptive message"
git push

# View recent commits
git log --oneline -10

# View diff
git diff
git diff --cached  # Staged changes
```

## Docker Commands

```bash
# Start all services
docker-compose up

# Start specific service
docker-compose up symbols

# Stop all services
docker-compose down

# View logs
docker-compose logs -f symbols

# Rebuild and start
docker-compose up --build

# Execute command in running container
docker-compose exec symbols sh
```

## Debugging Commands

```bash
# Check if port is in use
lsof -i :8000
netstat -tlnp | grep :8000

# Test HTTP endpoint
curl http://localhost:8000/health
curl -X POST http://localhost:8000/v1/symbols -H "Content-Type: application/json" -d '{"name":"test"}'

# View Go environment
go env

# Check workspace
go work edit -print
```

## Quick Reference

**Most common commands:**
1. `make contracts-all` (after editing proto files)
2. `cd services/symbols && make generate` (after editing wire.go)
3. `make test` (before committing)
4. `make build` (to verify build)
