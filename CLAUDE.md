# CLAUDE.md

Go CLI aggregating code files into LLM-optimized output.
Supports markdown/JSON/YAML with concurrent processing.

## Architecture

**Core**: `main.go`, `cli/`, `fileproc/`, `config/`, `shared/`, `testutil/`

**Modules**: Collection, processing, writers, file-type registry,
file-size/count caps

**Patterns**: Producer-consumer pipeline with ordered output, file-type
registry (configured once, read-only afterwards); see
`docs/audit/arch-profile.md`

## Commands

Tool versions live in `mise.toml`; run `make hooks-setup` once (mise install +
prek hooks).

```bash
make fmt && make lint && make test
./gibidify -source <dir> -format markdown --verbose
./gibidify -source <dir> -format json --log-level debug --verbose
```

## Config

`~/.config/gibidify/config.yaml`
Size limit 5MB, ignore dirs, custom types, max file count/total size caps

## Linting Standards (MANDATORY)

**Linter**: revive (comprehensive rule set migrated from golangci-lint)
**Command**: `make lint` (prek runs every hook: go vet, revive, staticcheck,
gosec, govulncheck, actionlint, zizmor, editorconfig-checker, markdownlint,
yamlfmt, yamllint)
**Complexity**: cognitive-complexity ≤15, cyclomatic ≤15, max-control-nesting ≤5
**Security**: unhandled errors, secure coding patterns, credential detection
**Performance**: optimize-operands-order, string-format, range optimizations
**Format**: line-length ≤120 chars, EditorConfig (LF, tabs), gofmt/goimports
**Testing**: error handling best practices, 0 tolerance policy

**CRITICAL**: All rules are non-negotiable unless the user explicitly asks to
change one (see revive.toml Restrictions below). `make fmt && make lint` must
show 0 issues.

## Testing

**Coverage**: run `go test -cover ./...` for current numbers (per-package; not
pinned in this doc to avoid drift)
**Patterns**: Table-driven tests, shared testutil helpers, mock objects, error
assertions
**Race detection**, benchmarks, comprehensive integration tests

## Development Patterns

**Logging**: Use `shared.GetLogger()` for all logging (stdlib `log/slog`
backend). Default level `warn`, set via `--log-level` flag
**Error Handling**: Use `shared.WrapError` family for structured errors with
context
**File reading**: Files are buffered whole, capped at `fileSizeLimit`
(`fileproc/processor.go`); there is no streaming path
**Context**: Use `shared.CheckContextCancellation` for standardized cancellation
**Testing**: Use `testutil.*` helpers for directory setup, error assertions
**Validation**: Centralized in `config/validation.go` with structured error
collection

## Standards

EditorConfig (LF, tabs), semantic commits, testing required, error wrapping

## revive.toml Restrictions

**Agents do not have permission** to modify `revive.toml` configuration unless
the user explicitly requests it.
The linting configuration is carefully tuned and should not be altered during
normal development.

## Status

**Health**: lint 0 issues; tests pass with `-race`. Run
`make lint && go test -race ./...` to verify.

**Done**: Deduplication, errors, benchmarks, config, optimization,
modularization, linting

## Workflow

1. `make fmt` first
2. Keep total coverage at or above the CI gate (60%, `build-test-publish.yml`)
3. Follow patterns
4. Update docs
