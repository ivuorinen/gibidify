.PHONY: all help install-tools lint lint-fix test coverage build clean all build-benchmark benchmark benchmark-collection benchmark-processing benchmark-concurrency benchmark-format security security-full vuln-check update-deps check-all dev-setup

# Default target shows help
.DEFAULT_GOAL := help

# All target runs full workflow
all: lint lint-fix test build

# Help target
help:
	@cat scripts/help.txt

# Install required tools
install-tools:
	@./scripts/install-tools.sh

# Run linters
lint:
	@./scripts/lint.sh

# Run linters with auto-fix
lint-fix:
	@./scripts/lint-fix.sh

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
	@revive -config revive.toml -formatter friendly -set_exit_status ./...

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

# Security targets
security:
	@echo "Running comprehensive security scan..."
	@./scripts/security-scan.sh

security-full:
	@echo "Running full security analysis..."
	@./scripts/security-scan.sh
	@echo "Running additional security checks..."
	@gosec -fmt=json -out=security-report.json ./...
	@staticcheck -checks=all ./...

vuln-check:
	@echo "Checking for dependency vulnerabilities..."
	@go install golang.org/x/vuln/cmd/govulncheck@latest
	@govulncheck ./...

# Update dependencies
update-deps:
	@echo "Updating Go dependencies..."
	@./scripts/update-deps.sh
