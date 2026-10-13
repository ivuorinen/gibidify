---
id: audit-ae75544d
auditor: audit
severity: low
category: conventions
area: git history
status: open
found: 2026-10-12
location: .goreleaser.yaml:238-260
location_sha: 7721bfa93ddd2d42
---

# Commit types misstate impact: breaking removal and toolchain bump

## Problem

A user-facing config removal was typed `refactor` without `!`, and a toolchain
bump was typed `feat`. 658eec6 "refactor: simplify gibidify down to its core
purpose (#245)" removed user-visible configuration (templates, metrics,
backpressure, resource-monitor timeouts, `filePatterns`, `supportedFormats`)
and streaming behavior, with no `!` or BREAKING footer. 4559d1b "feat: update
Go version from 1.25.6 to 1.25.7 (#172)" changes only `.go-version`.
goreleaser groups the changelog by these prefixes.

## Evidence

`git show 658eec6 --stat`: 80 files changed, 846 insertions(+), 11916
deletions(-); body: "Remove supportedFormats and filePatterns config entirely",
"remove metrics, templates, and benchmark subsystems". `git show 4559d1b
--stat`: `.go-version | 2 +-`. .goreleaser.yaml:250-258 maps `^feat` → Features,
`^refactor` → Refactoring.

## Impact

The next release's changelog will list a toolchain bump as a feature and hide
the removal of documented settings under "Refactoring"; users upgrading get no
warning that their config keys stopped working.

## Fix

History is not rewritten; record the removal in the next release notes
explicitly (a "Breaking changes" section listing the removed keys). Going
forward, use `refactor!:`/`BREAKING CHANGE:` for removals of documented behavior
and `chore(go):`/`build:` for toolchain bumps; the pr-lint action can enforce
this with a commitlint rule.

## Skipped (2026-10-12 fix pass)

Not applicable in a code change: the fix is a release-time action (write a
"Breaking changes" section in the next release notes) plus commit discipline
going forward. The 2026-10-12 fix pass adds its own user-visible changes for
those notes: the CWD config search path was removed, `-source`/`-destination`
accept `..`, `--verbose` now raises the log level, failures exit non-zero, YAML
scalars are always quoted, and symlinks/NUL-byte files are skipped. Close this
finding when the release notes ship.

## Skipped (2026-10-13 fix pass)

Still a release-time action. The 2026-10-13 fix pass adds these user-visible
changes for the same notes, the first three breaking:

- An unreadable, unparseable or invalid config file stops the run (exit 1)
  instead of falling back to defaults; dot-prefixed `ignoreDirectories` entries
  are accepted.
- A positional argument is an error instead of silently ending flag parsing.
- Credential files (`.env*`, `*.pem`, `*.key`, `id_rsa*`, `.npmrc`, …) are
  skipped unless `--include-sensitive` is given; `.git`/`.hg`/`.svn` are always
  skipped and `.git` left the default `ignoreDirectories` list.
- The output is written to a temporary file and renamed on success; an
  interrupted run keeps the previous output and exits 130. New output files
  are created 0600 (previously 0666 minus umask).
- YAML output is produced by yaml.v3 (always parseable; key order path,
  content, language); JSON no longer escapes `<`, `>`, `&`; invalid UTF-8 is
  replaced with U+FFFD and warned about.
- A deeper `.gitignore`'s `!pattern` re-includes files; an unreadable ignore
  file skips its directory (or fails the walk at the root).
- Oversized and credential files are logged and counted as skipped.
- Release images are one multi-platform manifest per tag (dockers_v2) and ship
  `THIRD_PARTY_NOTICES`.
