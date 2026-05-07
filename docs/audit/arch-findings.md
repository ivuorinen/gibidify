# Architecture Audit Findings
Generated: 2026-05-06
Last validated: 2026-05-06

## Summary
- Total: 5 | Open: 2 | Fixed: 2 | Invalid: 1

## Open Findings

### Medium

#### [ARCH-003] Resource monitoring split between `fileproc/` and `metrics/`
Category: general
Rule: arch-profile.md "Ambiguities & Contradictions"
Evidence: `fileproc/resource_monitor_*.go` (11 files) and `metrics/reporter.go` collect overlapping runtime data.
Fix: Move `resource_monitor_*` into a dedicated `resource/` package; have `metrics/` consume its public interface. Deferred — too large for a same-pass fix.

### Low

#### [ARCH-004] `cmd/benchmark/main.go` keeps flag results in package-level vars
Category: general
Rule: profile rule 11 — secondary entrypoints must not duplicate CLI flag parsing.
Evidence: Lines 15-37 declare seven package-level pointer vars whose values are only meaningful after `Parse()` runs.
Fix: Wrap the flags in a struct and parse via a function (mirroring `cli.Flags` and `cli.ParseFlags`). N-011 already corrected the parse call to error-returning form; this is the deeper structural cleanup. Not urgent.

## Fixed

### Pass 1 — 2026-05-06

#### [ARCH-001] `interfaces.go` defines orphaned interfaces and shadow types in `package main`
Fixed: 2026-05-06
Notes: Deleted `interfaces.go` (217 lines). `go build ./...` clean. The file contained no production-used symbols; the duplicate `WriteRequest`, `ResourceMetrics`, `FileProcessingResult`, `ProcessingMetrics` types are gone. Canonical `fileproc.WriteRequest` is unaffected.

#### [ARCH-005] `metrics` package imports `golang.org/x/text/cases` for cosmetic title-casing
Fixed: 2026-05-06
Notes: Replaced `cases.Title(language.English).String(phase)` with a small local `titleASCII` helper. Dropped `golang.org/x/text/cases` and `golang.org/x/text/language` imports from `metrics/reporter.go`. `templates/engine.go` retains the imports — that path renders user-supplied template variables and may legitimately need full Unicode title-casing.

## Invalid

### Pass 1 — 2026-05-06

#### [ARCH-002] `cli/processor_workers.go` parses file extension manually instead of using the registry
Notes: Reclassified after closer reading. The `format` value computed at line 163 is a metrics breakdown label (`"py"`, `"go"`, `"txt"`) — not a language identifier. Switching to `fileproc.Language()` would convert `"py"` → `"python"`, `"js"` → `"javascript"`, and return `""` for unmapped extensions like `.txt`, breaking the assertions in `metrics/collector_test.go:189-191, 290-291, 340`. The arch rule about registry encapsulation applies to language detection; this code is doing extension labeling.
