package main

import (
	"context"
	"errors"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ivuorinen/gibidify/cli"
	"github.com/ivuorinen/gibidify/shared"
	"github.com/ivuorinen/gibidify/testutil"
)

// withIsolatedFlags sets up isolated flag state for testing and returns a cleanup function.
// This helper saves the original os.Args and flag.CommandLine, resets CLI flags,
// and creates a fresh FlagSet to avoid conflicts between tests.
func withIsolatedFlags(t *testing.T) func() {
	t.Helper()

	oldArgs := os.Args
	oldFlag := flag.CommandLine

	cli.ResetFlags()
	flag.CommandLine = flag.NewFlagSet("test", flag.ContinueOnError)

	return func() {
		os.Args = oldArgs
		flag.CommandLine = oldFlag
	}
}

// TestRun_FlagParsingErrors tests error handling in flag parsing.
func TestRunFlagParsingErrors(t *testing.T) {
	// Test with isolated flag state to avoid conflicts with other tests
	t.Run("invalid_flag", func(t *testing.T) {
		cleanup := withIsolatedFlags(t)
		defer cleanup()

		os.Args = []string{"test", shared.TestCLIFlagNoUI, "value"}

		err := run(context.Background())
		if err == nil {
			t.Fatal("Expected error from invalid flag")
		}

		if !strings.Contains(err.Error(), shared.TestOpParsingFlags) {
			t.Errorf("Expected 'parsing flags' error, got: %v", err)
		}
	})

	t.Run("invalid_format", func(t *testing.T) {
		cleanup := withIsolatedFlags(t)
		defer cleanup()

		// Create temporary files for the test
		srcDir := t.TempDir()
		testutil.CreateTestFile(t, srcDir, shared.TestFileTXT, []byte("test"))

		outFile, outPath := testutil.CreateTempOutputFile(t, shared.TestMD)
		testutil.CloseFile(t, outFile)
		defer func() {
			if err := os.Remove(outPath); err != nil {
				t.Logf(shared.TestMsgFailedToRemoveTempFile, err)
			}
		}()

		os.Args = []string{
			"test", shared.TestCLIFlagSource, srcDir, shared.TestCLIFlagDestination, outPath,
			shared.TestCLIFlagFormat, "invalid", shared.TestCLIFlagNoUI,
		}

		err := run(context.Background())
		if err == nil {
			t.Fatal("Expected error from invalid format")
		}

		if !strings.Contains(err.Error(), shared.TestOpParsingFlags) {
			t.Errorf("Expected 'parsing flags' error, got: %v", err)
		}
	})
}

// TestRun_ProcessingErrors tests processing-related error paths.
func TestRunProcessingErrors(t *testing.T) {
	t.Run("nonexistent_source", func(t *testing.T) {
		cleanup := withIsolatedFlags(t)
		defer cleanup()

		outFile, outPath := testutil.CreateTempOutputFile(t, shared.TestMD)
		testutil.CloseFile(t, outFile)
		defer func() {
			if err := os.Remove(outPath); err != nil {
				t.Logf(shared.TestMsgFailedToRemoveTempFile, err)
			}
		}()

		// Use a path that doesn't exist (subpath under temp dir that was never created)
		nonExistentDir := filepath.Join(t.TempDir(), "nonexistent", "path")
		os.Args = []string{
			"test", shared.TestCLIFlagSource, nonExistentDir,
			shared.TestCLIFlagDestination, outPath, shared.TestCLIFlagNoUI,
		}

		err := run(context.Background())
		if err == nil {
			t.Fatal("Expected error from nonexistent source")
		}

		// Could be either parsing flags (validation) or processing error
		if !strings.Contains(err.Error(), shared.TestOpParsingFlags) && !strings.Contains(err.Error(), "processing") {
			t.Errorf("Expected error from parsing or processing, got: %v", err)
		}
	})

	t.Run("missing_source", func(t *testing.T) {
		cleanup := withIsolatedFlags(t)
		defer cleanup()

		outFile, outPath := testutil.CreateTempOutputFile(t, shared.TestMD)
		testutil.CloseFile(t, outFile)
		defer func() {
			if err := os.Remove(outPath); err != nil {
				t.Logf(shared.TestMsgFailedToRemoveTempFile, err)
			}
		}()

		os.Args = []string{"test", shared.TestCLIFlagDestination, outPath, shared.TestCLIFlagNoUI}

		err := run(context.Background())
		if err == nil {
			t.Fatal("Expected error from missing source")
		}

		// Should be a user error
		if !cli.IsUserError(err) {
			t.Errorf("Expected user error, got: %v", err)
		}
	})
}

