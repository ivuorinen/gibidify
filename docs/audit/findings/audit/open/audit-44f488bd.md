---
id: audit-44f488bd
auditor: audit
severity: low
category: maintainability
area: go.mod
status: open
found: 2026-10-12
location: go.mod:9
location_sha: d15c784d0baf9d80
---

# Ignore-file matching depends on go-gitignore, unchanged since 2021

## Problem

`github.com/sabhiram/go-gitignore` is required at pseudo-version
`v0.0.0-20210923224102-525f6e181f06`; `go list -m -versions` shows no tagged
releases and `go list -m -u` reports no newer version. It implements the
matching that decides what goes into the output.

## Evidence

go.mod:9 `github.com/sabhiram/go-gitignore v0.0.0-20210923224102-525f6e181f06`;
`go list -m -u -json` → `"Time": "2021-09-23T22:41:02Z"`, no `Update` field.

## Impact

gitignore semantics bugs (and directory-pattern handling, see the
gitignored-directories finding) will not be fixed upstream; the dependency is
effectively vendored code without the ability to patch it.

## Fix

Switch to a maintained matcher, e.g.
`github.com/go-git/go-git/v5/plumbing/format/gitignore` (used by go-git,
supports directory patterns and nested files via `ReadPatterns`), or vendor
go-gitignore under `internal/` with tests covering the patterns the walker
relies on.

## Skipped (2026-10-12 fix pass)

Not applied: both options add more than they remove. go-git pulls a large module
graph (go-billy and friends) for one package; vendoring copies third-party code
into the repo to maintain. The behaviour the walker relies on is now pinned by
tests instead — directory patterns via the trailing-slash form
(TestProdWalkerPrunesGitignoredDirectories), file patterns and negation
(TestProdWalkerWithIgnore) — so an upstream semantics problem or a replacement
would be caught. Revisit if a gitignore pattern the project needs fails one of
those tests.
