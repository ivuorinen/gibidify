package utils

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

const (
	windowsOS = "windows"
)

// validatePathTestCase represents a test case for path validation functions.
type validatePathTestCase struct {
	name        string
	path        string
	wantErr     bool
	errType     ErrorType
	errCode     string
	errContains string
}

// validateExpectedError validates expected error structure and content.
func validateExpectedError(t *testing.T, err error, validatorName string, testCase validatePathTestCase) {
	t.Helper()

	if err == nil {
		t.Errorf("%s() expected error, got nil", validatorName)

		return
	}

	var structErr *StructuredError
	if !errors.As(err, &structErr) {
		t.Errorf("Expected StructuredError, got %T", err)

		return
	}

	if structErr.Type != testCase.errType {
		t.Errorf("Expected error type %v, got %v", testCase.errType, structErr.Type)
	}
	if structErr.Code != testCase.errCode {
		t.Errorf("Expected error code %v, got %v", testCase.errCode, structErr.Code)
	}
	if testCase.errContains != "" && !strings.Contains(err.Error(), testCase.errContains) {
		t.Errorf("Error should contain %q, got: %v", testCase.errContains, err.Error())
	}
}

// testPathValidation is a helper function to test path validation functions without duplication.
func testPathValidation(
	t *testing.T,
	validatorName string,
	validatorFunc func(string) error,
	tests []validatePathTestCase,
) {
	t.Helper()

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				err := validatorFunc(tt.path)

				if tt.wantErr {
					validateExpectedError(t, err, validatorName, tt)

					return
				}

				if err != nil {
					t.Errorf("%s() unexpected error: %v", validatorName, err)
				}
			},
		)
	}
}

func TestGetAbsolutePath(t *testing.T) {
	// Get current working directory for tests
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	tests := createAbsolutePathTestCases(cwd)

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				verifyAbsolutePathResult(t, tt.path, tt.wantPrefix, tt.wantErr, tt.wantErrMsg, tt.skipWindows)
			},
		)
	}
}

// createAbsolutePathTestCases creates test cases for GetAbsolutePath.
func createAbsolutePathTestCases(cwd string) []struct {
	name        string
	path        string
	wantPrefix  string
	wantErr     bool
	wantErrMsg  string
	skipWindows bool
} {
	return []struct {
		name        string
		path        string
		wantPrefix  string
		wantErr     bool
		wantErrMsg  string
		skipWindows bool
	}{
		{
			name:       "absolute path unchanged",
			path:       cwd,
			wantPrefix: cwd,
			wantErr:    false,
		},
		{
			name:       "relative path current directory",
			path:       ".",
			wantPrefix: cwd,
			wantErr:    false,
		},
		{
			name:       "relative path parent directory",
			path:       "..",
			wantPrefix: filepath.Dir(cwd),
			wantErr:    false,
		},
		{
			name:       "relative path with file",
			path:       "test.txt",
			wantPrefix: filepath.Join(cwd, "test.txt"),
			wantErr:    false,
		},
		{
			name:       "relative path with subdirectory",
			path:       "subdir/file.go",
			wantPrefix: filepath.Join(cwd, "subdir", "file.go"),
			wantErr:    false,
		},
		{
			name:       "empty path",
			path:       "",
			wantPrefix: cwd,
			wantErr:    false,
		},
		{
			name:        "path with tilde",
			path:        "~/test",
			wantPrefix:  filepath.Join(cwd, "~", "test"),
			wantErr:     false,
			skipWindows: false,
		},
		{
			name:       "path with multiple separators",
			path:       "path//to///file",
			wantPrefix: filepath.Join(cwd, "path", "to", "file"),
			wantErr:    false,
		},
		{
			name:       "path with trailing separator",
			path:       "path/",
			wantPrefix: filepath.Join(cwd, "path"),
			wantErr:    false,
		},
	}
}

