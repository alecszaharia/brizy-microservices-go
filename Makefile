.PHONY: contracts-generate contracts-lint contracts-breaking contracts-format contracts-clean contracts-check contracts-all help

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

# Display help
help:
	@echo "Available targets:"
	@echo ""
	@echo "Contracts (proto/buf):"
	@echo "  contracts-generate  - Generate Go code, Connect RPC, and OpenAPI from proto files"
	@echo "  contracts-lint      - Lint protobuf files"
	@echo "  contracts-breaking  - Check for breaking changes against main branch"
	@echo "  contracts-format    - Format protobuf files"
	@echo "  contracts-clean     - Remove generated files in contracts/"
	@echo "  contracts-check     - Run lint and breaking change checks"
	@echo "  contracts-all       - Format, lint, and generate (full workflow)"