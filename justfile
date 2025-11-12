# Default recipe to display help information
default:
    @just --list

# Run the API (built version)
run:
    @echo "Building and running the API..."
    @go build -o bin/api cmd/server/main.go
    @./bin/api

# Run the API with hot reload (development)
dev:
    @echo "Starting API with hot reload..."
    @air

# Check for linting issues
lint:
    @echo "Running linting checks..."
    @golangci-lint run ./...

# Format the code
fmt:
    @echo "Formatting code..."
    @go fmt ./...
    @echo "Code formatted successfully!"

# Generate Swagger documentation
swagger:
    @echo "Generating Swagger documentation..."
    @swag init -g cmd/server/main.go -o docs
    @echo "Swagger documentation generated!"

# Run tests
test:
    @echo "Running tests..."
    @go test ./...

# Run database migrations up
migrate-up:
    @echo "Running database migrations..."
    @goose -dir internal/db/migrations postgres "$(DATABASE_DSN)" up

# Run database migrations down
migrate-down:
    @echo "Rolling back database migrations..."
    @goose -dir internal/db/migrations postgres "$(DATABASE_DSN)" down

# Build the API binary
build:
    @echo "Building API binary..."
    @go build -o bin/api cmd/server/main.go
    @echo "Binary built at: bin/api"

# Clean build artifacts
clean:
    @echo "Cleaning build artifacts..."
    @rm -rf bin/
    @echo "Cleaned!"
