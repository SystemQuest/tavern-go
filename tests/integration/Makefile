.PHONY: all build server test test-verbose clean help

# Default target
all: help

# Binary name
BINARY=server

# Build the test server
build:
	@echo "🔨 Building integration test server..."
	cd cmd/server && go build -o ../../$(BINARY) server.go
	@echo "✅ Server binary created"

# Start the test server
server: build
	@echo "🚀 Starting integration test server on http://localhost:5000"
	@echo "Press Ctrl+C to stop"
	@./$(BINARY)

# Run tavern tests
test:
	@echo "🧪 Running integration tests..."
	@if [ ! -f ../../bin/tavern ]; then \
		echo "Building tavern CLI..."; \
		cd ../.. && make build; \
	fi
	@echo "Testing response types..."
	@TEST_HOST=http://localhost:5000 SECOND_URL_PART=again ../../bin/tavern test_response_types.tavern.yaml
	@echo "Testing type tokens..."
	@TEST_HOST=http://localhost:5000 SECOND_URL_PART=again ../../bin/tavern test_typetokens.tavern.yaml
	@echo "Testing env var format..."
	@TEST_HOST=http://localhost:5000 SECOND_URL_PART=again ../../bin/tavern test_env_var_format.tavern.yaml
	@echo "Testing file uploads..."
	@TEST_HOST=http://localhost:5000 SECOND_URL_PART=again ../../bin/tavern test_files.tavern.yaml
	@echo "Testing regex validation..."
	@TEST_HOST=http://localhost:5000 SECOND_URL_PART=again ../../bin/tavern test_regex.tavern.yaml
	@echo "Testing anybool type matcher..."
	@TEST_HOST=http://localhost:5000 SECOND_URL_PART=again ../../bin/tavern test_anybool.tavern.yaml
	@echo "Testing type conversions..."
	@TEST_HOST=http://localhost:5000 SECOND_URL_PART=again ../../bin/tavern test_typeconvert.tavern.yaml
	@echo "Testing form encoded data..."
	@TEST_HOST=http://localhost:5000 SECOND_URL_PART=again ../../bin/tavern test_formencoded.tavern.yaml
	@echo "Testing nested variables..."
	@TEST_HOST=http://localhost:5000 SECOND_URL_PART=again ../../bin/tavern test_nested_vars.tavern.yaml
	@echo "Testing only stage keyword..."
	@TEST_HOST=http://localhost:5000 SECOND_URL_PART=again ../../bin/tavern test_only_stage.tavern.yaml
	@echo "Testing skipped stages..."
	@TEST_HOST=http://localhost:5000 SECOND_URL_PART=again ../../bin/tavern test_skipped_stages.tavern.yaml
	@echo "Testing strict key checks..."
	@TEST_HOST=http://localhost:5000 SECOND_URL_PART=again ../../bin/tavern test_strict_key_checks.tavern.yaml
	@echo "Testing multiple status codes..."
	@TEST_HOST=http://localhost:5000 BASE_URL=http://localhost:5000 SECOND_URL_PART=again ../../bin/tavern test_multiple_status_codes.tavern.yaml
	@echo "✅ All tests passed!"

