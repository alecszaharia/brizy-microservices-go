# Suggested Commands

## Root-Level Commands (Contracts/Proto)

Execute from repository root:

```bash
# Generate code from proto files (gRPC, Connect RPC, OpenAPI)
make contracts-generate

# Lint protobuf files
make contracts-lint

# Check for breaking API changes against main branch
make contracts-breaking

# Format proto files
make contracts-format

# Full workflow: format, lint, and generate
make contracts-all

# Clean generated files
make contracts-clean
```

## Service-Level Commands

Execute from `services/{service-name}/` directory:

```bash
cd services/symbols  # Or cd services/{service-name}

# Install required development tools
make init

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

# Full workflow: config + generate
make all
```

## Testing Commands

```bash
# Run all tests for a service
cd services/symbols && go test ./internal/...

# Run specific test
cd services/symbols && go test -run TestSymbolUseCase_Create ./internal/biz/

# Run tests with verbose output
cd services/symbols && go test -v ./internal/...

# Run tests with coverage
cd services/symbols && go test -coverprofile=coverage.out ./internal/...

# View coverage in browser
cd services/symbols && go tool cover -html=coverage.out
```

## Running the Service

```bash
# Build and run
cd services/symbols
make build
./bin/symbols -conf configs/config.yaml

# Run with Docker
docker-compose up symbols
```

## Common Development Workflow

```bash
# 1. Modify proto files in api/{service}/v1/
# 2. Generate contracts
make contracts-all

# 3. Modify service code
cd services/symbols

# 4. Update wire dependencies if needed (modify wire.go)
make generate

# 5. Run tests
make test

# 6. Build
make build
```

## Git Operations

```bash
# Standard git commands work as expected
git status
git add .
git commit -m "message"
git push
```
