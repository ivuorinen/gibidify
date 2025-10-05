.PHONY: help install-tools lint lint-fix lint-verbose test coverage build clean all build-benchmark benchmark benchmark-collection benchmark-processing benchmark-concurrency benchmark-format security security-full vuln-check check-all dev-setup

# Default target shows help
.DEFAULT_GOAL := help

# All target runs full workflow
all: lint test build

# Help target  
help:
	@cat scripts/help.txt

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
	@echo "Installing checkmake..."
	@go install github.com/checkmake/checkmake/cmd/checkmake@latest
	@echo "Installing shfmt..."
	@go install mvdan.cc/sh/v3/cmd/shfmt@latest
	@echo "Installing yamllint (Go-based)..."
	@go install github.com/excilsploft/yamllint@latest
	@echo "All tools installed successfully!"

# Run linters
lint:
	@./scripts/lint.sh

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
	@echo "Running shfmt formatting..."
	@shfmt -w -i 2 -ci .
	@echo "Running golangci-lint with --fix..."
	@golangci-lint run --fix ./...
	@echo "Auto-fix completed. Running final lint check..."
	@golangci-lint run ./...
	@echo "Running checkmake..."
	@checkmake --config=.checkmake Makefile
	@echo "Running yamllint..."
	@yamllint -c .yamllint .

# Run linters with verbose output
lint-verbose:
	@echo "Running golangci-lint (verbose)..."
	@golangci-lint run -v ./...
	@echo "Running checkmake (verbose)..."
	@checkmake --config=.checkmake --format="{{.Line}}:{{.Rule}}:{{.Violation}}" Makefile
	@echo "Running shfmt check (verbose)..."
	@shfmt -d .
	@echo "Running yamllint (verbose)..."
	@yamllint -c .yamllint -f parsable .

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

# Security targets
security:
	@echo "Running comprehensive security scan..."
	@./scripts/security-scan.sh

security-full:
	@echo "Running full security analysis..."
	@./scripts/security-scan.sh

vuln-check:
	@echo "Checking for dependency vulnerabilities..."
	@go install golang.org/x/vuln/cmd/govulncheck@latest
	@govulncheck ./...