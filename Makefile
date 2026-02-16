# AgentFramework Makefile
# Supports both desktop and CLI builds

.PHONY: all build clean test desktop cli run-desktop run-cli install help \
          test-unit test-integration test-platform test-all ci \
          lint security-check benchmark deps-install

# Variables
BINARY_NAME=agentframework
CLI_BINARY_NAME=af
DESKTOP_BINARY_NAME=AgentFramework
BUILD_DIR=build
CMD_DIR=cmd

# Go variables
GOCMD=go
GOFLAGS=-v
LDFLAGS=

# Detect OS
UNAME_S := $(shell uname -s)
ifeq ($(UNAME_S),Linux)
    TARGET_OS = linux
else ifeq ($(UNAME_S),Darwin)
    TARGET_OS = darwin
else
    TARGET_OS = windows
endif

all: build

## Build targets

build: desktop cli
	@echo "Built all targets"

desktop:
	@echo "Building desktop application..."
	@echo "Run 'make run-desktop' to start the desktop application"

cli:
	@echo "Building CLI application..."
	@mkdir -p $(BUILD_DIR)
	$(GOCMD) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(CLI_BINARY_NAME)$(EXTENSION) ./cmd/cli
	@echo "CLI binary built: $(BUILD_DIR)/$(CLI_BINARY_NAME)$(EXTENSION)"

## Run targets

run: run-desktop

run-desktop:
	@echo "Starting desktop application..."
	@if [ "$(TARGET_OS)" = "windows" ]; then \
		./$(BUILD_DIR)/bin/$(DESKTOP_BINARY_NAME)$(EXTENSION); \
	else \
		$(GOCMD) run github.com/wailsapp/wails/v2/cmd/wails; \
	fi

run-cli:
	@echo "Running CLI application..."
	@./$(BUILD_DIR)/$(CLI_BINARY_NAME)$(EXTENSION) $(ARGS)

## Install targets

install: install-cli install-desktop
	@echo "Installed all targets"

install-cli:
	@echo "Installing CLI binary..."
	@mkdir -p $(DESTDIR)usr/local/bin
	@cp $(BUILD_DIR)/$(CLI_BINARY_NAME)$(EXTENSION) $(DESTDIR)usr/local/bin/$(BINARY_NAME)$(EXTENSION)
	@echo "CLI binary installed to $(DESTDIR)usr/local/bin/$(BINARY_NAME)$(EXTENSION)"

install-desktop:
	@echo "Installing desktop application..."
	@echo "Desktop installation requires platform-specific steps"

## Test targets

test:
	@echo "Running tests..."
	$(GOCMD) test -v ./...

test-unit:
	@echo "Running unit tests..."
	$(GOCMD) test -v -race -coverprofile=coverage.out ./tests/unit/...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

test-integration:
	@echo "Running integration tests..."
	$(GOCMD) test -v -tags=integration ./tests/integration/...

test-platform:
	@echo "Running platform-specific tests..."
	$(GOCMD) test -v ./tests/unit/pkg/tools/sandbox/sys/... ./tests/unit/pkg/voice/... ./tests/unit/agent/...

test-all: test-unit test-integration test-platform
	@echo "All tests completed"

test-coverage:
	@echo "Running tests with coverage..."
	$(GOCMD) test -v -coverprofile=coverage.out -covermode=atomic ./...
	$(GOCMD) tool cover -html=coverage.out

## CI targets

ci:
	@echo "Running CI pipeline..."
	@bash ./ci/cross-platform-test.sh

ci-short:
	@echo "Running CI pipeline (quick)..."
	@bash ./ci/cross-platform-test.sh --skip-integration --skip-benchmarks

## Lint and security targets

lint:
	@echo "Running linters..."
	@echo "Running go vet..."
	$(GOCMD) vet ./...
	@echo "Checking code formatting..."
	@if [ -n "$$(gofmt -s -l .)" ]; then \
		echo "Error: Found unformatted files:"; \
		gofmt -s -l .; \
		exit 1; \
	fi
	@echo "Linting passed"

security-check:
	@echo "Running security checks..."
	@if command -v govulncheck >/dev/null 2>&1; then \
		govulncheck ./...; \
	else \
		echo "govulncheck not found, skipping vulnerability check"; \
	fi

## Benchmark targets

benchmark:
	@echo "Running benchmarks..."
	$(GOCMD) test -bench=. -benchmem -run=^$ ./tests/benchmarks/... | tee benchmark.txt

## Dependency targets

deps-install:
	@echo "Installing platform dependencies..."
	@bash ./ci/cross-platform-test.sh --skip-unit --skip-integration --skip-platform --skip-build --skip-security

deps-check:
	@echo "Checking dependencies..."
	@bash ./ci/cross-platform-test.sh --skip-unit --skip-integration --skip-platform --skip-build --skip-security --skip-deps

## Clean targets

clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out

clean-all: clean
	@echo "Cleaning all artifacts..."
	@rm -rf dist/

## Help target

help:
	@echo "AgentFramework Makefile"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Build targets:"
	@echo "  all              - Build all targets (default)"
	@echo "  build           - Build all targets"
	@echo "  desktop         - Build desktop application"
	@echo "  cli             - Build CLI application"
	@echo ""
	@echo "Run targets:"
	@echo "  run             - Run desktop application"
	@echo "  run-desktop     - Run desktop application"
	@echo "  run-cli         - Run CLI application"
	@echo ""
	@echo "Install targets:"
	@echo "  install         - Install all targets"
	@echo "  install-cli     - Install CLI binary"
	@echo "  install-desktop - Install desktop application"
	@echo ""
	@echo "Test targets:"
	@echo "  test            - Run all tests"
	@echo "  test-unit       - Run unit tests with coverage"
	@echo "  test-integration - Run integration tests"
	@echo "  test-platform   - Run platform-specific tests"
	@echo "  test-all        - Run all test suites"
	@echo "  test-coverage   - Run tests with coverage report"
	@echo ""
	@echo "CI targets:"
	@echo "  ci              - Run full CI pipeline"
	@echo "  ci-short        - Run CI pipeline (quick, skips integration/benchmarks)"
	@echo ""
	@echo "Lint and security:"
	@echo "  lint            - Run code linters and formatters"
	@echo "  security-check  - Run security vulnerability checks"
	@echo ""
	@echo "Benchmarking:"
	@echo "  benchmark       - Run benchmarks"
	@echo ""
	@echo "Dependency management:"
	@echo "  deps-install    - Install platform-specific dependencies"
	@echo "  deps-check      - Check if all dependencies are installed"
	@echo ""
	@echo "Clean targets:"
	@echo "  clean           - Clean build artifacts"
	@echo "  clean-all       - Clean all artifacts"
	@echo ""
	@echo "Help:"
	@echo "  help            - Show this help message"
	@echo ""
	@echo "Variables:"
	@echo "  BINARY_NAME     - Binary name (default: $(BINARY_NAME))"
	@echo "  BUILD_DIR       - Build directory (default: $(BUILD_DIR))"
	@echo "  DESTDIR         - Installation destination prefix"
	@echo "  ARGS            - Arguments to pass to run-cli"
