# Justfile for app-go-api
# Run 'just --list' to see all available commands

# Default recipe to display help information
default:
    @just --list

# ==============================================================================
# Development
# ==============================================================================

# Setup development environment
setup:
    @echo "Setting up development environment..."
    go mod download
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
    go install github.com/swaggo/swag/cmd/swag@latest
    go install github.com/pressly/goose/v3/cmd/goose@latest
    @echo "✅ Development environment ready!"

# Format all Go code
fmt:
    @echo "Formatting code..."
    @gofmt -w -s .
    @go fmt ./...
    @echo "✅ Code formatted successfully!"

# Run linter (matches CI configuration)
lint:
    @echo "Running linting checks..."
    @golangci-lint run --timeout=10m
    @echo "✅ Linting passed!"

# Run linter with auto-fix
lint-fix:
    @echo "Running linting with auto-fix..."
    @golangci-lint run --timeout=10m --fix
    @echo "✅ Auto-fix complete!"

# ==============================================================================
# Testing
# ==============================================================================

# Run all tests with race detection (like CI)
test:
    @echo "Running tests with race detection..."
    @go test -v -race ./...

# Run tests with coverage report
test-coverage:
    @echo "Running tests with coverage..."
    @mkdir -p coverage
    @go test -v -race -coverprofile=coverage/coverage.out -covermode=atomic ./...
    @go tool cover -html=coverage/coverage.out -o coverage/coverage.html
    @echo "✅ Coverage report generated: coverage/coverage.html"
    @./scripts/extract-coverage.sh

# Run tests for a specific package
test-pkg pkg:
    @echo "Running tests for {{pkg}}..."
    @go test -v -race ./{{pkg}}/...

# Run integration tests (requires DATABASE_URL)
test-integration:
    @echo "Running integration tests..."
    RUN_REPOSITORY_INTEGRATION=1 go test -v -race ./internal/repository/...

# Run E2E health check tests
test-e2e url="http://localhost:8080":
    @echo "Running E2E health check tests..."
    @API_URL={{url}} ./tests/e2e/health_check.sh

# ==============================================================================
# Building
# ==============================================================================

# Build the server binary
build:
    @echo "Building server binary..."
    @mkdir -p bin
    @go build -o bin/server cmd/server/main.go
    @echo "✅ Binary built: bin/server"

# Run the API (built version)
run:
    @echo "Building and running the API..."
    @go build -o bin/server cmd/server/main.go
    @./bin/server

# Run the API with hot reload (development)
dev:
    @echo "Starting API with hot reload..."
    @air

# ==============================================================================
# Docker
# ==============================================================================

# Build Docker image
docker-build tag="latest":
    @echo "Building Docker image..."
    @docker build -t app-go-api:{{tag}} .
    @echo "✅ Image built: app-go-api:{{tag}}"

# Run Docker container locally
docker-run:
    @echo "Running Docker container..."
    @docker run --rm -p 8080:8080 \
      -e DATABASE_URL=${DATABASE_URL:-postgres://postgres:postgres@host.docker.internal:5432/app?sslmode=disable} \
      -e REDIS_URL=${REDIS_URL:-redis://host.docker.internal:6379} \
      app-go-api:latest

# ==============================================================================
# Database
# ==============================================================================

# Generate Swagger documentation
swagger:
    @echo "Generating Swagger documentation..."
    @swag init -g cmd/server/main.go -o docs
    @echo "✅ Swagger documentation generated!"

# Run database migrations up
migrate-up:
    @echo "Running database migrations..."
    @goose -dir internal/db/migrations postgres "$(DATABASE_URL)" up
    @echo "✅ Migrations applied!"

# Run database migrations down
migrate-down:
    @echo "Rolling back database migrations..."
    @goose -dir internal/db/migrations postgres "$(DATABASE_URL)" down

# Create new migration file
migrate-create name:
    @echo "Creating migration: {{name}}..."
    @goose -dir internal/db/migrations create {{name}} sql

# ==============================================================================
# Quality Checks
# ==============================================================================

# Run all quality checks (like CI)
ci: fmt lint test
    @echo "✅ All quality checks passed!"

# Security scan with Trivy
security-scan:
    @echo "Running security scan..."
    @just docker-build scan
    @trivy image --severity HIGH,CRITICAL app-go-api:scan

# ==============================================================================
# Dependencies
# ==============================================================================

# Update dependencies
deps-update:
    @echo "Updating dependencies..."
    @go get -u ./...
    @go mod tidy
    @echo "✅ Dependencies updated!"

# Verify dependencies
deps-verify:
    @echo "Verifying dependencies..."
    @go mod verify
    @go mod tidy
    @echo "✅ Dependencies verified!"

# ==============================================================================
# Cleanup
# ==============================================================================

# Clean build artifacts and caches
clean:
    @echo "Cleaning build artifacts..."
    @rm -rf bin/ coverage/ *.out
    @go clean -testcache
    @echo "✅ Cleaned!"

# ==============================================================================
# Help
# ==============================================================================

# Show this help message
help:
    @echo "Available commands:"
    @just --list
