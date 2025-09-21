#!/usr/bin/env bash

# Shared security scanning functions

# Run gosec in parallel on all Go directories
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
      # Keep log for inspection/artifacts on failure
      exit_code=1
    else
      # Clean up log file if successful
      rm -f "gosec_${dir//\//_}.log"
    fi
  done

  return $exit_code
}

# If this file is sourced, export the functions
if [[ "${BASH_SOURCE[0]}" != "${0}" ]]; then
  export -f run_gosec_parallel
fi
