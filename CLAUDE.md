# CLAUDE.md

Go CLI aggregating code files into LLM-optimized output.
Supports markdown/JSON/YAML with concurrent processing.

## Architecture

**Core**: `main.go`, `cli/`, `fileproc/`, `config/`, `utils/`, `testutil/`, `cmd/`

**Advanced**: `metrics/`, `templates/`, `benchmark/`

**Modules**: Collection, processing, writers, registry (~63ns cache), resource limits, metrics, templating

**Patterns**: Producer-consumer, thread-safe registry, streaming, modular (50-200 lines)

## Commands

```bash
make lint-fix && make lint && make test
./gibidify -source <dir> -format markdown --verbose
./gibidify -source <dir> -format json --log-level debug --verbose
```

## Config

`~/.config/gibidify/config.yaml`
Size limit 5MB, ignore dirs, custom types, 100MB memory limit

## Linting Standards (MANDATORY)

**Linter**: revive (comprehensive rule set migrated from golangci-lint)
**Command**: `revive -config revive.toml ./...`
**Complexity**: cognitive-complexity ≤15, cyclomatic ≤15, max-control-nesting ≤5
**Security**: unhandled errors, secure coding patterns, credential detection
**Performance**: optimize-operands-order, string-format, range optimizations
**Format**: line-length ≤120 chars, EditorConfig (LF, tabs), gofmt/goimports
**Testing**: error handling best practices, 0 tolerance policy

**CRITICAL**: All rules non-negotiable. `make lint-fix && make lint` must show 0 issues.

## Testing

**Coverage**: 77.9% overall (utils 90.0%, cli 83.8%, config 77.0%, testutil 73.7%, fileproc 74.5%, metrics 96.0%, templates 87.3%)
**Patterns**: Table-driven tests, shared testutil helpers, mock objects, error assertions
**Race detection**, benchmarks, comprehensive integration tests

## Development Patterns

**Logging**: Use `utils.Logger()` for all logging (replaces logrus). Default WARN level, set via `--log-level` flag
**Error Handling**: Use `utils.WrapError` family for structured errors with context
**Streaming**: Use `utils.StreamContent/StreamLines` for consistent file processing
**Context**: Use `utils.CheckContextCancellation` for standardized cancellation
**Testing**: Use `testutil.*` helpers for directory setup, error assertions
**Validation**: Centralized in `config/validation.go` with structured error collection

## Standards

EditorConfig (LF, tabs), semantic commits, testing required, error wrapping

## revive.toml Restrictions

**AGENTS DO NOT HAVE PERMISSION** to modify `revive.toml` configuration unless user explicitly requests it.
The linting configuration is carefully tuned and should not be altered during normal development.

## Status

**Health: 9/10** - Production-ready with systematic deduplication complete

**Done**: Deduplication, errors, benchmarks, config, optimization, testing (77.9%), modularization, linting (0 issues), metrics system, templating

## Workflow

1. `make lint-fix` first
2. >80% coverage
3. Follow patterns
4. Update docs
