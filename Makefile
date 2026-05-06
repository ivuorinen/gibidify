# gibidify Makefile

.PHONY: help all build install
.PHONY: test test-verbose test-coverage
.PHONY: fmt lint lint-go lint-static lint-sec lint-yaml lint-actions lint-make lint-md
.PHONY: ci ci-lint ci-test
.PHONY: security security-full vuln-check
.PHONY: clean update-deps dev-setup pre-commit-setup
.PHONY: build-benchmark benchmark benchmark-go benchmark-all
.PHONY: benchmark-go-cli benchmark-go-fileproc benchmark-go-metrics benchmark-go-shared
.PHONY: benchmark-collection benchmark-processing benchmark-concurrency benchmark-format

# Tool versions (managed by Renovate)
# renovate: datasource=go depName=github.com/golangci/golangci-lint/v2/cmd/golangci-lint
GOLANGCI_LINT_VERSION := v2.10.1
# renovate: datasource=go depName=github.com/google/yamlfmt/cmd/yamlfmt
YAMLFMT_VERSION := v0.21.0
# renovate: datasource=go depName=github.com/rhysd/actionlint/cmd/actionlint
ACTIONLINT_VERSION := v1.7.11
# renovate: datasource=go depName=golang.org/x/tools/cmd/goimports
GOIMPORTS_VERSION := v0.42.0
# renovate: datasource=go depName=github.com/securego/gosec/v2/cmd/gosec
GOSEC_VERSION := v2.24.0
# renovate: datasource=go depName=honnef.co/go/tools/cmd/staticcheck
STATICCHECK_VERSION := v0.7.0
# renovate: datasource=go depName=github.com/mgechev/revive
# Pinned to v1.11.0 — newer revive added var-naming rules that flag the existing
# `shared` and `metrics` package names; revive.toml is intentionally off-limits.
REVIVE_VERSION := v1.11.0
# renovate: datasource=go depName=github.com/checkmake/checkmake/cmd/checkmake
CHECKMAKE_VERSION := v0.3.2
# govulncheck intentionally tracks @latest, not a pinned version:
# the vuln DB is fetched live from vuln.go.dev each run, but the binary
# controls reachability-analysis features and bug fixes. For a security
# scanner, "always run the newest one" beats reproducibility.
GOVULNCHECK_VERSION := latest
# renovate: datasource=npm depName=markdownlint-cli2
MARKDOWNLINT_CLI2_VERSION := 0.21.0

# Default goal — invoking `make` with no args prints the help table.
.DEFAULT_GOAL := help

LDFLAGS    := -s -w
LOCAL_PKG  := github.com/ivuorinen/gibidify

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-22s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

all: ci ## Run format, lint, and test (alias for ci)

# Build / install ------------------------------------------------------------

build: ## Build the gibidify binary
	go build -ldflags="$(LDFLAGS)" -o gibidify .

install: ## Install gibidify globally via go install
	go install $(LOCAL_PKG)@latest

# Test -----------------------------------------------------------------------

test: ## Run all tests with race detector
	go test -race ./...

test-verbose: ## Run tests with verbose output
	go test -race -v ./...

test-coverage: ## Run tests with coverage profile + HTML report
	go test -race -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -html=coverage.out -o coverage.html

# Format / lint --------------------------------------------------------------

fmt: ## Format Go code (gofmt + goimports)
	gofmt -w .
	go run golang.org/x/tools/cmd/goimports@$(GOIMPORTS_VERSION) -w -local $(LOCAL_PKG) .

lint: ## Run all linters via pre-commit (preferred)
	@pre-commit run --all-files

lint-go: ## Run only Go linters (vet + revive)
	go vet ./...
	go run github.com/mgechev/revive@$(REVIVE_VERSION) -config revive.toml -formatter friendly -set_exit_status ./...

lint-static: ## Run staticcheck
	go run honnef.co/go/tools/cmd/staticcheck@$(STATICCHECK_VERSION) ./...

