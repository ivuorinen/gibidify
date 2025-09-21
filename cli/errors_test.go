package cli

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/ivuorinen/gibidify/shared"
)

func TestNewErrorFormatter(t *testing.T) {
	ui := NewUIManager()
	formatter := NewErrorFormatter(ui)

	if formatter == nil {
		t.Error("NewErrorFormatter() returned nil")

		return
	}
	if formatter.ui != ui {
		t.Error("NewErrorFormatter() did not set ui manager correctly")
	}
}

func TestErrorFormatter_FormatError(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		expectedOutput []string // Substrings that should be present in output
	}{
		{
			name:           "nil error",
			err:            nil,
			expectedOutput: []string{}, // Should produce no output
		},
		{
			name: "structured error with context",
			err: &shared.StructuredError{
				Type:     shared.ErrorTypeFileSystem,
				Code:     shared.CodeFSAccess,
				Message:  "cannot access file",
				FilePath: "/test/path",
				Context: map[string]any{
					"permission": "0000",
					"owner":      "root",
				},
			},
			expectedOutput: []string{
				"✗ Error: cannot access file",
				"Type: FileSystem, Code: ACCESS_DENIED",
				"File: /test/path",
				"Context:",
				"permission: 0000",
				"owner: root",
				"⚠ Suggestions:",
				"Check if the path exists",
			},
		},
		{
			name: "validation error",
			err: &shared.StructuredError{
				Type:    shared.ErrorTypeValidation,
				Code:    shared.CodeValidationFormat,
				Message: "invalid output format",
			},
			expectedOutput: []string{
				"✗ Error: invalid output format",
				"Type: Validation, Code: FORMAT",
				"⚠ Suggestions:",
				"Use a supported format: markdown, json, yaml",
			},
		},
		{
			name: "processing error",
			err: &shared.StructuredError{
				Type:    shared.ErrorTypeProcessing,
				Code:    shared.CodeProcessingCollection,
				Message: "failed to collect files",
			},
			expectedOutput: []string{
				"✗ Error: failed to collect files",
				"Type: Processing, Code: COLLECTION",
				"⚠ Suggestions:",
				"Check if the source directory exists",
			},
		},
		{
			name: "I/O error",
			err: &shared.StructuredError{
				Type:    shared.ErrorTypeIO,
				Code:    shared.CodeIOFileCreate,
				Message: "cannot create output file",
			},
			expectedOutput: []string{
				"✗ Error: cannot create output file",
				"Type: IO, Code: FILE_CREATE",
				"⚠ Suggestions:",
				"Check if the destination directory exists",
			},
		},
		{
			name: "generic error with permission denied",
			err:  errors.New("permission denied: access to /secret/file"),
			expectedOutput: []string{
				"✗ Error: permission denied: access to /secret/file",
				"⚠ Suggestions:",
				"Check file/directory permissions",
				"Try running with appropriate privileges",
			},
		},
		{
			name: "generic error with file not found",
			err:  errors.New("no such file or directory"),
			expectedOutput: []string{
				"✗ Error: no such file or directory",
				"⚠ Suggestions:",
				"Verify the file/directory path is correct",
				"Check if the file exists",
			},
		},
		{
			name: "generic error with flag redefined",
			err:  errors.New("flag provided but not defined: -invalid"),
			expectedOutput: []string{
				"✗ Error: flag provided but not defined: -invalid",
				"⚠ Suggestions:",
				"Check your command line arguments",
				"Run with --help for usage information",
			},
		},
		{
			name: "unknown generic error",
			err:  errors.New("some unknown error"),
			expectedOutput: []string{
				"✗ Error: some unknown error",
				"⚠ Suggestions:",
				"Check your command line arguments",
				"Run with --help for usage information",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture output
			ui, output := createTestUI()
			formatter := NewErrorFormatter(ui)

			formatter.FormatError(tt.err)

			outputStr := output.String()

			// For nil error, output should be empty
			if tt.err == nil {
				if outputStr != "" {
					t.Errorf("Expected no output for nil error, got: %s", outputStr)
				}

				return
			}

			// Check that all expected substrings are present
			for _, expected := range tt.expectedOutput {
				if !strings.Contains(outputStr, expected) {
					t.Errorf("Output missing expected substring: %q\nFull output:\n%s", expected, outputStr)
				}
			}
		})
	}
}

