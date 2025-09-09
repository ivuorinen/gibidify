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

echo "Running gosec in parallel..."
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
if ! eclint check; then
  exit_code=1
fi

# Exit with failure status if any linter failed
exit $exit_code
