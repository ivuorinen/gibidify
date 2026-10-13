# gibidify Makefile

.PHONY: help all build install
.PHONY: test test-verbose test-coverage
.PHONY: fmt fmt-check lint lint-go lint-static lint-sec lint-yaml lint-yamllint lint-actions lint-make lint-md lint-shell lint-docker lint-sast
.PHONY: ci
.PHONY: security security-full vuln-check gosec-sarif vuln-sarif
.PHONY: clean update-deps dev-setup hooks-setup
.PHONY: benchmark-go benchmark-go-cli benchmark-go-fileproc benchmark-go-shared

# Tool versions live in mise.toml; every tool runs through mise.
MISE := mise exec --

# Default goal — invoking `make` with no args prints the help table.
.DEFAULT_GOAL := help

LDFLAGS    := -s -w
LOCAL_PKG  := github.com/ivuorinen/gibidify
# yamllint and markdownlint-cli2 ignore .gitignore, so they get the files git would track
# (wildcard drops paths deleted from the worktree but still in the index).
YAML_FILES  = $(wildcard $(shell git ls-files --cached --others --exclude-standard '*.yml' '*.yaml'))
# The findings INDEX.md is generated (findings.py index) with fixed-width paths.
MD_FILES    = $(filter-out docs/audit/findings/INDEX.md,$(wildcard $(shell git ls-files --cached --others --exclude-standard '*.md')))

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-22s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

all: ci ## Run format, lint, and test (alias for ci)

# Build / install ------------------------------------------------------------

build: ## Build the gibidify binary
	go build -ldflags="$(LDFLAGS)" -o gibidify .

install: ## Install the current checkout globally
	go install .

# Test -----------------------------------------------------------------------

test: ## Run all tests with race detector
	go test -race ./...

test-verbose: ## Run tests with verbose output
	go test -race -v ./...

test-coverage: ## Run tests with coverage profile + HTML report
	go test -race -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -html=coverage.out -o coverage.html

# Format / lint --------------------------------------------------------------

fmt: ## Format Go code (gofmt + goimports, mutating)
	gofmt -w .
	$(MISE) goimports -w -local $(LOCAL_PKG) .

fmt-check: ## Check Go formatting without modifying files (used in CI)
	@out=$$(gofmt -l .); test -z "$$out" || { printf 'Unformatted files:\n%s\nRun make fmt to fix\n' "$$out"; exit 1; }
	@out=$$($(MISE) goimports -l -local $(LOCAL_PKG) .); test -z "$$out" || { printf 'Files need import formatting:\n%s\nRun make fmt to fix\n' "$$out"; exit 1; }

lint: ## Run all linters via prek (preferred)
	@$(MISE) prek run --all-files

lint-go: ## Run only Go linters (vet + revive)
	go vet ./...
	$(MISE) revive -config revive.toml -formatter friendly -set_exit_status ./...

lint-static: ## Run staticcheck
	$(MISE) staticcheck ./...

lint-sec: ## Run gosec on code and tests (alias for `make security`)
	$(MISE) gosec -quiet -tests ./...

lint-yaml: ## Run only YAML formatter check (yamlfmt -lint)
	$(MISE) yamlfmt -lint -conf .yamlfmt.yml

lint-yamllint: ## Run yamllint in strict mode
	$(MISE) yamllint --strict $(YAML_FILES)

lint-actions: ## Run GitHub Actions linters (actionlint + zizmor)
	$(MISE) actionlint .github/workflows/*.yml
	$(MISE) zizmor --no-progress .github

lint-make: ## Run only Makefile linter (checkmake)
	$(MISE) checkmake --config=.checkmake Makefile

lint-md: ## Run only Markdown linter (markdownlint-cli2)
	$(MISE) markdownlint-cli2 $(MD_FILES)

lint-shell: ## Check shell script formatting (shfmt)
	$(MISE) shfmt -d .

lint-docker: ## Scan the Dockerfiles for misconfigurations (trivy)
	$(MISE) trivy config --exit-code 1 Dockerfile
	$(MISE) trivy config --exit-code 1 Dockerfile.release

lint-sast: ## Static analysis with opengrep registry rule packs
	$(MISE) opengrep scan --error --quiet \
		--config p/golang --config p/secrets --config p/github-actions \
		--config p/dockerfile --config p/security-audit --config p/trailofbits .

# CI -------------------------------------------------------------------------

ci: fmt-check lint test ## Run format check, lint, and test

# Security -------------------------------------------------------------------

security: ## Run gosec security scan on code and tests
	$(MISE) gosec -tests ./...

security-full: ## Run comprehensive multi-tool security scan
	@$(MISE) ./scripts/security-scan.sh

vuln-check: ## Check for dependency vulnerabilities (govulncheck)
	$(MISE) govulncheck ./...

gosec-sarif: ## CI: gosec scan as gosec-results.sarif
	$(MISE) gosec -tests -fmt sarif -out gosec-results.sarif ./...

vuln-sarif: ## CI: govulncheck scan as govulncheck-results.sarif
	$(MISE) govulncheck -format sarif ./... > govulncheck-results.sarif

# Maintenance ----------------------------------------------------------------

update-deps: ## Update Go dependencies to latest patch versions
	go get -u=patch ./...
	go mod tidy
	go mod verify
	@go list -u -m all | grep '\[' || true

clean: ## Remove build artifacts, coverage, test outputs
	rm -f gibidify coverage.out coverage.html test-results.json \
		gosec-report.json govulncheck-report.json gitleaks-report.json security-report.md
	go clean -testcache

# Development setup ----------------------------------------------------------

dev-setup: hooks-setup ## Set up development environment

hooks-setup: ## Install the tools and the prek git hooks
	mise install
	$(MISE) prek install --prepare-hooks

# Benchmarks (Go test benchmarks) --------------------------------------------

benchmark-go: ## Run all Go test benchmarks
	go test -bench=. -benchtime=100ms -run=^$$ ./...

benchmark-go-cli: ## Run CLI test benchmarks
	go test -bench=. -benchtime=100ms -run=^$$ ./cli/...

benchmark-go-fileproc: ## Run fileproc test benchmarks
	go test -bench=. -benchtime=100ms -run=^$$ ./fileproc/...

benchmark-go-shared: ## Run shared test benchmarks
	go test -bench=. -benchtime=100ms -run=^$$ ./shared/...
