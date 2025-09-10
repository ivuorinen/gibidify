# Code Style and Conventions

## EditorConfig Standards (MANDATORY)
- **Line endings**: LF (Unix-style)
- **Indentation**: Tabs (tab_width=2, indent_size=2)
- **Charset**: UTF-8
- **Final newline**: Required
- **Trailing whitespace**: Trimmed
- **Max line length**: 120 characters for Go files

## Go Code Style
- **Formatting**: gofmt, goimports, gofumpt
- **Import organization**: Local imports grouped (github.com/ivuorinen/gibidify)
- **Complexity limits**:
  - Cognitive complexity ≤ 15
  - Cyclomatic complexity ≤ 15  
  - Max control nesting ≤ 5

## Linting Standards (ZERO TOLERANCE)
- **Primary linter**: revive (comprehensive rule set)
- **Configuration**: revive.toml (DO NOT modify without explicit permission)
- **Security**: gosec for security scanning
- **All linting issues are BLOCKING errors**

## Development Patterns
- **Logging**: Use `shared.GetLogger()` (replaces logrus)
- **Error handling**: Use `shared.WrapError` family for structured errors
- **Streaming**: Use `shared.StreamContent/StreamLines` 
- **Context**: Use `shared.CheckContextCancellation`
- **Testing**: Use `testutil.*` helpers
- **Validation**: Centralized in `config/validation.go`

## File Organization
- **Module size**: 50-200 lines per module
- **Package structure**: Clean separation of concerns
- **Test files**: `*_test.go` with table-driven tests
- **Documentation**: Minimal unless explicitly requested

## Commit Standards
- **Semantic commits**: feat:, fix:, chore:, refactor:
- **No git commit**: Never commit automatically
- **No --no-verify**: Always run hooks
