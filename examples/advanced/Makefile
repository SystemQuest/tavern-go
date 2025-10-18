.PHONY: deps server test clean db-init quick-test help all docker-build docker-up docker-down docker-test docker-logs

# Default target
all: help

# Install Go dependencies
deps:
	@echo "📦 Installing dependencies..."
	@go get github.com/golang-jwt/jwt/v5
	@go get github.com/mattn/go-sqlite3
	@echo "✅ Dependencies installed"

# Initialize/reset the database
db-init:
	@echo "🗄️  Initializing database..."
	@rm -f numbers.db
	@echo "✅ Database reset"

# Start the test server
server: deps
	@echo "🚀 Starting advanced test server on http://localhost:5000"
	@echo "Press Ctrl+C to stop"
	@go run server.go jwt_validator.go

# Run tavern tests
test:
	@echo "🧪 Running Tavern tests..."
	@../../bin/tavern test_advanced.tavern.yaml

# Run tests with verbose output
test-verbose:
	@echo "🧪 Running Tavern tests (verbose)..."
	@../../bin/tavern -v test_advanced.tavern.yaml

# Clean up generated files
clean:
	@echo "🧹 Cleaning up..."
	@rm -f numbers.db
	@rm -f .server.pid
	@rm -f bin/server
	@echo "✅ Cleanup complete"

# Build the server binary
build: deps
	@echo "🔨 Building server..."
	@mkdir -p bin
	@go build -o bin/server server.go jwt_validator.go
	@echo "✅ Binary created at bin/server"

# Run the built binary
run-binary: build
	@echo "🚀 Starting server from binary..."
	@./bin/server

# Quick test - start server in background, run tests, stop server
quick-test: deps
	@echo "⚡ Quick test mode..."
	@echo "Cleaning database..."
	@rm -f numbers.db
	@echo "Starting server in background..."
	@go run server.go jwt_validator.go > /dev/null 2>&1 & echo $$! > .server.pid
	@sleep 3
	@echo "Running tests..."
	@../../bin/tavern test_advanced.tavern.yaml || (kill `cat .server.pid` 2>/dev/null && rm -f .server.pid && false)
	@echo "Stopping server..."
	@kill `cat .server.pid` 2>/dev/null || true
	@rm -f .server.pid
	@echo "✅ Quick test complete"

# Test login endpoint manually
test-login:
	@echo "🔐 Testing login endpoint..."
	@curl -X POST http://localhost:5000/login \
		-H 'Content-Type: application/json' \
		-d '{"user":"test-user","password":"correct-password"}' \
		| python3 -m json.tool || cat
	@echo ""

# Docker commands
docker-build:
	@echo "🐳 Building Docker image..."
	@docker compose build
	@echo "✅ Docker image built"

docker-up:
	@echo "🐳 Starting server with Docker Compose..."
	@docker compose up -d
	@echo "⏳ Waiting for server to be healthy..."
	@sleep 5
	@docker compose ps
	@echo "✅ Server is running at http://localhost:5000"

docker-down:
	@echo "🐳 Stopping Docker containers..."
	@docker compose down
	@echo "✅ Containers stopped"

docker-logs:
	@echo "📋 Docker container logs:"
	@docker compose logs -f

docker-test: docker-build docker-up
	@echo "🧪 Running Tavern tests against Docker container..."
	@sleep 3
	@../../bin/tavern test_advanced.tavern.yaml || (make docker-down && false)
	@make docker-down
	@echo "✅ Docker test complete"

# Show help
help:
	@echo "📖 Available commands:"
	@echo ""
	@echo "  make deps         - Install Go dependencies"
	@echo "  make server       - Start the test server"
	@echo "  make test         - Run Tavern tests (server must be running)"
	@echo "  make test-verbose - Run tests with verbose output"
	@echo "  make db-init      - Reset the database"
	@echo "  make build        - Build server binary"
	@echo "  make quick-test   - Auto start server, test, and stop"
	@echo "  make test-login   - Test login endpoint manually"
	@echo "  make clean        - Clean up temporary files"
	@echo ""
	@echo "🐳 Docker commands:"
	@echo "  make docker-build - Build Docker image"
	@echo "  make docker-up    - Start server in Docker"
	@echo "  make docker-down  - Stop Docker containers"
	@echo "  make docker-logs  - View Docker container logs"
	@echo "  make docker-test  - Build, start, test, and stop (automated)"
	@echo "  make help         - Show this help message"
	@echo ""
	@echo "💡 Typical workflow:"
	@echo "  1. make deps      - First time setup"
	@echo "  2. Terminal 1: make server"
	@echo "  3. Terminal 2: make test"
	@echo ""
	@echo "⚡ Or use: make quick-test (automated)"
	@echo ""
	@echo "🔐 Features:"
	@echo "  - JWT Authentication"
	@echo "  - SQLite Database"
	@echo "  - Multi-stage Tests"
	@echo "  - Variable Passing"
	@echo "  - YAML Anchors & Includes"
