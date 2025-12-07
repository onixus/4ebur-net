# Makefile for 4ebur-net

.PHONY: all build test lint clean run docker-build docker-run help

# Variables
BINARY_NAME=4ebur-net
GO=go
GOFLAGS=-v
LDFLAGS=-s -w
MAIN_PATH=cmd/proxy/main.go

# Build info
VERSION?=$(shell git describe --tags --always --dirty)
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')

all: test build

## build: Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	$(GO) build $(GOFLAGS) -ldflags="$(LDFLAGS) -X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME)" -o $(BINARY_NAME) $(MAIN_PATH)
	@echo "Build complete: $(BINARY_NAME)"

## build-all: Build for all platforms
build-all:
	@echo "Building for multiple platforms..."
	GOOS=linux GOARCH=amd64 $(GO) build -ldflags="$(LDFLAGS)" -o dist/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)
	GOOS=linux GOARCH=arm64 $(GO) build -ldflags="$(LDFLAGS)" -o dist/$(BINARY_NAME)-linux-arm64 $(MAIN_PATH)
	GOOS=darwin GOARCH=amd64 $(GO) build -ldflags="$(LDFLAGS)" -o dist/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)
	GOOS=darwin GOARCH=arm64 $(GO) build -ldflags="$(LDFLAGS)" -o dist/$(BINARY_NAME)-darwin-arm64 $(MAIN_PATH)
	GOOS=windows GOARCH=amd64 $(GO) build -ldflags="$(LDFLAGS)" -o dist/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PATH)
	@echo "Multi-platform build complete"

## test: Run tests
test:
	@echo "Running tests..."
	$(GO) test -v -race -coverprofile=coverage.out ./...
	@echo "Tests complete"

## test-coverage: Run tests with coverage report
test-coverage: test
	@echo "Generating coverage report..."
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

## bench: Run benchmarks
bench:
	@echo "Running benchmarks..."
	$(GO) test -bench=. -benchmem ./...

## lint: Run linters
lint:
	@echo "Running linters..."
	golangci-lint run --timeout=5m
	gofmt -s -l .
	$(GO) vet ./...

## fmt: Format code
fmt:
	@echo "Formatting code..."
	gofmt -s -w .
	goimports -w .

## clean: Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -f $(BINARY_NAME)
	rm -f coverage.out coverage.html
	rm -rf dist/
	@echo "Clean complete"

## run: Build and run
run: build
	@echo "Starting $(BINARY_NAME)..."
	./$(BINARY_NAME)

## docker-build: Build Docker image
docker-build:
	@echo "Building Docker image..."
	docker build -t onixus/$(BINARY_NAME):latest .
	@echo "Docker build complete"

## docker-build-alpine: Build Alpine Docker image
docker-build-alpine:
	@echo "Building Alpine Docker image..."
	docker build -f Dockerfile.alpine -t onixus/$(BINARY_NAME):alpine .
	@echo "Alpine Docker build complete"

## docker-run: Run Docker container
docker-run:
	@echo "Running Docker container..."
	docker run -d --name $(BINARY_NAME) -p 1488:1488 onixus/$(BINARY_NAME):latest

## docker-stop: Stop Docker container
docker-stop:
	@echo "Stopping Docker container..."
	docker stop $(BINARY_NAME)
	docker rm $(BINARY_NAME)

## docker-compose-up: Start with docker-compose
docker-compose-up:
	@echo "Starting with docker-compose..."
	docker-compose up -d

## docker-compose-down: Stop docker-compose
docker-compose-down:
	@echo "Stopping docker-compose..."
	docker-compose down

## install-tools: Install development tools
install-tools:
	@echo "Installing development tools..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install golang.org/x/tools/cmd/goimports@latest
	@echo "Tools installed"

## deps: Download dependencies
deps:
	@echo "Downloading dependencies..."
	$(GO) mod download
	$(GO) mod verify
	@echo "Dependencies ready"

## update-deps: Update dependencies
update-deps:
	@echo "Updating dependencies..."
	$(GO) get -u ./...
	$(GO) mod tidy
	@echo "Dependencies updated"

## help: Show this help
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

.DEFAULT_GOAL := help
