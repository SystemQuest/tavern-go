.PHONY: build test clean install examples lint fmt

BINARY_NAME=tavern
BUILD_DIR=bin
MAIN_PATH=./cmd/tavern

# Version information
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Build flags
LDFLAGS := -ldflags "\
	-X github.com/systemquest/tavern-go/pkg/version.Version=$(VERSION) \
	-X github.com/systemquest/tavern-go/pkg/version.GitCommit=$(COMMIT) \
	-X github.com/systemquest/tavern-go/pkg/version.BuildDate=$(BUILD_DATE)"

# Build the application
build:
	@echo "Building $(BINARY_NAME) $(VERSION)..."
	@mkdir -p $(BUILD_DIR)
	@go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# Install to GOPATH/bin
install:
	@echo "Installing $(BINARY_NAME)..."
	@go install $(MAIN_PATH)
	@echo "Install complete"

# Run tests
test:
	@echo "Running tests..."
	@go test -v -race -cover ./...

# Run tests with coverage
coverage:
	@echo "Running tests with coverage..."
	@go test -v -race -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# Run examples
examples: build
	@echo "Running examples..."
	@$(BUILD_DIR)/$(BINARY_NAME) examples/simple/test_example.tavern.yaml
	@echo "Examples complete"

# Lint code
lint:
	@echo "Linting..."
	@which golangci-lint > /dev/null || (echo "golangci-lint not installed. Run: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest" && exit 1)
	@golangci-lint run ./...

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...
	@gofmt -s -w .

# Tidy dependencies
tidy:
	@echo "Tidying dependencies..."
	@go mod tidy

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
	@echo "Clean complete"

# Build for multiple platforms
build-all:
	@echo "Building for multiple platforms..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)
	GOOS=linux GOARCH=arm64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 $(MAIN_PATH)
	GOOS=darwin GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)
	GOOS=darwin GOARCH=arm64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_PATH)
	GOOS=windows GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PATH)
	@echo "Build complete for all platforms"

# Run development server
dev: build
	@$(BUILD_DIR)/$(BINARY_NAME) --verbose

# Show help
help:
	@echo "Available targets:"
	@echo "  build       - Build the application"
	@echo "  install     - Install to GOPATH/bin"
	@echo "  test        - Run tests"
	@echo "  coverage    - Run tests with coverage"
	@echo "  examples    - Run example tests"
	@echo "  lint        - Lint code"
	@echo "  fmt         - Format code"
	@echo "  tidy        - Tidy dependencies"
	@echo "  clean       - Clean build artifacts"
	@echo "  build-all   - Build for multiple platforms"
	@echo "  help        - Show this help"