lint-sec: ## Run gosec (alias for `make security`)
	go run github.com/securego/gosec/v2/cmd/gosec@$(GOSEC_VERSION) -quiet ./...

lint-yaml: ## Run only YAML linter (yamlfmt -lint)
	go run github.com/google/yamlfmt/cmd/yamlfmt@$(YAMLFMT_VERSION) -lint -conf .yamlfmt.yml .

lint-actions: ## Run only GitHub Actions linter (actionlint)
	go run github.com/rhysd/actionlint/cmd/actionlint@$(ACTIONLINT_VERSION) .github/workflows/*.yml

lint-make: ## Run only Makefile linter (checkmake)
	go run github.com/checkmake/checkmake/cmd/checkmake@$(CHECKMAKE_VERSION) --config=.checkmake Makefile

lint-md: ## Run only Markdown linter (markdownlint-cli2)
	npx --yes markdownlint-cli2@$(MARKDOWNLINT_CLI2_VERSION) "*.md" "**/*.md"

# CI -------------------------------------------------------------------------

ci: fmt lint test ## Run format, lint, and test

ci-lint: ## CI: revive only with strict exit
	go run github.com/mgechev/revive@$(REVIVE_VERSION) -config revive.toml -formatter friendly -set_exit_status ./...

ci-test: ## CI: tests with coverage in JSON
	go test -race -coverprofile=coverage.out -json ./... > test-results.json

# Security -------------------------------------------------------------------

security: ## Run gosec security scan
	go run github.com/securego/gosec/v2/cmd/gosec@$(GOSEC_VERSION) ./...

security-full: ## Run comprehensive multi-tool security scan
	@./scripts/security-scan.sh

vuln-check: ## Check for dependency vulnerabilities (govulncheck)
	go run golang.org/x/vuln/cmd/govulncheck@$(GOVULNCHECK_VERSION) ./...

# Maintenance ----------------------------------------------------------------

update-deps: ## Update Go dependencies to latest patch versions
	go get -u=patch ./...
	go mod tidy
	go mod verify
	@go list -u -m all | grep '\[' || true

clean: ## Remove build artifacts, coverage, test outputs
	rm -f gibidify gibidify-benchmark coverage.out coverage.html test-results.json security-report.json
	go clean -testcache

# Development setup ----------------------------------------------------------

dev-setup: pre-commit-setup ## Set up development environment

pre-commit-setup: ## Install pre-commit hooks
	@command -v pre-commit >/dev/null 2>&1 || pip install pre-commit
	@pre-commit install
	@pre-commit install --install-hooks

# Benchmarks (gibidify-specific) ---------------------------------------------

build-benchmark: ## Build the gibidify-benchmark binary
	go build -ldflags="$(LDFLAGS)" -o gibidify-benchmark ./cmd/benchmark

benchmark: build-benchmark ## Run all custom benchmarks
	./gibidify-benchmark -type=all

benchmark-collection: build-benchmark ## Run file collection benchmarks
	./gibidify-benchmark -type=collection

benchmark-processing: build-benchmark ## Run file processing benchmarks
	./gibidify-benchmark -type=processing

benchmark-concurrency: build-benchmark ## Run concurrency benchmarks
	./gibidify-benchmark -type=concurrency

benchmark-format: build-benchmark ## Run format benchmarks
	./gibidify-benchmark -type=format

benchmark-go: ## Run all Go test benchmarks
	go test -bench=. -benchtime=100ms -run=^$$ ./...

benchmark-go-cli: ## Run CLI test benchmarks
	go test -bench=. -benchtime=100ms -run=^$$ ./cli/...

benchmark-go-fileproc: ## Run fileproc test benchmarks
	go test -bench=. -benchtime=100ms -run=^$$ ./fileproc/...

benchmark-go-metrics: ## Run metrics test benchmarks
	go test -bench=. -benchtime=100ms -run=^$$ ./metrics/...

benchmark-go-shared: ## Run shared test benchmarks
	go test -bench=. -benchtime=100ms -run=^$$ ./shared/...

benchmark-all: benchmark benchmark-go ## Run custom + Go test benchmarks
