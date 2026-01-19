GOHOSTOS:=$(shell go env GOHOSTOS)
GOPATH:=$(shell go env GOPATH)
VERSION=$(shell git describe --tags --always)

ifeq ($(GOHOSTOS), windows)
	#the `find.exe` is different from `find` in bash/shell.
	#to see https://docs.microsoft.com/en-us/windows-server/administration/windows-commands/find.
	#changed to use git-bash.exe to run find cli or other cli friendly, caused of every developer has a Git.
	#Git_Bash= $(subst cmd\,bin\bash.exe,$(dir $(shell where git)))
	Git_Bash=$(subst \,/, $(subst cmd\,bin\bash.exe,$(dir $(shell where git))))
	INTERNAL_PROTO_FILES=$(shell $(Git_Bash) -c "find internal -name *.proto")
	API_PROTO_FILES=$(shell $(Git_Bash) -c "find api -name *.proto")
else
	INTERNAL_PROTO_FILES=$(shell find internal -name *.proto)
	API_PROTO_FILES=$(shell find api -name *.proto)
endif

.PHONY: init config api build generate all test coverage lint lint-fix help

APP_NAME := symbols
BIN_DIR := bin

# init env
init:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install github.com/go-kratos/kratos/cmd/kratos/v2@latest
	go install github.com/go-kratos/kratos/cmd/protoc-gen-go-http/v2@latest
	go install github.com/google/gnostic/cmd/protoc-gen-openapi@latest
	go install github.com/google/wire/cmd/wire@latest
	go install github.com/envoyproxy/protoc-gen-validate@latest

# generate internal proto
config:
	buf dep update
	buf lint
	buf format -w
	buf generate

# build
build: service worker

service: generate
	go build -o $(BIN_DIR)/$(APP_NAME) ./cmd/$(APP_NAME)

worker: generate
	go build -o $(BIN_DIR)/$(APP_NAME)-worker ./cmd/$(APP_NAME)-worker

# generate wire files
generate:
	wire ./cmd/symbols
	wire ./cmd/symbols-worker
	go mod tidy

# generate all
all:
	make config;
	make generate;

# run unit tests
test:
	go test -v -race -coverprofile=coverage.out ./internal/...

# generate coverage report
coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# run linting
lint:
	@echo "Running golangci-lint on symbols service..."
	golangci-lint run ./...

# run linting with auto-fix
lint-fix:
	@echo "Running golangci-lint with auto-fix on symbols service..."
	golangci-lint run --fix ./...

# show help
help:
	@echo ''
	@echo 'Usage:'
	@echo ' make [target]'
	@echo ''
	@echo 'Targets:'
	@awk '/^[a-zA-Z\-\_0-9]+:/ { \
	helpMessage = match(lastLine, /^# (.*)/); \
		if (helpMessage) { \
			helpCommand = substr($$1, 0, index($$1, ":")); \
			helpMessage = substr(lastLine, RSTART + 2, RLENGTH); \
			printf "\033[36m%-22s\033[0m %s\n", helpCommand,helpMessage; \
		} \
	} \
	{ lastLine = $$0 }' $(MAKEFILE_LIST)

.DEFAULT_GOAL := help
