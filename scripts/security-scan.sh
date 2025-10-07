#!/bin/sh
set -eu

# Security Scanning Script for gibidify
# This script runs comprehensive security checks locally and in CI

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

cd "$PROJECT_ROOT"

echo "🔒 Starting comprehensive security scan for gibidify..."

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print status
print_status() {
  printf "${BLUE}[INFO]${NC} %s\n" "$1"
}

print_warning() {
  printf "${YELLOW}[WARN]${NC} %s\n" "$1"
}

print_error() {
  printf "${RED}[ERROR]${NC} %s\n" "$1"
}

print_success() {
  printf "${GREEN}[SUCCESS]${NC} %s\n" "$1"
}

# Run command with timeout if available, otherwise run directly
# Usage: run_with_timeout DURATION COMMAND [ARGS...]
run_with_timeout() {
  duration="$1"
  shift

  if command -v timeout >/dev/null 2>&1; then
    timeout "$duration" "$@"
  else
    # timeout not available, run command directly
    "$@"
  fi
}

# Check if required tools are installed
check_dependencies() {
  print_status "Checking security scanning dependencies..."

  missing_tools=""

  if ! command -v go >/dev/null 2>&1; then
    missing_tools="${missing_tools}go "
    print_error "Go is not installed. Please install Go first."
    print_error "Visit https://golang.org/doc/install for installation instructions."
    exit 1
  fi

  if ! command -v golangci-lint >/dev/null 2>&1; then
    print_warning "golangci-lint not found, installing..."
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
  fi

  if ! command -v gosec >/dev/null 2>&1; then
    print_warning "gosec not found, installing..."
    go install github.com/securego/gosec/v2/cmd/gosec@latest
  fi

  if ! command -v govulncheck >/dev/null 2>&1; then
    print_warning "govulncheck not found, installing..."
    go install golang.org/x/vuln/cmd/govulncheck@latest
  fi

  if ! command -v checkmake >/dev/null 2>&1; then
    print_warning "checkmake not found, installing..."
    go install github.com/checkmake/checkmake/cmd/checkmake@latest
  fi

  if ! command -v shfmt >/dev/null 2>&1; then
    print_warning "shfmt not found, installing..."
    go install mvdan.cc/sh/v3/cmd/shfmt@latest
  fi

  if ! command -v yamllint >/dev/null 2>&1; then
    print_warning "yamllint not found, attempting to install..."

    # Update PATH to include common user install directories
    export PATH="$HOME/.local/bin:$HOME/.cargo/bin:$PATH"

    installed=0

    # Try pipx first
    if command -v pipx >/dev/null 2>&1; then
      print_status "Attempting install with pipx..."
      if pipx install yamllint; then
        # Update PATH to include pipx bin directory
        export PATH="$(pipx environment --value PIPX_BIN_DIR 2>/dev/null || echo "$HOME/.local/bin"):$PATH"
        installed=1
      else
        print_warning "pipx install yamllint failed, trying next method..."
      fi
    fi

    # Try pip3 --user if pipx didn't work
    if [ "$installed" -eq 0 ] && command -v pip3 >/dev/null 2>&1; then
      print_status "Attempting install with pip3 --user..."
      if pip3 install --user yamllint; then
        installed=1
      else
        print_warning "pip3 install yamllint failed, trying next method..."
      fi
    fi

    # Try apt-get with smart sudo handling
    if [ "$installed" -eq 0 ] && command -v apt-get >/dev/null 2>&1; then
      sudo_cmd=""
      can_use_apt=false

      # Check if running as root
      if [ "$(id -u)" -eq 0 ]; then
        print_status "Running as root, no sudo needed for apt-get..."
        sudo_cmd=""
        can_use_apt=true
      elif command -v sudo >/dev/null 2>&1; then
        # Try non-interactive sudo first
        if sudo -n true 2>/dev/null; then
          print_status "Attempting install with apt-get (sudo cached)..."
          sudo_cmd="sudo"
          can_use_apt=true
        elif [ -t 0 ]; then
          # TTY available, allow interactive sudo
          print_status "Attempting install with apt-get (may prompt for sudo)..."
          sudo_cmd="sudo"
          can_use_apt=true
        else
          print_warning "apt-get available but sudo not accessible (non-interactive, no cache). Skipping apt-get."
          can_use_apt=false
        fi
      else
        print_warning "apt-get available but sudo not found. Skipping apt-get."
        can_use_apt=false
      fi

      # Attempt apt-get only if we have permission to use it
      if [ "$can_use_apt" = true ]; then
        if [ -n "$sudo_cmd" ]; then
          if run_with_timeout 300 ${sudo_cmd:+"$sudo_cmd"} apt-get update; then
            if run_with_timeout 300 ${sudo_cmd:+"$sudo_cmd"} apt-get install -y yamllint; then
              installed=1
            else
              print_warning "apt-get install yamllint failed or timed out"
            fi
          else
            print_warning "apt-get update failed or timed out"
          fi
        else
          # Running as root without sudo
          if run_with_timeout 300 apt-get update; then
            if run_with_timeout 300 apt-get install -y yamllint; then
              installed=1
            else
              print_warning "apt-get install yamllint failed or timed out"
            fi
          else
            print_warning "apt-get update failed or timed out"
          fi
        fi
      fi
    fi

    # Final check with updated PATH
    if ! command -v yamllint >/dev/null 2>&1; then
      print_error "yamllint installation failed or yamllint still not found in PATH."
      print_error "Please install yamllint manually using one of:"
      print_error "  - pipx install yamllint"
      print_error "  - pip3 install --user yamllint"
      print_error "  - sudo apt-get install yamllint (Debian/Ubuntu)"
      print_error "  - brew install yamllint (macOS)"
      exit 1
    fi

    print_status "yamllint successfully installed and found in PATH"
  fi

  if [ -n "$missing_tools" ]; then
    print_error "Missing required tools: $missing_tools"
    print_error "Please install the missing tools and try again."
    exit 1
  fi

  print_success "All dependencies are available"
}

