# CLAUDE.md

Go CLI that aggregates code files into LLM-optimized output. Supports markdown/JSON/YAML with concurrent processing.

## Architecture (40 files, 189KB, 6.8K lines)

**Core**: `main.go` (37 lines), `cli/` (4 files), `fileproc/` (22 files), `config/` (3 files), `utils/` (4 files), `testutil/` (2 files)

**Key modules**: File collection, processing, writers (markdown/JSON/YAML), registry with caching, back-pressure management

**Patterns**: Producer-consumer pools, thread-safe registry (~63ns lookups), streaming with back-pressure, modular files (50-200 lines), progress bars, enhanced errors

## Commands

```bash
make lint-fix && make lint && make test  # Essential workflow
./gibidify -source <dir> -format markdown --no-colors --no-progress --verbose
```

## Config

XDG config paths: `~/.config/gibidify/config.yaml`

**Key settings**: File size limit (5MB), ignore dirs, custom file types, back-pressure (100MB memory limit)

## Quality

**CRITICAL**: `make lint-fix && make lint` (0 issues), max 120 chars, EditorConfig compliance, 30+ linters

## Testing

**Coverage**: 84%+ (utils 90.9%, testutil 84.2%, fileproc 83.8%), race detection, benchmarks, testutil helpers

## Standards

EditorConfig (LF, tabs), semantic commits, testing required, linting must pass

## Status

**Health: 10/10** - Production-ready, 84%+ coverage, modular architecture, memory-optimized

**Completed**: Structured errors, benchmarking, config validation, memory optimization, code modularization, CLI enhancements (progress bars, colors, enhanced errors)

**Next**: Security hardening, documentation, output customization

## Workflow

1. `make lint-fix` before changes 2. >80% coverage 3. Follow patterns 4. Update docs 5. Security/performance
