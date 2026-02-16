# AgentFramework Makefile
# Supports both desktop and CLI builds

.PHONY: all build clean test desktop cli run-desktop run-cli install help

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

test-coverage:
	@echo "Running tests with coverage..."
	$(GOCMD) test -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out

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
	@echo "Targets:"
	@echo "  all              - Build all targets (default)"
	@echo "  build           - Build all targets"
	@echo "  desktop         - Build desktop application"
	@echo "  cli             - Build CLI application"
	@echo "  run             - Run desktop application"
	@echo "  run-desktop     - Run desktop application"
	@echo "  run-cli         - Run CLI application"
	@echo "  install         - Install all targets"
	@echo "  install-cli     - Install CLI binary"
	@echo "  install-desktop - Install desktop application"
	@echo "  test            - Run tests"
	@echo "  test-coverage  - Run tests with coverage"
	@echo "  clean           - Clean build artifacts"
	@echo "  clean-all       - Clean all artifacts"
	@echo "  help            - Show this help message"
	@echo ""
	@echo "Variables:"
	@echo "  BINARY_NAME     - Binary name (default: $(BINARY_NAME))"
	@echo "  BUILD_DIR       - Build directory (default: $(BUILD_DIR))"
	@echo "  DESTDIR         - Installation destination prefix"
	@echo "  ARGS            - Arguments to pass to run-cli"