// verifyAbsolutePathResult verifies the result of GetAbsolutePath.
func verifyAbsolutePathResult(
	t *testing.T,
	path, wantPrefix string,
	wantErr bool,
	wantErrMsg string,
	skipWindows bool,
) {
	t.Helper()

	if skipWindows && runtime.GOOS == windowsOS {
		t.Skip("Skipping test on Windows")
	}

	got, err := GetAbsolutePath(path)

	if wantErr {
		verifyExpectedError(t, err, wantErrMsg)

		return
	}

	if err != nil {
		t.Errorf("GetAbsolutePath() unexpected error = %v", err)

		return
	}

	verifyAbsolutePathOutput(t, got, wantPrefix)
}

// verifyExpectedError verifies that an expected error occurred.
func verifyExpectedError(t *testing.T, err error, wantErrMsg string) {
	t.Helper()

	if err == nil {
		t.Error("GetAbsolutePath() error = nil, wantErr true")

		return
	}

	if wantErrMsg != "" && !strings.Contains(err.Error(), wantErrMsg) {
		t.Errorf("GetAbsolutePath() error = %v, want error containing %v", err, wantErrMsg)
	}
}

// verifyAbsolutePathOutput verifies the output of GetAbsolutePath.
func verifyAbsolutePathOutput(t *testing.T, got, wantPrefix string) {
	t.Helper()

	// Clean the expected path for comparison
	wantClean := filepath.Clean(wantPrefix)
	gotClean := filepath.Clean(got)

	if gotClean != wantClean {
		t.Errorf("GetAbsolutePath() = %v, want %v", gotClean, wantClean)
	}

	// Verify the result is actually absolute
	if !filepath.IsAbs(got) {
		t.Errorf("GetAbsolutePath() returned non-absolute path: %v", got)
	}
}

func TestGetAbsolutePathSpecialCases(t *testing.T) {
	if runtime.GOOS == windowsOS {
		t.Skip("Skipping Unix-specific tests on Windows")
	}

	tests := []struct {
		name    string
		setup   func(*testing.T) (string, func())
		path    string
		wantErr bool
	}{
		{
			name:    "symlink to directory",
			setup:   setupSymlinkToDirectory,
			path:    "",
			wantErr: false,
		},
		{
			name:    "broken symlink",
			setup:   setupBrokenSymlink,
			path:    "",
			wantErr: false, // filepath.Abs still works with broken symlinks
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				verifySpecialCaseAbsolutePath(t, tt.setup, tt.path, tt.wantErr)
			},
		)
	}
}

// setupSymlinkToDirectory creates a symlink pointing to a directory.
func setupSymlinkToDirectory(t *testing.T) (string, func()) {
	t.Helper()
	tmpDir := t.TempDir()
	target := filepath.Join(tmpDir, "target")
	link := filepath.Join(tmpDir, "link")

	if err := os.Mkdir(target, 0o750); err != nil {
		t.Fatalf("Failed to create target directory: %v", err)
	}
	if err := os.Symlink(target, link); err != nil {
		t.Fatalf("Failed to create symlink: %v", err)
	}

	return link, func() {}
}

// setupBrokenSymlink creates a broken symlink.
func setupBrokenSymlink(t *testing.T) (string, func()) {
	t.Helper()
	tmpDir := t.TempDir()
	link := filepath.Join(tmpDir, "broken_link")

	if err := os.Symlink("/nonexistent/path", link); err != nil {
		t.Fatalf("Failed to create broken symlink: %v", err)
	}

	return link, func() {}
}

// verifySpecialCaseAbsolutePath verifies GetAbsolutePath with special cases.
func verifySpecialCaseAbsolutePath(t *testing.T, setup func(*testing.T) (string, func()), path string, wantErr bool) {
	t.Helper()
	testPath, cleanup := setup(t)
	defer cleanup()

	if path == "" {
		path = testPath
	}

	got, err := GetAbsolutePath(path)
	if (err != nil) != wantErr {
		t.Errorf("GetAbsolutePath() error = %v, wantErr %v", err, wantErr)

		return
	}

	if err == nil && !filepath.IsAbs(got) {
		t.Errorf("GetAbsolutePath() returned non-absolute path: %v", got)
	}
}

