package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/ivuorinen/gibidify/cli"
	"github.com/ivuorinen/gibidify/config"
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
				// Set args with missing source
				os.Args = []string{
					"gibidify", testutil.TestCLIFlagDestination, testutil.TestOutputMD, testutil.TestCLIFlagNoUI,
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
				srcDir := t.TempDir()
				outFile, outPath := testutil.CreateTempOutputFile(t, testutil.TestOutputMD)
				testutil.CloseFile(t, outFile)
				// Set args with invalid format
				os.Args = []string{
					"gibidify", testutil.TestCLIFlagSource, srcDir, testutil.TestCLIFlagDestination, outPath,
					testutil.TestCLIFlagFormat, "invalid", testutil.TestCLIFlagNoUI,
				}
			},
			expectError: true,
			errorSubstr: testutil.TestOpParsingFlags,
		},
		{
			name: "Invalid source directory",
			setup: func(t *testing.T) {
				// Reset flags and set args with non-existent source
				cli.ResetFlags()
				outFile, outPath := testutil.CreateTempOutputFile(t, testutil.TestOutputMD)
				testutil.CloseFile(t, outFile)
				os.Args = []string{
					"gibidify", testutil.TestCLIFlagSource, "/nonexistent/directory",
					testutil.TestCLIFlagDestination, outPath, testutil.TestCLIFlagNoUI,
				}
			},
			expectError: true,
			errorSubstr: testutil.TestOpParsingFlags,
		},
		{
			name: "Valid run with minimal setup",
			setup: func(t *testing.T) {
				// Reset flags
				cli.ResetFlags()

				// Create valid setup
				srcDir := testutil.SetupTempDirWithStructure(t, []testutil.DirSpec{
					{
						Path: "",
						Files: []testutil.FileSpec{
							{Name: testutil.TestFileTXT, Content: testutil.TestContent},
						},
					},
				})

				outFile, outPath := testutil.CreateTempOutputFile(t, testutil.TestOutputMD)
				testutil.CloseFile(t, outFile)

				os.Args = []string{
					"gibidify", testutil.TestCLIFlagSource, srcDir,
					testutil.TestCLIFlagDestination, outPath, testutil.TestCLIFlagNoUI,
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

		// Create minimal valid setup
		srcDir := testutil.SetupTempDirWithStructure(t, []testutil.DirSpec{
			{
				Path: "",
				Files: []testutil.FileSpec{
					{Name: testutil.TestFileTXT, Content: testutil.TestContent},
				},
			},
		})

		outFile, outPath := testutil.CreateTempOutputFile(t, "test_output.json")
		testutil.CloseFile(t, outFile)

		// Set minimal required args
		os.Args = []string{
			"gibidify", testutil.TestCLIFlagSource, srcDir,
			testutil.TestCLIFlagDestination, outPath, testutil.TestCLIFlagNoUI,
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

	// Create valid setup
	srcDir := testutil.SetupTempDirWithStructure(t, []testutil.DirSpec{
		{
			Path: "",
			Files: []testutil.FileSpec{
				{Name: testutil.TestFileGo, Content: testutil.LiteralPackageMain + "\nfunc main() {}"},
			},
		},
	})

	outFile, outPath := testutil.CreateTempOutputFile(t, testutil.TestOutputMD)
	testutil.CloseFile(t, outFile)

	os.Args = []string{
		"gibidify", testutil.TestCLIFlagSource, srcDir,
		testutil.TestCLIFlagDestination, outPath, testutil.TestCLIFlagNoUI,
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

// TestRunHelpIsNotAnError verifies -h prints usage and returns nil (exit 0).
func TestRunHelpIsNotAnError(t *testing.T) {
	restore := testutil.SuppressAllOutput(t)
	defer restore()

	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	cli.ResetFlags()
	os.Args = []string{"gibidify", "-h"}

	testutil.AssertNoError(t, run(context.Background()), "run() with -h")
}

// TestRunEnforcesConfiguredMaxConcurrency verifies the config file's
// maxConcurrency caps -concurrency. Flags are parsed before the config loads,
// so run() must re-check after loading.
func TestRunEnforcesConfiguredMaxConcurrency(t *testing.T) {
	restore := testutil.SuppressAllOutput(t)
	defer restore()

	xdg := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", xdg)
	// The config store is global; do not leak maxConcurrency into later tests.
	t.Cleanup(config.Reset)
	cfgDir := testutil.CreateTestDirectory(t, xdg, shared.AppName)
	testutil.CreateTestFile(t, cfgDir, "config.yaml", []byte("maxConcurrency: 2\n"))

	srcDir := t.TempDir()
	testutil.CreateTestFile(t, srcDir, testutil.TestFileGo, []byte(testutil.LiteralPackageMain))

	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	cli.ResetFlags()
	os.Args = []string{
		"gibidify", testutil.TestCLIFlagSource, srcDir,
		testutil.TestCLIFlagDestination, filepath.Join(t.TempDir(), "out.json"),
		testutil.TestCLIFlagConcurrency, "3", testutil.TestCLIFlagNoUI,
	}

	err := run(context.Background())
	testutil.AssertExpectedError(t, err, "run() above maxConcurrency")
	testutil.AssertErrorContains(t, err, "exceeds maximum", "run() maxConcurrency error")
}

// TestRunFailsOnInvalidConfig verifies an invalid config file aborts the run as
// a user error before any output is written. It used to fall back to defaults,
// which dropped the user's ignoreDirectories and included what they excluded.
func TestRunFailsOnInvalidConfig(t *testing.T) {
	restore := testutil.SuppressAllOutput(t)
	defer restore()

	xdg := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", xdg)
	t.Cleanup(config.Reset)
	cfgDir := testutil.CreateTestDirectory(t, xdg, shared.AppName)
	testutil.CreateTestFile(t, cfgDir, "config.yaml", []byte("ignoreDirectories: [secrets\n"))

	srcDir := t.TempDir()
	testutil.CreateTestFile(t, srcDir, testutil.TestFileGo, []byte(testutil.LiteralPackageMain))
	outPath := filepath.Join(t.TempDir(), "out.json")

	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	cli.ResetFlags()
	os.Args = []string{
		"gibidify", testutil.TestCLIFlagSource, srcDir,
		testutil.TestCLIFlagDestination, outPath, testutil.TestCLIFlagNoUI,
	}

	err := run(context.Background())
	testutil.AssertErrorContains(t, err, "parsing config file", "run() with a malformed config")
	if !cli.IsUserError(err) {
		t.Errorf("config error must be a user error (exit 1), got %v", err)
	}
	if _, statErr := os.Stat(outPath); !os.IsNotExist(statErr) {
		t.Errorf("no output may be written when the config is invalid (stat err: %v)", statErr)
	}
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

			// Create valid setup
			srcDir := testutil.SetupTempDirWithStructure(t, []testutil.DirSpec{
				{
					Path: "",
					Files: []testutil.FileSpec{
						{Name: testutil.TestFileTXT, Content: testutil.TestContent},
					},
				},
			})

			outFile, outPath := testutil.CreateTempOutputFile(t, testutil.TestOutputMD)
			testutil.CloseFile(t, outFile)

			// Set args with log level
			os.Args = []string{
				"gibidify", testutil.TestCLIFlagSource, srcDir, testutil.TestCLIFlagDestination, outPath,
				"-log-level", tt.logLevel, testutil.TestCLIFlagNoUI,
			}

			// Run
			ctx := context.Background()
			err := run(ctx)

			// Should succeed
			testutil.AssertNoError(t, err, "run() with log level "+tt.logLevel)
		})
	}
}