# Run gosec security scanner
run_gosec() {
  print_status "Running gosec security scanner..."

  if gosec -fmt=json -out=gosec-report.json -stdout -verbose=text ./...; then
    print_success "gosec scan completed successfully"
  else
    print_error "gosec found security issues!"
    if [ -f "gosec-report.json" ]; then
      echo "Detailed report saved to gosec-report.json"
    fi
    return 1
  fi
}

# Run vulnerability check
run_govulncheck() {
  print_status "Running govulncheck for dependency vulnerabilities..."

  # govulncheck with -json always exits 0, so we need to check the output
  # Redirect stderr to separate file to avoid corrupting JSON output
  govulncheck -json ./... >govulncheck-report.json 2>govulncheck-errors.log

  # Check if there were errors during execution
  if [ -s govulncheck-errors.log ]; then
    print_warning "govulncheck produced errors (see govulncheck-errors.log)"
  fi

  # Use jq to detect finding entries in the JSON output
  # govulncheck emits a stream of Message objects, need to slurp and filter for Finding field
  if command -v jq >/dev/null 2>&1; then
    # First validate JSON is parseable
    if ! jq -s '.' govulncheck-report.json >/dev/null 2>&1; then
      print_error "govulncheck report contains malformed JSON"
      echo "Unable to parse govulncheck-report.json"
      return 1
    fi

    # JSON is valid, now check for findings
    if jq -s -e 'map(select(.Finding)) | length > 0' govulncheck-report.json >/dev/null 2>&1; then
      print_error "Vulnerabilities found in dependencies!"
      echo "Detailed report saved to govulncheck-report.json"
      return 1
    else
      print_success "No known vulnerabilities found in dependencies"
    fi
  else
    # Fallback to grep if jq is not available (case-insensitive to match "Finding")
    if grep -qi '"finding":' govulncheck-report.json 2>/dev/null; then
      print_error "Vulnerabilities found in dependencies!"
      echo "Detailed report saved to govulncheck-report.json"
      return 1
    else
      print_success "No known vulnerabilities found in dependencies"
    fi
  fi
}

# Run enhanced golangci-lint with security focus
run_security_lint() {
  print_status "Running security-focused linting..."

  security_linters="gosec,gocritic,bodyclose,rowserrcheck,misspell,unconvert,unparam,unused,errcheck,ineffassign,staticcheck"

  if golangci-lint run --enable="$security_linters" --timeout=5m; then
    print_success "Security linting passed"
  else
    print_error "Security linting found issues!"
    return 1
  fi
}

