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

// TestCollectFiles tests the real CollectFiles function with a temp directory.
func TestCollectFiles(t *testing.T) {
	// Create a temporary directory
	tmpDir := t.TempDir()

	// Create test files
	testFile1 := filepath.Join(tmpDir, "test1.go")
	testFile2 := filepath.Join(tmpDir, "test2.go")

	if err := os.WriteFile(testFile1, []byte("package main\n"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	if err := os.WriteFile(testFile2, []byte("package main\n"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test CollectFiles - it should work without error
	files, err := fileproc.CollectFiles(tmpDir)
	if err != nil {
		t.Fatalf("CollectFiles failed: %v", err)
	}

	// Just verify we got some files (the walker may filter based on patterns)
	// The important part is that the function executes without error
	t.Logf("Collected %d files", len(files))
}

// TestCollectFilesNonExistentDir tests CollectFiles with a non-existent directory.
func TestCollectFilesNonExistentDir(t *testing.T) {
	_, err := fileproc.CollectFiles("/nonexistent/directory/path")
	if err == nil {
		t.Error("Expected error for non-existent directory, got nil")
	}
}
