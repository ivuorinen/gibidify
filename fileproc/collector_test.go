package fileproc_test

import (
	"os"
	"testing"

	fileproc "github.com/ivuorinen/gibidify/fileproc"
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
