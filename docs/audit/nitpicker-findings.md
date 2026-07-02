# Nitpicker Findings
Generated: 2026-05-06
Last validated: 2026-07-02

This pass ran in `security + docs + architecture` mode. Critical/High findings from
the specialist auditors were incorporated; the per-tool reports live in
`docs/audit/security-findings.md`, `docs/audit/doc-findings.md`,
`docs/audit/arch-findings.md`, and `docs/audit/arch-profile.md`.

## Summary
- Total: 25 | Open: 1 | Fixed: 22 | Invalid: 2
- Open by severity: Medium 1 (deferred refactor)

## Open Findings

### Medium

#### [N-009] Resource monitoring split between `fileproc/` and `metrics/`
Category: maintainability
Area: fileproc/resource_monitor_*.go (11 files), metrics/reporter.go
Problem: Two packages collect overlapping runtime metrics. Ambiguity about where to add new operational metrics.
Evidence: `ls fileproc/resource_monitor_*` produces 11 files; `metrics/reporter.go` consumes related data.
Impact: Cross-package coupling; risk of duplicate counters drifting.
Fix: Move `resource_monitor_*` into a dedicated `resource/` package with a small public surface, and have `metrics/` import that surface. Deferred — too large for a same-pass fix; needs a dedicated PR with test migration.

## Fixed

### Pass 1 — 2026-05-06

#### [N-001] CLAUDE.md references the non-existent `utils/` package throughout
Fixed: 2026-05-06
Notes: Replaced `utils/` → `shared/` and `utils.X` → `shared.X` for `GetLogger`, `WrapError`, `StreamContent`, `StreamLines`, `CheckContextCancellation`. Verified all symbols exist in `shared/`. Dropped the per-package coverage table in favor of "run `go test -cover ./...`".

#### [N-002] `examples/basic-usage.md` JSON example shows fields the writer does not emit
Fixed: 2026-05-06
Notes: Replaced fake `{ size, metadata, total_files, total_size, processing_time }` with the real shape `{ prefix, suffix, files: [{ path, language, content }] }` matching `fileproc/formats.go` and `fileproc/json_writer.go`.

#### [N-003] GitHub Actions shell injection via `${{ github.ref_name }}` in `run:` steps
Fixed: 2026-05-06
Notes: Bound `github.ref_name`, `github.token`, `github.actor`, `github.repository`, `matrix.goos`, `matrix.goarch`, and the windows `.exe` ext to step `env:` blocks; rewrote each `run:` script to reference `"$REF_NAME"`, `"$GH_TOKEN"`, etc. instead of GitHub-Actions interpolation. Files: `.github/workflows/build-test-publish.yml` build and docker jobs.

#### [N-004] Dockerfile uses `useradd` on Alpine — image build will fail
Fixed: 2026-05-06
Notes: Replaced `useradd -ms /bin/bash gibidify` with `adduser -D -s /bin/sh gibidify`. Switched `COPY` + `chmod` to `COPY --chmod=0755`. Added `HEALTHCHECK NONE` (closes SEC-005). Reordered `USER gibidify` to come after `COPY` so the file actually lands with root ownership before the user switch.

#### [N-005] `interfaces.go` defines orphaned interfaces and shadow types in `package main`
Fixed: 2026-05-06
Notes: Deleted `interfaces.go` (217 lines). `go build ./...` clean afterward; nothing depended on it. The canonical types (`fileproc.WriteRequest`, `metrics.*`) are unaffected.

#### [N-006] README and TODO advertise stale codebase metrics (92 files, 21.5K lines)
Fixed: 2026-05-06
Notes: Removed the file/line counts from `README.md:17` and `TODO.md:8`. Did not replace with current counts to avoid the same drift recurring. Future maintainers should add `cloc` to CI if they want a maintained number.

#### [N-007] `examples/basic-usage.md` Dockerfile recipe pinned to `golang:1.25-alpine`
Fixed: 2026-05-06
Notes: Updated to `golang:1.26-alpine` to match `go.mod` (`go 1.26.2`).

#### [N-008] README's flag list omits `--no-ui`
Fixed: 2026-05-06
Notes: Added `- --no-ui: disable all UI output (implies --no-colors and --no-progress).` to the Flags section in README.md.

