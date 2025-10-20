.PHONY: server test clean deps

# Start the test server
server:
	@echo "Starting remember example server on :5000..."
	@go run server.go

# Run tavern tests
test:
	@echo "Running tavern tests..."
	@tavern-go test test_server.tavern.yaml -v

# Install dependencies
deps:
	@echo "Installing dependencies..."
	@go get github.com/gorilla/sessions

# Clean up
clean:
	@echo "Cleaning up..."
	@rm -f cookies.txt

# Run server and tests (in separate terminals)
help:
	@echo "Remember Cookie Example - Makefile"
	@echo ""
	@echo "Usage:"
	@echo "  make server    - Start the test server on :5000"
	@echo "  make test      - Run tavern tests (requires server running)"
	@echo "  make deps      - Install Go dependencies"
	@echo "  make clean     - Clean up temporary files"
	@echo ""
	@echo "Example workflow:"
	@echo "  Terminal 1: make server"
	@echo "  Terminal 2: make test"
