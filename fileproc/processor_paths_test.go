package fileproc_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/ivuorinen/gibidify/fileproc"
	"github.com/ivuorinen/gibidify/shared"
	"github.com/ivuorinen/gibidify/testutil"
)

// TestProcessorSkipsDirectory verifies a path that opens as a directory is
// skipped as not a regular file and produces no output.
func TestProcessorSkipsDirectory(t *testing.T) {
	testutil.ResetConfig(t, "")
	root := t.TempDir()
	dir := testutil.CreateTestDirectory(t, root, "subdir")

	reqs, err := processOne(context.Background(), fileproc.NewFileProcessor(root), dir)
	if !errors.Is(err, fileproc.ErrSkipped) {
		t.Fatalf("expected ErrSkipped, got %v", err)
	}
	testutil.AssertErrorContains(t, err, "not a regular file", "process directory")
	if len(reqs) != 0 {
		t.Errorf("directory produced output: %+v", reqs)
	}
}

// TestProcessorUnrelatableRootKeepsFullPath verifies that when the file cannot
// be made relative to the root, the request carries the path as given.
func TestProcessorUnrelatableRootKeepsFullPath(t *testing.T) {
	testutil.ResetConfig(t, "")
	path := testutil.CreateTestFile(t, t.TempDir(), "a.go", []byte(testContent))

	// filepath.Rel fails between a relative root and an absolute path.
	reqs, err := processOne(context.Background(), fileproc.NewFileProcessor("relative-root"), path)
	if err != nil {
		t.Fatalf(testutil.TestMsgUnexpectedError, err)
	}
	if len(reqs) != 1 {
		t.Fatalf("got %d requests, want 1", len(reqs))
	}
	if reqs[0].Path != path || reqs[0].Content != testContent {
		t.Fatalf("request = %+v, want path %q with the file content", reqs[0], path)
	}
}

// TestProdWalkerUnresolvableRoot verifies a relative root that cannot be made
// absolute, because the working directory was removed, fails the walk.
func TestProdWalkerUnresolvableRoot(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("relies on getcwd failing for a removed working directory (Linux)")
	}
	gone := filepath.Join(t.TempDir(), "gone")
	if err := os.Mkdir(gone, 0o750); err != nil {
		t.Fatal(err)
	}
	t.Chdir(gone)
	if err := os.Remove(gone); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Getwd(); err == nil {
		t.Skip("getcwd still succeeds for a removed directory on this system")
	}

	found, err := fileproc.NewProdWalker(fileproc.WalkOptions{}).Walk("src")
	var se *shared.StructuredError
	if !errors.As(err, &se) || se.Code != shared.CodeFSPathResolution {
		t.Fatalf("error = %v, want code %s", err, shared.CodeFSPathResolution)
	}
	if found != nil {
		t.Fatalf("found = %v, want nil", found)
	}
}
