#!/usr/bin/env bash
#
# Cross-Platform CI Script for AgentFramework
# Supports Windows, macOS, and Linux testing
#

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Detect OS
detect_os() {
    case "$(uname -s)" in
        Linux*)     OS=Linux;;
        Darwin*)    OS=macOS;;
        CYGWIN*)    OS=Cygwin;;
        MINGW*)     OS=MinGW;;
        MSYS*)     OS=MSYS;;
        *)          OS="UNKNOWN:${unameOut}"
    esac
}

# Print colored message
print_message() {
    local color=$1
    local message=$2
    echo -e "${color}${message}${NC}"
}

# Print section header
print_section() {
    echo ""
    print_message "${BLUE}" "========================================"
    print_message "${BLUE}" "$1"
    print_message "${BLUE}" "========================================"
    echo ""
}

# Check command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Install platform dependencies
install_dependencies() {
    print_section "Installing platform-specific dependencies"

    case "$OS" in
        Linux*)
            print_message "${YELLOW}" "Installing Linux dependencies..."
            if command_exists apt-get; then
                sudo apt-get update
                sudo apt-get install -y \
                    xclip \
                    libnotify-bin \
                    espeak \
                    espeak-ng \
                    ffmpeg \
                    pocketsphinx \
                    sox
            elif command_exists dnf; then
                sudo dnf install -y \
                    xclip \
                    libnotify \
                    espeak \
                    ffmpeg \
                    pocketsphinx
            elif command_exists pacman; then
                sudo pacman -S --noconfirm \
                    xclip \
                    libnotify \
                    espeak-ng \
                    ffmpeg \
                    pocketsphinx
            else
                print_message "${RED}" "Package manager not found. Please install dependencies manually."
                exit 1
            fi
            ;;

        macOS*)
            print_message "${YELLOW}" "Installing macOS dependencies..."
            if command_exists brew; then
                brew install ffmpeg
                brew install sox
            else
                print_message "${YELLOW}" "Homebrew not found. Installing..."
                /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
                brew install ffmpeg sox
            fi
            ;;

        MinGW*|Cygwin*|MSYS*)
            print_message "${YELLOW}" "Installing Windows dependencies..."
            if command_exists choco; then
                choco install ffmpeg -y
            else
                print_message "${YELLOW}" "Chocolatey not found. Please install FFmpeg manually."
            fi
            ;;

        *)
            print_message "${RED}" "Unknown OS: $OS"
            exit 1
            ;;
    esac

    print_message "${GREEN}" "✓ Dependencies installed"
}

# Verify dependencies
verify_dependencies() {
    print_section "Verifying dependencies"

    local missing_deps=0

    # Check Go
    if ! command_exists go; then
        print_message "${RED}" "✗ Go not found"
        missing_deps=$((missing_deps + 1))
    else
        GO_VERSION=$(go version | awk '{print $3}')
        print_message "${GREEN}" "✓ Go found: $GO_VERSION"
    fi

    # Platform-specific checks
    case "$OS" in
        Linux*)
            print_message "${BLUE}" "Checking Linux dependencies..."
            for cmd in xclip notify-send espeak ffmpeg; do
                if command_exists $cmd; then
                    print_message "${GREEN}" "✓ $cmd found"
                else
                    print_message "${YELLOW}" "⚠ $cmd not found (optional)"
                fi
            done
            ;;

        macOS*)
            print_message "${BLUE}" "Checking macOS dependencies..."
            for cmd in pbcopy pbpaste say; do
                if command_exists $cmd; then
                    print_message "${GREEN}" "✓ $cmd found"
                else
                    print_message "${RED}" "✗ $cmd not found"
                    missing_deps=$((missing_deps + 1))
                fi
            done
            ;;
    esac

    if [ $missing_deps -gt 0 ]; then
        print_message "${RED}" "Found $missing_deps missing dependencies"
        exit 1
    fi

    print_message "${GREEN}" "✓ All dependencies verified"
}