#### [N-011] `cmd/benchmark/main.go:38` calls `flag.Parse()` directly
Fixed: 2026-05-06
Notes: Switched to `flag.CommandLine.Parse(os.Args[1:])` with explicit error propagation, mirroring the pattern in `cli/flags.go`. Package-level flag vars retained — refactoring those is N-013, deferred.

#### [N-012] CLAUDE.md "Status" line claims `Health: 9/10`
Fixed: 2026-05-06
Notes: Replaced with verifiable status: "lint 0 issues; tests pass with `-race`. Run `make lint && go test -race ./...` to verify." Removed the stale 77.9% coverage number from the Done line as well.

#### [SEC-004] Stale `gibidify-benchmark` build artifact pinned to vulnerable Go stdlib
Fixed: 2026-05-06
Notes: Removed `gibidify-benchmark`, `gibidify`, `coverage.out`, `coverage_backup.out` from the working tree. All four are gitignored. Future runs of grype will no longer match the 19 stdlib CVEs against a stale binary.

#### [ARCH-005] `metrics` package imports `golang.org/x/text/cases` for cosmetic title-casing
Fixed: 2026-05-06
Notes: Replaced `cases.Title(language.English).String(phase)` with a local `titleASCII` helper that uppercases the first byte (phase names are ASCII constants like "collection", "processing"). Dropped both `golang.org/x/text/cases` and `golang.org/x/text/language` imports from `metrics/reporter.go`. `templates/engine.go` keeps them — that path renders user-supplied template variables and may legitimately need full Unicode title-casing.

### Pass 2 — 2026-07-02

#### [N-014] Entire `templates/` package (~1,150 LOC) is dead — zero importers
Category: maintainability
Area: templates/, config/getters.go, config/loader.go, shared/constants.go, README.md, CLAUDE.md, config.example.yaml
Fixed: 2026-07-02
Notes: `templates/` had no importers anywhere (not even tests outside the package); no writer read any `output.template/metadata/markdown/custom/variables` config. Deleted the 3-file package, its 24 config getters (`OutputTemplate`, `TemplateMetadata*`, `TemplateMarkdown*`, `TemplateCustom*`, `TemplateVariables`) + `metadataBool`/`markdownBool` helpers, ~40 orphaned `shared` constants/keys/defaults (including `TemplateFmtTimestamp` and two engine-test message constants), and all matching `viper.SetDefault` calls in `config/loader.go`. Pruned the misleading `output:` block from `config.example.yaml` and README, and the templating mentions from CLAUDE.md. `go mod tidy` moved `golang.org/x/text` to indirect (templates was its last direct consumer). Kept the real `GetCustomLanguages` getter/test. Confirmed with user before deleting a documented feature. build/vet/`-race` tests all pass.

#### [N-015] Duplicate custom-extension validators in `config/validation.go`
Category: maintainability
Area: config/validation.go:187-233
Fixed: 2026-07-02
Notes: `validateCustomImageExtensions` and `validateCustomBinaryExtensions` were byte-identical except the config key. Collapsed into one `validateCustomExtensions(key string)` called with each key from `ValidateFileTypesConfig`.

#### [N-016] Dead `goto`/`select` defensive branch in `resource_monitor_types.go`
Category: maintainability
Area: fileproc/resource_monitor_types.go:99-107
Fixed: 2026-07-02
Notes: The pre-fill loop used a non-blocking `select` with `default: goto rateLimitFull`, but `rateLimitChan` is made with capacity `rateLimitFilesPerSec` and the refill goroutine starts only afterward, so every send always succeeds. Replaced with a plain blocking send loop; removed the label and dead branch.

#### [N-017] `ValidateSourcePath` cwd-escape check lacks a path-separator boundary
Category: security
Area: shared/paths.go:89
Fixed: 2026-07-02
Notes: `strings.HasPrefix(abs, cwdAbs)` would treat `/cwd-evil` as inside `/cwd`. Not currently exploitable (only relative paths reach it and `..` is already rejected), but hardened as trust-boundary defense-in-depth: `abs == cwdAbs || strings.HasPrefix(abs, cwdAbs+string(filepath.Separator))`.

