.PHONY: help docker-build docker-up docker-down docker-test docker-logs docker-clean

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

docker-build: ## Build the Docker image
	@echo "Building cookie server Docker image..."
	docker compose build

docker-up: ## Start the cookie server
	@echo "Starting cookie server..."
	docker compose up -d
	@echo "Waiting for server to be healthy..."
	@timeout=30; \
	while [ $$timeout -gt 0 ]; do \
		if docker compose ps | grep -q "healthy"; then \
			echo "Server is healthy!"; \
			break; \
		fi; \
		echo "Waiting... ($$timeout seconds remaining)"; \
		sleep 1; \
		timeout=$$((timeout - 1)); \
	done; \
	if [ $$timeout -eq 0 ]; then \
		echo "Timeout waiting for server to be healthy"; \
		exit 1; \
	fi

docker-down: ## Stop the cookie server
	@echo "Stopping cookie server..."
	docker compose down

docker-test: ## Run tavern tests against the cookie server
	@echo "Running tavern tests..."
	@if ! docker compose ps | grep -q "Up"; then \
		echo "Error: Server is not running. Run 'make docker-up' first."; \
		exit 1; \
	fi
	cd ../.. && ./bin/tavern examples/cookies/test_cookies.tavern.yaml

docker-logs: ## Show cookie server logs
	docker compose logs -f cookie-server

docker-clean: docker-down ## Clean up Docker resources
	@echo "Cleaning up Docker resources..."
	docker compose down -v --rmi local
	@echo "Cleanup complete!"

run-local: ## Run server locally (for development)
	@echo "Starting cookie server locally..."
	go run server.go

test-local: ## Run tests against local server
	@echo "Running tavern tests against localhost..."
	cd ../.. && ./bin/tavern examples/cookies/test_cookies.tavern.yaml
