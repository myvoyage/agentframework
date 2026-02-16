#!/bin/bash
# AgentFramework CLI Test Script
# Tests basic functionality of the CLI interface

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test counter
TESTS_PASSED=0
TESTS_FAILED=0

# Function to print test results
print_result() {
    if [ $1 -eq 0 ]; then
        echo -e "${GREEN}✓ PASS${NC}: $2"
        ((TESTS_PASSED++))
    else
        echo -e "${RED}✗ FAIL${NC}: $2"
        ((TESTS_FAILED++))
    fi
}

# Function to run a test
run_test() {
    local test_name="$1"
    local command="$2"

    echo -e "${YELLOW}Running: ${test_name}${NC}"
    if eval "$command" > /dev/null 2>&1; then
        print_result 0 "${test_name}"
    else
        print_result 1 "${test_name}"
    fi
}

echo "========================================"
echo "AgentFramework CLI Tests"
echo "========================================"
echo ""

# Check if CLI binary exists
if [ ! -f "./build/af" ] && [ ! -f "./build/af.exe" ]; then
    echo -e "${RED}Error: CLI binary not found${NC}"
    echo "Please run 'make cli' first"
    exit 1
fi

# Determine binary name
BINARY="./build/af"
if [ -f "./build/af.exe" ]; then
    BINARY="./build/af.exe"
fi

# Test 1: Help command
run_test "Help command" "$BINARY --help"

# Test 2: Version command
run_test "Version command" "$BINARY version"

# Test 3: Config get command
run_test "Config get" "$BINARY config get"

# Test 4: Workflow list command
run_test "Workflow list" "$BINARY workflow list"

# Test 5: Skill list command
run_test "Skill list" "$BINARY skill list"

# Test 6: Skill info command
run_test "Skill info" "$BINARY skill info"

# Test 7: File ls command
run_test "File ls" "$BINARY file ls"

# Test 8: JSON output format
run_test "JSON output" "$BINARY -o json workflow list"

# Test 9: Custom config flag
run_test "Custom config flag" "$BINARY -c host.yaml workflow list"

# Test 10: Verbose flag
run_test "Verbose flag" "$BINARY -v workflow list"

echo ""
echo "========================================"
echo "Test Results"
echo "========================================"
echo -e "${GREEN}Passed: ${TESTS_PASSED}${NC}"
echo -e "${RED}Failed: ${TESTS_FAILED}${NC}"

if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "${GREEN}All tests passed!${NC}"
    exit 0
else
    echo -e "${RED}Some tests failed${NC}"
    exit 1
fi