# Run tests with verbose output
test-verbose:
	@echo "🧪 Running integration tests (verbose)..."
	@if [ ! -f ../../bin/tavern ]; then \
		echo "Building tavern CLI..."; \
		cd ../.. && make build; \
	fi
	@echo "Testing response types..."
	@TEST_HOST=http://localhost:5000 SECOND_URL_PART=again ../../bin/tavern -v test_response_types.tavern.yaml
	@echo "Testing type tokens..."
	@TEST_HOST=http://localhost:5000 SECOND_URL_PART=again ../../bin/tavern -v test_typetokens.tavern.yaml
	@echo "Testing env var format..."
	@TEST_HOST=http://localhost:5000 SECOND_URL_PART=again ../../bin/tavern -v test_env_var_format.tavern.yaml
	@echo "Testing file uploads..."
	@TEST_HOST=http://localhost:5000 SECOND_URL_PART=again ../../bin/tavern -v test_files.tavern.yaml
	@echo "Testing regex validation..."
	@TEST_HOST=http://localhost:5000 SECOND_URL_PART=again ../../bin/tavern -v test_regex.tavern.yaml
	@echo "Testing anybool type matcher..."
	@TEST_HOST=http://localhost:5000 SECOND_URL_PART=again ../../bin/tavern -v test_anybool.tavern.yaml
	@echo "Testing type conversions..."
	@TEST_HOST=http://localhost:5000 SECOND_URL_PART=again ../../bin/tavern -v test_typeconvert.tavern.yaml
	@echo "Testing form encoded data..."
	@TEST_HOST=http://localhost:5000 SECOND_URL_PART=again ../../bin/tavern -v test_formencoded.tavern.yaml
	@echo "Testing nested variables..."
	@TEST_HOST=http://localhost:5000 SECOND_URL_PART=again ../../bin/tavern -v test_nested_vars.tavern.yaml
	@echo "Testing only stage keyword..."
	@TEST_HOST=http://localhost:5000 SECOND_URL_PART=again ../../bin/tavern -v test_only_stage.tavern.yaml
	@echo "Testing skipped stages..."
	@TEST_HOST=http://localhost:5000 SECOND_URL_PART=again ../../bin/tavern -v test_skipped_stages.tavern.yaml
	@echo "Testing strict key checks..."
	@TEST_HOST=http://localhost:5000 SECOND_URL_PART=again ../../bin/tavern -v test_strict_key_checks.tavern.yaml
	@echo "Testing multiple status codes..."
	@TEST_HOST=http://localhost:5000 BASE_URL=http://localhost:5000 SECOND_URL_PART=again ../../bin/tavern -v test_multiple_status_codes.tavern.yaml

# Quick test - start server in background, run tests, stop server
quick-test: build
	@echo "⚡ Quick test mode..."
	@echo "Starting server in background..."
	@./$(BINARY) > /dev/null 2>&1 & echo $$! > .server.pid
	@sleep 2
	@echo "Running tests..."
	@$(MAKE) test || true
	@echo "Stopping server..."
	@kill `cat .server.pid` 2>/dev/null || true
	@rm -f .server.pid
	@echo "✅ Quick test complete"

# Clean up
clean:
	@echo "🧹 Cleaning up..."
	@rm -f $(BINARY) .server.pid
	@echo "✅ Cleanup complete"

# Show help
help:
	@echo "📖 Available commands:"
	@echo ""
	@echo "  make build        - Build the test server"
	@echo "  make server       - Start the test server"
	@echo "  make test         - Run integration tests (server must be running)"
	@echo "  make test-verbose - Run tests with verbose output"
	@echo "  make quick-test   - Auto start server, test, and stop"
	@echo "  make clean        - Clean up binaries and temporary files"
	@echo "  make help         - Show this help message"
	@echo ""
	@echo "💡 Typical workflow:"
	@echo "  Terminal 1: make server"
	@echo "  Terminal 2: make test"
	@echo ""
	@echo "⚡ Or use: make quick-test (automated)"
	@echo ""
	@echo "📋 Test files:"
	@echo "  - test_response_types.tavern.yaml   (list/array responses)"
	@echo "  - test_typetokens.tavern.yaml       (!anything marker)"
	@echo "  - test_env_var_format.tavern.yaml   (env vars in includes)"
	@echo "  - test_files.tavern.yaml            (file uploads)"
	@echo "  - test_regex.tavern.yaml            (regex validation)"
	@echo "  - test_anybool.tavern.yaml          (!anybool matcher)"
	@echo "  - test_typeconvert.tavern.yaml      (type conversions)"
	@echo "  - test_formencoded.tavern.yaml      (form encoded data)"
	@echo "  - test_nested_vars.tavern.yaml      (nested variable access)"
	@echo "  - test_only_stage.tavern.yaml       (only keyword isolation)"
	@echo "  - test_skipped_stages.tavern.yaml   (skip keyword)"
	@echo "  - test_strict_key_checks.tavern.yaml (strict mode & _xfail)"
	@echo "  - test_multiple_status_codes.tavern.yaml (multiple status codes)"

