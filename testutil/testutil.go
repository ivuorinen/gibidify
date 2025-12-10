// Package testutil provides common testing utilities and helper functions.
//
// Testing Patterns and Conventions:
//
//	File Setup:
//	  - Use CreateTestFile() for individual files
//	  - Use CreateTestFiles() for multiple files from FileSpec
//	  - Use CreateTestDirectoryStructure() for complex directory trees
//	  - Use SetupTempDirWithStructure() for complete test environments
//
//	Error Assertions:
//	  - Use AssertError() for conditional error checking
//	  - Use AssertNoError() when expecting success
//	  - Use AssertExpectedError() when expecting failure
//	  - Use AssertErrorContains() for substring validation
//
//	Configuration:
//	  - Use ResetViperConfig() to reset between tests
//	  - Remember to call config.LoadConfig() after ResetViperConfig()
//
//	Best Practices:
//	  - Always use t.Helper() in test helper functions
//	  - Use descriptive operation names in assertions
//	  - Prefer table-driven tests for multiple scenarios
//	  - Use testutil.ErrTestError for standard test errors
package testutil

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/spf13/viper"

	"github.com/ivuorinen/gibidify/config"
	"github.com/ivuorinen/gibidify/shared"
)

// SuppressLogs suppresses logger output during testing to keep test output clean.
// Returns a function that should be called to restore the original log output.
func SuppressLogs(t *testing.T) func() {
	t.Helper()
	logger := shared.GetLogger()

	// Capture original output by temporarily setting it to discard
	logger.SetOutput(io.Discard)

	// Return function to restore original settings (stderr)
	return func() {
		logger.SetOutput(os.Stderr)
	}
}

// OutputRestoreFunc represents a function that restores output after suppression.
type OutputRestoreFunc func()

// SuppressAllOutput suppresses both stdout and stderr during testing.
// This captures all output including UI messages, progress bars, and direct prints.
// Returns a function that should be called to restore original output.
func SuppressAllOutput(t *testing.T) OutputRestoreFunc {
	t.Helper()

	// Save original stdout and stderr
	originalStdout := os.Stdout
	originalStderr := os.Stderr

	// Suppress logger output as well
	logger := shared.GetLogger()
	logger.SetOutput(io.Discard)

	// Open /dev/null for safe redirection
	devNull, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	if err != nil {
		t.Fatalf("Failed to open devnull: %v", err)
	}

	// Redirect both stdout and stderr to /dev/null
	os.Stdout = devNull
	os.Stderr = devNull

	// Return restore function
	return func() {
		// Close devNull first
		if devNull != nil {
			_ = devNull.Close() // Ignore close errors in cleanup
		}

		// Restore original outputs
		os.Stdout = originalStdout
		os.Stderr = originalStderr
		logger.SetOutput(originalStderr)
	}
}

// CaptureOutput captures both stdout and stderr during test execution.
// Returns the captured output as strings and a restore function.
func CaptureOutput(t *testing.T) (getStdout func() string, getStderr func() string, restore OutputRestoreFunc) {
	t.Helper()

	// Save original outputs
	originalStdout := os.Stdout
	originalStderr := os.Stderr

	// Create pipes for stdout
	stdoutReader, stdoutWriter, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create stdout pipe: %v", err)
	}

	// Create pipes for stderr
	stderrReader, stderrWriter, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create stderr pipe: %v", err)
	}

	// Redirect outputs
	os.Stdout = stdoutWriter
	os.Stderr = stderrWriter

	// Suppress logger output to stderr
	logger := shared.GetLogger()
	logger.SetOutput(stderrWriter)

	// Buffers to collect output
	var stdoutBuf, stderrBuf bytes.Buffer

	// Start goroutines to read from pipes
	stdoutDone := make(chan struct{})
	stderrDone := make(chan struct{})

	go func() {
		defer close(stdoutDone)
		_, _ = io.Copy(&stdoutBuf, stdoutReader) //nolint:errcheck // Ignore errors during test output capture shutdown
	}()

	go func() {
		defer close(stderrDone)
		_, _ = io.Copy(&stderrBuf, stderrReader) //nolint:errcheck // Ignore errors during test output capture shutdown
	}()

	return func() string {
			return stdoutBuf.String()
		}, func() string {
			return stderrBuf.String()
		}, func() {
			// Close writers first to signal EOF
			_ = stdoutWriter.Close() // Ignore close errors in cleanup
			_ = stderrWriter.Close() // Ignore close errors in cleanup

			// Wait for readers to finish
			<-stdoutDone
			<-stderrDone

			// Close readers
			_ = stdoutReader.Close() // Ignore close errors in cleanup
			_ = stderrReader.Close() // Ignore close errors in cleanup

			// Restore original outputs
			os.Stdout = originalStdout
			os.Stderr = originalStderr
			logger.SetOutput(originalStderr)
		}
}

