package main

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/ivuorinen/gibidify/config"
	"github.com/ivuorinen/gibidify/testutil"
)

const (
	testFileCount = 1000
)

// TestMain configures test-time flags for packages.
func TestMain(m *testing.M) {
	// Inform packages that we're running under tests so they can adjust noisy logging.
	// The config package will suppress the specific info-level message about missing config
	// while still allowing tests to enable debug/info level logging when needed.
	config.SetRunningInTest(true)
	os.Exit(m.Run())
}

// TestIntegrationFullCLI simulates a full run of the CLI application using adaptive concurrency.
func TestIntegrationFullCLI(t *testing.T) {
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
	data, err := os.ReadFile(outFilePath) // #nosec G304 - test file path is controlled
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}
	output := string(data)
	testutil.VerifyContentContains(t, output, []string{"PREFIX", "Hello World", "SUFFIX"})
}

// TestIntegrationCancellation verifies that the application correctly cancels processing when the context times out.
func TestIntegrationCancellation(t *testing.T) {
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

	// Create a context with a very short timeout to force cancellation.
	ctx, cancel := context.WithTimeout(
		t.Context(),
		1*time.Millisecond,
	)
	defer cancel()

	// Run the application; we expect an error due to cancellation.
	runErr := run(ctx)
	if runErr == nil {
		t.Error("Expected Run to fail due to cancellation, but it succeeded")
	}
}
