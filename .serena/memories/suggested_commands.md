# Essential Development Commands

## Linting and Code Quality
```bash
# Auto-fix linting issues (ALWAYS run first)
make lint-fix

# Check linting (must show 0 issues)
make lint

# CI-specific linting
make ci-lint
```

## Testing
```bash
# Run all tests with race detection
make test

# Run tests with coverage report
make coverage

# CI-specific testing
make ci-test
```

## Build and Run
```bash
# Build the application
make build

# Run the application
./gibidify -source <dir> -format markdown --verbose
./gibidify -source <dir> -format json --log-level debug --verbose
```

## Benchmarking
```bash
# Build and run all benchmarks
make benchmark

# Specific benchmark types
make benchmark-collection
make benchmark-processing
make benchmark-concurrency
make benchmark-format
```

## Security
```bash
# Run security scan
make security

# Full security analysis
make security-full

# Vulnerability check
make vuln-check
```

## Development Workflow
```bash
# Complete development cycle
make all  # Runs: lint lint-fix test build

# Clean build artifacts
make clean
```

## System Commands (Darwin)
- `which <command>` - Find command path location
- `rg` - Code search (preferred over grep)
- `fd` - File search (preferred over find)
- `pwd` - Check current directory
- `ls` - List directory contents