// TestRun_MarkdownExecution tests successful markdown execution.
func TestRunMarkdownExecution(t *testing.T) {
	// Suppress all output for cleaner test output
	restore := testutil.SuppressAllOutput(t)
	defer restore()

	cleanup := withIsolatedFlags(t)
	defer cleanup()

	// Create test environment
	srcDir := t.TempDir()
	testutil.CreateTestFiles(t, srcDir, []testutil.FileSpec{
		{Name: shared.TestFileMainGo, Content: shared.LiteralPackageMain + "\nfunc main() {}"},
		{Name: shared.TestFileHelperGo, Content: shared.LiteralPackageMain + "\nfunc help() {}"},
	})

	// Use non-existent output path to verify run() actually creates it
	outPath := filepath.Join(t.TempDir(), "output.md")
	defer func() {
		if err := os.Remove(outPath); err != nil && !os.IsNotExist(err) {
			t.Logf(shared.TestMsgFailedToRemoveTempFile, err)
		}
	}()

	os.Args = []string{
		"test", shared.TestCLIFlagSource, srcDir, shared.TestCLIFlagDestination, outPath,
		shared.TestCLIFlagFormat, "markdown", shared.TestCLIFlagNoUI,
	}

	err := run(context.Background())
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify output file was created with content
	info, err := os.Stat(outPath)
	if os.IsNotExist(err) {
		t.Fatal("Output file was not created")
	}
	if err != nil {
		t.Fatalf("Failed to stat output file: %v", err)
	}
	if info.Size() == 0 {
		t.Error("Output file is empty, expected content")
	}
}

// TestRun_JSONExecution tests successful JSON execution.
func TestRunJSONExecution(t *testing.T) {
	// Suppress all output for cleaner test output
	restore := testutil.SuppressAllOutput(t)
	defer restore()

	cleanup := withIsolatedFlags(t)
	defer cleanup()

	// Create test environment with unique directories
	srcDir := t.TempDir()
	testutil.CreateTestFile(t, srcDir, shared.TestFileMainGo, []byte(shared.LiteralPackageMain))

	// Use non-existent output path to verify run() actually creates it
	outPath := filepath.Join(t.TempDir(), "output.json")
	defer func() {
		if err := os.Remove(outPath); err != nil && !os.IsNotExist(err) {
			t.Logf(shared.TestMsgFailedToRemoveTempFile, err)
		}
	}()

	// Set CLI args with fresh paths
	os.Args = []string{
		"test", shared.TestCLIFlagSource, srcDir, shared.TestCLIFlagDestination, outPath,
		shared.TestCLIFlagFormat, "json", shared.TestCLIFlagNoUI,
	}

	err := run(context.Background())
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify output file was created with content
	info, err := os.Stat(outPath)
	if os.IsNotExist(err) {
		t.Fatal("Output file was not created")
	}
	if err != nil {
		t.Fatalf("Failed to stat output file: %v", err)
	}
	if info.Size() == 0 {
		t.Error("Output file is empty, expected content")
	}
}

