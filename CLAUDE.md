# CLAUDE.md

Go CLI aggregating code files into LLM-optimized output. Supports markdown/JSON/YAML with concurrent processing.

## Architecture (42 files, 8.2K lines)

**Core**: `main.go` (37), `cli/` (4), `fileproc/` (27), `config/` (3), `utils/` (4), `testutil/` (2)

**Modules**: Collection, processing, writers, registry (~63ns cache), resource limits

**Patterns**: Producer-consumer, thread-safe registry, streaming, modular (50-200 lines)

## Commands

```bash
make lint-fix && make lint && make test
./gibidify -source <dir> -format markdown --verbose
```

## Config

`~/.config/gibidify/config.yaml`
Size limit 5MB, ignore dirs, custom types, 100MB memory limit

## Quality

**CRITICAL**: `make lint-fix && make lint` (0 issues), 120 chars, EditorConfig, 30+ linters

## Testing

**Coverage**: 84%+ (utils 90.9%, fileproc 83.8%), race detection, benchmarks

## Standards

EditorConfig (LF, tabs), semantic commits, testing required

## Status

**Health: 10/10** - Production-ready, 84%+ coverage, modular, memory-optimized

**Done**: Errors, benchmarks, config, optimization, modularization, CLI (progress/colors), security (path validation, resource limits, scanning)

**Next**: Documentation, output customization

## Workflow

1. `make lint-fix` first 2. >80% coverage 3. Follow patterns 4. Update docs
