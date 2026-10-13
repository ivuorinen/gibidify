package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"testing"
	"time"

	"github.com/ivuorinen/gibidify/fileproc"
	"github.com/ivuorinen/gibidify/shared"
	"github.com/ivuorinen/gibidify/testutil"
)

// pipelineFlags returns quiet JSON flags for src and dest.
func pipelineFlags(src, dest string, concurrency int) *Flags {
	return &Flags{
		SourceDir:   src,
		Destination: dest,
		Format:      shared.FormatJSON,
		Concurrency: concurrency,
		NoUI:        true,
	}
}

// runWithTimeout runs Process and fails the test if it does not return in time,
// so a pipeline deadlock fails fast instead of hanging the suite.
func runWithTimeout(t *testing.T, flags *Flags) error {
	t.Helper()
	done := make(chan error, 1)
	go func() { done <- NewProcessor(flags).Process(context.Background()) }()

	select {
	case err := <-done:
		return err
	case <-time.After(20 * time.Second):
		t.Fatal("Process did not return: pipeline deadlock")

		return nil
	}
}

// outputPaths returns the file paths listed in a JSON output file.
func outputPaths(t *testing.T, path string) []string {
	t.Helper()
	data := testutil.ReadFile(t, path)
	var out fileproc.OutputData
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatalf("parse output: %v\n%s", err, data)
	}
	paths := make([]string, 0, len(out.Files))
	for _, f := range out.Files {
		paths = append(paths, f.Path)
	}

	return paths
}

// TestProcessFailsWhenAFileCannotBeRead verifies a per-file failure makes the
// run return an error (non-zero exit) while the readable files are still written.
func TestProcessFailsWhenAFileCannotBeRead(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root can read mode-000 files")
	}
	restore := testutil.SuppressAllOutput(t)
	defer restore()
	testutil.ResetConfig(t, "")

	src := t.TempDir()
	testutil.CreateTestFile(t, src, "a.go", []byte("package a\n"))
	locked := testutil.CreateTestFile(t, src, "b.go", []byte("package b\n"))
	if err := os.Chmod(locked, 0o000); err != nil {
		t.Fatal(err)
	}
	dest := filepath.Join(t.TempDir(), "out.json")

	if err := runWithTimeout(t, pipelineFlags(src, dest, 2)); err == nil {
		t.Fatal("expected an error when a file cannot be read")
	}
	if got := outputPaths(t, dest); !slices.Equal(got, []string{"a.go"}) {
		t.Errorf("expected the readable file in the output, got %v", got)
	}
}

// TestProcessWriteFailureReturnsErrorWithoutHanging guards the deadlock where a
// failed writer stopped draining while more files than the channel buffer queued.
func TestProcessWriteFailureReturnsErrorWithoutHanging(t *testing.T) {
	if _, err := os.Stat("/dev/full"); err != nil {
		t.Skip("/dev/full not available")
	}
	restore := testutil.SuppressAllOutput(t)
	defer restore()
	testutil.ResetConfig(t, "")

	src := t.TempDir()
	for i := range 3 * channelBuffer {
		testutil.CreateTestFile(t, src, fmt.Sprintf("f%03d.go", i), []byte("package p\n"))
	}

	if err := runWithTimeout(t, pipelineFlags(src, "/dev/full", 2)); err == nil {
		t.Fatal("expected an error writing to a full device")
	}
}

// TestProcessDoesNotIngestOwnOutput verifies a destination inside the source
// tree is excluded, so a second run lists the same files as the first.
func TestProcessDoesNotIngestOwnOutput(t *testing.T) {
	restore := testutil.SuppressAllOutput(t)
	defer restore()
	testutil.ResetConfig(t, "")

	src := t.TempDir()
	testutil.CreateTestFile(t, src, "a.go", []byte("package a\n"))
	dest := filepath.Join(src, "out.json")

	for run := 1; run <= 2; run++ {
		if err := runWithTimeout(t, pipelineFlags(src, dest, 2)); err != nil {
			t.Fatalf("run %d: %v", run, err)
		}
		if got := outputPaths(t, dest); !slices.Equal(got, []string{"a.go"}) {
			t.Errorf("run %d: expected [a.go], got %v", run, got)
		}
	}
}

