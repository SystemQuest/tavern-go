.PHONY: server test clean all help

# Default target
all: help

# Start the test server
server:
	@echo "🚀 Starting test server on http://localhost:5000"
	@echo "Press Ctrl+C to stop"
	@go run server.go

# Run tavern tests
test:
	@echo "🧪 Running Tavern tests..."
	@../../tavern-go test_server.tavern.yaml

# Run tests with verbose output
test-verbose:
	@echo "🧪 Running Tavern tests (verbose)..."
	@../../tavern-go -v test_server.tavern.yaml

# Clean (placeholder for now)
clean:
	@echo "🧹 Cleaning up..."
	@echo "Nothing to clean yet"

# Build the server binary
build:
	@echo "🔨 Building server..."
	@go build -o bin/server server.go
	@echo "✅ Binary created at bin/server"

# Run the built binary
run-binary: build
	@echo "🚀 Starting server from binary..."
	@./bin/server

# Quick test - start server in background, run tests, stop server
quick-test:
	@echo "⚡ Quick test mode..."
	@echo "Starting server in background..."
	@go run server.go > /dev/null 2>&1 & echo $$! > .server.pid
	@sleep 2
	@echo "Running tests..."
	@../../tavern-go test_server.tavern.yaml || true
	@echo "Stopping server..."
	@kill `cat .server.pid` 2>/dev/null || true
	@rm -f .server.pid
	@echo "✅ Quick test complete"

# Show help
help:
	@echo "📖 Available commands:"
	@echo ""
	@echo "  make server       - Start the test server"
	@echo "  make test         - Run Tavern tests (server must be running)"
	@echo "  make test-verbose - Run tests with verbose output"
	@echo "  make build        - Build server binary"
	@echo "  make quick-test   - Auto start server, test, and stop"
	@echo "  make clean        - Clean up temporary files"
	@echo "  make help         - Show this help message"
	@echo ""
	@echo "💡 Typical workflow:"
	@echo "  Terminal 1: make server"
	@echo "  Terminal 2: make test"
	@echo ""
	@echo "⚡ Or use: make quick-test (automated)"