#### [N-019] `github.com/sirupsen/logrus` replaceable by stdlib `log/slog`
Category: maintainability
Area: shared/logger.go, go.mod
Fixed: 2026-07-02
Notes: logrus was encapsulated behind the `Logger` interface in a single file (`shared/logger.go`); the rest of the codebase already used `shared.GetLogger()`. Reimplemented `logService` on `log/slog` (`TextHandler` with a shared swappable writer + `slog.LevelVar` so `SetOutput`/`SetLevel` still propagate to `WithFields`-derived loggers). Preserved the prior output contract: timestamp dropped and level lowercased (`level=error`) via `ReplaceAttr`, so `errors_test.go`'s format assertion passes unchanged. `Logger` interface unchanged → no other files touched. `go mod tidy` dropped logrus (its only transitive dep, `golang.org/x/sys`, is still pulled by color/progressbar/term). Direct deps 6 → 5. All `-race` tests pass.

#### [N-020] `github.com/spf13/viper` replaced by an in-house yaml.v3 store
Category: maintainability
Area: config/store.go (new), config/loader.go, config/getters.go, config/validation.go, testutil/testutil.go, 12 test files, go.mod
Fixed: 2026-07-02
Notes: viper was used only as a global dotted-key store with defaults + `IsSet` semantics (no env-binding/watching). A plain struct can't represent `IsSet` (validate-only-if-user-set), so replaced it with a 210-line `config/store.go`: two nested `map[string]any` trees (explicit values shadow defaults), yaml.v3 file loading over search paths, and the exact surface the app used (`Set`/`SetDefault`/`IsSet`/`Reset`/`AddConfigPath`/`ReadInConfig`/`FileUsed`/`SetConfigFile` + typed `GetInt`/`GetInt64`/`GetBool`/`GetStringSlice`/`GetStringMapString`). Nested-map (not flattened) model so map-valued keys like `fileTypes.customLanguages` retrieve correctly. `getters.go`/`validation.go` call the package-level funcs; tests migrated mechanically (`viper.X`→`config.X`, internal `config`-package test uses bare names). Preserved exact behavior: validate only on file load, fallback-to-defaults on validation failure. Direct deps 5 → 4; indirect 17 → 9 (dropped afero, cast, pflag, gotenv, locafero, fsnotify, mapstructure, go-toml, subosito, go-viper). Smoke-tested a real XDG config load end-to-end; all `-race` tests pass, revive clean (except the pre-existing package-name warning).

### Pass 3 — 2026-07-02

#### [N-021] `benchmark/` + `cmd/benchmark/` shipped a second binary of dev tooling
Category: maintainability
Area: benchmark/, cmd/benchmark/, Makefile
Fixed: 2026-07-02
Notes: A custom benchmarking harness (~750 prod + ~1,270 test LOC) built as a separate `gibidify-benchmark` binary — dev tooling, not the app's aggregate→format→write purpose. Deleted both packages and the empty `cmd/`. Removed the `build-benchmark`/`benchmark`/`benchmark-collection|processing|concurrency|format`/`benchmark-all` Makefile targets and the `gibidify-benchmark` clean entry; kept `benchmark-go-*` (`go test -bench`). User-approved.

#### [N-022] `metrics/` package (Collector/Reporter) over-engineered for a CLI
Category: maintainability
Area: metrics/ (deleted), cli/processor_*.go, shared/constants.go
Fixed: 2026-07-02
Notes: ~914 prod + ~1,000 test LOC of real-time stats collection, phase timing, and reporters wired through `cli/processor_*`. Replaced with a single `shared.GetLogger().Infof("Processed %d files in %s", ...)` summary line in `Process`. Deleted the package, the `metricsCollector`/`metricsReporter` fields + all call sites, `processor_stats.go`, and the `MetricsPhase*` constants. User-approved.

