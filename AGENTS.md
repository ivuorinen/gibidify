# AGENTS

This repo is a Go CLI that aggregates code files into a single text output. The
main entry point is `main.go` with packages under `config` and `fileproc`.
Tests exist for each package, and CI workflows live in `.github/workflows`.

## Contributions
- Look for additional `AGENTS.md` files under `.github` first.
- Use Semantic Commit messages and PR titles.
- Run `go test ./...` and linting for code changes. Docs-only changes skip this.
- Use Yarn if installing Node packages.
- Follow `.editorconfig` and formatting via pre-commit.