// CreateTestFile creates a test file with the given content and returns its path.
func CreateTestFile(t *testing.T, dir, filename string, content []byte) string {
	t.Helper()
	filePath := filepath.Join(dir, filename)
	if err := os.WriteFile(filePath, content, shared.TestFilePermission); err != nil {
		t.Fatalf("Failed to write file %s: %v", filePath, err)
	}

	return filePath
}

// CreateTempOutputFile creates a temporary output file and returns the file handle and path.
func CreateTempOutputFile(t *testing.T, pattern string) (file *os.File, path string) {
	t.Helper()
	outFile, err := os.CreateTemp(t.TempDir(), pattern)
	if err != nil {
		t.Fatalf("Failed to create temp output file: %v", err)
	}
	path = outFile.Name()

	return outFile, path
}

// CreateTestDirectory creates a test directory and returns its path.
func CreateTestDirectory(t *testing.T, parent, name string) string {
	t.Helper()
	dirPath := filepath.Join(parent, name)
	if err := os.Mkdir(dirPath, shared.TestDirPermission); err != nil {
		t.Fatalf("Failed to create directory %s: %v", dirPath, err)
	}

	return dirPath
}

// FileSpec represents a file specification for creating test files.
type FileSpec struct {
	Name    string
	Content string
}

// CreateTestFiles creates multiple test files from specifications.
func CreateTestFiles(t *testing.T, rootDir string, fileSpecs []FileSpec) []string {
	t.Helper()
	createdFiles := make([]string, 0, len(fileSpecs))
	for _, spec := range fileSpecs {
		filePath := CreateTestFile(t, rootDir, spec.Name, []byte(spec.Content))
		createdFiles = append(createdFiles, filePath)
	}

	return createdFiles
}

// ResetViperConfig resets Viper configuration and optionally sets a config path.
func ResetViperConfig(t *testing.T, configPath string) {
	t.Helper()
	viper.Reset()
	if configPath != "" {
		viper.AddConfigPath(configPath)
	}
	config.LoadConfig()
}

// SetViperKeys sets specific configuration keys for testing.
func SetViperKeys(t *testing.T, keyValues map[string]any) {
	t.Helper()
	viper.Reset()
	for key, value := range keyValues {
		viper.Set(key, value)
	}
	config.LoadConfig()
}

// ApplyBackpressureOverrides applies backpressure configuration overrides for testing.
// This is a convenience wrapper around SetViperKeys specifically for backpressure tests.
func ApplyBackpressureOverrides(t *testing.T, overrides map[string]any) {
	t.Helper()
	SetViperKeys(t, overrides)
}

// SetupCLIArgs configures os.Args for CLI testing.
func SetupCLIArgs(srcDir, outFilePath, prefix, suffix string, concurrency int) {
	os.Args = []string{
		"gibidify",
		"-source", srcDir,
		"-destination", outFilePath,
		"-prefix", prefix,
		"-suffix", suffix,
		"-concurrency", strconv.Itoa(concurrency),
		"-no-ui", // Suppress UI output during tests
	}
}

// VerifyContentContains checks that content contains all expected substrings.
func VerifyContentContains(t *testing.T, content string, expectedSubstrings []string) {
	t.Helper()
	for _, expected := range expectedSubstrings {
		if !strings.Contains(content, expected) {
			t.Errorf("Content missing expected substring: %s", expected)
		}
	}
}

// MustSucceed fails the test if the error is not nil.
func MustSucceed(t *testing.T, err error, operation string) {
	t.Helper()
	if err != nil {
		t.Fatalf(shared.TestMsgOperationFailed, operation, err)
	}
}

// CloseFile closes a file and reports errors to the test.
func CloseFile(t *testing.T, file *os.File) {
	t.Helper()
	if err := file.Close(); err != nil {
		t.Errorf("Failed to close file: %v", err)
	}
}

