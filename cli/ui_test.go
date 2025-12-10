package cli

import (
	"os"
	"strings"
	"testing"

	"github.com/ivuorinen/gibidify/shared"
)

func TestNewUIManager(t *testing.T) {
	ui := NewUIManager()

	if ui == nil {
		t.Error("NewUIManager() returned nil")

		return
	}
	if ui.output == nil {
		t.Error("NewUIManager() did not set output")

		return
	}
	if ui.output != os.Stderr {
		t.Error("NewUIManager() should default output to os.Stderr")
	}
}

func TestUIManagerSetColorOutput(t *testing.T) {
	ui := NewUIManager()

	// Test enabling colors
	ui.SetColorOutput(true)
	if !ui.enableColors {
		t.Error("SetColorOutput(true) did not enable colors")
	}

	// Test disabling colors
	ui.SetColorOutput(false)
	if ui.enableColors {
		t.Error("SetColorOutput(false) did not disable colors")
	}
}

func TestUIManagerSetProgressOutput(t *testing.T) {
	ui := NewUIManager()

	// Test enabling progress
	ui.SetProgressOutput(true)
	if !ui.enableProgress {
		t.Error("SetProgressOutput(true) did not enable progress")
	}

	// Test disabling progress
	ui.SetProgressOutput(false)
	if ui.enableProgress {
		t.Error("SetProgressOutput(false) did not disable progress")
	}
}

func TestUIManagerStartProgress(t *testing.T) {
	tests := []struct {
		name        string
		total       int
		description string
		enabled     bool
		expectBar   bool
	}{
		{
			name:        "valid progress with enabled progress",
			total:       10,
			description: shared.TestProgressMessage,
			enabled:     true,
			expectBar:   true,
		},
		{
			name:        "disabled progress should not create bar",
			total:       10,
			description: shared.TestProgressMessage,
			enabled:     false,
			expectBar:   false,
		},
		{
			name:        "zero total should not create bar",
			total:       0,
			description: shared.TestProgressMessage,
			enabled:     true,
			expectBar:   false,
		},
		{
			name:        "negative total should not create bar",
			total:       -1,
			description: shared.TestProgressMessage,
			enabled:     true,
			expectBar:   false,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				ui, _ := createTestUI() //nolint:errcheck // Test helper output buffer not used in this test
				ui.SetProgressOutput(tt.enabled)

				ui.StartProgress(tt.total, tt.description)

				if tt.expectBar && ui.progressBar == nil {
					t.Error("StartProgress() should have created progress bar but didn't")
				}
				if !tt.expectBar && ui.progressBar != nil {
					t.Error("StartProgress() should not have created progress bar but did")
				}
			},
		)
	}
}

func TestUIManagerUpdateProgress(t *testing.T) {
	ui, _ := createTestUI() //nolint:errcheck // Test helper output buffer not used in this test
	ui.SetProgressOutput(true)

	// Test with no progress bar (should not panic)
	ui.UpdateProgress(1)

	// Test with progress bar
	ui.StartProgress(10, "Test progress")
	if ui.progressBar == nil {
		t.Fatal("StartProgress() did not create progress bar")
	}

	// Should not panic
	ui.UpdateProgress(1)
	ui.UpdateProgress(5)
}

func TestUIManagerFinishProgress(t *testing.T) {
	ui, _ := createTestUI() //nolint:errcheck // Test helper output buffer not used in this test
	ui.SetProgressOutput(true)

	// Test with no progress bar (should not panic)
	ui.FinishProgress()

	// Test with progress bar
	ui.StartProgress(10, "Test progress")
	if ui.progressBar == nil {
		t.Fatal("StartProgress() did not create progress bar")
	}

	ui.FinishProgress()
	if ui.progressBar != nil {
		t.Error("FinishProgress() should have cleared progress bar")
	}
}

// testPrintMethod is a helper function to test UI print methods without duplication.
type printMethodTest struct {
	name         string
	enableColors bool
	format       string
	args         []any
	expectedText string
}

func testPrintMethod(
	t *testing.T,
	methodName string,
	printFunc func(*UIManager, string, ...any),
	tests []printMethodTest,
) {
	t.Helper()

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				ui, output := createTestUI()
				ui.SetColorOutput(tt.enableColors)

				printFunc(ui, tt.format, tt.args...)

				if !tt.enableColors {
					outputStr := output.String()
					if !strings.Contains(outputStr, tt.expectedText) {
						t.Errorf("%s() output %q should contain %q", methodName, outputStr, tt.expectedText)
					}
				}
			},
		)
	}

	// Test color method separately (doesn't capture output but shouldn't panic)
	t.Run(
		methodName+" with colors should not panic", func(_ *testing.T) {
			ui, _ := createTestUI() //nolint:errcheck // Test helper output buffer not used in this test
			ui.SetColorOutput(true)
			// Should not panic
			printFunc(ui, "Test message")
		},
	)
}

