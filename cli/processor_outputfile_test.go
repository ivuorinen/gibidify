package cli

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ivuorinen/gibidify/shared"
	"github.com/ivuorinen/gibidify/testutil"
)

// devNull returns os.DevNull, skipping the test where it is not a character
// device.
func devNull(t *testing.T) string {
	t.Helper()
	info, err := os.Stat(os.DevNull)
	if err != nil || info.Mode()&os.ModeCharDevice == 0 {
		t.Skipf("%s is not a character device", os.DevNull)
	}

	return os.DevNull
}

// assertStillDevice fails the test if path is no longer a character device,
// meaning a run renamed a file over it.
func assertStillDevice(t *testing.T, path string) {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil || info.Mode()&os.ModeCharDevice == 0 {
		t.Errorf("%s is no longer a character device (stat error: %v)", path, err)
	}
}

// TestProcessToDeviceWritesInPlace verifies a device destination is written
// directly on success and on cancellation, never replaced by a renamed file.
func TestProcessToDeviceWritesInPlace(t *testing.T) {
	dest := devNull(t)
	src := t.TempDir()
	testutil.CreateTestFile(t, src, "a.go", []byte("package a\n"))

	t.Run("success", func(t *testing.T) {
		testutil.AssertNoError(t, runWithTimeout(t, pipelineFlags(src, dest, 1)), "Process")
		assertStillDevice(t, dest)
	})

	t.Run("canceled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		_, err := NewProcessor(pipelineFlags(src, dest, 1)).processFiles(ctx, nil)
		if !errors.Is(err, context.Canceled) {
			t.Errorf("processFiles() error = %v, want context.Canceled", err)
		}
		assertStillDevice(t, dest)
	})
}

// TestCreateOutputFileDirectoryDestination verifies a directory destination is
// reported as a create failure instead of being written into.
func TestCreateOutputFileDirectoryDestination(t *testing.T) {
	dest := t.TempDir()
	out, err := NewProcessor(pipelineFlags(t.TempDir(), dest, 1)).createOutputFile()
	if out != nil {
		_ = out.Close()
		t.Error("createOutputFile() returned a file for a directory destination")
	}
	assertErrorCode(t, err, shared.CodeIOFileCreate)
}

// TestOutputFileCommitRenameFailure verifies a failed rename reports the
// destination and removes the temporary output.
func TestOutputFileCommitRenameFailure(t *testing.T) {
	dir := t.TempDir()
	tmp := testutil.CreateTestFile(t, dir, ".out.json.tmp", []byte("{}"))
	dest := filepath.Join(dir, "missing", "out.json")

	err := (&outputFile{tmpPath: tmp}).commit(dest)
	testutil.AssertErrorContains(t, err, "replacing "+dest, "commit")
	if _, statErr := os.Stat(tmp); !os.IsNotExist(statErr) {
		t.Errorf("temporary output %s not removed after a failed commit (stat error: %v)", tmp, statErr)
	}
}

// TestOutputFileDiscardRemoveFailure verifies a temporary output that cannot be
// removed is logged rather than silently kept.
func TestOutputFileDiscardRemoveFailure(t *testing.T) {
	logs := captureLogs(t)
	// A non-empty directory cannot be removed with os.Remove.
	tmp := testutil.CreateTestDirectory(t, t.TempDir(), "partial")
	testutil.CreateTestFile(t, tmp, "keep.txt", []byte("x"))

	(&outputFile{tmpPath: tmp}).discard()

	if !strings.Contains(logs.String(), "Failed to remove partial output") {
		t.Errorf("log does not report the removal failure:\n%s", logs.String())
	}
	if _, err := os.Stat(tmp); err != nil {
		t.Errorf("stat %s: %v", tmp, err)
	}
}