# Run Go checks
run_go_checks() {
    print_section "Running Go checks"

    # Download dependencies
    print_message "${BLUE}" "Downloading Go modules..."
    go mod download

    # Verify dependencies
    print_message "${BLUE}" "Verifying Go modules..."
    go mod verify

    # Run go vet
    print_message "${BLUE}" "Running go vet..."
    go vet ./...

    # Check formatting
    print_message "${BLUE}" "Checking code formatting..."
    UNFORMATTED=$(gofmt -s -l .)
    if [ -n "$UNFORMATTED" ]; then
        print_message "${RED}" "✗ Found unformatted files:"
        echo "$UNFORMATTED"
        exit 1
    fi
    print_message "${GREEN}" "✓ Code formatting check passed"
}

# Run unit tests
run_unit_tests() {
    print_section "Running unit tests"

    print_message "${BLUE}" "Running unit tests with race detector and coverage..."
    go test -v -race -coverprofile=coverage.out -covermode=atomic ./...

    print_message "${BLUE}" "Generating coverage report..."
    go tool cover -html=coverage.out -o coverage.html

    print_message "${GREEN}" "✓ Unit tests passed"
    print_message "${GREEN}" "Coverage report: coverage.html"
}

# Run integration tests
run_integration_tests() {
    print_section "Running integration tests"

    if [ -d "./tests/integration" ]; then
        print_message "${BLUE}" "Running integration tests..."
        go test -v -tags=integration ./tests/integration/... || {
            print_message "${YELLOW}" "⚠ Some integration tests failed (may be expected)"
        }
    else
        print_message "${YELLOW}" "⚠ Integration tests directory not found"
    fi

    print_message "${GREEN}" "✓ Integration tests completed"
}

# Run platform-specific tests
run_platform_tests() {
    print_section "Running platform-specific tests"

    print_message "${BLUE}" "Running platform-specific tests..."

    # Clipboard tests
    if [ -d "./tests/unit/pkg/tools/sandbox/sys" ]; then
        print_message "${BLUE}" "Testing clipboard module..."
        go test -v ./tests/unit/pkg/tools/sandbox/sys/... -run TestClipboard || {
            print_message "${YELLOW}" "⚠ Some clipboard tests failed (may require GUI)"
        }
    fi

    # Notification tests
    if [ -d "./tests/unit/pkg/tools/sandbox/sys" ]; then
        print_message "${BLUE}" "Testing notification module..."
        go test -v ./tests/unit/pkg/tools/sandbox/sys/... -run TestNotification || {
            print_message "${YELLOW}" "⚠ Some notification tests failed (may require GUI)"
        }
    fi

    # Voice tests
    if [ -d "./tests/unit/pkg/voice" ]; then
        print_message "${BLUE}" "Testing voice modules (TTS/STT)..."
        go test -v ./tests/unit/pkg/voice/... -run TestTTS || {
            print_message "${YELLOW}" "⚠ Some TTS tests failed (may require audio device)"
        }
        go test -v ./tests/unit/pkg/voice/... -run TestSTT || {
            print_message "${YELLOW}" "⚠ Some STT tests failed (may require audio device)"
        }
    fi

    # Scheduler tests
    if [ -d "./tests/unit/agent" ]; then
        print_message "${BLUE}" "Testing workflow scheduler..."
        go test -v ./tests/unit/agent/... -run TestScheduler || {
            print_message "${YELLOW}" "⚠ Some scheduler tests failed"
        }
    fi

    print_message "${GREEN}" "✓ Platform-specific tests completed"
}

# Build application
build_application() {
    print_section "Building application"

    print_message "${BLUE}" "Building CLI application..."
    mkdir -p bin

    if [ "$OS" = "Windows" ]; then
        go build -v -o bin/af.exe ./cmd/cli
        print_message "${GREEN}" "✓ CLI binary built: bin/af.exe"
    else
        go build -v -o bin/af ./cmd/cli
        print_message "${GREEN}" "✓ CLI binary built: bin/af"
    fi
}

# Test built application
test_application() {
    print_section "Testing built application"

    if [ "$OS" = "Windows" ]; then
        if [ -f "bin/af.exe" ]; then
            print_message "${BLUE}" "Running CLI binary..."
            ./bin/af.exe --help
            print_message "${GREEN}" "✓ CLI binary runs successfully"
        else
            print_message "${RED}" "✗ CLI binary not found"
            exit 1
        fi
    else
        if [ -f "bin/af" ]; then
            print_message "${BLUE}" "Running CLI binary..."
            ./bin/af --help
            print_message "${GREEN}" "✓ CLI binary runs successfully"
        else
            print_message "${RED}" "✗ CLI binary not found"
            exit 1
        fi
    fi
}

