# Codebase Structure

## Core Directories

### Main Application
- **main.go**: Application entrypoint with main() and run() functions
- **interfaces.go**: Core interface definitions

### Core Packages
- **cli/**: Command-line interface, flags, UI, processor logic
  - processor_*.go: Collection, processing, workers, stats, types
  - ui.go: Progress bars and colored output
  - flags.go: Command-line flag handling
  - errors.go: CLI-specific error handling

- **shared/**: Shared utilities (formerly utils/)
  - logger.go: Centralized logging with shared.GetLogger()
  - errors.go: Structured error handling with WrapError family
  - writers.go: File streaming and writing utilities
  - paths.go: Path manipulation utilities
  - conversions.go: Data conversion helpers

- **config/**: Configuration management
  - YAML-based configuration with validation
  - validation.go: Centralized validation logic

- **fileproc/**: File processing logic
  - Resource monitoring and validation
  - File type detection and filtering

### Advanced Features
- **metrics/**: Performance tracking and profiling
  - Real-time processing statistics
  - Memory usage tracking and recommendations

- **templates/**: Template system
  - 4 built-in templates: default, minimal, detailed, compact
  - Custom template support with variable substitution

- **testutil/**: Testing utilities and helpers
  - Shared test infrastructure
  - Mock objects and assertions

### Supporting Infrastructure
- **cmd/**: Additional commands (benchmark/)
- **scripts/**: Development and CI scripts
- **examples/**: Usage examples
- **.github/**: GitHub Actions and workflows

## Configuration Files
- **revive.toml**: Linting configuration (PROTECTED - do not modify)
- **.editorconfig**: Code formatting rules (MANDATORY compliance)
- **Makefile**: Build and development targets
- **go.mod/go.sum**: Go module dependencies
