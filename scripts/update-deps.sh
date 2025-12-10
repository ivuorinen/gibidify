#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

cd "$PROJECT_ROOT" || {
  echo "Failed to change directory to $PROJECT_ROOT"
  exit 1
}

# shellcheck source=scripts/install-tools.sh
source "$SCRIPT_DIR/install-tools.sh"

# Track overall exit status
exit_code=0

print_status "=== Updating Go Dependencies ==="

# Function to handle rollback if needed
rollback() {
  if [[ -f go.mod.backup && -f go.sum.backup ]]; then
    print_warning "Rolling back changes due to errors..."
    mv go.mod.backup go.mod
    mv go.sum.backup go.sum
    print_success "Rollback completed"
  fi
  return 0
}

# Function to cleanup backup files
cleanup() {
  if [[ -f go.mod.backup ]]; then
    rm go.mod.backup
  fi
  if [[ -f go.sum.backup ]]; then
    rm go.sum.backup
  fi
  return 0
}

# Trap to ensure cleanup on exit
trap cleanup EXIT

print_status "Creating backup of go.mod and go.sum..."
cp go.mod go.mod.backup
cp go.sum go.sum.backup

print_status "Checking current module status..."
if ! go mod verify; then
  print_error "Current module verification failed"
  exit_code=1
  exit $exit_code
fi

print_status "Updating dependencies with 'go get -u'..."
if ! go get -u ./...; then
  print_error "Failed to update dependencies"
  rollback
  exit_code=1
  exit $exit_code
fi

print_status "Running 'go mod tidy'..."
if ! go mod tidy; then
  print_error "Failed to tidy module dependencies"
  rollback
  exit_code=1
  exit $exit_code
fi

print_status "Verifying updated dependencies..."
if ! go mod verify; then
  print_error "Module verification failed after updates"
  rollback
  exit_code=1
  exit $exit_code
fi

print_status "Running vulnerability check..."
if command -v govulncheck >/dev/null 2>&1; then
  if ! govulncheck ./...; then
    print_warning "Vulnerability check failed - review output above"
    print_warning "Consider updating specific vulnerable packages or pinning versions"
    # Don't fail the script for vulnerabilities, just warn
  fi
else
  print_warning "govulncheck not found - install with: go install golang.org/x/vuln/cmd/govulncheck@latest"
fi

print_status "Running basic build test..."
if ! go build ./...; then
  print_error "Build failed after dependency updates"
  rollback
  exit_code=1
  exit $exit_code
fi

print_status "Running quick test to ensure functionality..."
if ! go test -short ./...; then
  print_error "Tests failed after dependency updates"
  rollback
  exit_code=1
  exit $exit_code
fi

if [[ $exit_code -eq 0 ]]; then
  print_success "Dependencies updated successfully!"
  print_success "Review the changes with 'git diff' before committing"
  cleanup
else
  print_error "Dependency update failed"
fi

exit $exit_code