func TestErrorFormatter_suggestFileAccess(t *testing.T) {
	ui, output := createTestUI()
	formatter := NewErrorFormatter(ui)

	// Create a temporary file to test with existing file
	tempDir := t.TempDir()
	tempFile, err := os.Create(tempDir + "/testfile")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	if err := tempFile.Close(); err != nil {
		t.Errorf("Failed to close temp file: %v", err)
	}

	tests := []struct {
		name           string
		filePath       string
		expectedOutput []string
	}{
		{
			name:     "empty file path",
			filePath: "",
			expectedOutput: []string{
				"Check if the path exists:",
				"Verify read permissions",
			},
		},
		{
			name:     "existing file",
			filePath: tempFile.Name(),
			expectedOutput: []string{
				"Check if the path exists:",
				"Path exists but may not be accessible",
				"Mode:",
			},
		},
		{
			name:     "nonexistent file",
			filePath: "/nonexistent/file",
			expectedOutput: []string{
				"Check if the path exists:",
				"Verify read permissions",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output.Reset()
			formatter.suggestFileAccess(tt.filePath)

			outputStr := output.String()
			for _, expected := range tt.expectedOutput {
				if !strings.Contains(outputStr, expected) {
					t.Errorf("Output missing expected substring: %q\nFull output:\n%s", expected, outputStr)
				}
			}
		})
	}
}

func TestErrorFormatter_suggestFileNotFound(t *testing.T) {
	// Create a test directory with some files
	tempDir := t.TempDir()
	testFiles := []string{"similar-file.txt", "another-similar.go", "different.md"}
	for _, filename := range testFiles {
		file, err := os.Create(fmt.Sprintf("%s/%s", tempDir, filename))
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", filename, err)
		}
		if err := file.Close(); err != nil {
			t.Errorf("Failed to close test file %s: %v", filename, err)
		}
	}

	ui, output := createTestUI()
	formatter := NewErrorFormatter(ui)

	tests := []struct {
		name           string
		filePath       string
		expectedOutput []string
	}{
		{
			name:     "empty file path",
			filePath: "",
			expectedOutput: []string{
				"Check if the file/directory exists:",
			},
		},
		{
			name:     "file with similar matches",
			filePath: tempDir + "/similar",
			expectedOutput: []string{
				"Check if the file/directory exists:",
				"Similar files in",
				"similar-file.txt",
			},
		},
		{
			name:     "nonexistent directory",
			filePath: "/nonexistent/dir/file.txt",
			expectedOutput: []string{
				"Check if the file/directory exists:",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output.Reset()
			formatter.suggestFileNotFound(tt.filePath)

			outputStr := output.String()
			for _, expected := range tt.expectedOutput {
				if !strings.Contains(outputStr, expected) {
					t.Errorf("Output missing expected substring: %q\nFull output:\n%s", expected, outputStr)
				}
			}
		})
	}
}

func TestErrorFormatter_provideSuggestions(t *testing.T) {
	ui, output := createTestUI()
	formatter := NewErrorFormatter(ui)

	tests := []struct {
		name              string
		err               *shared.StructuredError
		expectSuggestions []string
	}{
		{
			name: "filesystem error",
			err: &shared.StructuredError{
				Type: shared.ErrorTypeFileSystem,
				Code: shared.CodeFSAccess,
			},
			expectSuggestions: []string{"Suggestions:", "Check if the path exists"},
		},
		{
			name: "validation error",
			err: &shared.StructuredError{
				Type: shared.ErrorTypeValidation,
				Code: shared.CodeValidationFormat,
			},
			expectSuggestions: []string{"Suggestions:", "Use a supported format"},
		},
		{
			name: "processing error",
			err: &shared.StructuredError{
				Type: shared.ErrorTypeProcessing,
				Code: shared.CodeProcessingCollection,
			},
			expectSuggestions: []string{"Suggestions:", "Check if the source directory exists"},
		},
		{
			name: "I/O error",
			err: &shared.StructuredError{
				Type: shared.ErrorTypeIO,
				Code: shared.CodeIOWrite,
			},
			expectSuggestions: []string{"Suggestions:", "Check available disk space"},
		},
		{
			name: "unknown error type",
			err: &shared.StructuredError{
				Type: shared.ErrorTypeUnknown,
			},
			expectSuggestions: []string{"Check your command line arguments"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output.Reset()
			formatter.provideSuggestions(tt.err)

			outputStr := output.String()
			for _, expected := range tt.expectSuggestions {
				if !strings.Contains(outputStr, expected) {
					t.Errorf("Output missing expected substring: %q\nFull output:\n%s", expected, outputStr)
				}
			}
		})
	}
}

