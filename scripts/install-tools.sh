#!/usr/bin/env bash
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

# Check if required tools are installed
check_dependencies() {
  print_status "Checking dependencies..."

  local missing_tools=()

  if ! command -v go &>/dev/null; then
    missing_tools+=("go")
  fi

  # Check that tools are installed:

  if [[ ${#missing_tools[@]} -ne 0 ]]; then
    print_error "Missing required tools: ${missing_tools[*]}"
    print_error "Please install the missing tools and try again."
    exit 1
  fi

  # Security tools

  if ! command -v gosec &>/dev/null; then
    print_warning "gosec not found, installing..."
    go install github.com/securego/gosec/v2/cmd/gosec@v2.22.8
  fi

  if ! command -v govulncheck &>/dev/null; then
    print_warning "govulncheck not found, installing..."
    go install golang.org/x/vuln/cmd/govulncheck@v1.1.4
  fi

  # Linting tools

  if ! command -v revive &>/dev/null; then
    print_warning "revive not found, installing..."
    go install github.com/mgechev/revive@v1.11.0
  fi

  if ! command -v gocyclo &>/dev/null; then
    print_warning "gocyclo not found, installing..."
    go install github.com/fzipp/gocyclo/cmd/gocyclo@v0.6.0
  fi

  if ! command -v checkmake &>/dev/null; then
    print_warning "checkmake not found, installing..."
    go install github.com/mrtazz/checkmake/cmd/checkmake@v0.2.2
  fi

  if ! command -v eclint &>/dev/null; then
    print_warning "eclint not found, installing..."
    go install gitlab.com/greut/eclint/cmd/eclint@v0.5.1
  fi

  if ! command -v staticcheck &>/dev/null; then
    print_warning "staticcheck not found, installing..."
    go install honnef.co/go/tools/cmd/staticcheck@v0.6.1
  fi

  # Formatting tools

  if ! command -v gofumpt &>/dev/null; then
    print_warning "gofumpt not found, installing..."
    go install mvdan.cc/gofumpt@v0.8.0
  fi

  if ! command -v goimports &>/dev/null; then
    print_warning "goimports not found, installing..."
    go install golang.org/x/tools/cmd/goimports@v0.36.0
  fi

  if ! command -v shfmt &>/dev/null; then
    print_warning "shfmt not found, installing..."
    go install mvdan.cc/sh/v3/cmd/shfmt@v3.12.0
  fi

  if ! command -v yamlfmt &>/dev/null; then
    print_warning "yamlfmt not found, installing..."
    go install github.com/google/yamlfmt/cmd/yamlfmt@v0.4.0
  fi

  print_success "All dependencies are available"
  return 0
}

# ---

# If this file is sourced, export the functions
if [[ "${BASH_SOURCE[0]}" != "${0}" ]]; then
  export -f check_dependencies print_error print_warning print_success print_status
fi

# if this file is executed, execute the function
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
  SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
  PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

  cd "$PROJECT_ROOT" || {
    echo "Failed to change directory to $PROJECT_ROOT"
    exit 1
  }

  echo "Installing dev tools for gibidify..."

  check_dependencies
fi
