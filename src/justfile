# Golang REST API Justfile
# Run 'just --list' to see all available commands

BINARY_NAME := "kronox-api"
MAIN_PATH := "./cmd/kronox-api"
DOCKER_IMAGE := "kronox-api"
DOCKER_TAG := "latest"

# Default recipe (runs when you just type 'just')
default:
    @just --list

# Development commands
# Run the application in development mode with live reload (requires air)
dev:
    #!/usr/bin/env bash
    if ! command -v air &> /dev/null; then
        echo "Air not found. Installing..."
        go install github.com/cosmtrek/air@latest
    fi
    air -c .air.toml

# Build the application
build:
    @echo "Building {{BINARY_NAME}}..."
    go build -o bin/{{BINARY_NAME}} {{MAIN_PATH}}

# Run the application
run: build
    @echo "Starting {{BINARY_NAME}}..."
    ./bin/{{BINARY_NAME}}

# Clean up build artifacts
clean:
    @echo "Cleaning..."
    rm -rf bin/
    go clean

# Run tests
test:
    @echo "Running tests..."
    go test -v ./...

# Run tests with coverage
test-cover:
    @echo "Running tests with coverage..."
    go test -v -coverprofile=coverage.out ./...
    go tool cover -html=coverage.out -o coverage.html
    @echo "Coverage report generated: coverage.html"

# Run tests with race detection
test-race:
    @echo "Running tests with race detection..."
    go test -race -v ./...

# Run benchmarks
bench:
    @echo "Running benchmarks..."
    go test -bench=. -benchmem ./...

# Code quality commands
# Format code
fmt:
    @echo "Formatting code..."
    go fmt ./...

# Run linter
lint:
    @echo "Running linter..."
    golangci-lint run

# Fix linting issues
lint-fix:
    @echo "Fixing linting issues..."
    golangci-lint run --fix

# Code checks
vet:
    @echo "Vetting code..."
    go vet ./...

# Run all quality checks
check: fmt vet lint test
    @echo "All checks passed!"

# Dependency management
# Tidy dependencies
tidy:
    @echo "Tidying dependencies..."
    go mod tidy

# Update dependencies
update:
    @echo "Updating dependencies..."
    go get -u ./...
    go mod tidy

# Vendor dependencies
vendor:
    @echo "Vendoring dependencies..."
    go mod vendor

# Docker commands
# Build Docker image
docker-build:
    @echo "Building Docker image..."
    docker build -t {{DOCKER_IMAGE}}:{{DOCKER_TAG}} .

# Run Docker container
docker-run: docker-build
    @echo "Running with Docker..."
    docker run -p 5055:5055 {{DOCKER_IMAGE}}:{{DOCKER_TAG}}

# Push Docker image to registry
docker-push: docker-build
    @echo "Pushing Docker image..."
    docker push {{DOCKER_IMAGE}}:{{DOCKER_TAG}}

# Development environment setup
# Install development tools
install-tools:
    @echo "Installing development tools..."
    go install github.com/air-verse/air@latest
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Initialize project (run once for new projects)
init: install-tools tidy
    @echo "Project initialized!"

# Production commands
# Build for production (with optimizations)
build-prod:
    @echo "Building for production..."
    CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o bin/{{BINARY_NAME}} {{MAIN_PATH}}

# Release build with version
build-release version:
    @echo "Building release {{version}}..."
    CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s -X main.version={{version}}" -o bin/{{BINARY_NAME}}-{{version}} {{MAIN_PATH}}

# Security commands
# Check for vulnerabilities
security-check:
    @echo "Checking for security vulnerabilities..."
    govulncheck ./...

# Generate API documentation (if using swag)
docs:
    @echo "Generating API documentation..."
    swag init -g {{MAIN_PATH}}/main.go

# Load testing (using hey - install with: go install github.com/rakyll/hey@latest)
load-test url="http://localhost:5055/health":
    @echo "Running load test against {{url}}..."
    hey -n 1000 -c 10 {{url}}

# Utility commands
# Show project structure
tree:
    tree -I 'bin|vendor|.git'

# Count lines of code
loc:
    @echo "Lines of code:"
    find . -name "*.go" -not -path "./vendor/*" | xargs wc -l | tail -1
