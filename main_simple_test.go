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
	"github.com/ivuorinen/gibidify/testutil"
)

// TestRun_FlagParsingErrors tests error handling in flag parsing.
func TestRun_FlagParsingErrors(t *testing.T) {
	// Test in complete isolation with separate process to avoid flag conflicts
	t.Run("invalid_flag", func(t *testing.T) {
		// Save original state
		oldArgs := os.Args
		oldFlag := flag.CommandLine
		defer func() {
			os.Args = oldArgs
			flag.CommandLine = oldFlag
		}()

		// Create fresh flag set
		flag.CommandLine = flag.NewFlagSet("test", flag.ContinueOnError)
		os.Args = []string{"test", "-nonexistent-flag", "value"}

		err := run(context.Background())
		if err == nil {
			t.Fatal("Expected error from invalid flag")
		}

		if !strings.Contains(err.Error(), "parsing flags") {
			t.Errorf("Expected 'parsing flags' error, got: %v", err)
		}
	})

	t.Run("invalid_format", func(t *testing.T) {
		// Save original state
		oldArgs := os.Args
		oldFlag := flag.CommandLine
		defer func() {
			os.Args = oldArgs
			flag.CommandLine = oldFlag
		}()

		// Create fresh flag set
		flag.CommandLine = flag.NewFlagSet("test", flag.ContinueOnError)

		// Create temporary files for the test
		srcDir := t.TempDir()
		testutil.CreateTestFile(t, srcDir, "test.txt", []byte("test"))

		outFile, outPath := testutil.CreateTempOutputFile(t, "test.md")
		testutil.CloseFile(t, outFile)
		defer func() {
			if err := os.Remove(outPath); err != nil {
				t.Logf("Failed to remove temp file: %v", err)
			}
		}()

		os.Args = []string{"test", "-source", srcDir, "-destination", outPath, "-format", "invalid"}

		err := run(context.Background())
		if err == nil {
			t.Fatal("Expected error from invalid format")
		}

		if !strings.Contains(err.Error(), "parsing flags") {
			t.Errorf("Expected 'parsing flags' error, got: %v", err)
		}
	})
}

