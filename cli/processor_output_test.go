package cli

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ivuorinen/gibidify/shared"
	"github.com/ivuorinen/gibidify/testutil"
)

// captureLogs sends the shared logger's warn-and-above output to a buffer for
// the test.
func captureLogs(t *testing.T) *bytes.Buffer {
	t.Helper()
	buf := &bytes.Buffer{}
	logger := shared.GetLogger()
	logger.SetLevel(shared.LogLevelWarn)
	logger.SetOutput(buf)
	t.Cleanup(func() { logger.SetOutput(os.Stderr) })

	return buf
}

// assertErrorCode fails the test unless err is a StructuredError with code.
func assertErrorCode(t *testing.T, err error, code string) {
	t.Helper()
	var structErr *shared.StructuredError
	if !errors.As(err, &structErr) {
		t.Fatalf("error = %v (%T), want a StructuredError with code %s", err, err, code)
	}
	if structErr.Code != code {
		t.Errorf("error code = %s, want %s (error: %v)", structErr.Code, code, err)
	}
}

// assertNoOutput fails the test if path exists.
func assertNoOutput(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("output %s exists after a failed run (stat error: %v)", path, err)
	}
}

// TestProcessStopsBeforeWriting verifies Process returns the collection or
// pre-validation error without creating the output.
func TestProcessStopsBeforeWriting(t *testing.T) {
	tests := []struct {
		name   string
		config map[string]any
		source func(t *testing.T) string
		code   string
	}{
		{
			name:   "source cannot be walked",
			source: func(t *testing.T) string { t.Helper(); return filepath.Join(t.TempDir(), "missing") },
			code:   shared.CodeProcessingCollection,
		},
		{
			name:   "too many files",
			config: map[string]any{shared.ConfigKeyResourceLimitsMaxFiles: 1},
			source: func(t *testing.T) string {
				t.Helper()
				dir := t.TempDir()
				testutil.CreateTestFile(t, dir, "a.go", []byte("package a\n"))
				testutil.CreateTestFile(t, dir, "b.go", []byte("package b\n"))

				return dir
			},
			code: shared.CodeResourceLimitFiles,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testutil.SetConfigKeys(t, tt.config)
			dest := filepath.Join(t.TempDir(), "out.json")
			err := runWithTimeout(t, pipelineFlags(tt.source(t), dest, 1))
			assertErrorCode(t, err, tt.code)
			assertNoOutput(t, dest)
		})
	}
}

// TestValidateFileCollectionTotalSize verifies the total size cap rejects the
// collection once the running total passes it.
func TestValidateFileCollectionTotalSize(t *testing.T) {
	testutil.SetConfigKeys(t, map[string]any{shared.ConfigKeyResourceLimitsMaxTotalSize: 15})
	dir := t.TempDir()
	files := []string{
		testutil.CreateTestFile(t, dir, "a.txt", []byte("0123456789")),
		testutil.CreateTestFile(t, dir, "b.txt", []byte("0123456789")),
	}

	err := NewProcessor(pipelineFlags(dir, "", 1)).validateFileCollection(files)
	assertErrorCode(t, err, shared.CodeResourceLimitTotalSize)
	testutil.AssertErrorContains(t, err, "total file size (20 bytes)", "validateFileCollection")
}

// TestValidateFileCollectionUnstattableFile verifies a file that vanished
// after collection does not fail pre-validation but is reported.
func TestValidateFileCollectionUnstattableFile(t *testing.T) {
	testutil.SetConfigKeys(t, nil)
	logs := captureLogs(t)
	dir := t.TempDir()
	files := []string{testutil.CreateTestFile(t, dir, "a.txt", []byte("x")), filepath.Join(dir, "gone.txt")}

	err := NewProcessor(pipelineFlags(dir, "", 1)).validateFileCollection(files)
	testutil.AssertNoError(t, err, "validateFileCollection")
	if !strings.Contains(logs.String(), "Could not stat 1 files") {
		t.Errorf("log does not report the unstattable file:\n%s", logs.String())
	}
}

// TestProcessFilesSetupErrors verifies processFiles fails before starting the
// pipeline when the source cannot be resolved or the output cannot be opened.
func TestProcessFilesSetupErrors(t *testing.T) {
	t.Run("source cannot be resolved", func(t *testing.T) {
		removedWorkingDir(t)
		dest := filepath.Join(t.TempDir(), "out.json")
		_, err := NewProcessor(pipelineFlags("src", dest, 1)).processFiles(context.Background(), nil)
		assertErrorCode(t, err, shared.CodeFSPathResolution)
		assertNoOutput(t, dest)
	})

	t.Run("output directory missing", func(t *testing.T) {
		dest := filepath.Join(t.TempDir(), "missing", "out.json")
		_, err := NewProcessor(pipelineFlags(t.TempDir(), dest, 1)).processFiles(context.Background(), nil)
		assertErrorCode(t, err, shared.CodeIOFileCreate)
	})
}

// TestProcessFilesCanceledWithNothingToSend verifies a run canceled before it
// starts fails even when there was no file to send, and leaves no output.
func TestProcessFilesCanceledWithNothingToSend(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	dir := t.TempDir()
	dest := filepath.Join(dir, "out.json")

	_, err := NewProcessor(pipelineFlags(dir, dest, 1)).processFiles(ctx, nil)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("processFiles() error = %v, want context.Canceled", err)
	}
	assertNoOutput(t, dest)
	if entries, _ := os.ReadDir(dir); len(entries) != 0 {
		t.Errorf("temporary output left behind: %v", entries)
	}
}
