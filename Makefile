.PHONY: help install-tools lint lint-fix lint-verbose test coverage build clean all build-benchmark benchmark benchmark-collection benchmark-processing benchmark-concurrency benchmark-format

# Default target shows help
.DEFAULT_GOAL := help

# All target runs full workflow
all: lint test build

# Help target
help:
	@echo "Available targets:"
	@echo "  install-tools  - Install required linting and development tools"
	@echo "  lint          - Run all linters"
	@echo "  lint-fix      - Run linters with auto-fix enabled"
	@echo "  lint-verbose  - Run linters with verbose output"
	@echo "  test          - Run tests"
	@echo "  coverage      - Run tests with coverage"
	@echo "  build         - Build the application"
	@echo "  clean         - Clean build artifacts"
	@echo "  all           - Run lint, test, and build"
	@echo ""
	@echo "Benchmark targets:"
	@echo "  build-benchmark        - Build the benchmark binary"
	@echo "  benchmark              - Run all benchmarks"
	@echo "  benchmark-collection   - Run file collection benchmarks"
	@echo "  benchmark-processing   - Run file processing benchmarks"
	@echo "  benchmark-concurrency  - Run concurrency benchmarks"
	@echo "  benchmark-format       - Run format benchmarks"
	@echo ""
	@echo "Run 'make <target>' to execute a specific target."

# Install required tools
install-tools:
	@echo "Installing golangci-lint..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "Installing gofumpt..."
	@go install mvdan.cc/gofumpt@latest
	@echo "Installing goimports..."
	@go install golang.org/x/tools/cmd/goimports@latest
	@echo "Installing staticcheck..."
	@go install honnef.co/go/tools/cmd/staticcheck@latest
	@echo "Installing gosec..."
	@go install github.com/securego/gosec/v2/cmd/gosec@latest
	@echo "Installing gocyclo..."
	@go install github.com/fzipp/gocyclo/cmd/gocyclo@latest
	@echo "All tools installed successfully!"

# Run linters
lint:
	@echo "Running golangci-lint..."
	@golangci-lint run ./...

# Run linters with auto-fix
lint-fix:
	@echo "Running gofumpt..."
	@gofumpt -l -w .
	@echo "Running goimports..."
	@goimports -w -local github.com/ivuorinen/gibidify .
	@echo "Running go fmt..."
	@go fmt ./...
	@echo "Running go mod tidy..."
	@go mod tidy
	@echo "Running golangci-lint with --fix..."
	@golangci-lint run --fix ./...
	@echo "Auto-fix completed. Running final lint check..."
	@golangci-lint run ./...

# Run linters with verbose output
lint-verbose:
	@golangci-lint run -v ./...

# Run tests
test:
	@echo "Running tests..."
	@go test -race -v ./...

# Run tests with coverage
coverage:
	@echo "Running tests with coverage..."
	@go test -race -coverprofile=coverage.out -covermode=atomic ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Build the application
build:
	@echo "Building gibidify..."
	@go build -ldflags="-s -w" -o gibidify .
	@echo "Build complete: ./gibidify"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -f gibidify gibidify-benchmark
	@rm -f coverage.out coverage.html
	@echo "Clean complete"

# CI-specific targets
.PHONY: ci-lint ci-test

ci-lint:
	@golangci-lint run --out-format=github-actions ./...

ci-test:
	@go test -race -coverprofile=coverage.out -json ./... > test-results.json

# Build benchmark binary
build-benchmark:
	@echo "Building gibidify-benchmark..."
	@go build -ldflags="-s -w" -o gibidify-benchmark ./cmd/benchmark
	@echo "Build complete: ./gibidify-benchmark"

# Run benchmarks
benchmark: build-benchmark
	@echo "Running all benchmarks..."
	@./gibidify-benchmark -type=all

# Run specific benchmark types
benchmark-collection: build-benchmark
	@echo "Running file collection benchmarks..."
	@./gibidify-benchmark -type=collection

benchmark-processing: build-benchmark
	@echo "Running file processing benchmarks..."
	@./gibidify-benchmark -type=processing

benchmark-concurrency: build-benchmark
	@echo "Running concurrency benchmarks..."
	@./gibidify-benchmark -type=concurrency

benchmark-format: build-benchmark
	@echo "Running format benchmarks..."
	@./gibidify-benchmark -type=format