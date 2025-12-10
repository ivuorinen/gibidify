#!/usr/bin/env bash

# Gosec security scanner script for individual Go files
# Runs gosec on each Go directory and reports issues per file

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# If NO_COLOR is set, disable colors
if [[ -n "${NO_COLOR:-}" ]]; then
  RED=''
  GREEN=''
  YELLOW=''
  BLUE=''
  NC=''
fi

# Function to print status
print_status() {
  local msg="$1"
  echo -e "${BLUE}[INFO]${NC} $msg"
  return 0
}

print_warning() {
  local msg="$1"
  echo -e "${YELLOW}[WARN]${NC} $msg" >&2
  return 0
}

print_error() {
  local msg="$1"
  echo -e "${RED}[ERROR]${NC} $msg" >&2
  return 0
}

print_success() {
  local msg="$1"
  echo -e "${GREEN}[SUCCESS]${NC} $msg"
  return 0
}

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

cd "$PROJECT_ROOT" || {
  print_error "Failed to change directory to $PROJECT_ROOT"
  exit 1
}

# Check if gosec is available
if ! command -v gosec &>/dev/null; then
  print_error "gosec not found. Please install it first:"
  print_error "go install github.com/securego/gosec/v2/cmd/gosec@latest"
  exit 1
fi

# Check if jq is available
if ! command -v jq &>/dev/null; then
  print_error "jq not found. Please install it first:"
  print_error "brew install jq  # on macOS"
  print_error "apt-get install jq  # on Ubuntu/Debian"
  exit 1
fi

# Get all Go files and unique directories
GO_FILES=$(find . -name "*.go" -not -path "./.*" | sort)
TOTAL_FILES=$(echo "$GO_FILES" | wc -l | tr -d ' ')

DIRECTORIES=$(echo "$GO_FILES" | xargs -n1 dirname | sort -u)
TOTAL_DIRS=$(echo "$DIRECTORIES" | wc -l | tr -d ' ')

print_status "Found $TOTAL_FILES Go files in $TOTAL_DIRS directories"
print_status "Running gosec security scan..."

ISSUES_FOUND=0
FILES_WITH_ISSUES=0
CURRENT_DIR=0

# Create a temporary directory for reports
TEMP_DIR=$(mktemp -d)
trap 'rm -rf "$TEMP_DIR"' EXIT

# Process each directory
while IFS= read -r dir; do
  CURRENT_DIR=$((CURRENT_DIR + 1))
  echo -ne "\r${BLUE}[PROGRESS]${NC} Scanning $CURRENT_DIR/$TOTAL_DIRS: $dir                    "

  # Run gosec on the directory
  REPORT_FILE="$TEMP_DIR/$(echo "$dir" | tr '/' '_' | tr '.' '_').json"
  if gosec -fmt=json "$dir" >"$REPORT_FILE" 2>/dev/null; then
    # Check for issues in all files in this directory
    ISSUES=$(jq -r '.Issues // [] | length' "$REPORT_FILE" 2>/dev/null || echo "0")

    if [[ "$ISSUES" -gt 0 ]]; then
      echo # New line after progress
      print_warning "Found $ISSUES security issue(s) in directory $dir:"

      # Group issues by file and display them
      jq -r '.Issues[] | "\(.file)|\(.rule_id)|\(.details)|\(.line)"' "$REPORT_FILE" 2>/dev/null | while IFS='|' read -r file rule details line; do
        if [[ -n "$file" ]]; then
          # Only count each file once
          if ! grep -q "$file" "$TEMP_DIR/processed_files.txt" 2>/dev/null; then
            echo "$file" >>"$TEMP_DIR/processed_files.txt"
            FILES_WITH_ISSUES=$((FILES_WITH_ISSUES + 1))
          fi
          echo "  $file:$line → $rule: $details"
        fi
      done

      ISSUES_FOUND=$((ISSUES_FOUND + ISSUES))
      echo
    fi
  else
    echo # New line after progress
    print_error "Failed to scan directory $dir"
  fi
done <<<"$DIRECTORIES"

echo # Final new line after progress

# Count actual files with issues
if [[ -f "$TEMP_DIR/processed_files.txt" ]]; then
  FILES_WITH_ISSUES=$(wc -l <"$TEMP_DIR/processed_files.txt" | tr -d ' ')
fi

# Summary
print_status "Gosec scan completed!"
print_status "Directories scanned: $TOTAL_DIRS"
print_status "Files scanned: $TOTAL_FILES"

if [[ $ISSUES_FOUND -eq 0 ]]; then
  print_success "No security issues found! 🎉"
  exit 0
else
  print_warning "Found $ISSUES_FOUND security issue(s) in $FILES_WITH_ISSUES file(s)"
  print_status "Review the issues above and fix them before proceeding"
  exit 1
fi
