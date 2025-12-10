package main

import (
	"context"
	"flag"
	"os"
	"testing"

	"github.com/ivuorinen/gibidify/cli"
	"github.com/ivuorinen/gibidify/shared"
	"github.com/ivuorinen/gibidify/testutil"
)

// TestRunErrorPaths tests various error paths in the run() function.
func TestRunErrorPaths(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(t *testing.T)
		expectError bool
		errorSubstr string
	}{
		{
			name: "Invalid flags - missing source",
			setup: func(_ *testing.T) {
				// Reset flags and set invalid args
				cli.ResetFlags()
				flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
				// Set args with missing source
				os.Args = []string{
					"gibidify", shared.TestCLIFlagDestination, shared.TestOutputMD, shared.TestCLIFlagNoUI,
				}
			},
			expectError: true,
			errorSubstr: "parsing flags",
		},
		{
			name: "Invalid flags - invalid format",
			setup: func(t *testing.T) {
				// Reset flags and set invalid args
				cli.ResetFlags()
				flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
				srcDir := t.TempDir()
				outFile, outPath := testutil.CreateTempOutputFile(t, shared.TestOutputMD)
				testutil.CloseFile(t, outFile)
				// Set args with invalid format
				os.Args = []string{
					"gibidify", shared.TestCLIFlagSource, srcDir, shared.TestCLIFlagDestination, outPath,
					shared.TestCLIFlagFormat, "invalid", shared.TestCLIFlagNoUI,
				}
			},
			expectError: true,
			errorSubstr: shared.TestOpParsingFlags,
		},
		{
			name: "Invalid source directory",
			setup: func(t *testing.T) {
				// Reset flags and set args with non-existent source
				cli.ResetFlags()
				flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
				outFile, outPath := testutil.CreateTempOutputFile(t, shared.TestOutputMD)
				testutil.CloseFile(t, outFile)
				os.Args = []string{
					"gibidify", shared.TestCLIFlagSource, "/nonexistent/directory",
					shared.TestCLIFlagDestination, outPath, shared.TestCLIFlagNoUI,
				}
			},
			expectError: true,
			errorSubstr: shared.TestOpParsingFlags, // Flag validation catches this, not processing
		},
		{
			name: "Valid run with minimal setup",
			setup: func(t *testing.T) {
				// Reset flags
				cli.ResetFlags()
				flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

				// Create valid setup
				srcDir := testutil.SetupTempDirWithStructure(t, []testutil.DirSpec{
					{
						Path: "",
						Files: []testutil.FileSpec{
							{Name: shared.TestFileTXT, Content: shared.TestContent},
						},
					},
				})

				outFile, outPath := testutil.CreateTempOutputFile(t, shared.TestOutputMD)
				testutil.CloseFile(t, outFile)

				os.Args = []string{
					"gibidify", shared.TestCLIFlagSource, srcDir,
					shared.TestCLIFlagDestination, outPath, shared.TestCLIFlagNoUI,
				}
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Suppress all output for cleaner test output
			restore := testutil.SuppressAllOutput(t)
			defer restore()

			// Setup test case
			tt.setup(t)

			// Run the function
			ctx := context.Background()
			err := run(ctx)

			// Check expectations
			if tt.expectError {
				testutil.AssertExpectedError(t, err, "run() with error case")
				if tt.errorSubstr != "" {
					testutil.AssertErrorContains(t, err, tt.errorSubstr, "run() error content")
				}
			} else {
				testutil.AssertNoError(t, err, "run() success case")
			}
		})
	}
}

// TestRunFlagParsing tests the flag parsing path in run() function.
func TestRunFlagParsing(t *testing.T) {
	// Suppress logs for cleaner test output
	restoreLogs := testutil.SuppressLogs(t)
	defer restoreLogs()

	// Save original args
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Test with empty args (should use defaults)
	t.Run("default args", func(t *testing.T) {
		// Reset flags
		cli.ResetFlags()
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

		// Create minimal valid setup
		srcDir := testutil.SetupTempDirWithStructure(t, []testutil.DirSpec{
			{
				Path: "",
				Files: []testutil.FileSpec{
					{Name: shared.TestFileTXT, Content: shared.TestContent},
				},
			},
		})

		outFile, outPath := testutil.CreateTempOutputFile(t, "test_output.json")
		testutil.CloseFile(t, outFile)

		// Set minimal required args
		os.Args = []string{
			"gibidify", shared.TestCLIFlagSource, srcDir,
			shared.TestCLIFlagDestination, outPath, shared.TestCLIFlagNoUI,
		}

		// Run and verify it works with defaults
		ctx := context.Background()
		err := run(ctx)
		testutil.AssertNoError(t, err, "run() with default flags")
	})
}

// TestRunWithCanceledContext tests run() with pre-canceled context.
func TestRunWithCanceledContext(t *testing.T) {
	// Suppress logs for cleaner test output
	restoreLogs := testutil.SuppressLogs(t)
	defer restoreLogs()

	// Save original args
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Reset flags
	cli.ResetFlags()
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	// Create valid setup
	srcDir := testutil.SetupTempDirWithStructure(t, []testutil.DirSpec{
		{
			Path: "",
			Files: []testutil.FileSpec{
				{Name: shared.TestFileGo, Content: shared.LiteralPackageMain + "\nfunc main() {}"},
			},
		},
	})

	outFile, outPath := testutil.CreateTempOutputFile(t, shared.TestOutputMD)
	testutil.CloseFile(t, outFile)

	os.Args = []string{
		"gibidify", shared.TestCLIFlagSource, srcDir,
		shared.TestCLIFlagDestination, outPath, shared.TestCLIFlagNoUI,
	}

	// Create canceled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Run with canceled context
	err := run(ctx)

	// Should get processing error due to canceled context
	testutil.AssertExpectedError(t, err, "run() with canceled context")
	testutil.AssertErrorContains(t, err, "processing", "run() canceled context error")
}

// TestRunLogLevel tests the log level setting in run().
func TestRunLogLevel(t *testing.T) {
	// Suppress logs for cleaner test output
	restoreLogs := testutil.SuppressLogs(t)
	defer restoreLogs()

	// Save original args
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	tests := []struct {
		name     string
		logLevel string
	}{
		{"debug level", "debug"},
		{"info level", "info"},
		{"warn level", "warn"},
		{"error level", "error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset flags
			cli.ResetFlags()
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

			// Create valid setup
			srcDir := testutil.SetupTempDirWithStructure(t, []testutil.DirSpec{
				{
					Path: "",
					Files: []testutil.FileSpec{
						{Name: shared.TestFileTXT, Content: shared.TestContent},
					},
				},
			})

			outFile, outPath := testutil.CreateTempOutputFile(t, shared.TestOutputMD)
			testutil.CloseFile(t, outFile)

			// Set args with log level
			os.Args = []string{
				"gibidify", shared.TestCLIFlagSource, srcDir, shared.TestCLIFlagDestination, outPath,
				"-log-level", tt.logLevel, shared.TestCLIFlagNoUI,
			}

			// Run
			ctx := context.Background()
			err := run(ctx)

			// Should succeed
			testutil.AssertNoError(t, err, "run() with log level "+tt.logLevel)
		})
	}
}