func TestMissingSourceError(t *testing.T) {
	err := NewMissingSourceError()

	if err == nil {
		t.Error("NewMissingSourceError() returned nil")

		return
	}

	expectedMsg := "source directory is required"
	if err.Error() != expectedMsg {
		t.Errorf("MissingSourceError.Error() = %v, want %v", err.Error(), expectedMsg)
	}

	// Test type assertion
	var cliErr *MissingSourceError
	if !errors.As(err, &cliErr) {
		t.Error("NewMissingSourceError() did not return *MissingSourceError type")
	}
}

func TestIsUserError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "CLI missing source error",
			err:      NewMissingSourceError(),
			expected: true,
		},
		{
			name: "validation structured error",
			err: &shared.StructuredError{
				Type: shared.ErrorTypeValidation,
			},
			expected: true,
		},
		{
			name: "validation format structured error",
			err: &shared.StructuredError{
				Code: shared.CodeValidationFormat,
			},
			expected: true,
		},
		{
			name: "validation size structured error",
			err: &shared.StructuredError{
				Code: shared.CodeValidationSize,
			},
			expected: true,
		},
		{
			name: "non-validation structured error",
			err: &shared.StructuredError{
				Type: shared.ErrorTypeFileSystem,
			},
			expected: false,
		},
		{
			name:     "generic error with flag keyword",
			err:      errors.New("flag provided but not defined"),
			expected: true,
		},
		{
			name:     "generic error with usage keyword",
			err:      errors.New("usage: command [options]"),
			expected: true,
		},
		{
			name:     "generic error with invalid argument",
			err:      errors.New("invalid argument provided"),
			expected: true,
		},
		{
			name:     "generic error with file not found",
			err:      errors.New("file not found"),
			expected: true,
		},
		{
			name:     "generic error with permission denied",
			err:      errors.New("permission denied"),
			expected: true,
		},
		{
			name:     "system error not user-facing",
			err:      errors.New("internal system error"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsUserError(tt.err)
			if result != tt.expected {
				t.Errorf("IsUserError(%v) = %v, want %v", tt.err, result, tt.expected)
			}
		})
	}
}

// Helper functions for testing

// createTestUI creates a UIManager with captured output for testing.
func createTestUI() (*UIManager, *bytes.Buffer) {
	output := &bytes.Buffer{}
	ui := &UIManager{
		enableColors:   false, // Disable colors for consistent testing
		enableProgress: false, // Disable progress for testing
		output:         output,
	}

	return ui, output
}

// TestErrorFormatterIntegration tests the complete error formatting workflow.
func TestErrorFormatterIntegration(t *testing.T) {
	ui, output := createTestUI()
	formatter := NewErrorFormatter(ui)

	// Test a complete workflow with a complex structured error
	structuredErr := &shared.StructuredError{
		Type:     shared.ErrorTypeFileSystem,
		Code:     shared.CodeFSNotFound,
		Message:  "source directory not found",
		FilePath: "/missing/directory",
		Context: map[string]any{
			"attempted_path": "/missing/directory",
			"current_dir":    "/working/dir",
		},
	}

	formatter.FormatError(structuredErr)
	outputStr := output.String()

	// Verify all components are present
	expectedComponents := []string{
		"✗ Error: source directory not found",
		"Type: FileSystem, Code: NOT_FOUND",
		"File: /missing/directory",
		"Context:",
		"attempted_path: /missing/directory",
		"current_dir: /working/dir",
		"⚠ Suggestions:",
		"Check if the file/directory exists",
	}

	for _, expected := range expectedComponents {
		if !strings.Contains(outputStr, expected) {
			t.Errorf("Integration test output missing expected component: %q\nFull output:\n%s", expected, outputStr)
		}
	}
}

