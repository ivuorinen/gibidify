# Architecture Profile
Generated: 2026-05-06

## Detected Patterns

### Pipe and Filter — High confidence
Evidence:
- Producer-consumer pipeline: `fileproc/walker.go` (collector) → `fileproc/processor.go` (filter/transformer) → `fileproc/{json,markdown,yaml}_writer.go` (sink), connected by Go channels
- `fileproc/processor.go:55,67,80` — functions take `outCh chan<- WriteRequest`; pipeline stages decoupled by channel hand-off
- `fileproc/backpressure.go` — explicit back-pressure management between stages (`maxPendingFiles`, `maxPendingWrites`)
- `interfaces.go:13-30` — `Processor` interface plus `FileProcessorInterface` describes the staged transformation
- CLAUDE.md acknowledges this: "Producer-consumer, … streaming"

### Plugin / Registry — High confidence
Evidence:
- `fileproc/registry.go:19-32` — `FileTypeRegistry` with `RegistryStats`, accessed via `DefaultRegistry()`; ~63ns lookup mentioned in CLAUDE.md
- `fileproc/extensions.go` — extensible file-type registration (custom + disabled extensions, custom languages)
- `fileproc/formats.go:19-23` — `FormatWriter` interface implemented by three writers (`json_writer.go`, `markdown_writer.go`, `yaml_writer.go`); writers selected at runtime by format flag
- Configuration-driven extension: `output.markdown.*`, `output.metadata.*`, `output.variables.*` in `config.example.yaml`

### Layered / Functional Modular — Medium confidence
Evidence:
- Top-level Go packages organized by responsibility, not by feature: `cli/` (presentation), `fileproc/` (domain/pipeline), `config/` (configuration), `metrics/` (observability), `templates/` (templating), `shared/` (cross-cutting utilities), `testutil/` (test support)
- Dependency direction broadly inward: `cli/` and `cmd/benchmark/` depend on `fileproc`, `config`, `shared`; `fileproc/` depends on `config`, `shared`; `shared/` has no inward dependency on application packages
- No `domain/`, `application/`, `infrastructure/` rings — so it is layered-by-function, not Onion or Clean

### Repository Pattern — Low confidence
Evidence:
- `fileproc/registry.go` is registry-style but for file-type lookup, not data access
- No persistent store, no `*Repository` types
- Listed as "low" because the registry concept is present but not the canonical data-access repository

## Detected Combination

**Custom hybrid: Pipe-and-Filter pipeline + Plugin/Registry extension + Functional Layered organization**

This is not a canonical named combination. It is a CLI batch-processing tool whose primary structural backbone is the file-processing pipeline; layering is shallow and follows Go's package-as-module convention.

## Inferred Structural Rules

The following rules are inferred from the detected combination and should be enforced by `arch-auditor`:

1. **Pipeline stage decoupling**: Stages communicate only via channels (`chan WriteRequest`, `chan FileData`-equivalent). A stage must not call into a downstream stage directly.
2. **Pipeline direction**: Walker → Processor → Writer. No back-edges. Writers must not import or call back into walker/processor.
3. **Format-writer interchangeability**: All three writers (`json_writer.go`, `markdown_writer.go`, `yaml_writer.go`) implement `FormatWriter` and must be selectable purely from the `-format` flag — no caller may switch on concrete writer type.
4. **Registry encapsulation**: File-type detection goes through `DefaultRegistry()` / `FileTypeRegistry`. No package outside `fileproc/` should directly inspect file extensions for language detection.
5. **`shared/` is leaf-level**: `shared/` may not import any application package (`cli/`, `fileproc/`, `config/`, `metrics/`, `templates/`, `cmd/`, `benchmark/`, `testutil/`). Imports must flow inward toward `shared/`.
6. **`config/` does not import `fileproc/`**: Configuration owns no domain knowledge. `fileproc/` may read config; the inverse is forbidden.
7. **`testutil/` is test-only**: All exported symbols in `testutil/` must be used only from `_test.go` files. No production code may import `testutil/`.
8. **Logging through `shared.GetLogger()`**: No package may import `logrus` or `log` directly for application-level logging. The interface is `shared.Logger`.
9. **Errors through `shared.WrapError` family**: Domain code returns wrapped structured errors via `shared.WrapError*`. Bare `fmt.Errorf` for new error chains is acceptable; wrapping existing errors must use the helpers.
10. **Streaming for file content**: File content is processed via `shared.StreamContent`/`StreamLines` rather than full-buffer `os.ReadFile`. Large-file handling is non-negotiable per CLAUDE.md "memory-optimized".
11. **CLI entrypoint is single**: `main.go` and `cli/` own program startup. `cmd/benchmark/` is the only secondary entrypoint and must not duplicate CLI flag parsing.
12. **Plugin extension points are config-driven**: Custom file extensions, custom languages, and template variables are added through `config.yaml`, not by recompilation of `fileproc/`.

## Ambiguities & Contradictions

- **`interfaces.go` lives in package `main` at the repo root.** It defines cross-cutting interfaces (`Processor`, `FileProcessorInterface`, `ResourceMonitorInterface`, `MetricsCollectorInterface`) that conceptually belong in `shared/` or a dedicated `interfaces/` package. Putting them in `package main` makes them un-importable by sibling packages — anything that wanted to satisfy these interfaces from `cli/` or `fileproc/` would have a circular import. This is a structural smell flagged for arch-auditor.
- **`benchmark/` and `cmd/benchmark/`**: two top-level locations referencing benchmarking. Their relationship needs verification (library vs entrypoint) — currently ambiguous from naming alone.
- **`metrics/` vs `fileproc/resource_monitor_*.go`**: there is overlap between `metrics/` (observability) and the `resource_monitor_*` family inside `fileproc/`. Whether resource monitoring is part of the pipeline or part of metrics is not consistently chosen.

## Drift

(first run — no prior profile)
