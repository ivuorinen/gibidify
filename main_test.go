package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ivuorinen/gibidify/cli"
	"github.com/ivuorinen/gibidify/testutil"
)

const (
	testFileCount = 1000
)

// resetFlagState resets the global flag state to allow multiple test runs.
func resetFlagState() {
	// Reset both the flag.CommandLine and cli global state for clean testing
	cli.ResetFlags()
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
}

// TestIntegrationFullCLI simulates a full run of the CLI application using adaptive concurrency.
func TestIntegrationFullCLI(t *testing.T) {
	// Suppress logs for cleaner test output
	restore := testutil.SuppressAllOutput(t)
	defer restore()

	resetFlagState()
	srcDir := setupTestFiles(t)
	outFilePath := setupOutputFile(t)
	setupCLIArgs(srcDir, outFilePath)

	// Run the application with a background context.
	ctx := t.Context()
	if runErr := run(ctx); runErr != nil {
		t.Fatalf("Run failed: %v", runErr)
	}

	verifyOutput(t, outFilePath)
}

// setupTestFiles creates test files and returns the source directory.
func setupTestFiles(t *testing.T) string {
	t.Helper()
	srcDir := t.TempDir()

	// Create two test files.
	testutil.CreateTestFiles(t, srcDir, []testutil.FileSpec{
		{Name: "file1.txt", Content: "Hello World"},
		{Name: "file2.go", Content: "package main\nfunc main() {}"},
	})

	return srcDir
}

// setupOutputFile creates a temporary output file and returns its path.
func setupOutputFile(t *testing.T) string {
	t.Helper()
	outFile, outFilePath := testutil.CreateTempOutputFile(t, "gibidify_output.txt")
	testutil.CloseFile(t, outFile)

	return outFilePath
}

// setupCLIArgs configures the CLI arguments for testing.
func setupCLIArgs(srcDir, outFilePath string) {
	testutil.SetupCLIArgs(srcDir, outFilePath, "PREFIX", "SUFFIX", 2)
}

// verifyOutput checks that the output file contains expected content.
func verifyOutput(t *testing.T, outFilePath string) {
	t.Helper()
	data, err := os.ReadFile(outFilePath)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}
	output := string(data)
	testutil.VerifyContentContains(t, output, []string{"PREFIX", "Hello World", "SUFFIX"})
}

// TestIntegrationCancellation verifies that the application correctly cancels processing when the context times out.
func TestIntegrationCancellation(t *testing.T) {
	// Suppress logs for cleaner test output
	restore := testutil.SuppressAllOutput(t)
	defer restore()

	resetFlagState()
	// Create a temporary source directory with many files to simulate a long-running process.
	srcDir := t.TempDir()

	// Create a large number of small files.
	for i := range testFileCount {
		fileName := fmt.Sprintf("file%d.txt", i)
		testutil.CreateTestFile(t, srcDir, fileName, []byte("Content"))
	}

	// Create a temporary output file.
	outFile, outFilePath := testutil.CreateTempOutputFile(t, "gibidify_output.txt")
	testutil.CloseFile(t, outFile)
	defer func() {
		if removeErr := os.Remove(outFilePath); removeErr != nil {
			t.Fatalf("cleanup output file: %v", removeErr)
		}
	}()

	// Set up CLI arguments.
	testutil.SetupCLIArgs(srcDir, outFilePath, "PREFIX", "SUFFIX", 2)

	// Create a context with a short timeout to force cancellation.
	ctx, cancel := context.WithTimeout(
		t.Context(),
		5*time.Millisecond,
	)
	defer cancel()

	// Run the application; we expect an error due to cancellation.
	runErr := run(ctx)
	if runErr == nil {
		t.Error("Expected Run to fail due to cancellation, but it succeeded")
	}
}

// BenchmarkRun benchmarks the run() function performance.
func BenchmarkRun(b *testing.B) {
	// Save original args and flags
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	ctx := context.Background()

	for b.Loop() {
		// Create fresh directories for each iteration
		srcDir := b.TempDir()
		outDir := b.TempDir()

		// Create benchmark files
		files := map[string]string{
			"bench1.go":  "package main\n// Benchmark file 1",
			"bench2.txt": "Benchmark content file 2",
			"bench3.md":  "# Benchmark markdown file",
		}

		for name, content := range files {
			filePath := filepath.Join(srcDir, name)
			if err := os.WriteFile(filePath, []byte(content), 0o600); err != nil {
				b.Fatalf("Failed to create benchmark file %s: %v", name, err)
			}
		}

		outFilePath := filepath.Join(outDir, "output.md")

		// Reset flags for each iteration
		cli.ResetFlags()
		flag.CommandLine = flag.NewFlagSet("bench", flag.ContinueOnError)

		os.Args = []string{"gibidify", "-source", srcDir, "-destination", outFilePath, "-no-ui"}

		if err := run(ctx); err != nil {
			b.Fatalf("run() failed in benchmark: %v", err)
		}
	}
}

// BenchmarkRunLargeFiles benchmarks the run() function with larger files.
func BenchmarkRunLargeFiles(b *testing.B) {
	// Save original args
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	largeContent := strings.Repeat("This is a large file for benchmarking purposes.\n", 1000)
	ctx := context.Background()

	for b.Loop() {
		// Create fresh directories for each iteration
		srcDir := b.TempDir()
		outDir := b.TempDir()

		// Create large benchmark files
		files := map[string]string{
			"large1.go":  "package main\n" + largeContent,
			"large2.txt": largeContent,
			"large3.md":  "# Large File\n" + largeContent,
		}

		for name, content := range files {
			filePath := filepath.Join(srcDir, name)
			if err := os.WriteFile(filePath, []byte(content), 0o600); err != nil {
				b.Fatalf("Failed to create large benchmark file %s: %v", name, err)
			}
		}

		outFilePath := filepath.Join(outDir, "output.md")

		// Reset flags for each iteration
		cli.ResetFlags()
		flag.CommandLine = flag.NewFlagSet("bench", flag.ContinueOnError)

		os.Args = []string{"gibidify", "-source", srcDir, "-destination", outFilePath, "-no-ui"}

		if err := run(ctx); err != nil {
			b.Fatalf("run() failed in large files benchmark: %v", err)
		}
	}
}
