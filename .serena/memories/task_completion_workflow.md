# Task Completion Workflow

## Mandatory Steps After Code Changes

### 1. Auto-fix First (ALWAYS)
```bash
make lint-fix
```
- Runs gofumpt, goimports, go fmt, go mod tidy
- Fixes shfmt, revive, gosec, checkmake, yamlfmt, eclint issues
- Must be run BEFORE manual linting

### 2. Linting Verification (ZERO TOLERANCE)
```bash
make lint
```
- Must show 0 issues (all linting issues are blocking)
- Runs revive, gosec, checkmake, shfmt, yamllint, eclint
- Never modify revive.toml unless explicitly requested

### 3. Testing (Required)
```bash
make test
```
- Run all tests with race detection
- Fix all test failures and warnings
- Maintain >80% coverage target

### 4. Final Build Verification
```bash
make build
```
- Ensure clean build with no errors
- Verify application functionality

## Critical Rules
- **EditorConfig violations are BLOCKING**: Always follow .editorconfig rules
- **Never commit automatically**: Wait for explicit user request
- **Use full paths**: Always use absolute paths when cd'ing
- **Tool verification**: Always use `which <command>` to verify tool paths
- **Memory efficiency**: Prefer mcp tools, use rg/fd over grep/find

## Error Handling
- All linting errors must be fixed (no exceptions)
- All test failures must be resolved
- Use structured error handling with shared.WrapError
- Check context cancellation with shared.CheckContextCancellation

## Quality Gates
1. EditorConfig compliance ✓
2. Zero linting issues ✓  
3. All tests passing ✓
4. Clean build ✓
5. >80% test coverage maintained ✓
