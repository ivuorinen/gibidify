package fileproc_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ivuorinen/gibidify/fileproc"
)

func TestCollectFilesWithFakeWalker(t *testing.T) {
	// Instead of using the production walker, use FakeWalker.
	expectedFiles := []string{
		"/path/to/file1.txt",
		"/path/to/file2.go",
	}
	fake := fileproc.FakeWalker{
		Files: expectedFiles,
		Err:   nil,
	}

	// Use fake.Walk directly.
	files, err := fake.Walk("dummyRoot")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(files) != len(expectedFiles) {
		t.Fatalf("Expected %d files, got %d", len(expectedFiles), len(files))
	}

	for i, f := range files {
		if f != expectedFiles[i] {
			t.Errorf("Expected file %s, got %s", expectedFiles[i], f)
		}
	}
}

func TestCollectFilesError(t *testing.T) {
	// Fake walker returns an error.
	fake := fileproc.FakeWalker{
		Files: nil,
		Err:   os.ErrNotExist,
	}

	_, err := fake.Walk("dummyRoot")
	if err == nil {
		t.Fatal("Expected an error, got nil")
	}
}

// TestCollectFiles tests the actual CollectFiles function with a real directory.
func TestCollectFiles(t *testing.T) {
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
	files, err := fileproc.CollectFiles(tmpDir)
	if err != nil {
		t.Fatalf("CollectFiles failed: %v", err)
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
	_, err := fileproc.CollectFiles("/non/existent/directory")
	if err == nil {
		t.Error("Expected error for non-existent directory, got nil")
	}
}

// TestCollectFiles_EmptyDirectory tests CollectFiles with an empty directory.
func TestCollectFilesEmptyDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	// Don't create any files

	files, err := fileproc.CollectFiles(tmpDir)
	if err != nil {
		t.Fatalf("CollectFiles failed on empty directory: %v", err)
	}

	if len(files) != 0 {
		t.Errorf("Expected 0 files in empty directory, got %d", len(files))
	}
}
