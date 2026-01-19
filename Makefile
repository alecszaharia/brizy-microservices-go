.PHONY: contracts-generate contracts-lint contracts-breaking contracts-format contracts-clean contracts-check contracts-all lint lint-fix lint-all lint-all-fix lint-platform lint-platform-fix help

# init env
init:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install github.com/envoyproxy/protoc-gen-validate@latest
	go install github.com/go-kratos/kratos/cmd/protoc-gen-go-http/v2@latest
	go install github.com/google/gnostic/cmd/protoc-gen-openapi@latest
	go install connectrpc.com/connect/cmd/protoc-gen-connect-go@latest
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest
	go install github.com/bufbuild/buf/cmd/buf@v1.62.0
	go install github.com/go-kratos/kratos/cmd/kratos/v2@latest
	go install github.com/google/wire/cmd/wire@latest

# Contracts related

# Generate code from protobuf definitions
contracts-deps:
	buf dep update

contracts-generate: contracts-deps
	buf generate

# Lint protobuf files
contracts-lint: contracts-deps
	buf lint

# Check for breaking changes
contracts-breaking: contracts-deps
	buf breaking --against '.git#branch=main'

# Format protobuf files
contracts-format: contracts-deps
	buf format -w

# Clean generated files
contracts-clean:
	rm -rf contracts/*

# Run all checks (lint + breaking)
contracts-check: contracts-lint contracts-breaking

# Full workflow: format, lint, and generate
contracts-all: contracts-format contracts-lint contracts-generate


# Linting (golangci-lint)

lint-platform:
	@echo "Running golangci-lint on platform module..."
	cd platform && golangci-lint run ./...

lint-platform-fix:
	@echo "Running golangci-lint with auto-fix on platform module..."
	cd platform && golangci-lint run --fix ./...

lint:
	@echo "ERROR: 'make lint' is ambiguous in this monorepo."
	@echo "Please use: make lint-platform, make lint-all, or cd to a specific module"
	@exit 1

lint-fix:
	@echo "ERROR: 'make lint-fix' is ambiguous in this monorepo."
	@echo "Please use: make lint-platform-fix, make lint-all-fix, or cd to a specific module"
	@exit 1

lint-all:
	@echo "Linting all modules..."
	@echo "→ Platform module..."
	@cd platform && golangci-lint run ./...
	@echo "→ Contracts skipped (generated code)"
	@for service in services/*; do \
		if [ -d "$$service" ] && [ -f "$$service/.golangci.yml" ]; then \
			echo "→ Linting $$service..."; \
			cd $$service && golangci-lint run ./... && cd ../..; \
		fi \
	done
	@echo "✓ All modules linted successfully"

lint-all-fix:
	@echo "Linting all modules with auto-fix..."
	@cd platform && golangci-lint run --fix ./...
	@for service in services/*; do \
		if [ -d "$$service" ] && [ -f "$$service/.golangci.yml" ]; then \
			cd $$service && golangci-lint run --fix ./... && cd ../..; \
		fi \
	done
	@echo "✓ All modules linted and fixed"

# Display help
help:
	@echo "Available targets:"
	@echo ""
	@echo "Setup:"
	@echo "  init                - Install all required tools (buf, protoc plugins, wire, golangci-lint)"
	@echo ""
	@echo "Contracts (proto/buf):"
	@echo "  contracts-generate  - Generate Go code, Connect RPC, and OpenAPI from proto files"
	@echo "  contracts-lint      - Lint protobuf files"
	@echo "  contracts-breaking  - Check for breaking changes against main branch"
	@echo "  contracts-format    - Format protobuf files"
	@echo "  contracts-clean     - Remove generated files in contracts/"
	@echo "  contracts-check     - Run lint and breaking change checks"
	@echo "  contracts-all       - Format, lint, and generate (full workflow)"
	@echo ""
	@echo "Linting (Go code):"
	@echo "  lint-platform       - Run golangci-lint on platform module"
	@echo "  lint-platform-fix   - Run golangci-lint with auto-fix on platform"
	@echo "  lint-all            - Run golangci-lint on all modules (platform + services)"
	@echo "  lint-all-fix        - Run golangci-lint with auto-fix on all modules"