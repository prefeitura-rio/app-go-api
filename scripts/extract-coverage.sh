#!/bin/bash
set -euo pipefail

# Check if coverage.out exists
if [ ! -f coverage.out ]; then
    echo "Error: coverage.out not found. Run tests first with: go test -coverprofile=coverage.out ./..." >&2
    exit 1
fi

# Extract coverage percentage (only output the number to stdout)
coverage=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')

# Output only the number (for capture in CI)
echo "${coverage}"

# Export to GitHub Actions if running in CI
if [ -n "${GITHUB_ENV:-}" ]; then
    echo "COVERAGE=${coverage}" >> "$GITHUB_ENV"
fi

exit 0