func TestUIManagerPrintSuccess(t *testing.T) {
	tests := []printMethodTest{
		{
			name:         "success without colors",
			enableColors: false,
			format:       "Operation completed successfully",
			args:         []any{},
			expectedText: "✓ Operation completed successfully",
		},
		{
			name:         "success with args without colors",
			enableColors: false,
			format:       "Processed %d files in %s",
			args:         []any{5, "project"},
			expectedText: "✓ Processed 5 files in project",
		},
	}

	testPrintMethod(
		t, "PrintSuccess", func(ui *UIManager, format string, args ...any) {
			ui.PrintSuccess(format, args...)
		}, tests,
	)
}

func TestUIManagerPrintError(t *testing.T) {
	tests := []printMethodTest{
		{
			name:         "error without colors",
			enableColors: false,
			format:       "Operation failed",
			args:         []any{},
			expectedText: "✗ Operation failed",
		},
		{
			name:         "error with args without colors",
			enableColors: false,
			format:       "Failed to process %d files",
			args:         []any{3},
			expectedText: "✗ Failed to process 3 files",
		},
	}

	testPrintMethod(
		t, "PrintError", func(ui *UIManager, format string, args ...any) {
			ui.PrintError(format, args...)
		}, tests,
	)
}

func TestUIManagerPrintWarning(t *testing.T) {
	tests := []printMethodTest{
		{
			name:         "warning without colors",
			enableColors: false,
			format:       "This is a warning",
			args:         []any{},
			expectedText: "⚠ This is a warning",
		},
		{
			name:         "warning with args without colors",
			enableColors: false,
			format:       "Found %d potential issues",
			args:         []any{2},
			expectedText: "⚠ Found 2 potential issues",
		},
	}

	testPrintMethod(
		t, "PrintWarning", func(ui *UIManager, format string, args ...any) {
			ui.PrintWarning(format, args...)
		}, tests,
	)
}

func TestUIManagerPrintInfo(t *testing.T) {
	tests := []printMethodTest{
		{
			name:         "info without colors",
			enableColors: false,
			format:       "Information message",
			args:         []any{},
			expectedText: "ℹ Information message",
		},
		{
			name:         "info with args without colors",
			enableColors: false,
			format:       "Processing file %s",
			args:         []any{"example.go"},
			expectedText: "ℹ Processing file example.go",
		},
	}

	testPrintMethod(
		t, "PrintInfo", func(ui *UIManager, format string, args ...any) {
			ui.PrintInfo(format, args...)
		}, tests,
	)
}

func TestUIManagerPrintHeader(t *testing.T) {
	tests := []struct {
		name         string
		enableColors bool
		format       string
		args         []any
		expectedText string
	}{
		{
			name:         "header without colors",
			enableColors: false,
			format:       "Main Header",
			args:         []any{},
			expectedText: "Main Header",
		},
		{
			name:         "header with args without colors",
			enableColors: false,
			format:       "Processing %s Module",
			args:         []any{"CLI"},
			expectedText: "Processing CLI Module",
		},
		{
			name:         "header with colors",
			enableColors: true,
			format:       "Build Results",
			args:         []any{},
			expectedText: "Build Results",
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				ui, output := createTestUI()
				ui.SetColorOutput(tt.enableColors)

				ui.PrintHeader(tt.format, tt.args...)

				outputStr := output.String()
				if !strings.Contains(outputStr, tt.expectedText) {
					t.Errorf("PrintHeader() output %q should contain %q", outputStr, tt.expectedText)
				}
			},
		)
	}
}

// colorTerminalTestCase represents a test case for color terminal detection.
type colorTerminalTestCase struct {
	name          string
	term          string
	ci            string
	githubActions string
	noColor       string
	forceColor    string
	expected      bool
}

// clearColorTerminalEnvVars clears all environment variables used for terminal color detection.
func clearColorTerminalEnvVars(t *testing.T) {
	t.Helper()
	envVars := []string{"TERM", "CI", "GITHUB_ACTIONS", "NO_COLOR", "FORCE_COLOR"}
	for _, envVar := range envVars {
		if err := os.Unsetenv(envVar); err != nil {
			t.Logf("Failed to unset %s: %v", envVar, err)
		}
	}
}