func TestGetAbsolutePathConcurrency(_ *testing.T) {
	// Test that GetAbsolutePath is safe for concurrent use
	paths := []string{".", "..", "test.go", "subdir/file.txt", "/tmp/test"}
	done := make(chan bool)

	for _, p := range paths {
		go func(path string) {
			_, _ = GetAbsolutePath(path) // nolint:errcheck // intentionally ignoring errors in concurrent test
			done <- true
		}(p)
	}

	// Wait for all goroutines to complete
	for range paths {
		<-done
	}
}

func TestGetAbsolutePathErrorFormatting(t *testing.T) {
	// This test verifies error message formatting
	// We need to trigger an actual error from filepath.Abs
	// On Unix systems, we can't easily trigger filepath.Abs errors
	// so we'll just verify the error wrapping works correctly

	// Create a test that would fail if filepath.Abs returns an error
	path := "test/path"
	got, err := GetAbsolutePath(path)
	if err != nil {
		// If we somehow get an error, verify it's properly formatted
		if !strings.Contains(err.Error(), "failed to get absolute path for") {
			t.Errorf("Error message format incorrect: %v", err)
		}
		if !strings.Contains(err.Error(), path) {
			t.Errorf("Error message should contain original path: %v", err)
		}
	} else if !filepath.IsAbs(got) {
		// Normal case - just verify we got a valid absolute path
		t.Errorf("Expected absolute path, got: %v", got)
	}
}

// BenchmarkGetAbsolutePath benchmarks the GetAbsolutePath function.
func BenchmarkGetAbsolutePath(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = GetAbsolutePath("test/path/file.go") // nolint:errcheck // benchmark test
	}
}

// BenchmarkGetAbsolutePathAbs benchmarks with already absolute path.
func BenchmarkGetAbsolutePathAbs(b *testing.B) {
	absPath := "/home/user/test/file.go"
	if runtime.GOOS == windowsOS {
		absPath = "C:\\Users\\test\\file.go"
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = GetAbsolutePath(absPath) // nolint:errcheck // benchmark test
	}
}

// BenchmarkGetAbsolutePathCurrent benchmarks with current directory.
func BenchmarkGetAbsolutePathCurrent(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = GetAbsolutePath(".") // nolint:errcheck // benchmark test
	}
}

func TestValidateSourcePath(t *testing.T) {
	// Create test directories for validation
	tmpDir := t.TempDir()
	validDir := filepath.Join(tmpDir, "validdir")
	validFile := filepath.Join(tmpDir, "validfile.txt")

	// Create test directory and file
	if err := os.Mkdir(validDir, 0o750); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	if err := os.WriteFile(validFile, []byte("test"), 0o600); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tests := []validatePathTestCase{
		{
			name:        "empty path",
			path:        "",
			wantErr:     true,
			errType:     ErrorTypeValidation,
			errCode:     CodeValidationRequired,
			errContains: "source path is required",
		},
		{
			name:        "path traversal with double dots",
			path:        "../../../etc/passwd",
			wantErr:     true,
			errType:     ErrorTypeValidation,
			errCode:     CodeValidationPath,
			errContains: "path traversal attempt detected",
		},
		{
			name:        "path traversal in middle",
			path:        "valid/../../../secrets",
			wantErr:     true,
			errType:     ErrorTypeValidation,
			errCode:     CodeValidationPath,
			errContains: "path traversal attempt detected",
		},
		{
			name:        "nonexistent directory",
			path:        "/nonexistent/directory",
			wantErr:     true,
			errType:     ErrorTypeFileSystem,
			errCode:     CodeFSNotFound,
			errContains: "source directory does not exist",
		},
		{
			name:        "file instead of directory",
			path:        validFile,
			wantErr:     true,
			errType:     ErrorTypeValidation,
			errCode:     CodeValidationPath,
			errContains: "source path must be a directory",
		},
		{
			name:    "valid directory (absolute)",
			path:    validDir,
			wantErr: false,
		},
		{
			name:    "valid directory (relative)",
			path:    ".",
			wantErr: false,
		},
		{
			name:    "valid directory (current)",
			path:    tmpDir,
			wantErr: false,
		},
	}

	// Save and restore current directory for relative path tests
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer func() {
		// Need to use os.Chdir here since t.Chdir only works in the current function context
		if err := os.Chdir(originalWd); err != nil { // nolint:usetesting // needed in defer function
			t.Logf("Failed to restore working directory: %v", err)
		}
	}()
	t.Chdir(tmpDir)

	testPathValidation(t, "ValidateSourcePath", ValidateSourcePath, tests)
}

