# Replace check_secrets with gitleaks

## Problem

The `check_secrets` function in `scripts/security-scan.sh` uses hand-rolled regex
patterns that produce false positives. The pattern `key\s*[:=]\s*['"][^'"]{8,}['"]`
matches every `configKey: "backpressure.maxPendingFiles"` line in
`config/getters_test.go` (40+ matches), causing `make security-full` to fail.

The git history check (`git log --oneline -10 | grep -i "key|token"`) also matches
on benign commit messages containing words like "key" or "token".

## Decision

Replace the custom `check_secrets` function with
[gitleaks](https://github.com/gitleaks/gitleaks), a widely adopted Go-based secret
scanner with built-in rules for AWS keys, GitHub tokens, private keys, high-entropy
strings, and more.

## Approach

- **Drop-in replacement**: Only the `check_secrets` function body changes. The
  function signature and return behavior (0 = clean, 1 = findings) remain identical.
- **`go run` invocation**: Use `go run github.com/gitleaks/gitleaks/v8@latest` so
  the tool is fetched automatically if not cached. No changes to `install-tools.sh`.
- **Working tree scan only**: Use `gitleaks dir` to scan current files. No git
  history scanning (matches current script behavior scope).
- **Config file**: A `.gitleaks.toml` at the project root extends gitleaks' built-in
  rules with an allowlist to suppress known false positives in test files.
- **CI unaffected**: `.github/workflows/security.yml` runs its own inline steps
  (gosec, govulncheck, checkmake, shfmt, yamllint, Trivy) and does not call
  `security-scan.sh` or `check_secrets`.

## Files Changed

| File | Change |
|------|--------|
| `scripts/security-scan.sh` | Replace `check_secrets` function body |
| `.gitleaks.toml` | New file -- gitleaks configuration with allowlist |

## Verification

```bash
make security-full   # should pass end-to-end
```