# Check for potential secrets
check_secrets() {
  print_status "Scanning for potential secrets and sensitive data..."

  # POSIX-compatible secrets_found flag using a temp file
  secrets_found_file="$(mktemp)" || {
    print_error "Failed to create temporary file with mktemp"
    exit 1
  }
  if [ -z "$secrets_found_file" ]; then
    print_error "mktemp returned empty path"
    exit 1
  fi
  # Clean up temp file on exit
  trap 'rm -f "$secrets_found_file"' EXIT

  # Common secret patterns
  patterns='password\s*[:=]\s*['\''"][^'\''"]''{3,}['\''"]
secret\s*[:=]\s*['\''"][^'\''"]''{3,}['\''"]
key\s*[:=]\s*['\''"][^'\''"]''{8,}['\''"]
token\s*[:=]\s*['\''"][^'\''"]''{8,}['\''"]
api_?key\s*[:=]\s*['\''"][^'\''"]''{8,}['\''"]
aws_?access_?key
aws_?secret
AKIA[0-9A-Z]{16}
github_?token
private_?key'

  # Check each pattern using printf and a pipe (POSIX)
  printf '%s\n' "$patterns" | while IFS= read -r pattern; do
    if [ -n "$pattern" ] && grep -r -i -E "$pattern" --include="*.go" . 2>/dev/null | grep -q .; then
      print_warning "Potential secret pattern found: $pattern"
      touch "$secrets_found_file"
    fi
  done

  if [ -f "$secrets_found_file" ]; then
    secrets_found=true
  else
    secrets_found=false
  fi

  # Check git history for secrets (last 10 commits)
  if git log --oneline -10 2>/dev/null | grep -i -E "(password|secret|key|token)" >/dev/null 2>&1; then
    print_warning "Potential secrets mentioned in recent commit messages"
    secrets_found=true
  fi

  if [ "$secrets_found" = true ]; then
    print_warning "Potential secrets detected. Please review manually."
    return 1
  else
    print_success "No obvious secrets detected"
  fi
}

# Check for hardcoded network addresses
check_hardcoded_addresses() {
  print_status "Checking for hardcoded network addresses..."

  addresses_found=false

  # Look for IP addresses (excluding common safe ones)
  if grep -r -E "([0-9]{1,3}\.){3}[0-9]{1,3}" --include="*.go" . 2>/dev/null |
    grep -v -E "(127\.0\.0\.1|0\.0\.0\.0|255\.255\.255\.255|localhost)" >/dev/null 2>&1; then
    print_warning "Hardcoded IP addresses found:"
    grep -r -E "([0-9]{1,3}\.){3}[0-9]{1,3}" --include="*.go" . 2>/dev/null |
      grep -v -E "(127\.0\.0\.1|0\.0\.0\.0|255\.255\.255\.255|localhost)" || true
    addresses_found=true
  fi

  # Look for URLs (excluding documentation examples and comments)
  if grep -r -E "https?://[^/\s]+" --include="*.go" . 2>/dev/null |
    grep -v -E "(example\.com|localhost|127\.0\.0\.1|\$\{|//.*https?://)" >/dev/null 2>&1; then
    print_warning "Hardcoded URLs found:"
    grep -r -E "https?://[^/\s]+" --include="*.go" . 2>/dev/null |
      grep -v -E "(example\.com|localhost|127\.0\.0\.1|\$\{|//.*https?://)" || true
    addresses_found=true
  fi

  if [ "$addresses_found" = true ]; then
    print_warning "Hardcoded network addresses detected. Please review."
    return 1
  else
    print_success "No hardcoded network addresses found"
  fi
}

# Check Docker security (if Dockerfile exists)
check_docker_security() {
  if [ -f "Dockerfile" ]; then
    print_status "Checking Docker security..."

    # Basic Dockerfile security checks
    docker_issues=false

    if grep -q "^USER root" Dockerfile; then
      print_warning "Dockerfile runs as root user"
      docker_issues=true
    fi

    if ! grep -q "^USER " Dockerfile; then
      print_warning "Dockerfile doesn't specify a non-root user"
      docker_issues=true
    fi

    if grep -q "RUN.*wget\|RUN.*curl" Dockerfile && ! grep -q "rm.*wget\|rm.*curl" Dockerfile; then
      print_warning "Dockerfile may leave curl/wget installed"
      docker_issues=true
    fi

    if [ "$docker_issues" = true ]; then
      print_warning "Docker security issues detected"
      return 1
    else
      print_success "Docker security check passed"
    fi
  else
    print_status "No Dockerfile found, skipping Docker security check"
  fi
}

# Check file permissions
check_file_permissions() {
  print_status "Checking file permissions..."

  perm_issues=false

  # Check for overly permissive files (using octal for cross-platform compatibility)
  # -perm -002 finds files writable by others (works on both BSD and GNU find)
  if find . -type f -perm -002 -not -path "./.git/*" 2>/dev/null | grep -q .; then
    print_warning "World-writable files found:"
    find . -type f -perm -002 -not -path "./.git/*" 2>/dev/null || true
    perm_issues=true
  fi

  # Check for executable files that shouldn't be
  # -perm -111 finds files executable by anyone (works on both BSD and GNU find)
  if find . -type f -name "*.go" -perm -111 -not -path "./.git/*" 2>/dev/null | grep -q .; then
    print_warning "Executable Go files found (should not be executable):"
    find . -type f -name "*.go" -perm -111 -not -path "./.git/*" 2>/dev/null || true
    perm_issues=true
  fi

  if [ "$perm_issues" = true ]; then
    print_warning "File permission issues detected"
    return 1
  else
    print_success "File permissions check passed"
  fi
}

# Check Makefile with checkmake
check_makefile() {
  if [ -f "Makefile" ]; then
    print_status "Checking Makefile with checkmake..."

    if checkmake --config=.checkmake Makefile; then
      print_success "Makefile check passed"
    else
      print_error "Makefile issues detected!"
      return 1
    fi
  else
    print_status "No Makefile found, skipping checkmake"
  fi
}

# Check shell scripts with shfmt
check_shell_scripts() {
  print_status "Checking shell script formatting..."

  if find . -name "*.sh" -type f 2>/dev/null | head -1 | grep -q .; then
    if shfmt -d .; then
      print_success "Shell script formatting check passed"
    else
      print_error "Shell script formatting issues detected!"
      return 1
    fi
  else
    print_status "No shell scripts found, skipping shfmt check"
  fi
}

# Check YAML files
check_yaml_files() {
  print_status "Checking YAML files..."

  if find . \( -name "*.yml" -o -name "*.yaml" \) -type f 2>/dev/null | head -1 | grep -q .; then
    if yamllint .; then
      print_success "YAML files check passed"
    else
      print_error "YAML file issues detected!"
      return 1
    fi
  else
    print_status "No YAML files found, skipping yamllint check"
  fi
}

# Generate security report
generate_report() {
  print_status "Generating security scan report..."

  report_file="security-report.md"

  cat >"$report_file" <<EOF
# Security Scan Report

**Generated:** $(date)
**Project:** gibidify
**Scan Type:** Comprehensive Security Analysis

## Scan Results

### Security Tools Used
- gosec (Go security analyzer)
- govulncheck (Vulnerability database checker)
- golangci-lint (Static analysis with security linters)
- checkmake (Makefile linting)
- shfmt (Shell script formatting)
- yamllint (YAML file validation)
- Custom secret detection
- Custom network address detection
- Docker security checks
- File permission checks

### Files Generated
- \`gosec-report.json\` - Detailed gosec security findings
- \`govulncheck-report.json\` - Dependency vulnerability report

### Recommendations
1. Review all security findings in the generated reports
2. Address any HIGH or MEDIUM severity issues immediately
3. Consider implementing additional security measures for LOW severity issues
4. Regularly update dependencies to patch known vulnerabilities
5. Run security scans before each release

### Next Steps
- Fix any identified vulnerabilities
- Update security scanning in CI/CD pipeline
- Consider adding security testing to the test suite
- Review and update security documentation

---
*This report was generated automatically by the gibidify security scanning script.*
EOF

  print_success "Security report generated: $report_file"
}

# Main execution
main() {
  echo "🔒 gibidify Security Scanner"
  echo "=========================="
  echo

  exit_code=0

  check_dependencies
  echo

  # Run all security checks
  run_gosec || exit_code=1
  echo

  run_govulncheck || exit_code=1
  echo

  run_security_lint || exit_code=1
  echo

  check_secrets || exit_code=1
  echo

  check_hardcoded_addresses || exit_code=1
  echo

  check_docker_security || exit_code=1
  echo

  check_file_permissions || exit_code=1
  echo

  check_makefile || exit_code=1
  echo

  check_shell_scripts || exit_code=1
  echo

  check_yaml_files || exit_code=1
  echo

  generate_report
  echo

  if [ "$exit_code" -eq 0 ]; then
    print_success "🎉 All security checks passed!"
  else
    print_error "❌ Security issues detected. Please review the reports and fix identified issues."
    print_status "Generated reports:"
    print_status "- gosec-report.json (if exists)"
    print_status "- govulncheck-report.json (if exists)"
    print_status "- security-report.md"
  fi

  exit "$exit_code"
}

# Run main function
main "$@"