// BaseName returns the base name of a file path (filename without directory).
func BaseName(path string) string {
	return filepath.Base(path)
}

// Advanced directory setup patterns.

// DirSpec represents a directory specification for creating test directory structures.
type DirSpec struct {
	Path  string
	Files []FileSpec
}

// CreateTestDirectoryStructure creates multiple directories with files.
func CreateTestDirectoryStructure(t *testing.T, rootDir string, dirSpecs []DirSpec) []string {
	t.Helper()
	createdPaths := make([]string, 0)

	for _, dirSpec := range dirSpecs {
		dirPath := filepath.Join(rootDir, dirSpec.Path)
		if err := os.MkdirAll(dirPath, shared.TestDirPermission); err != nil {
			t.Fatalf("Failed to create directory structure %s: %v", dirPath, err)
		}
		createdPaths = append(createdPaths, dirPath)

		// Create files in the directory
		for _, fileSpec := range dirSpec.Files {
			filePath := CreateTestFile(t, dirPath, fileSpec.Name, []byte(fileSpec.Content))
			createdPaths = append(createdPaths, filePath)
		}
	}

	return createdPaths
}

// SetupTempDirWithStructure creates a temp directory with a structured layout.
func SetupTempDirWithStructure(t *testing.T, dirSpecs []DirSpec) string {
	t.Helper()
	rootDir := t.TempDir()
	CreateTestDirectoryStructure(t, rootDir, dirSpecs)

	return rootDir
}

// Error assertion helpers - safe to use across packages.

// AssertError checks if an error matches the expected state.
// If wantErr is true, expects err to be non-nil.
// If wantErr is false, expects err to be nil and fails if it's not.
func AssertError(t *testing.T, err error, wantErr bool, operation string) {
	t.Helper()
	if (err != nil) != wantErr {
		if wantErr {
			t.Errorf(shared.TestMsgOperationNoError, operation)
		} else {
			t.Errorf("Operation %s unexpected error: %v", operation, err)
		}
	}
}

// AssertNoError fails the test if err is not nil.
func AssertNoError(t *testing.T, err error, operation string) {
	t.Helper()
	if err != nil {
		t.Errorf(shared.TestMsgOperationFailed, operation, err)
	}
}

// AssertExpectedError fails the test if err is nil when an error is expected.
func AssertExpectedError(t *testing.T, err error, operation string) {
	t.Helper()
	if err == nil {
		t.Errorf(shared.TestMsgOperationNoError, operation)
	}
}

// AssertErrorContains checks that error contains the expected substring.
func AssertErrorContains(t *testing.T, err error, expectedSubstring, operation string) {
	t.Helper()
	if err == nil {
		t.Errorf("Operation %s expected error containing %q but got none", operation, expectedSubstring)

		return
	}
	if !strings.Contains(err.Error(), expectedSubstring) {
		t.Errorf("Operation %s error %q should contain %q", operation, err.Error(), expectedSubstring)
	}
}

// ValidateErrorCase checks error expectations and optionally validates error message content.
// This is a comprehensive helper that combines error checking with substring matching.
func ValidateErrorCase(t *testing.T, err error, wantErr bool, errContains string, operation string) {
	t.Helper()
	if wantErr {
		if err == nil {
			t.Errorf("%s: expected error but got none", operation)

			return
		}
		if errContains != "" && !strings.Contains(err.Error(), errContains) {
			t.Errorf("%s: expected error containing %q, got: %v", operation, errContains, err)
		}
	} else {
		if err != nil {
			t.Errorf("%s: unexpected error: %v", operation, err)
		}
	}
}

// VerifyStructuredError validates StructuredError properties.
// This helper ensures structured errors have the expected Type and Code values.
func VerifyStructuredError(t *testing.T, err error, expectedType shared.ErrorType, expectedCode string) {
	t.Helper()
	var structErr *shared.StructuredError
	if !errors.As(err, &structErr) {
		t.Errorf("expected StructuredError, got: %T", err)

		return
	}
	if structErr.Type != expectedType {
		t.Errorf("expected Type %v, got %v", expectedType, structErr.Type)
	}
	if structErr.Code != expectedCode {
		t.Errorf("expected Code %q, got %q", expectedCode, structErr.Code)
	}
}
