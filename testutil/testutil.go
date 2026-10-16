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
//	  - Use AssertNoError() when expecting success
//	  - Use AssertExpectedError() when expecting failure
//	  - Use AssertErrorContains() for substring validation
//
//	Configuration:
//	  - Use ResetConfig() to reset and reload between tests (it calls
//	    config.LoadConfig() itself)
//	  - Use SetConfigKeys() to start from defaults plus explicit values
//
//	Best Practices:
//	  - Always use t.Helper() in test helper functions
//	  - Use descriptive operation names in assertions
//	  - Prefer table-driven tests for multiple scenarios
//	  - Use testutil.ErrTestError for standard test errors
package testutil

import (
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/ivuorinen/gibidify/config"
	"github.com/ivuorinen/gibidify/shared"
)

// RunIsolated runs m with XDG_CONFIG_HOME and HOME pointed at an empty
// temporary directory and is meant to be the whole body of a package's
// TestMain. config.LoadConfig searches both, so without it the suite read the
// developer's own ~/.config/gibidify/config.yaml: a valid restrictive config
// failed 11 tests, and a passing run tested that config instead of the defaults.
// Tests that need a config file still pass their own path, which is searched
// first.
func RunIsolated(m *testing.M) {
	dir, err := os.MkdirTemp("", "gibidify-test-home-*")
	if err != nil {
		panic("creating isolated config home: " + err.Error())
	}
	defer func() {
		if err := os.RemoveAll(dir); err != nil {
			panic("removing isolated config home: " + err.Error())
		}
	}()

	for _, key := range []string{"XDG_CONFIG_HOME", "HOME"} {
		if err := os.Setenv(key, dir); err != nil {
			panic("setting " + key + ": " + err.Error())
		}
	}

	// TestMain returning hands m.Run's result to os.Exit.
	m.Run()
}

// SuppressLogs suppresses logger output during testing to keep test output clean.
// Returns a function that restores the logger output in effect when
// SuppressLogs was called.
func SuppressLogs(t testing.TB) func() {
	t.Helper()
	logger := shared.GetLogger()
	previous := logger.Output()

	logger.SetOutput(io.Discard)

	return func() {
		logger.SetOutput(previous)
	}
}

// OutputRestoreFunc represents a function that restores output after suppression.
type OutputRestoreFunc func()

// SuppressAllOutput suppresses both stdout and stderr during testing.
// This captures all output including UI messages, progress bars, and direct prints.
// Returns a function that should be called to restore original output.
func SuppressAllOutput(t testing.TB) OutputRestoreFunc {
	t.Helper()

	// Save original stdout and stderr
	originalStdout := os.Stdout
	originalStderr := os.Stderr

	// Suppress logger output as well
	logger := shared.GetLogger()
	originalLogOutput := logger.Output()
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
		logger.SetOutput(originalLogOutput)
	}
}

// CreateTestFile creates a test file with the given content and returns its path.
func CreateTestFile(t testing.TB, dir, filename string, content []byte) string {
	t.Helper()
	filePath := filepath.Join(dir, filename)
	if err := os.WriteFile(filePath, content, TestFilePermission); err != nil {
		t.Fatalf("Failed to write file %s: %v", filePath, err)
	}

	return filePath
}

// ReadFile returns the contents of a file the test created, failing the test
// if it cannot be read.
func ReadFile(t testing.TB, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		t.Fatalf("Failed to read file %s: %v", path, err)
	}

	return data
}

// CreateTempOutputFile creates a temporary output file and returns the file handle and path.
func CreateTempOutputFile(t testing.TB, pattern string) (file *os.File, path string) {
	t.Helper()
	outFile, err := os.CreateTemp(t.TempDir(), pattern)
	if err != nil {
		t.Fatalf("Failed to create temp output file: %v", err)
	}
	path = outFile.Name()

	return outFile, path
}

// CreateTestDirectory creates a test directory and returns its path.
func CreateTestDirectory(t testing.TB, parent, name string) string {
	t.Helper()
	dirPath := filepath.Join(parent, name)
	if err := os.Mkdir(dirPath, TestDirPermission); err != nil {
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
func CreateTestFiles(t testing.TB, rootDir string, fileSpecs []FileSpec) []string {
	t.Helper()
	createdFiles := make([]string, 0, len(fileSpecs))
	for _, spec := range fileSpecs {
		filePath := CreateTestFile(t, rootDir, spec.Name, []byte(spec.Content))
		createdFiles = append(createdFiles, filePath)
	}

	return createdFiles
}

// ResetConfig clears the config store, searches configPath first when it is
// set, and loads the configuration, failing the test if it does not load.
// The store is cleared again when the test ends.
func ResetConfig(t testing.TB, configPath string) {
	t.Helper()
	config.Reset()
	t.Cleanup(config.Reset)
	if configPath != "" {
		config.AddConfigPath(configPath)
	}
	if err := config.LoadConfig(); err != nil {
		t.Fatalf("loading config: %v", err)
	}
}

// SetConfigKeys resets the config store to defaults and sets keyValues as
// explicit values, without reading any config file. The store is cleared
// again when the test ends.
func SetConfigKeys(t testing.TB, keyValues map[string]any) {
	t.Helper()
	config.Reset()
	t.Cleanup(config.Reset)
	config.SetDefaultConfig()
	for key, value := range keyValues {
		config.Set(key, value)
	}
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
func VerifyContentContains(t testing.TB, content string, expectedSubstrings []string) {
	t.Helper()
	for _, expected := range expectedSubstrings {
		if !strings.Contains(content, expected) {
			t.Errorf("Content missing expected substring: %s", expected)
		}
	}
}

// MustSucceed fails the test if the error is not nil.
func MustSucceed(t testing.TB, err error, operation string) {
	t.Helper()
	if err != nil {
		t.Fatalf(TestMsgOperationFailed, operation, err)
	}
}

// CloseFile closes a file and reports errors to the test.
func CloseFile(t testing.TB, file *os.File) {
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
func CreateTestDirectoryStructure(t testing.TB, rootDir string, dirSpecs []DirSpec) []string {
	t.Helper()
	createdPaths := make([]string, 0)

	for _, dirSpec := range dirSpecs {
		dirPath := filepath.Join(rootDir, dirSpec.Path)
		if err := os.MkdirAll(dirPath, TestDirPermission); err != nil {
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
func SetupTempDirWithStructure(t testing.TB, dirSpecs []DirSpec) string {
	t.Helper()
	rootDir := t.TempDir()
	CreateTestDirectoryStructure(t, rootDir, dirSpecs)

	return rootDir
}

// Error assertion helpers - safe to use across packages.

// AssertNoError fails the test if err is not nil.
func AssertNoError(t testing.TB, err error, operation string) {
	t.Helper()
	if err != nil {
		t.Errorf(TestMsgOperationFailed, operation, err)
	}
}

// AssertExpectedError fails the test if err is nil when an error is expected.
func AssertExpectedError(t testing.TB, err error, operation string) {
	t.Helper()
	if err == nil {
		t.Errorf(TestMsgOperationNoError, operation)
	}
}

// AssertErrorContains checks that error contains the expected substring.
func AssertErrorContains(t testing.TB, err error, expectedSubstring, operation string) {
	t.Helper()
	if err == nil {
		t.Errorf("Operation %s expected error containing %q but got none", operation, expectedSubstring)

		return
	}
	if !strings.Contains(err.Error(), expectedSubstring) {
		t.Errorf("Operation %s error %q should contain %q", operation, err.Error(), expectedSubstring)
	}
}
