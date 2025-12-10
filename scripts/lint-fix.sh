#!/usr/bin/env bash

# Enable strict error handling
set -euo pipefail
IFS=$'\n\t'
shopt -s globstar nullglob

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

cd "$PROJECT_ROOT" || {
  echo "Failed to change directory to $PROJECT_ROOT"
  exit 1
}

# shellcheck source=scripts/install-tools.sh
source "$SCRIPT_DIR/install-tools.sh"
# shellcheck source=scripts/security.sh
source "$SCRIPT_DIR/security.sh"

check_dependencies

echo "Running gofumpt..."
gofumpt -l -w .
echo "Running goimports..."
goimports -w -local github.com/ivuorinen/gibidify .
echo "Running go fmt..."
go fmt ./...
echo "Running go mod tidy..."
go mod tidy
echo "Running shfmt formatting..."
shfmt -w -i 2 -ci .
echo "Running revive linter..."
revive -config revive.toml -formatter friendly -set_exit_status ./...
echo "Running gosec security linter in parallel..."
if ! run_gosec_parallel; then
  echo "gosec found security issues"
  exit 1
fi
echo "Auto-fix completed. Running final lint check..."
revive -config revive.toml -formatter friendly -set_exit_status ./...
if ! run_gosec_parallel; then
  echo "Final gosec check found security issues"
  exit 1
fi
echo "Running checkmake..."
checkmake --config=.checkmake Makefile
echo "Running yamlfmt..."
yamlfmt -conf .yamlfmt.yml -gitignore_excludes -dstar ./**/*.{yaml,yml}
echo "Running eclint fix..."
eclint -fix *.go *.md benchmark/ cli/ cmd/ config/ fileproc/ gibidiutils/ metrics/ shared/ templates/ testutil/ scripts/ Makefile