// TestErrorFormatter_SuggestPathResolution tests the suggestPathResolution function.
func TestErrorFormatter_SuggestPathResolution(t *testing.T) {
	tests := []struct {
		name           string
		filePath       string
		expectedOutput []string
	}{
		{
			name:     "with file path",
			filePath: "relative/path/file.txt",
			expectedOutput: []string{
				"Use an absolute path instead of relative",
				"Try:",
			},
		},
		{
			name:     "empty file path",
			filePath: "",
			expectedOutput: []string{
				"Use an absolute path instead of relative",
			},
		},
		{
			name:     "current directory reference",
			filePath: "./file.txt",
			expectedOutput: []string{
				"Use an absolute path instead of relative",
				"Try:",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture output
			output := &bytes.Buffer{}
			ui := &UIManager{output: output}
			formatter := NewErrorFormatter(ui)

			// Call the method
			formatter.suggestPathResolution(tt.filePath)

			// Check output
			outputStr := output.String()
			for _, expected := range tt.expectedOutput {
				if !strings.Contains(outputStr, expected) {
					t.Errorf("suggestPathResolution output missing: %q\nFull output: %q", expected, outputStr)
				}
			}
		})
	}
}

// TestErrorFormatter_SuggestFileSystemGeneral tests the suggestFileSystemGeneral function.
func TestErrorFormatter_SuggestFileSystemGeneral(t *testing.T) {
	tests := []struct {
		name           string
		filePath       string
		expectedOutput []string
	}{
		{
			name:     "with file path",
			filePath: "/path/to/file.txt",
			expectedOutput: []string{
				"Check file/directory permissions",
				"Verify the path is correct",
				"Path: /path/to/file.txt",
			},
		},
		{
			name:     "empty file path",
			filePath: "",
			expectedOutput: []string{
				"Check file/directory permissions",
				"Verify the path is correct",
			},
		},
		{
			name:     "relative path",
			filePath: "../parent/file.txt",
			expectedOutput: []string{
				"Check file/directory permissions",
				"Verify the path is correct",
				"Path: ../parent/file.txt",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture output
			output := &bytes.Buffer{}
			ui := &UIManager{output: output}
			formatter := NewErrorFormatter(ui)

			// Call the method
			formatter.suggestFileSystemGeneral(tt.filePath)

			// Check output
			outputStr := output.String()
			for _, expected := range tt.expectedOutput {
				if !strings.Contains(outputStr, expected) {
					t.Errorf("suggestFileSystemGeneral output missing: %q\nFull output: %q", expected, outputStr)
				}
			}

			// When no file path is provided, should not contain "Path:" line
			if tt.filePath == "" && strings.Contains(outputStr, "Path:") {
				t.Error("suggestFileSystemGeneral should not include Path line when filePath is empty")
			}
		})
	}
}

// TestErrorFormatter_SuggestionFunctions_Integration tests the integration of suggestion functions.
func TestErrorFormatter_SuggestionFunctions_Integration(t *testing.T) {
	// Test that suggestion functions work as part of the full error formatting workflow
	tests := []struct {
		name                string
		err                 *shared.StructuredError
		expectedSuggestions []string
	}{
		{
			name: "filesystem path resolution error",
			err: &shared.StructuredError{
				Type:     shared.ErrorTypeFileSystem,
				Code:     shared.CodeFSPathResolution,
				Message:  "path resolution failed",
				FilePath: "relative/path",
			},
			expectedSuggestions: []string{
				"Use an absolute path instead of relative",
				"Try:",
			},
		},
		{
			name: "filesystem unknown error",
			err: &shared.StructuredError{
				Type:     shared.ErrorTypeFileSystem,
				Code:     "UNKNOWN_FS_ERROR", // This will trigger default case
				Message:  "unknown filesystem error",
				FilePath: "/some/path",
			},
			expectedSuggestions: []string{
				"Check file/directory permissions",
				"Verify the path is correct",
				"Path: /some/path",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture output
			output := &bytes.Buffer{}
			ui := &UIManager{output: output}
			formatter := NewErrorFormatter(ui)

			// Format the error (which should include suggestions)
			formatter.FormatError(tt.err)

			// Check that expected suggestions are present
			outputStr := output.String()
			for _, expected := range tt.expectedSuggestions {
				if !strings.Contains(outputStr, expected) {
					t.Errorf("Integrated suggestion missing: %q\nFull output: %q", expected, outputStr)
				}
			}
		})
	}
}
