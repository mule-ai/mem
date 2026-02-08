# Variables
BINARY_NAME=mem
MAIN_PATH=./cmd/mem
BIN_DIR=./bin
COVERAGE_FILE=coverage.out
COVERAGE_HTML=coverage.html

# Build variables
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME?=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS=-ldflags "-X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME)"

# Go variables
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=gofmt
GOLINT=golangci-lint

# Default target
.PHONY: all
all: build

## build: Build the binary
.PHONY: build
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BIN_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BIN_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "Built $(BIN_DIR)/$(BINARY_NAME)"

## build-local: Build binary for local installation
.PHONY: build-local
build-local:
	@echo "Building $(BINARY_NAME) for local use..."
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) $(MAIN_PATH)
	@echo "Built ./$(BINARY_NAME)"

## install: Install the binary to $GOPATH/bin or /usr/local/bin
.PHONY: install
install:
	@echo "Installing $(BINARY_NAME)..."
	$(GOBUILD) $(LDFLAGS) -o $$GOPATH/bin/$(BINARY_NAME) $(MAIN_PATH)
	@echo "Installed to $$GOPATH/bin/$(BINARY_NAME)"

## install-system: Install system-wide (requires sudo)
.PHONY: install-system
install-system: build
	@echo "Installing $(BINARY_NAME) system-wide..."
	sudo cp $(BIN_DIR)/$(BINARY_NAME) /usr/local/bin/
	@echo "Installed to /usr/local/bin/$(BINARY_NAME)"

## clean: Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning..."
	@rm -f $(BIN_DIR)/$(BINARY_NAME)
	@rm -f $(BINARY_NAME)
	@rm -f $(COVERAGE_FILE)
	@rm -f $(COVERAGE_HTML)
	@echo "Clean complete"

## test: Run all tests
.PHONY: test
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

## test-short: Run short tests only
.PHONY: test-short
test-short:
	@echo "Running short tests..."
	$(GOTEST) -short -v ./...

## test-coverage: Run tests with coverage
.PHONY: test-coverage
test-coverage:
	@echo "Running tests with coverage..."
	$(GOTEST) -coverprofile=$(COVERAGE_FILE) -covermode=atomic ./...
	@echo "Coverage report generated: $(COVERAGE_FILE)"

## test-coverage-html: Generate HTML coverage report
.PHONY: test-coverage-html
test-coverage-html: test-coverage
	@echo "Generating HTML coverage report..."
	$(GOCMD) tool cover -html=$(COVERAGE_FILE) -o $(COVERAGE_HTML)
	@echo "HTML coverage report: $(COVERAGE_HTML)"

## test-race: Run tests with race detector
.PHONY: test-race
test-race:
	@echo "Running tests with race detector..."
	$(GOTEST) -race -v ./...

## benchmark: Run benchmarks
.PHONY: benchmark
benchmark:
	@echo "Running benchmarks..."
	$(GOTEST) -bench=. -benchmem ./...

## lint: Run linter
.PHONY: lint
lint:
	@echo "Running linter..."
	$(GOLINT) run ./...

## fmt: Format code
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	$(GOFMT) -s -w .
	@echo "Code formatted"

## fmt-check: Check if code is formatted
.PHONY: fmt-check
fmt-check:
	@echo "Checking code formatting..."
	@test -z "$$($(GOFMT) -l .)" || (echo "Code is not formatted. Run 'make fmt'" && exit 1)
	@echo "Code is properly formatted"

## vet: Run go vet
.PHONY: vet
vet:
	@echo "Running go vet..."
	$(GOCMD) vet ./...

## deps: Download dependencies
.PHONY: deps
deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

## deps-update: Update dependencies
.PHONY: deps-update
deps-update:
	@echo "Updating dependencies..."
	$(GOGET) -u ./...
	$(GOMOD) tidy

## mod-verify: Verify dependencies
.PHONY: mod-verify
mod-verify:
	@echo "Verifying dependencies..."
	$(GOMOD) verify

## run: Run the application
.PHONY: run
run: build-local
	@echo "Running $(BINARY_NAME)..."
	./$(BINARY_NAME)

## help: Show this help message
.PHONY: help
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'

## ci: Run CI checks (fmt, vet, test)
.PHONY: ci
ci: fmt-check vet test
	@echo "CI checks passed!"

## pre-commit: Run pre-commit checks
.PHONY: pre-commit
pre-commit: fmt-check vet test-short
	@echo "Pre-commit checks passed!"

## release: Build release binaries for multiple platforms
.PHONY: release
release:
	@echo "Building release binaries..."
	@mkdir -p $(BIN_DIR)
	@echo "Building for Linux (amd64)..."
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BIN_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)
	@echo "Building for Linux (arm64)..."
	GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BIN_DIR)/$(BINARY_NAME)-linux-arm64 $(MAIN_PATH)
	@echo "Building for macOS (amd64)..."
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BIN_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)
	@echo "Building for macOS (arm64)..."
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BIN_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_PATH)
	@echo "Release binaries built in $(BIN_DIR)/"

## docker-build: Build Docker image
.PHONY: docker-build
docker-build:
	@echo "Building Docker image..."
	docker build -t $(BINARY_NAME):$(VERSION) .
	docker tag $(BINARY_NAME):$(VERSION) $(BINARY_NAME):latest

## docker-run: Run Docker container
.PHONY: docker-run
docker-run:
	@echo "Running Docker container..."
	docker run --rm -it $(BINARY_NAME):latest

## clean-all: Clean everything including cached data
.PHONY: clean-all
clean-all: clean
	@echo "Cleaning all caches..."
	$(GOCMD) clean -cache -testcache -modcache -i
	@echo "All cleaned"

## tools: Install development tools
.PHONY: tools
	@echo "Installing development tools..."
	$(GOGET) -u github.com/golangci/golangci-lint/cmd/golangci-lint
	@echo "Development tools installed"