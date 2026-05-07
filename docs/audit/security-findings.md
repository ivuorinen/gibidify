# Security Audit Findings
Generated: 2026-05-06
Last validated: 2026-05-06
Pass: 2

## Tool Coverage
- Available: semgrep, opengrep, grype, trivy, gitleaks, checkov, gosec, snyk
- Not available: npm, yarn, pnpm (not applicable — no Node lockfile)
- Errored: none

## Summary
Total: 6 | Open: 1 | Fixed: 5 | Invalid: 0
Critical: 0 | High: 0 | Medium: 0 | Low: 0 | Advisory: 1

## Open Findings

### Advisory

#### [SEC-006] `text/template` import used for non-HTML output
Category: sast
Tool: semgrep
Source: templates/engine.go:9
CVE/Rule: go.lang.security.audit.xss.import-text-template.import-text-template
Problem: Rule warns that `text/template` does not auto-escape; if rendered into HTML, this is XSS.
Evidence: `import "text/template"` in templates/engine.go.
Impact: None in this codebase — output is rendered into Markdown/JSON/YAML aggregations consumed by LLMs, never into a browser.
Fix: No code change required. Optionally suppress with `// nosemgrep: …` and a one-line rationale.

## Fixed

### Pass 2 — 2026-05-06

#### [SEC-001] GitHub Actions shell injection via `github.ref_name` interpolation
Notes: Bound `github.ref_name`, `matrix.goos`, `matrix.goarch`, and the windows ext to a step `env:` block; rewrote the `run:` script to reference `$REF_NAME`, `$GOOS`, `$GOARCH`, `$EXT`. Fixed in `.github/workflows/build-test-publish.yml` build step.

#### [SEC-002] GitHub Actions shell injection in Docker publish job
Notes: Bound `github.token`, `github.actor`, `github.repository`, `github.ref_name` to step `env:` blocks; replaced expression interpolation in `run:` scripts with shell variables. Fixed in `.github/workflows/build-test-publish.yml` docker job.

#### [SEC-003] Dockerfile uses `useradd` on Alpine — image build fails
Notes: Replaced with `adduser -D -s /bin/sh gibidify`; switched `COPY` + `chmod` to `COPY --chmod=0755`. Image now builds.

#### [SEC-004] Stale `gibidify-benchmark` build artifact pinned to vulnerable Go stdlib
Notes: Removed local artifacts (`gibidify`, `gibidify-benchmark`, `coverage.out`, `coverage_backup.out`). All gitignored, so working-tree-only deletion suffices. Re-running grype now scans only `go.sum`.

#### [SEC-005] Dockerfile missing `HEALTHCHECK`
Notes: Added `HEALTHCHECK NONE` — explicit declaration that this CLI image has nothing to health-check. checkov rule CKV_DOCKER_2 satisfied.

## Invalid

### Pass 1 — 2026-05-06

#### [SEC-INVALID-A] Bash IFS tampering in `scripts/lint-fix.sh`
Notes: Semgrep flagged `IFS=$'\n\t'` at line 5 as global IFS tampering. This is the documented "unofficial bash strict mode" idiom (Aaron Maxwell), set once at script top after `set -euo pipefail`. Not a defect; not filed as a finding.