// setColorTerminalTestEnv sets up environment variables for a test case.
func setColorTerminalTestEnv(t *testing.T, testCase colorTerminalTestCase) {
	t.Helper()

	envSettings := map[string]string{
		"TERM":           testCase.term,
		"CI":             testCase.ci,
		"GITHUB_ACTIONS": testCase.githubActions,
		"NO_COLOR":       testCase.noColor,
		"FORCE_COLOR":    testCase.forceColor,
	}

	for key, value := range envSettings {
		if value != "" {
			t.Setenv(key, value)
		}
	}
}

func TestIsColorTerminal(t *testing.T) {
	// Save original environment
	originalEnv := map[string]string{
		"TERM":           os.Getenv("TERM"),
		"CI":             os.Getenv("CI"),
		"GITHUB_ACTIONS": os.Getenv("GITHUB_ACTIONS"),
		"NO_COLOR":       os.Getenv("NO_COLOR"),
		"FORCE_COLOR":    os.Getenv("FORCE_COLOR"),
	}

	defer func() {
		// Restore original environment
		for key, value := range originalEnv {
			setEnvOrUnset(key, value)
		}
	}()

	tests := []colorTerminalTestCase{
		{
			name:     "dumb terminal",
			term:     "dumb",
			expected: false,
		},
		{
			name:     "empty term",
			term:     "",
			expected: false,
		},
		{
			name:          "github actions with CI",
			term:          shared.TestTerminalXterm256,
			ci:            "true",
			githubActions: "true",
			expected:      true,
		},
		{
			name:     "CI without github actions",
			term:     shared.TestTerminalXterm256,
			ci:       "true",
			expected: false,
		},
		{
			name:     "NO_COLOR set",
			term:     shared.TestTerminalXterm256,
			noColor:  "1",
			expected: false,
		},
		{
			name:       "FORCE_COLOR set",
			term:       shared.TestTerminalXterm256,
			forceColor: "1",
			expected:   true,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				clearColorTerminalEnvVars(t)
				setColorTerminalTestEnv(t, tt)

				result := isColorTerminal()
				if result != tt.expected {
					t.Errorf("isColorTerminal() = %v, want %v", result, tt.expected)
				}
			},
		)
	}
}

func TestIsInteractiveTerminal(_ *testing.T) {
	// This test is limited because we can't easily mock os.Stderr.Stat()
	// but we can at least verify it doesn't panic and returns a boolean
	result := isInteractiveTerminal()

	// Result should be a boolean (true or false, both are valid)
	// result is already a boolean, so this check is always satisfied
	_ = result
}

func TestUIManagerprintf(t *testing.T) {
	ui, output := createTestUI()

	ui.printf("Hello %s", "world")

	expected := "Hello world"
	if output.String() != expected {
		t.Errorf("printf() = %q, want %q", output.String(), expected)
	}
}

// Helper function to set environment variable or unset if empty.
func setEnvOrUnset(key, value string) {
	if value == "" {
		if err := os.Unsetenv(key); err != nil {
			// In tests, environment variable errors are not critical,
			// but we should still handle them to avoid linting issues
			_ = err // explicitly ignore error
		}
	} else {
		if err := os.Setenv(key, value); err != nil {
			// In tests, environment variable errors are not critical,
			// but we should still handle them to avoid linting issues
			_ = err // explicitly ignore error
		}
	}
}

// Integration test for UI workflow.
func TestUIManagerIntegration(t *testing.T) {
	ui, output := createTestUI() //nolint:errcheck // Test helper, output buffer is used
	ui.SetColorOutput(false)     // Disable colors for consistent output
	ui.SetProgressOutput(false)  // Disable progress for testing

	// Simulate a complete UI workflow
	ui.PrintHeader("Starting Processing")
	ui.PrintInfo("Initializing system")
	ui.StartProgress(3, shared.TestProgressMessage)
	ui.UpdateProgress(1)
	ui.PrintInfo("Processing file 1")
	ui.UpdateProgress(1)
	ui.PrintWarning("Skipping invalid file")
	ui.UpdateProgress(1)
	ui.FinishProgress()
	ui.PrintSuccess("Processing completed successfully")

	outputStr := output.String()

	expectedStrings := []string{
		"Starting Processing",
		"ℹ Initializing system",
		"ℹ Processing file 1",
		"⚠ Skipping invalid file",
		"✓ Processing completed successfully",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(outputStr, expected) {
			t.Errorf("Integration test output missing expected string: %q\nFull output:\n%s", expected, outputStr)
		}
	}
}
