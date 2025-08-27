#!/usr/bin/env bash

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

cd "$PROJECT_ROOT" || {
  echo "Failed to change directory to $PROJECT_ROOT"
  exit 1
}

# shellcheck source=scripts/install-tools.sh
source "$SCRIPT_DIR/install-tools.sh"

check_dependencies

echo "Linting..."

# Track overall exit status
exit_code=0

echo "Running revive..."
if ! revive -config revive.toml -formatter friendly -set_exit_status ./...; then
  exit_code=1
fi

echo "Running gosec..."
if ! gosec -fmt=text -quiet ./...; then
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
if ! yamllint -c .yamllint .github/workflows/*.yml ./*.yaml .yamllint; then
  exit_code=1
fi

echo "Running eclint..."
if ! eclint; then
  exit_code=1
fi

# Exit with failure status if any linter failed
exit $exit_code
