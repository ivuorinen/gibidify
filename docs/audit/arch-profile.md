# Architecture Profile

Generated: 2026-10-13

## Detected Patterns

### Pipe and Filter — High confidence

Evidence:

- `cli/processor_workers.go` `sendFiles` feeds `jobs chan fileJob` to a
  worker pool (`startWorkers`, `p.flags.Concurrency` goroutines); each worker
  runs the `fileproc.ProcessFileContext` filter and hands its result back on a
  per-job `out` channel.
- `sequence` reorders the per-job channels into collection order and forwards
  them on `writeCh chan fileproc.WriteRequest` to the sink,
  `fileproc.StartWriter` (`cli/processor_processing.go` `processFiles`).
- Stages share no state other than the channels; shutdown is channel-closure
  driven (`close(jobs)`, `close(order)`, `close(writeCh)`), and the writer
  drains `writeCh` after an error (`fileproc/writer.go:25-29`).

### Strategy + Registry — Medium confidence

Evidence:

- `fileproc.FormatWriter` (`formats.go:19-23`) with three implementations
  (`JSONWriter`, `MarkdownWriter`, `YAMLWriter`) plus `unsupportedFormatWriter`,
  chosen in one place by format string (`StartWriter`, `writer.go:41-61`).
- `fileproc.FileTypeRegistry` singleton (`registry.go`) holding image, binary
  and language extension maps, extended and reduced from configuration
  (`ConfigureFromSettings`, `config.go`).

### Layered by function — Medium confidence

Evidence:

- Production import graph (`go list`): `main → cli, config, shared`;
  `cli → config, fileproc, shared`; `fileproc → config, shared`;
  `config → shared`; `shared →` (none); `testutil → config, shared`.
- Dependencies point one way, `cli` → `fileproc` → `config` → `shared`, with
  no cycles and no back-edges. There are no `domain/` / `application/` /
  `infrastructure/` rings, so this is function-layered packaging, not Onion or
  Clean Architecture.

## Detected Combination

Custom hybrid: Pipe-and-Filter pipeline + Strategy/Registry extension points +
function-layered packages. A single-binary batch CLI; no canonical combination
applies.

## Inferred Structural Rules

1. `shared/` is a leaf: it imports no application package.
2. `config/` imports only `shared/`; it holds no file-processing knowledge.
    `fileproc/` receives configuration values through `config` getters read in
    its constructors (`NewFileFilter`, `NewFileProcessor`) or as arguments
    (`ConfigureFromSettings`); `config/` never imports `fileproc/`.
3. `fileproc/` never imports `cli/`. Pipeline stages in `fileproc/` take and
    return channels or values; they do not call one another.
4. Pipeline order is walker → processor → sequencer → writer, with no
    back-edges. Output order equals collection order regardless of worker
    scheduling.
5. Every pipeline goroutine is joined before `processFiles` returns, and the
    output file is closed only after the writer has finished.
6. Output formats are selected only through `StartWriter`; no caller switches
    on a concrete writer type.
7. The file-type registry is configured before any worker starts and is
    read-only afterwards (stated at `fileproc/registry.go:17-19`); lookups
    take no lock.
8. Logging goes through `shared.GetLogger()`; no production package imports
    stdlib `log` or writes diagnostics to stderr directly. User-facing status
    output goes through `cli.UIManager`.
9. Errors crossing package boundaries are `shared.StructuredError` values
    (`WrapError` / `NewStructuredError`), so `cli.ErrorFormatter` can
    classify them.
10. File content is buffered whole, capped at `fileSizeLimit`; there is no
    streaming path (`fileproc/processor.go` `readCapped`).
11. `testutil/` is imported only by `_test.go` files.

## Ambiguities & Contradictions

- Pipeline orchestration lives in `cli/` (`processor_processing.go`,
  `processor_workers.go`: channel wiring, worker pool, sequencing), while every
  stage lives in `fileproc/`. The presentation package therefore owns the
  pipeline's concurrency design, and `fileproc` cannot be driven end to end
  without `cli`.
- `fileproc/` holds every stage in one package — collection (walker, filters,
  ignore rules), classification (registry), per-file processing and output
  formatting (writers) — so the stage boundaries are file-level, not
  package-level.
- Process-wide mutable singletons: `config.cfg`, the `fileproc` registry,
  `cli.globalFlags`/`flagsParsed` and the `shared` logger. Tests reset them
  through exported hooks (`config.Reset`, `fileproc.ResetRegistryForTesting`,
  `cli.ResetFlags`) compiled into the production packages.
- `fileproc.Walker` is an interface with no consumer (`CollectFiles` uses
  `*ProdWalker` directly); see audit-12c8d813.

## Drift

Compared with the profile last written in e8f04c5 (2026-05-07), which predates
the 658eec6 simplification:

- Removed packages: `metrics/`, `templates/`, `benchmark/`, `cmd/benchmark/`.
  The rules that referenced them (single secondary entrypoint, template
  variables) no longer apply.
- Removed mechanisms: back-pressure (`fileproc/backpressure.go`), resource
  monitoring (`fileproc/resource_monitor_*.go`), `RegistryStats`, and the
  streaming helpers (`shared.StreamContent`/`StreamLines`). The old rule
  "Streaming for file content" is inverted: content is now buffered whole
  (rule 10).
- Pipeline orchestration moved into `cli/` with an explicit sequencer that makes
  output order deterministic (rules 4 and 5); the old profile had no ordering
  rule.
- Logging moved from logrus to `log/slog` behind `shared.Logger` (rule 8).
- The "Plugin / Registry — High" pattern is downgraded to "Strategy + Registry
  — Medium": the configurable output sections it cited (`output.markdown.*`,
  `output.variables.*`) were removed, leaving only extension lists and the
  format switch.
- Repository Pattern (Low) dropped: no data-access abstraction exists.
