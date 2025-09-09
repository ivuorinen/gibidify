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

# shellcheck source=scripts/install-tools.sh
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
echo "Running gosec security linter in parallel..."
run_gosec_parallel() {
  local exit_code=0
  local pids=()
  local go_dirs=("./benchmark" "./cli" "./cmd" "./config" "./fileproc" "./metrics" "./templates" "./testutil" "./utils" ".")

  # Start gosec for each directory in background
  for dir in "${go_dirs[@]}"; do
    if [[ "$dir" == "." ]]; then
      # For root directory, scan only .go files directly (not subdirectories)
      gosec -fmt=text -quiet -exclude-dir=benchmark -exclude-dir=cli -exclude-dir=cmd -exclude-dir=config -exclude-dir=fileproc -exclude-dir=metrics -exclude-dir=templates -exclude-dir=testutil -exclude-dir=utils . >"gosec_${dir//\//_}.log" 2>&1 &
    else
      gosec -fmt=text -quiet "$dir" >"gosec_${dir//\//_}.log" 2>&1 &
    fi
    pids+=($!)
  done

  # Wait for all gosec processes to complete and check their exit codes
  for i in "${!pids[@]}"; do
    local pid="${pids[$i]}"
    local dir="${go_dirs[$i]}"
    if ! wait "$pid"; then
      echo "gosec failed for directory: $dir"
      cat "gosec_${dir//\//_}.log"
      exit_code=1
    fi
    # Clean up log file if successful
    rm -f "gosec_${dir//\//_}.log"
  done

  return $exit_code
}

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
eclint fix
