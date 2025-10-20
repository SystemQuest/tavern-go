.PHONY: server test clean all help deps

# Default target
all: help

# Install dependencies
deps:
	@echo "📦 Installing dependencies..."
	@go get github.com/google/uuid

# Start the test server
server:
	@echo "🚀 Starting !anything example server on http://localhost:5001"
	@echo "Press Ctrl+C to stop"
	@go run server.go

# Run tavern tests (requires server to be running)
test:
	@echo "🧪 Running Tavern tests..."
	@../../bin/tavern test_anything.tavern.yaml

# Run tests with verbose output
test-verbose:
	@echo "🧪 Running Tavern tests (verbose)..."
	@../../bin/tavern -v test_anything.tavern.yaml

# Start server in background and run tests
test-all:
	@echo "🚀 Starting server and running tests..."
	@go run server.go & SERVER_PID=$$!; \
	sleep 2; \
	../../bin/tavern test_anything.tavern.yaml; \
	TEST_RESULT=$$?; \
	kill $$SERVER_PID 2>/dev/null || true; \
	exit $$TEST_RESULT

# Clean (placeholder)
clean:
	@echo "🧹 Cleaning up..."
	@echo "Nothing to clean yet"

# Help
help:
	@echo "Available targets:"
	@echo "  deps          - Install Go dependencies (uuid)"
	@echo "  server        - Start the test server on port 5001"
	@echo "  test          - Run Tavern tests (server must be running separately)"
	@echo "  test-verbose  - Run tests with verbose output"
	@echo "  test-all      - Start server, run tests, and cleanup"
	@echo "  clean         - Clean up (placeholder)"
	@echo "  help          - Show this help message"
	@echo ""
	@echo "Quick start:"
	@echo "  1. Install dependencies:  make deps"
	@echo "  2. Terminal 1:            make server"
	@echo "  3. Terminal 2:            make test"
	@echo "  Or run everything:        make test-all"
