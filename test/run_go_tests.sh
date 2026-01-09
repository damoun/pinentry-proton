#!/usr/bin/env bash
# Run Go tests (unit and integration)

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "${SCRIPT_DIR}")"

cd "${PROJECT_ROOT}"

echo "Running Go unit tests..."
go test -v ./internal/... || exit 1

echo ""
echo "Running integration tests..."
go test -v ./test/... || exit 1

echo ""
echo "Running tests with race detector..."
go test -race ./internal/... ./test/... || exit 1

echo ""
echo "All Go tests passed!"
