#!/usr/bin/env bash

# Enable strict error handling
set -euo pipefail
IFS=$'\n\t'

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

cd "$PROJECT_ROOT" || {
  echo "Failed to change directory to $PROJECT_ROOT"
  exit 1
}

source "$SCRIPT_DIR/install-tools.sh"

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
echo "Running gosec security linter..."
gosec -fmt=text -quiet ./...
echo "Auto-fix completed. Running final lint check..."
revive -config revive.toml -formatter friendly -set_exit_status ./...
gosec -fmt=text -quiet ./...
echo "Running checkmake..."
checkmake --config=.checkmake Makefile
echo "Running yamlfmt..."
yamlfmt -conf .yamlfmt.yml -gitignore_excludes -dstar ./**/*.{yaml,yml}
echo "Running eclint fix..."
eclint -fix