// TestRun_ErrorWrapping tests that errors are properly wrapped.
func TestRunErrorWrapping(t *testing.T) {
	cleanup := withIsolatedFlags(t)
	defer cleanup()

	os.Args = []string{"test", "-invalid-flag"}

	err := run(context.Background())
	if err == nil {
		t.Fatal("Expected error")
	}

	// Should wrap with proper context
	if !strings.Contains(err.Error(), "parsing flags:") {
		t.Errorf("Error not properly wrapped, got: %v", err)
	}
}

// TestRun_HappyPathWithDefaultConfig tests successful execution with default configuration.
// This validates that run() completes successfully when given valid inputs,
// implicitly exercising the config loading path without directly verifying it.
func TestRunHappyPathWithDefaultConfig(t *testing.T) {
	// Suppress all output for cleaner test output
	restore := testutil.SuppressAllOutput(t)
	defer restore()

	cleanup := withIsolatedFlags(t)
	defer cleanup()

	// Create valid test setup
	srcDir := t.TempDir()
	testutil.CreateTestFile(t, srcDir, shared.TestFileGo, []byte(shared.LiteralPackageMain))

	outFile, outPath := testutil.CreateTempOutputFile(t, shared.TestMD)
	testutil.CloseFile(t, outFile)
	defer func() {
		if err := os.Remove(outPath); err != nil {
			t.Logf(shared.TestMsgFailedToRemoveTempFile, err)
		}
	}()

	os.Args = []string{
		"test", shared.TestCLIFlagSource, srcDir, shared.TestCLIFlagDestination, outPath, shared.TestCLIFlagNoUI,
	}

	err := run(context.Background())
	if err != nil {
		t.Fatalf("run() failed: %v", err)
	}
}

// TestErrorClassification tests user vs system error classification.
func TestErrorClassification(t *testing.T) {
	tests := []struct {
		name      string
		err       error
		isUserErr bool
	}{
		{
			name:      "nil_error",
			err:       nil,
			isUserErr: false,
		},
		{
			name:      "cli_missing_source",
			err:       cli.NewCLIMissingSourceError(),
			isUserErr: true,
		},
		{
			name:      "flag_error",
			err:       errors.New("flag: invalid argument"),
			isUserErr: true,
		},
		{
			name:      "permission_denied",
			err:       errors.New("permission denied"),
			isUserErr: true,
		},
		{
			name:      "file_not_found",
			err:       errors.New("file not found"),
			isUserErr: true,
		},
		{
			name:      "generic_system_error",
			err:       errors.New("internal system failure"),
			isUserErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isUser := cli.IsUserError(tt.err)
			if isUser != tt.isUserErr {
				t.Errorf("IsUserError(%v) = %v, want %v", tt.err, isUser, tt.isUserErr)
			}
		})
	}
}

// TestRun_ContextCancellation tests context cancellation handling.
func TestRunContextCancellation(t *testing.T) {
	// Suppress all output for cleaner test output
	restore := testutil.SuppressAllOutput(t)
	defer restore()

	cleanup := withIsolatedFlags(t)
	defer cleanup()

	// Create test environment
	srcDir := t.TempDir()
	testutil.CreateTestFile(t, srcDir, shared.TestFileGo, []byte(shared.LiteralPackageMain))

	outFile, outPath := testutil.CreateTempOutputFile(t, shared.TestMD)
	testutil.CloseFile(t, outFile)
	defer func() {
		if err := os.Remove(outPath); err != nil {
			t.Logf(shared.TestMsgFailedToRemoveTempFile, err)
		}
	}()

	os.Args = []string{
		"test", shared.TestCLIFlagSource, srcDir, shared.TestCLIFlagDestination, outPath, shared.TestCLIFlagNoUI,
	}

	// Create pre-canceled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := run(ctx)
	// Assert that canceled context causes an error
	if err == nil {
		t.Error("Expected error with canceled context, got nil")
	} else if !errors.Is(err, context.Canceled) && !strings.Contains(err.Error(), "context canceled") {
		t.Errorf("Expected context.Canceled error, got: %v", err)
	}
}
