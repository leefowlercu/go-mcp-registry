.PHONY: help test test-verbose test-cover test-integration test-all build examples build-all fmt vet lint check deps tidy update-deps clean coverage ci run-list run-get run-paginate

# Default target
help: ## Display available make targets
	@echo "Available targets:"
	@echo "  help                 Display available make targets"
	@echo "  test                 Run unit tests"
	@echo "  test-verbose         Run unit tests with verbose output"
	@echo "  test-cover           Run unit tests with coverage"
	@echo "  test-integration     Run integration tests (requires network access)"
	@echo "  test-all             Run both unit and integration tests"
	@echo "  build                Build all packages"
	@echo "  examples             Build all example programs"
	@echo "  build-all            Build everything (packages + examples)"
	@echo "  fmt                  Format code using gofmt"
	@echo "  vet                  Run go vet for static analysis"
	@echo "  lint                 Run formatting and static analysis"
	@echo "  check                Run all quality checks (format, vet, test)"
	@echo "  deps                 Download dependencies"
	@echo "  tidy                 Tidy go.mod and go.sum"
	@echo "  update-deps          Update dependencies to latest versions"
	@echo "  clean                Clean build artifacts and test cache"
	@echo "  coverage             Generate HTML coverage report"
	@echo "  ci                   Run full CI pipeline"
	@echo "  run-list             Run the list example"
	@echo "  run-get              Run the get example with default server"
	@echo "  run-paginate         Run the paginate example"

# Variables
GO_FILES := $(shell find . -name '*.go' -not -path './test/*' -not -path './.git/*')
EXAMPLES_DIR := ./examples
BUILD_DIR := ./build
COVERAGE_FILE := coverage.out

# Test targets
test: ## Run unit tests
	@echo "Running unit tests..."
	go test ./...

test-verbose: ## Run unit tests with verbose output
	@echo "Running unit tests with verbose output..."
	go test -v ./...

test-cover: ## Run unit tests with coverage
	@echo "Running unit tests with coverage..."
	go test -cover ./...
	go test -coverprofile=$(COVERAGE_FILE) ./...

test-integration: ## Run integration tests (requires network access)
	@echo "Running integration tests..."
	INTEGRATION_TESTS=true go test -v ./test/integration/

test-all: test test-integration ## Run both unit and integration tests

# Build targets
build: ## Build all packages
	@echo "Building all packages..."
	go build ./...

examples: ## Build all example programs
	@echo "Building examples..."
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/list $(EXAMPLES_DIR)/list
	go build -o $(BUILD_DIR)/get $(EXAMPLES_DIR)/get
	go build -o $(BUILD_DIR)/paginate $(EXAMPLES_DIR)/paginate

build-all: build examples ## Build everything (packages + examples)

# Code quality targets
fmt: ## Format code using gofmt
	@echo "Formatting code..."
	gofmt -s -w .

vet: ## Run go vet for static analysis
	@echo "Running go vet..."
	go vet ./...

lint: fmt vet ## Run formatting and static analysis

check: lint test ## Run all quality checks (format, vet, test)

# Dependency management
deps: ## Download dependencies
	@echo "Downloading dependencies..."
	go mod download

tidy: ## Tidy go.mod and go.sum
	@echo "Tidying dependencies..."
	go mod tidy

update-deps: ## Update dependencies to latest versions
	@echo "Updating dependencies..."
	go get -u ./...
	go mod tidy

# Development targets
clean: ## Clean build artifacts and test cache
	@echo "Cleaning build artifacts..."
	go clean -testcache -cache
	rm -rf $(BUILD_DIR)
	rm -f $(COVERAGE_FILE) coverage.html

coverage: test-cover ## Generate HTML coverage report
	@echo "Generating HTML coverage report..."
	go tool cover -html=$(COVERAGE_FILE) -o coverage.html
	@echo "Coverage report generated: coverage.html"

ci: deps lint test build ## Run full CI pipeline

# Example runners
run-list: ## Run the list example
	@echo "Running list example..."
	go run $(EXAMPLES_DIR)/list/

run-get: ## Run the get example with default server
	@echo "Running get example with default server..."
	go run $(EXAMPLES_DIR)/get/ "ai.waystation/gmail"

run-paginate: ## Run the paginate example
	@echo "Running paginate example..."
	go run $(EXAMPLES_DIR)/paginate/