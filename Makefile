.PHONY: contracts-generate contracts-lint contracts-breaking contracts-format contracts-clean contracts-check contracts-all help

# Contracts related

# Generate code from protobuf definitions
contracts-generate:
	buf generate

# Lint protobuf files
contracts-lint:
	buf lint

# Check for breaking changes
contracts-breaking:
	buf breaking --against '.git#branch=main'

# Format protobuf files
contracts-format:
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