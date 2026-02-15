#!/usr/bin/env bash
set -euo pipefail

echo "=== Stage7 CI Demo: Production End-to-End (Stage 7) ==="

# Environment defaults
BRIDGE_PORT=${BRIDGE_PORT:-8080}
BUILD_CMD="go build -tags eino -o bin/stage7 ./..."
TEST_CMD="go test ./..."

echo "[1/4] Building Stage 7 artifacts..."
eval "$BUILD_CMD"

echo "[2/4] Running unit/integration tests..."
eval "$TEST_CMD"

echo "[3/4] Optional: Launch HTTP RPC bridge for integration tests (if a binary exists)"
BRIDGE_BIN="./bin/stage7"
if [ -x "$BRIDGE_BIN" ]; then
  "$BRIDGE_BIN" &
  BRIDGE_PID=$!
  echo "Bridge started with PID $BRIDGE_PID"
  sleep 2
  if command -v curl >/dev/null 2>&1; then
    echo "Pinging bridge endpoints..."
    curl -sSf http://localhost:"$BRIDGE_PORT"/health >/dev/null && echo 'health: PASS' || echo 'health: FAIL'
    curl -sSf http://localhost:"$BRIDGE_PORT"/invoke_tool || true
    curl -sSf http://localhost:"$BRIDGE_PORT"/run_pipeline || true
  fi
  kill "$BRIDGE_PID" || true
  wait "$BRIDGE_PID" 2>/dev/null || true
else
  echo "No Stage7 bridge binary found at $BRIDGE_BIN, skipping bridge run."
fi

echo "[4/4] CI demo completed."
