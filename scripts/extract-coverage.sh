#!/bin/bash
set -euo pipefail

# Run tests with coverage
echo "Running tests with coverage..."
go test -race -coverprofile=coverage.out -covermode=atomic ./...

# Extract coverage percentage
coverage=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')

echo "Coverage: ${coverage}%"

# Export to GitHub Actions if running in CI
if [ -n "${GITHUB_ENV:-}" ]; then
    echo "COVERAGE=${coverage}" >> "$GITHUB_ENV"
fi

# Generate HTML report if not in CI
if [ -z "${CI:-}" ]; then
    mkdir -p coverage
    go tool cover -html=coverage.out -o coverage/coverage.html
    echo "HTML report: coverage/coverage.html"
fi

exit 0
