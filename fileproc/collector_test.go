package fileproc_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ivuorinen/gibidify/config"
	"github.com/ivuorinen/gibidify/fileproc"
	"github.com/ivuorinen/gibidify/shared"
)

// TestCollectFiles tests the actual CollectFiles function with a real directory.
func TestCollectFiles(t *testing.T) {
	// Collection filters by the configured file-size limit; load defaults so it is non-zero.
	config.Reset()
	config.SetDefaultConfig()

	// Create a temporary directory with test files
	tmpDir := t.TempDir()

	// Create test files with known supported extensions
	testFiles := map[string]string{
		"test1.go": "package main\n\nfunc main() {\n\t// Go file\n}",
		"test2.py": "# Python file\nprint('hello world')",
		"test3.js": "// JavaScript file\nconsole.log('hello');",
	}

	for name, content := range testFiles {
		filePath := filepath.Join(tmpDir, name)
		if err := os.WriteFile(filePath, []byte(content), 0o600); err != nil {
			t.Fatalf("Failed to create test file %s: %v", name, err)
		}
	}

	// Test CollectFiles
	files, skipped, err := fileproc.CollectFiles(tmpDir, fileproc.WalkOptions{})
	if err != nil {
		t.Fatalf("CollectFiles failed: %v", err)
	}
	if skipped != 0 {
		t.Errorf("skipped = %d, want 0", skipped)
	}

	// Verify we got the expected number of files
	if len(files) != len(testFiles) {
		t.Errorf("Expected %d files, got %d", len(testFiles), len(files))
	}

	// Verify all expected files are found
	foundFiles := make(map[string]bool)
	for _, file := range files {
		foundFiles[file] = true
	}

	for expectedFile := range testFiles {
		expectedPath := filepath.Join(tmpDir, expectedFile)
		if !foundFiles[expectedPath] {
			t.Errorf("Expected file %s not found in results", expectedPath)
		}
	}
}

// TestCollectFiles_NonExistentDirectory tests CollectFiles with a non-existent directory.
func TestCollectFilesNonExistentDirectory(t *testing.T) {
	_, _, err := fileproc.CollectFiles("/non/existent/directory", fileproc.WalkOptions{})
	if err == nil {
		t.Error("Expected error for non-existent directory, got nil")
	}
}

// TestCollectFiles_EmptyDirectory tests CollectFiles with an empty directory.
func TestCollectFilesEmptyDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	// Don't create any files

	files, _, err := fileproc.CollectFiles(tmpDir, fileproc.WalkOptions{})
	if err != nil {
		t.Fatalf("CollectFiles failed on empty directory: %v", err)
	}

	if len(files) != 0 {
		t.Errorf("Expected 0 files in empty directory, got %d", len(files))
	}
}

// TestCollectFilesReportsOversizedFiles verifies a file over fileSizeLimit is
// left out, counted, and logged at warn level with its path. It used to vanish
// with no log line while the summary reported 0 skipped.
func TestCollectFilesReportsOversizedFiles(t *testing.T) {
	config.Reset()
	config.SetDefaultConfig()
	config.Set(shared.ConfigKeyFileSizeLimit, shared.ConfigFileSizeLimitMin)

	var logs strings.Builder
	logger := shared.GetLogger()
	logger.SetLevel(shared.LogLevelWarn)
	logger.SetOutput(&logs)
	t.Cleanup(func() {
		logger.SetOutput(os.Stderr)
		config.Reset()
	})

	tmpDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmpDir, "small.go"), []byte("package a\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	big := bytes.Repeat([]byte("x"), shared.ConfigFileSizeLimitMin+1)
	if err := os.WriteFile(filepath.Join(tmpDir, "big.txt"), big, 0o600); err != nil {
		t.Fatal(err)
	}

	files, skipped, err := fileproc.CollectFiles(tmpDir, fileproc.WalkOptions{})
	if err != nil {
		t.Fatalf("CollectFiles: %v", err)
	}
	if len(files) != 1 || filepath.Base(files[0]) != "small.go" {
		t.Errorf("files = %v, want only small.go", files)
	}
	if skipped != 1 {
		t.Errorf("skipped = %d, want 1", skipped)
	}
	if !strings.Contains(logs.String(), "big.txt") || !strings.Contains(logs.String(), "exceeds fileSizeLimit") {
		t.Errorf("expected a warning naming big.txt, got: %q", logs.String())
	}
}