// TestProcessOutputOrderIsDeterministic verifies output follows collection
// order regardless of which worker finishes first.
func TestProcessOutputOrderIsDeterministic(t *testing.T) {
	restore := testutil.SuppressAllOutput(t)
	defer restore()
	testutil.ResetConfig(t, "")

	src := t.TempDir()
	want := make([]string, 0, 200)
	for i := range 200 {
		name := fmt.Sprintf("f%03d.go", i)
		want = append(want, name)
		testutil.CreateTestFile(t, src, name, []byte(fmt.Sprintf("package p // %d\n", i)))
	}
	dest := filepath.Join(t.TempDir(), "out.json")

	if err := runWithTimeout(t, pipelineFlags(src, dest, 8)); err != nil {
		t.Fatal(err)
	}
	if got := outputPaths(t, dest); !slices.Equal(got, want) {
		t.Errorf("output order differs from collection order:\n got %v\nwant %v", got, want)
	}
}

// TestProcessCanceledKeepsPreviousOutput verifies a canceled run leaves an
// existing destination byte-for-byte intact and no temporary file behind.
// os.Create used to truncate the destination before processing started, so an
// interrupted re-run destroyed the previous good output.
func TestProcessCanceledKeepsPreviousOutput(t *testing.T) {
	restore := testutil.SuppressAllOutput(t)
	defer restore()
	testutil.ResetConfig(t, "")

	src := t.TempDir()
	testutil.CreateTestFile(t, src, "a.go", []byte("package a\n"))
	outDir := t.TempDir()
	previous := []byte(`{"previous":"run"}`)
	dest := testutil.CreateTestFile(t, outDir, "out.json", previous)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := NewProcessor(pipelineFlags(src, dest, 2)).Process(ctx); err == nil {
		t.Fatal("expected an error for a canceled run")
	}
	if got := testutil.ReadFile(t, dest); string(got) != string(previous) {
		t.Errorf("destination changed: got %q, want %q", got, previous)
	}
	entries, err := os.ReadDir(outDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Errorf("expected only out.json in %s, found %d entries (temporary file left behind?)", outDir, len(entries))
	}
}

// TestProcessReplacesOutputAndKeepsMode verifies a successful run replaces an
// existing destination and keeps its permissions.
func TestProcessReplacesOutputAndKeepsMode(t *testing.T) {
	restore := testutil.SuppressAllOutput(t)
	defer restore()
	testutil.ResetConfig(t, "")

	src := t.TempDir()
	testutil.CreateTestFile(t, src, "a.go", []byte("package a\n"))
	dest := testutil.CreateTestFile(t, t.TempDir(), "out.json", []byte("old"))
	// Any mode other than the 0600 default proves it was carried over.
	if err := os.Chmod(dest, 0o400); err != nil {
		t.Fatal(err)
	}

	if err := runWithTimeout(t, pipelineFlags(src, dest, 2)); err != nil {
		t.Fatal(err)
	}
	if got := outputPaths(t, dest); !slices.Equal(got, []string{"a.go"}) {
		t.Errorf("output files = %v, want [a.go]", got)
	}
	info, err := os.Stat(dest)
	if err != nil {
		t.Fatal(err)
	}
	if runtime.GOOS != "windows" && info.Mode().Perm() != 0o400 {
		t.Errorf("mode = %v, want 0400 kept from the replaced file", info.Mode().Perm())
	}
}

// TestProcessCanceledRemovesPartialOutput verifies a canceled run returns an
// error and leaves no truncated output behind.
func TestProcessCanceledRemovesPartialOutput(t *testing.T) {
	restore := testutil.SuppressAllOutput(t)
	defer restore()
	testutil.ResetConfig(t, "")

	src := t.TempDir()
	testutil.CreateTestFile(t, src, "a.go", []byte("package a\n"))
	dest := filepath.Join(t.TempDir(), "out.json")

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := NewProcessor(pipelineFlags(src, dest, 2)).Process(ctx); err == nil {
		t.Fatal("expected an error for a canceled run")
	}
	if _, err := os.Stat(dest); !os.IsNotExist(err) {
		t.Errorf("expected partial output to be removed, stat err: %v", err)
	}
}
