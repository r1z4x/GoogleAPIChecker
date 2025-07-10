# Google API Checker Makefile

.PHONY: build clean test run help

# Binary name
BINARY_NAME=googleapichecker

# Build flags
LDFLAGS=-ldflags "-X main.Version=$(shell git describe --tags --always --dirty)"

# Default target
all: build

# Build the application
build:
	@echo "🔨 Building Google API Checker..."
	go build $(LDFLAGS) -o $(BINARY_NAME) .
	@echo "✅ Build completed!"

# Clean build artifacts
clean:
	@echo "🧹 Cleaning build artifacts..."
	rm -f $(BINARY_NAME)
	rm -f *.json
	@echo "✅ Clean completed!"

# Run tests
test:
	@echo "🧪 Running tests..."
	go test -v ./...
	@echo "✅ Tests completed!"

# Install dependencies
deps:
	@echo "📦 Installing dependencies..."
	go mod tidy
	go mod download
	@echo "✅ Dependencies installed!"

# Run the application (requires API token)
run:
	@echo "🚀 Running Google API Checker..."
	@if [ -z "$(TOKEN)" ]; then \
		echo "❌ Error: TOKEN environment variable is required"; \
		echo "Usage: make run TOKEN=your_google_api_token"; \
		exit 1; \
	fi
	./$(BINARY_NAME) --token $(TOKEN)

# Run with custom parameters
run-custom:
	@echo "🚀 Running Google API Checker with custom parameters..."
	@if [ -z "$(TOKEN)" ]; then \
		echo "❌ Error: TOKEN environment variable is required"; \
		echo "Usage: make run-custom TOKEN=your_token THREADS=20 OUTPUT=my_results.json"; \
		exit 1; \
	fi
	./$(BINARY_NAME) --token $(TOKEN) --threads $(THREADS) --output $(OUTPUT)

# Show help
help:
	@echo "Google API Checker - Available Commands:"
	@echo ""
	@echo "  make build          - Build the application"
	@echo "  make clean          - Clean build artifacts"
	@echo "  make test           - Run tests"
	@echo "  make deps           - Install dependencies"
	@echo "  make run            - Run with API token (TOKEN=your_token)"
	@echo "  make run-custom     - Run with custom parameters"
	@echo "  make help           - Show this help"
	@echo ""
	@echo "Examples:"
	@echo "  make run TOKEN=your_google_api_token"
	@echo "  make run-custom TOKEN=your_token THREADS=20 OUTPUT=results.json"

# Development helpers
dev-setup: deps build
	@echo "✅ Development environment setup completed!"

# Quick test run (simulated)
test-run:
	@echo "🧪 Running test with simulated data..."
	@echo "This will run with simulated API responses for testing purposes"
	./$(BINARY_NAME) --token test-token --threads 5 --output test_results.json

# Format code
fmt:
	@echo "🎨 Formatting code..."
	go fmt ./...
	@echo "✅ Code formatting completed!"

# Lint code
lint:
	@echo "🔍 Linting code..."
	golangci-lint run
	@echo "✅ Linting completed!"

# Install development tools
install-tools:
	@echo "🛠️ Installing development tools..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "✅ Development tools installed!" 