#### [N-023] resource_monitor + backpressure machinery replaced by simple caps
Category: maintainability
Area: fileproc/resource_monitor_*.go + backpressure.go (deleted), fileproc/processor.go, cli/processor_*.go, config/*
Fixed: 2026-07-02
Notes: ~1,360 prod + ~2,500 test LOC of rate limiting, graceful degradation, emergency stop, concurrent-read semaphores, hard-memory-limit checks, per-file/overall timeout contexts, and channel back-pressure — heavy for a local-file CLI. Deleted the 6 resource_monitor files + backpressure.go; slimmed `fileproc/processor.go` to size-cap validation + in-memory/streaming + ctx cancellation; replaced back-pressure channels with fixed-buffer channels (`channelBuffer=64`). Kept the real safety caps (FileSizeLimit in `file_filters`, MaxFiles/MaxTotalSize in `processor_collection`, always applied). Removed the ~40 now-orphaned config getters/validators/defaults/keys/Min-Max constants and `CodeResourceLimitTimeout`, and the backpressure/resource-limit sections of `config.example.yaml`/README/CLAUDE. Fixed a latent test-ordering bug (`TestCollectFiles` relied on another test loading config defaults) by giving it explicit setup. User-approved. Net for Pass 3: prod LOC ~8,700 → 5,437 (−37%); all `-race` tests pass, smoke-tested end-to-end.

### Pass 4 — 2026-07-02

#### [N-024] FileTypeRegistry cache was premature optimization over O(1) map lookups
Category: performance
Area: fileproc/cache.go (deleted), fileproc/registry.go, fileproc/detection.go, shared/constants.go
Fixed: 2026-07-02
Notes: The registry wrapped an extCache/resultCache + RWMutex + LRU eviction + RegistryStats layer around detection that is already a direct lookup into three small maps; `Stats()`/`CacheInfo()`/`RegistryStats` had zero non-test callers, so the whole layer was invisible to the app's output. Deleted `cache.go` (132 LOC), folded a lock-free `getFileTypeResult` into `registry.go` (maps are read-only after `ConfigureFromSettings`, so no mutex is needed), dropped the `invalidateCache()` calls from `Add*`/`Disable*`, and removed the orphaned `FileTypeRegistryMaxCacheSize` constant and the cache/Stats tests. Detection output is unchanged (smoke-tested: `.png` excluded as image, code files included). Prod LOC 5,437 → 5,281; fileproc test time ~2.0s → 1.1s; all `-race` tests pass, revive clean apart from the pre-existing package-name warning.

## Invalid

### Pass 1 — 2026-05-06

#### [N-010] `cli/processor_workers.go` parses extensions instead of using the file-type registry
Notes: On closer look, the `format` value here is a metrics breakdown label (e.g. `"py"`, `"go"`, `"txt"`), not a language identifier. The registry's `Language()` returns language names (`"python"`, `"go"`) and `""` for unmapped extensions like `.txt`, which would change the metrics output and break tests in `metrics/collector_test.go` that assert on bare extensions. The arch rule about registry encapsulation applies to language detection; this code is doing extension labeling. Reclassified as Invalid rather than fixed.

### Pass 2 — 2026-07-02

#### [N-018] Claimed data race on `fileproc/cache.go` extension-map reads
Notes: An auditor flagged unlocked reads of `imageExts`/`binaryExts`/`languageMap` in `getFileTypeResult` as a critical race. Not a race: `ConfigureFromSettings` (the only mutator, via `Add*`/`Disable*`) runs inside `NewProcessor` before `Process()` starts any worker, so the maps are read-only during concurrency and concurrent map reads are safe in Go — consistent with `-race` tests passing. Also rejected the sibling suggestion to convert the manual `RLock`/`RUnlock` in `getNormalizedExtension` to `defer`: the manual unlock is deliberate so the expensive `normalizeExtension` compute runs outside the lock. Documented the read-only-after-config invariant with a comment instead of adding locks.

## Notes

- `revive -config revive.toml ./...` reports only the pre-existing `avoid meaningless package names` on `package shared` (present on HEAD, out of scope — renaming a repo-wide package is not a requested simplification); `gofmt -l` clean; `go test -race ./...` passes after this pass — verified 2026-07-02.
- Pass 2 rejected as non-defects (silence = approval): `cli/flags.go` global flag cache (intentional, has `ResetFlags` for tests), `processor_workers.go:54` `fileCtx, fileCancel := ctx, func(){}` (idiomatic), `metrics.FinalMetrics`/`QuickStats` (unused by production but tested public API — kept), `shared.Logger` single-impl interface (testability seam — kept), `testutil` ignored `Close()` errors (test cleanup — acceptable).
- TODO.md still lists the templates feature under completed history; left intact as a historical changelog rather than rewritten.
- `kics.config` (called out as orphaned in a previous pass) was retained: deleting committed config without explicit user approval is out of scope. Either wire it into CI or remove in a follow-up.
- Deferred to a separate PR: N-009 (resource_monitor refactor — large structural change).
- Deferred but not filed as a finding here: N-013 (cmd/benchmark package-level flag vars) — quality-of-life only, not a defect.