func TestValidateDestinationPath(t *testing.T) {
	tmpDir := t.TempDir()
	existingDir := filepath.Join(tmpDir, "existing")
	existingFile := filepath.Join(tmpDir, "existing.txt")
	validDest := filepath.Join(tmpDir, "output.txt")

	// Create test directory and file
	if err := os.Mkdir(existingDir, 0o750); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	if err := os.WriteFile(existingFile, []byte("test"), 0o600); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tests := []validatePathTestCase{
		{
			name:        "empty path",
			path:        "",
			wantErr:     true,
			errType:     ErrorTypeValidation,
			errCode:     CodeValidationRequired,
			errContains: "destination path is required",
		},
		{
			name:        "path traversal attack",
			path:        "../../../tmp/malicious.txt",
			wantErr:     true,
			errType:     ErrorTypeValidation,
			errCode:     CodeValidationPath,
			errContains: "path traversal attempt detected",
		},
		{
			name:        "destination is existing directory",
			path:        existingDir,
			wantErr:     true,
			errType:     ErrorTypeValidation,
			errCode:     CodeValidationPath,
			errContains: "destination cannot be a directory",
		},
		{
			name:        "parent directory doesn't exist",
			path:        "/nonexistent/dir/output.txt",
			wantErr:     true,
			errType:     ErrorTypeFileSystem,
			errCode:     CodeFSNotFound,
			errContains: "destination parent directory does not exist",
		},
		{
			name:    "valid destination path",
			path:    validDest,
			wantErr: false,
		},
		{
			name:    "overwrite existing file (should be valid)",
			path:    existingFile,
			wantErr: false,
		},
	}

	testPathValidation(t, "ValidateDestinationPath", ValidateDestinationPath, tests)
}

func TestValidateConfigPath(t *testing.T) {
	tests := []validatePathTestCase{
		{
			name:    "empty path (allowed for config)",
			path:    "",
			wantErr: false,
		},
		{
			name:        "path traversal attack",
			path:        "../../../etc/passwd",
			wantErr:     true,
			errType:     ErrorTypeValidation,
			errCode:     CodeValidationPath,
			errContains: "path traversal attempt detected",
		},
		{
			name:        "complex path traversal",
			path:        "config/../../../secrets/config.yaml",
			wantErr:     true,
			errType:     ErrorTypeValidation,
			errCode:     CodeValidationPath,
			errContains: "path traversal attempt detected",
		},
		{
			name:    "valid config path",
			path:    "config.yaml",
			wantErr: false,
		},
		{
			name:    "valid absolute config path",
			path:    "/etc/myapp/config.yaml",
			wantErr: false,
		},
		{
			name:    "config in subdirectory",
			path:    "configs/production.yaml",
			wantErr: false,
		},
	}

	testPathValidation(t, "ValidateConfigPath", ValidateConfigPath, tests)
}

func TestGetBaseName(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "simple filename",
			path:     "/path/to/file.txt",
			expected: "file.txt",
		},
		{
			name:     "directory path",
			path:     "/path/to/directory",
			expected: "directory",
		},
		{
			name:     "root path",
			path:     "/",
			expected: "/",
		},
		{
			name:     "current directory",
			path:     ".",
			expected: "output", // Special case: . returns "output"
		},
		{
			name:     "empty path",
			path:     "",
			expected: "output", // Special case: empty returns "output"
		},
		{
			name:     "path with trailing separator",
			path:     "/path/to/dir/",
			expected: "dir",
		},
		{
			name:     "relative path",
			path:     "subdir/file.go",
			expected: "file.go",
		},
		{
			name:     "single filename",
			path:     "README.md",
			expected: "README.md",
		},
		{
			name:     "path with spaces",
			path:     "/path/to/my file.txt",
			expected: "my file.txt",
		},
		{
			name:     "path with special characters",
			path:     "/path/to/file-name_123.ext",
			expected: "file-name_123.ext",
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				result := GetBaseName(tt.path)
				if result != tt.expected {
					t.Errorf("GetBaseName(%q) = %q, want %q", tt.path, result, tt.expected)
				}
			},
		)
	}
}

