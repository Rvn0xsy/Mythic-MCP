.PHONY: all build test test-unit test-integration test-e2e lint fmt clean coverage help

# Build variables
BINARY_NAME=mythic-mcp
VERSION?=dev
BUILD_DIR=bin
GO_FILES=$(shell find . -name '*.go' -type f)

# Go commands
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=$(GOCMD) fmt

all: lint test build ## Run all checks and build

build: ## Build the server binary
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -ldflags "-X main.Version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/mythic-mcp
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

test: test-unit ## Run all tests

test-unit: ## Run unit tests
	@echo "Running unit tests..."
	$(GOTEST) -v -coverprofile=coverage.out -covermode=atomic ./pkg/... ./cmd/...

test-integration: ## Run integration tests (no Mythic required)
	@echo "Running integration tests..."
	$(GOTEST) -v -tags=integration ./tests/integration/...

test-e2e: ## Run E2E tests (requires Mythic)
	@echo "Running E2E tests..."
	$(GOTEST) -v -timeout 20m -tags=e2e ./tests/integration/...

coverage: test-unit ## Generate coverage report
	@echo "Generating coverage report..."
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"
	@$(GOCMD) tool cover -func=coverage.out | grep total

lint: ## Run linter
	@echo "Running linter..."
	@which golangci-lint > /dev/null || (echo "golangci-lint not installed. Run: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest" && exit 1)
	golangci-lint run --timeout 5m

fmt: ## Format code
	@echo "Formatting code..."
	$(GOFMT) ./...

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
	@rm -f mythic-mcp

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

help: ## Show this help
	@echo "Mythic MCP Server - Makefile Commands"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
