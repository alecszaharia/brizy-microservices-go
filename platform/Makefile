.PHONY: init tidy lint lint-fix

# init env
init:
	go get github.com/go-kratos/kratos/v2/middleware@latest
	go get github.com/google/uuid@latest

tidy:
	go mod tidy

# Lint Go code
lint:
	@echo "Running golangci-lint on platform..."
	golangci-lint run ./...

# Lint with auto-fix
lint-fix:
	@echo "Running golangci-lint with auto-fix on platform..."
	golangci-lint run --fix ./...