// Security-focused integration tests.
func TestPathValidationIntegration(t *testing.T) {
	tmpDir := t.TempDir()
	validSourceDir := filepath.Join(tmpDir, "source")
	validDestFile := filepath.Join(tmpDir, "output.txt")

	// Create source directory
	if err := os.Mkdir(validSourceDir, 0o750); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	// Test complete validation workflow
	tests := []struct {
		name            string
		sourcePath      string
		destPath        string
		configPath      string
		expectSourceErr bool
		expectDestErr   bool
		expectConfigErr bool
	}{
		{
			name:            "valid paths",
			sourcePath:      validSourceDir,
			destPath:        validDestFile,
			configPath:      "config.yaml",
			expectSourceErr: false,
			expectDestErr:   false,
			expectConfigErr: false,
		},
		{
			name:            "source path traversal attack",
			sourcePath:      "../../../etc",
			destPath:        validDestFile,
			configPath:      "config.yaml",
			expectSourceErr: true,
			expectDestErr:   false,
			expectConfigErr: false,
		},
		{
			name:            "destination path traversal attack",
			sourcePath:      validSourceDir,
			destPath:        "../../../tmp/malicious.txt",
			configPath:      "config.yaml",
			expectSourceErr: false,
			expectDestErr:   true,
			expectConfigErr: false,
		},
		{
			name:            "config path traversal attack",
			sourcePath:      validSourceDir,
			destPath:        validDestFile,
			configPath:      "../../../etc/passwd",
			expectSourceErr: false,
			expectDestErr:   false,
			expectConfigErr: true,
		},
		{
			name:            "multiple path traversal attacks",
			sourcePath:      "../../../var",
			destPath:        "../../../tmp/bad.txt",
			configPath:      "../../../etc/config",
			expectSourceErr: true,
			expectDestErr:   true,
			expectConfigErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				// Test source validation
				sourceErr := ValidateSourcePath(tt.sourcePath)
				if (sourceErr != nil) != tt.expectSourceErr {
					t.Errorf("Source validation: expected error %v, got %v", tt.expectSourceErr, sourceErr)
				}

				// Test destination validation
				destErr := ValidateDestinationPath(tt.destPath)
				if (destErr != nil) != tt.expectDestErr {
					t.Errorf("Destination validation: expected error %v, got %v", tt.expectDestErr, destErr)
				}

				// Test config validation
				configErr := ValidateConfigPath(tt.configPath)
				if (configErr != nil) != tt.expectConfigErr {
					t.Errorf("Config validation: expected error %v, got %v", tt.expectConfigErr, configErr)
				}
			},
		)
	}
}

// Benchmark the validation functions for performance.
func BenchmarkValidateSourcePath(b *testing.B) {
	tmpDir := b.TempDir()
	validDir := filepath.Join(tmpDir, "testdir")
	if err := os.Mkdir(validDir, 0o750); err != nil {
		b.Fatalf("Failed to create test directory: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ValidateSourcePath(validDir) // nolint:errcheck // benchmark test
	}
}

func BenchmarkValidateDestinationPath(b *testing.B) {
	tmpDir := b.TempDir()
	validDest := filepath.Join(tmpDir, "output.txt")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ValidateDestinationPath(validDest) // nolint:errcheck // benchmark test
	}
}

func BenchmarkGetBaseName(b *testing.B) {
	path := "/very/long/path/to/some/deeply/nested/file.txt"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GetBaseName(path)
	}
}
