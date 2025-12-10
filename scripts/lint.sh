#!/usr/bin/env bash

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

echo "Linting..."

# Track overall exit status
exit_code=0

echo "Running revive..."
if ! revive -config revive.toml -formatter friendly -set_exit_status ./...; then
  exit_code=1
fi

echo "Running gosec in parallel..."

if ! run_gosec_parallel; then
  exit_code=1
fi

echo "Running checkmake..."
if ! checkmake --config=.checkmake Makefile; then
  exit_code=1
fi

echo "Running shfmt check..."
if ! shfmt -d .; then
  exit_code=1
fi

echo "Running yamllint..."
if command -v yamllint >/dev/null 2>&1; then
  # Python yamllint supports the .yamllint config; lint the whole repo
  if ! yamllint -c .yamllint .; then
    exit_code=1
  fi
elif command -v yaml-lint >/dev/null 2>&1; then
  # Go yaml-lint has different flags and no .yamllint support; use its defaults
  if ! yaml-lint .; then
    exit_code=1
  fi
else
  echo "YAML linter not found (yamllint or yaml-lint); skipping."
fi

echo "Running editorconfig-checker..."
if ! eclint *.go *.md benchmark/ cli/ cmd/ config/ fileproc/ gibidiutils/ metrics/ shared/ templates/ testutil/ scripts/ Makefile; then
  exit_code=1
fi

# Exit with failure status if any linter failed
exit $exit_code
