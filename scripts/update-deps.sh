#!/bin/bash

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Track overall exit status
exit_code=0

echo -e "${BLUE}=== Updating Go Dependencies ===${NC}"

# Function to handle rollback if needed
rollback() {
  if [[ -f go.mod.backup && -f go.sum.backup ]]; then
    echo -e "${YELLOW}Rolling back changes due to errors...${NC}"
    mv go.mod.backup go.mod
    mv go.sum.backup go.sum
    echo -e "${GREEN}Rollback completed${NC}"
  fi
}

# Function to cleanup backup files
cleanup() {
  if [[ -f go.mod.backup ]]; then
    rm go.mod.backup
  fi
  if [[ -f go.sum.backup ]]; then
    rm go.sum.backup
  fi
}

# Trap to ensure cleanup on exit
trap cleanup EXIT

echo "Creating backup of go.mod and go.sum..."
cp go.mod go.mod.backup
cp go.sum go.sum.backup

echo "Checking current module status..."
if ! go mod verify; then
  echo -e "${RED}Current module verification failed${NC}"
  exit_code=1
  exit $exit_code
fi

echo "Updating dependencies with 'go get -u'..."
if ! go get -u ./...; then
  echo -e "${RED}Failed to update dependencies${NC}"
  rollback
  exit_code=1
  exit $exit_code
fi

echo "Running 'go mod tidy'..."
if ! go mod tidy; then
  echo -e "${RED}Failed to tidy module dependencies${NC}"
  rollback
  exit_code=1
  exit $exit_code
fi

echo "Verifying updated dependencies..."
if ! go mod verify; then
  echo -e "${RED}Module verification failed after updates${NC}"
  rollback
  exit_code=1
  exit $exit_code
fi

echo "Running vulnerability check..."
if command -v govulncheck >/dev/null 2>&1; then
  if ! govulncheck ./...; then
    echo -e "${YELLOW}Vulnerability check failed - review output above${NC}"
    echo -e "${YELLOW}Consider updating specific vulnerable packages or pinning versions${NC}"
    # Don't fail the script for vulnerabilities, just warn
  fi
else
  echo -e "${YELLOW}govulncheck not found - install with: go install golang.org/x/vuln/cmd/govulncheck@latest${NC}"
fi

echo "Running basic build test..."
if ! go build ./...; then
  echo -e "${RED}Build failed after dependency updates${NC}"
  rollback
  exit_code=1
  exit $exit_code
fi

echo "Running quick test to ensure functionality..."
if ! go test -short ./...; then
  echo -e "${RED}Tests failed after dependency updates${NC}"
  rollback
  exit_code=1
  exit $exit_code
fi

if [[ $exit_code -eq 0 ]]; then
  echo -e "${GREEN}Dependencies updated successfully!${NC}"
  echo -e "${GREEN}Review the changes with 'git diff' before committing${NC}"
  cleanup
else
  echo -e "${RED}Dependency update failed${NC}"
fi

exit $exit_code
