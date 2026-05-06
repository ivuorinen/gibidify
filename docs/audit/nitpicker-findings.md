# Nitpicker Findings
Generated: 2026-05-06
Last validated: 2026-05-06

This pass ran in `security + docs + architecture` mode. Critical/High findings from
the specialist auditors were incorporated; the per-tool reports live in
`docs/audit/security-findings.md`, `docs/audit/doc-findings.md`,
`docs/audit/arch-findings.md`, and `docs/audit/arch-profile.md`.

## Summary
- Total: 13 | Open: 1 | Fixed: 11 | Invalid: 1
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

## Invalid

### Pass 1 — 2026-05-06

#### [N-010] `cli/processor_workers.go` parses extensions instead of using the file-type registry
Notes: On closer look, the `format` value here is a metrics breakdown label (e.g. `"py"`, `"go"`, `"txt"`), not a language identifier. The registry's `Language()` returns language names (`"python"`, `"go"`) and `""` for unmapped extensions like `.txt`, which would change the metrics output and break tests in `metrics/collector_test.go` that assert on bare extensions. The arch rule about registry encapsulation applies to language detection; this code is doing extension labeling. Reclassified as Invalid rather than fixed.

## Notes

- `make lint` exits 0 and `go test -race ./...` passes after this pass — verified.
- `kics.config` (called out as orphaned in the previous pass) was retained: deleting committed config without explicit user approval is out of scope. Either wire it into CI or remove in a follow-up.
- Deferred to a separate PR: N-009 (resource_monitor refactor — large structural change).
- Deferred but not filed as a finding here: N-013 (cmd/benchmark package-level flag vars) — quality-of-life only, not a defect.