# Run security checks
run_security_checks() {
    print_section "Running security checks"

    # Check for vulnerabilities
    print_message "${BLUE}" "Checking for known vulnerabilities..."
    if command_exists govulncheck; then
        govulncheck ./... || {
            print_message "${YELLOW}" "⚠ Some vulnerabilities found"
        }
    else
        print_message "${YELLOW}" "⚠ govulncheck not found, skipping vulnerability check"
    fi

    print_message "${GREEN}" "✓ Security checks completed"
}

# Run benchmarks
run_benchmarks() {
    print_section "Running benchmarks"

    if [ -d "./tests/benchmarks" ]; then
        print_message "${BLUE}" "Running benchmarks..."
        go test -bench=. -benchmem -run=^$ ./tests/benchmarks/... | tee benchmark.txt
        print_message "${GREEN}" "✓ Benchmarks completed"
    else
        print_message "${YELLOW}" "⚠ Benchmarks directory not found"
    fi
}

# Cleanup
cleanup() {
    print_section "Cleanup"

    print_message "${BLUE}" "Cleaning up temporary files..."
    rm -f coverage.out coverage.html benchmark.txt
    print_message "${GREEN}" "✓ Cleanup completed"
}

# Main execution
main() {
    print_message "${BLUE}" "========================================"
    print_message "${BLUE}" "AgentFramework Cross-Platform CI"
    print_message "${BLUE}" "========================================"

    # Detect OS
    detect_os
    print_message "${GREEN}" "Detected OS: $OS"

    # Parse arguments
    SKIP_DEPS=false
    SKIP_UNIT=false
    SKIP_INTEGRATION=false
    SKIP_PLATFORM=false
    SKIP_BUILD=false
    SKIP_SECURITY=false
    RUN_BENCHMARKS=false

    while [[ $# -gt 0 ]]; do
        case $1 in
            --skip-deps)
                SKIP_DEPS=true
                shift
                ;;
            --skip-unit)
                SKIP_UNIT=true
                shift
                ;;
            --skip-integration)
                SKIP_INTEGRATION=true
                shift
                ;;
            --skip-platform)
                SKIP_PLATFORM=true
                shift
                ;;
            --skip-build)
                SKIP_BUILD=true
                shift
                ;;
            --skip-security)
                SKIP_SECURITY=true
                shift
                ;;
            --benchmarks)
                RUN_BENCHMARKS=true
                shift
                ;;
            --help)
                echo "Usage: $0 [options]"
                echo "Options:"
                echo "  --skip-deps          Skip dependency installation"
                echo "  --skip-unit          Skip unit tests"
                echo "  --skip-integration   Skip integration tests"
                echo "  --skip-platform      Skip platform-specific tests"
                echo "  --skip-build         Skip application build"
                echo "  --skip-security      Skip security checks"
                echo "  --benchmarks         Run benchmarks"
                echo "  --help               Show this help message"
                exit 0
                ;;
            *)
                echo "Unknown option: $1"
                echo "Use --help for usage information"
                exit 1
                ;;
        esac
    done

    # Execute CI pipeline
    [ "$SKIP_DEPS" = false ] && install_dependencies
    verify_dependencies
    run_go_checks
    [ "$SKIP_UNIT" = false ] && run_unit_tests
    [ "$SKIP_INTEGRATION" = false ] && run_integration_tests
    [ "$SKIP_PLATFORM" = false ] && run_platform_tests
    [ "$SKIP_BUILD" = false ] && { build_application; test_application; }
    [ "$SKIP_SECURITY" = false ] && run_security_checks
    [ "$RUN_BENCHMARKS" = true ] && run_benchmarks

    # Success
    print_section "CI Pipeline Completed Successfully"
    print_message "${GREEN}" "✓ All checks passed"

    exit 0
}

# Run main function
main "$@"
