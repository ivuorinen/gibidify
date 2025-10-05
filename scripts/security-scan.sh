#!/bin/bash
set -euo pipefail

# Security Scanning Script for gibidify
# This script runs comprehensive security checks locally and in CI

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
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
	echo -e "${BLUE}[INFO]${NC} $1"
}

print_warning() {
	echo -e "${YELLOW}[WARN]${NC} $1"
}

print_error() {
	echo -e "${RED}[ERROR]${NC} $1"
}

print_success() {
	echo -e "${GREEN}[SUCCESS]${NC} $1"
}

# Check if required tools are installed
check_dependencies() {
	print_status "Checking security scanning dependencies..."

	local missing_tools=()

	if ! command -v go &>/dev/null; then
		missing_tools+=("go")
	fi

	if ! command -v golangci-lint &>/dev/null; then
		print_warning "golangci-lint not found, installing..."
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	fi

	if ! command -v gosec &>/dev/null; then
		print_warning "gosec not found, installing..."
		go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
	fi

	if ! command -v govulncheck &>/dev/null; then
		print_warning "govulncheck not found, installing..."
		go install golang.org/x/vuln/cmd/govulncheck@latest
	fi

	if ! command -v checkmake &>/dev/null; then
		print_warning "checkmake not found, installing..."
		go install github.com/checkmake/checkmake/cmd/checkmake@latest
	fi

	if ! command -v shfmt &>/dev/null; then
		print_warning "shfmt not found, installing..."
		go install mvdan.cc/sh/v3/cmd/shfmt@latest
	fi

	if ! command -v yamllint &>/dev/null; then
		print_warning "yamllint not found, installing..."
		go install github.com/excilsploft/yamllint@latest
	fi

	if [ ${#missing_tools[@]} -ne 0 ]; then
		print_error "Missing required tools: ${missing_tools[*]}"
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

	if govulncheck -json ./... >govulncheck-report.json 2>&1; then
		print_success "No known vulnerabilities found in dependencies"
	else
		if grep -q '"finding"' govulncheck-report.json 2>/dev/null; then
			print_error "Vulnerabilities found in dependencies!"
			echo "Detailed report saved to govulncheck-report.json"
			return 1
		else
			print_success "No vulnerabilities found"
		fi
	fi
}

# Run enhanced golangci-lint with security focus
run_security_lint() {
	print_status "Running security-focused linting..."

	local security_linters="gosec,gocritic,bodyclose,rowserrcheck,misspell,unconvert,unparam,unused,errcheck,ineffassign,staticcheck"

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

	local secrets_found=false

	# Common secret patterns
	local patterns=(
		"password\s*[:=]\s*['\"][^'\"]{3,}['\"]"
		"secret\s*[:=]\s*['\"][^'\"]{3,}['\"]"
		"key\s*[:=]\s*['\"][^'\"]{8,}['\"]"
		"token\s*[:=]\s*['\"][^'\"]{8,}['\"]"
		"api_?key\s*[:=]\s*['\"][^'\"]{8,}['\"]"
		"aws_?access_?key"
		"aws_?secret"
		"AKIA[0-9A-Z]{16}" # AWS Access Key pattern
		"github_?token"
		"private_?key"
	)

	for pattern in "${patterns[@]}"; do
		if grep -r -i -E "$pattern" --include="*.go" . 2>/dev/null; then
			print_warning "Potential secret pattern found: $pattern"
			secrets_found=true
		fi
	done

	# Check git history for secrets (last 10 commits)
	if git log --oneline -10 | grep -i -E "(password|secret|key|token)" >/dev/null 2>&1; then
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

	local addresses_found=false

	# Look for IP addresses (excluding common safe ones)
	if grep -r -E "([0-9]{1,3}\.){3}[0-9]{1,3}" --include="*.go" . |
		grep -v -E "(127\.0\.0\.1|0\.0\.0\.0|255\.255\.255\.255|localhost)" >/dev/null 2>&1; then
		print_warning "Hardcoded IP addresses found:"
		grep -r -E "([0-9]{1,3}\.){3}[0-9]{1,3}" --include="*.go" . |
			grep -v -E "(127\.0\.0\.1|0\.0\.0\.0|255\.255\.255\.255|localhost)" || true
		addresses_found=true
	fi

	# Look for URLs (excluding documentation examples)
	if grep -r -E "https?://[^/\s]+" --include="*.go" . |
		grep -v -E "(example\.com|localhost|127\.0\.0\.1|\$\{)" >/dev/null 2>&1; then
		print_warning "Hardcoded URLs found:"
		grep -r -E "https?://[^/\s]+" --include="*.go" . |
			grep -v -E "(example\.com|localhost|127\.0\.0\.1|\$\{)" || true
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
		local docker_issues=false

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

	local perm_issues=false

	# Check for overly permissive files
	if find . -type f -perm /o+w -not -path "./.git/*" | grep -q .; then
		print_warning "World-writable files found:"
		find . -type f -perm /o+w -not -path "./.git/*" || true
		perm_issues=true
	fi

	# Check for executable files that shouldn't be
	if find . -type f -name "*.go" -perm /a+x | grep -q .; then
		print_warning "Executable Go files found (should not be executable):"
		find . -type f -name "*.go" -perm /a+x || true
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

	if find . -name "*.sh" -type f | head -1 | grep -q .; then
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

	if find . -name "*.yml" -o -name "*.yaml" -type f | head -1 | grep -q .; then
		if yamllint -c .yamllint .; then
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

	local report_file="security-report.md"

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

	local exit_code=0

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

	if [ $exit_code -eq 0 ]; then
		print_success "🎉 All security checks passed!"
	else
		print_error "❌ Security issues detected. Please review the reports and fix identified issues."
		print_status "Generated reports:"
		print_status "- gosec-report.json (if exists)"
		print_status "- govulncheck-report.json (if exists)"
		print_status "- security-report.md"
	fi

	exit $exit_code
}

# Run main function
main "$@"