// TestRun_ProcessingErrors tests processing-related error paths.
func TestRun_ProcessingErrors(t *testing.T) {
	t.Run("nonexistent_source", func(t *testing.T) {
		// Save original state
		oldArgs := os.Args
		oldFlag := flag.CommandLine
		defer func() {
			os.Args = oldArgs
			flag.CommandLine = oldFlag
		}()

		// Create fresh flag set
		flag.CommandLine = flag.NewFlagSet("test", flag.ContinueOnError)

		outFile, outPath := testutil.CreateTempOutputFile(t, "test.md")
		testutil.CloseFile(t, outFile)
		defer func() {
			if err := os.Remove(outPath); err != nil {
				t.Logf("Failed to remove temp file: %v", err)
			}
		}()

		// Create a temporary directory and immediately remove it to ensure it doesn't exist
		nonExistentDir := filepath.Join(t.TempDir(), "nonexistent", "path")
		os.Args = []string{"test", "-source", nonExistentDir, "-destination", outPath}

		err := run(context.Background())
		if err == nil {
			t.Fatal("Expected error from nonexistent source")
		}

		// Could be either parsing flags (validation) or processing error
		if !strings.Contains(err.Error(), "parsing flags") && !strings.Contains(err.Error(), "processing") {
			t.Errorf("Expected error from parsing or processing, got: %v", err)
		}
	})

	t.Run("missing_source", func(t *testing.T) {
		// Save original state
		oldArgs := os.Args
		oldFlag := flag.CommandLine
		defer func() {
			os.Args = oldArgs
			flag.CommandLine = oldFlag
		}()

		// Create fresh flag set
		flag.CommandLine = flag.NewFlagSet("test", flag.ContinueOnError)

		outFile, outPath := testutil.CreateTempOutputFile(t, "test.md")
		testutil.CloseFile(t, outFile)
		defer func() {
			if err := os.Remove(outPath); err != nil {
				t.Logf("Failed to remove temp file: %v", err)
			}
		}()

		os.Args = []string{"test", "-destination", outPath}

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
func TestRun_MarkdownExecution(t *testing.T) {
	// Suppress logs for cleaner test output
	restoreLogs := testutil.SuppressLogs(t)
	defer restoreLogs()

	// Save original state
	oldArgs := os.Args
	oldFlag := flag.CommandLine
	defer func() {
		os.Args = oldArgs
		flag.CommandLine = oldFlag
	}()

	// Reset global flag state and create fresh flag set
	cli.ResetFlags()
	flag.CommandLine = flag.NewFlagSet("test", flag.ContinueOnError)

	// Create test environment
	srcDir := t.TempDir()
	testutil.CreateTestFiles(t, srcDir, []testutil.FileSpec{
		{Name: "main.go", Content: "package main\nfunc main() {}"},
		{Name: "helper.go", Content: "package main\nfunc help() {}"},
	})

	outFile, outPath := testutil.CreateTempOutputFile(t, "test.md")
	testutil.CloseFile(t, outFile)
	defer func() {
		if err := os.Remove(outPath); err != nil {
			t.Logf("Failed to remove temp file: %v", err)
		}
	}()

	os.Args = []string{"test", "-source", srcDir, "-destination", outPath, "-format", "markdown"}

	err := run(context.Background())
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify output file was created
	if _, err := os.Stat(outPath); os.IsNotExist(err) {
		t.Error("Output file was not created")
	}
}

// TestRun_JSONExecution tests successful JSON execution.
func TestRun_JSONExecution(t *testing.T) {
	// Suppress logs for cleaner test output
	restoreLogs := testutil.SuppressLogs(t)
	defer restoreLogs()

	// Save original state
	oldArgs := os.Args
	oldFlag := flag.CommandLine
	defer func() {
		os.Args = oldArgs
		flag.CommandLine = oldFlag
	}()

	// Reset global flag state and create fresh flag set for complete isolation
	cli.ResetFlags()
	flag.CommandLine = flag.NewFlagSet("test", flag.ContinueOnError)

	// Create test environment with unique directories
	srcDir := t.TempDir()
	testutil.CreateTestFile(t, srcDir, "main.go", []byte("package main"))

	outFile, outPath := testutil.CreateTempOutputFile(t, "test.json")
	testutil.CloseFile(t, outFile)
	defer func() {
		if err := os.Remove(outPath); err != nil {
			t.Logf("Failed to remove temp file: %v", err)
		}
	}()

	// Set CLI args with fresh paths
	os.Args = []string{"test", "-source", srcDir, "-destination", outPath, "-format", "json"}

	err := run(context.Background())
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify output file was created
	if _, err := os.Stat(outPath); os.IsNotExist(err) {
		t.Error("Output file was not created")
	}
}

// TestRun_ErrorWrapping tests that errors are properly wrapped.
func TestRun_ErrorWrapping(t *testing.T) {
	// Save original state
	oldArgs := os.Args
	oldFlag := flag.CommandLine
	defer func() {
		os.Args = oldArgs
		flag.CommandLine = oldFlag
	}()

	// Reset global flag state and test flag parsing error wrapping
	cli.ResetFlags()
	flag.CommandLine = flag.NewFlagSet("test", flag.ContinueOnError)
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

// TestRun_ConfigurationCalls tests that config loading is called.
func TestRun_ConfigurationCalls(t *testing.T) {
	// Suppress logs for cleaner test output
	restoreLogs := testutil.SuppressLogs(t)
	defer restoreLogs()

	// Save original state
	oldArgs := os.Args
	oldFlag := flag.CommandLine
	defer func() {
		os.Args = oldArgs
		flag.CommandLine = oldFlag
	}()

	// Reset global flag state and create fresh flag set
	cli.ResetFlags()
	flag.CommandLine = flag.NewFlagSet("test", flag.ContinueOnError)

	// Create valid test setup
	srcDir := t.TempDir()
	testutil.CreateTestFile(t, srcDir, "test.go", []byte("package main"))

	outFile, outPath := testutil.CreateTempOutputFile(t, "test.md")
	testutil.CloseFile(t, outFile)
	defer func() {
		if err := os.Remove(outPath); err != nil {
			t.Logf("Failed to remove temp file: %v", err)
		}
	}()

	os.Args = []string{"test", "-source", srcDir, "-destination", outPath}

	err := run(context.Background())
	if err != nil {
		t.Fatalf("run() failed: %v", err)
	}

	// If we get here, config.LoadConfig() was called without error
	t.Log("Configuration loading completed successfully")
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
			err:       cli.NewMissingSourceError(),
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
func TestRun_ContextCancellation(t *testing.T) {
	// Suppress logs for cleaner test output
	restoreLogs := testutil.SuppressLogs(t)
	defer restoreLogs()

	// Save original state
	oldArgs := os.Args
	oldFlag := flag.CommandLine
	defer func() {
		os.Args = oldArgs
		flag.CommandLine = oldFlag
	}()

	// Reset global flag state and create fresh flag set
	cli.ResetFlags()
	flag.CommandLine = flag.NewFlagSet("test", flag.ContinueOnError)

	// Create test environment
	srcDir := t.TempDir()
	testutil.CreateTestFile(t, srcDir, "test.go", []byte("package main"))

	outFile, outPath := testutil.CreateTempOutputFile(t, "test.md")
	testutil.CloseFile(t, outFile)
	defer func() {
		if err := os.Remove(outPath); err != nil {
			t.Logf("Failed to remove temp file: %v", err)
		}
	}()

	os.Args = []string{"test", "-source", srcDir, "-destination", outPath}

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
