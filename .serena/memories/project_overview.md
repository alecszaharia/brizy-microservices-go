# Project Overview

## Purpose
This is a **Brizy microservices monorepo** built with Go. The project provides cloud-native microservices for managing Brizy platform entities (currently includes a symbol-service for symbol management).

## Tech Stack

### Core Technologies
- **Language**: Go 1.25 with workspaces
- **Framework**: go-kratos/kratos (cloud-native microservices framework)
- **Architecture**: Clean Architecture with dependency injection
- **ORM**: GORM for database access
- **API**: Protocol Buffers (gRPC, Connect RPC, HTTP/JSON)
- **Dependency Injection**: Google Wire
- **Containerization**: Docker & Docker Compose

### Development Tools
- **buf**: Protobuf linting, generation, and breaking change detection
- **protoc-gen-go**: Go code generation from proto files
- **protoc-gen-go-grpc**: gRPC service generation
- **protoc-gen-go-http**: Kratos HTTP bindings
- **protoc-gen-openapi**: OpenAPI specification generation
- **protoc-gen-validate**: Proto validation rules
- **wire**: Dependency injection code generation

## Key Features
- Clean Architecture with clear layer separation (service → biz → data)
- Dual transport support (gRPC and HTTP/JSON)
- Connect RPC for browser-friendly gRPC
- Automatic API generation from protobuf definitions
- Offset-based pagination utilities
- Request ID middleware with context propagation
