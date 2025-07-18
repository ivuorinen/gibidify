// Package testutil provides common testing utilities and helper functions.
package testutil

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/spf13/viper"

	"github.com/ivuorinen/gibidify/config"
)

const (
	// FilePermission is the default file permission for test files.
	FilePermission = 0o644
	// DirPermission is the default directory permission for test directories.
	DirPermission = 0o755
)

// CreateTestFile creates a test file with the given content and returns its path.
func CreateTestFile(t *testing.T, dir, filename string, content []byte) string {
	t.Helper()
	filePath := filepath.Join(dir, filename)
	if err := os.WriteFile(filePath, content, FilePermission); err != nil {
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
	if err := os.Mkdir(dirPath, DirPermission); err != nil {
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

// SetupCLIArgs configures os.Args for CLI testing.
func SetupCLIArgs(srcDir, outFilePath, prefix, suffix string, concurrency int) {
	os.Args = []string{
		"gibidify",
		"-source", srcDir,
		"-destination", outFilePath,
		"-prefix", prefix,
		"-suffix", suffix,
		"-concurrency", strconv.Itoa(concurrency),
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
		t.Fatalf("Operation %s failed: %v", operation, err)
	}
}

// CloseFile closes a file and reports errors to the test.
func CloseFile(t *testing.T, file *os.File) {
	t.Helper()
	if err := file.Close(); err != nil {
		t.Errorf("Failed to close file: %v", err)
	}
}
