package cli

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/fatih/color"
	"github.com/stretchr/testify/assert"

	"github.com/ivuorinen/gibidify/gibidiutils"
)

func TestNewErrorFormatter(t *testing.T) {
	ui := &UIManager{
		output: &bytes.Buffer{},
	}

	ef := NewErrorFormatter(ui)

	assert.NotNil(t, ef)
	assert.Equal(t, ui, ef.ui)
}

func TestFormatError(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		expectedOutput []string
		notExpected    []string
	}{
		{
			name:           "nil error",
			err:            nil,
			expectedOutput: []string{},
		},
		{
			name: "structured error",
			err: gibidiutils.NewStructuredError(
				gibidiutils.ErrorTypeFileSystem,
				gibidiutils.CodeFSNotFound,
				"file not found",
				"/test/file.txt",
				map[string]interface{}{"size": 1024},
			),
			expectedOutput: []string{
				gibidiutils.IconError + " Error",
				"FileSystem",
				"file not found",
				"/test/file.txt",
				"NOT_FOUND",
			},
		},
		{
			name:           "generic error",
			err:            errors.New("something went wrong"),
			expectedOutput: []string{gibidiutils.IconError + " Error", "something went wrong"},
		},
		{
			name: "wrapped structured error",
			err: gibidiutils.WrapError(
				errors.New("inner error"),
				gibidiutils.ErrorTypeValidation,
				gibidiutils.CodeValidationRequired,
				"validation failed",
			),
			expectedOutput: []string{
				gibidiutils.IconError + " Error",
				"validation failed",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			ui := &UIManager{
				enableColors: false,
				output:       buf,
			}
			prev := color.NoColor
			color.NoColor = true
			t.Cleanup(func() { color.NoColor = prev })

			ef := NewErrorFormatter(ui)
			ef.FormatError(tt.err)

			output := buf.String()
			for _, expected := range tt.expectedOutput {
				assert.Contains(t, output, expected)
			}
			for _, notExpected := range tt.notExpected {
				assert.NotContains(t, output, notExpected)
			}
		})
	}
}

func TestFormatStructuredError(t *testing.T) {
	tests := []struct {
		name           string
		err            *gibidiutils.StructuredError
		expectedOutput []string
	}{
		{
			name: "filesystem error",
			err: gibidiutils.NewStructuredError(
				gibidiutils.ErrorTypeFileSystem,
				gibidiutils.CodeFSPermission,
				"permission denied",
				"/etc/shadow",
				nil,
			),
			expectedOutput: []string{
				"FileSystem",
				"permission denied",
				"/etc/shadow",
				"PERMISSION_DENIED",
				"Suggestions:",
			},
		},
		{
			name: "validation error",
			err: gibidiutils.NewStructuredError(
				gibidiutils.ErrorTypeValidation,
				gibidiutils.CodeValidationFormat,
				"invalid format",
				"",
				map[string]interface{}{"format": "xml"},
			),
			expectedOutput: []string{
				"Validation",
				"invalid format",
				"FORMAT",
				"Suggestions:",
			},
		},
		{
			name: "processing error",
			err: gibidiutils.NewStructuredError(
				gibidiutils.ErrorTypeProcessing,
				gibidiutils.CodeProcessingFileRead,
				"failed to read file",
				"large.bin",
				nil,
			),
			expectedOutput: []string{
				"Processing",
				"failed to read file",
				"large.bin",
				"FILE_READ",
				"Suggestions:",
			},
		},
		{
			name: "IO error",
			err: gibidiutils.NewStructuredError(
				gibidiutils.ErrorTypeIO,
				gibidiutils.CodeIOFileWrite,
				"disk full",
				"/output/result.txt",
				nil,
			),
			expectedOutput: []string{
				"IO",
				"disk full",
				"/output/result.txt",
				"FILE_WRITE",
				"Suggestions:",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			ui := &UIManager{
				enableColors: false,
				output:       buf,
			}
			prev := color.NoColor
			color.NoColor = true
			t.Cleanup(func() { color.NoColor = prev })

			ef := &ErrorFormatter{ui: ui}
			ef.formatStructuredError(tt.err)

			output := buf.String()
			for _, expected := range tt.expectedOutput {
				assert.Contains(t, output, expected)
			}
		})
	}
}

func TestFormatGenericError(t *testing.T) {
	buf := &bytes.Buffer{}
	ui := &UIManager{
		enableColors: false,
		output:       buf,
	}
	prev := color.NoColor
	color.NoColor = true
	t.Cleanup(func() { color.NoColor = prev })

	ef := &ErrorFormatter{ui: ui}
	ef.formatGenericError(errors.New("generic error message"))

	output := buf.String()
	assert.Contains(t, output, gibidiutils.IconError+" Error")
	assert.Contains(t, output, "generic error message")
}

func TestProvideSuggestions(t *testing.T) {
	tests := []struct {
		name           string
		err            *gibidiutils.StructuredError
		expectedSugges []string
	}{
		{
			name: "filesystem permission error",
			err: gibidiutils.NewStructuredError(
				gibidiutils.ErrorTypeFileSystem,
				gibidiutils.CodeFSPermission,
				"permission denied",
				"/root/file",
				nil,
			),
			expectedSugges: []string{
				"Check file/directory permissions",
				"Verify the path is correct",
			},
		},
		{
			name: "filesystem not found error",
			err: gibidiutils.NewStructuredError(
				gibidiutils.ErrorTypeFileSystem,
				gibidiutils.CodeFSNotFound,
				"file not found",
				"/missing/file",
				nil,
			),
			expectedSugges: []string{
				"Check if the file/directory exists: /missing/file",
			},
		},
		{
			name: "validation format error",
			err: gibidiutils.NewStructuredError(
				gibidiutils.ErrorTypeValidation,
				gibidiutils.CodeValidationFormat,
				"unsupported format",
				"",
				nil,
			),
			expectedSugges: []string{
				"Use a supported format: markdown, json, yaml",
				"Example: -format markdown",
			},
		},
		{
			name: "validation path error",
			err: gibidiutils.NewStructuredError(
				gibidiutils.ErrorTypeValidation,
				gibidiutils.CodeValidationPath,
				"invalid path",
				"../../etc",
				nil,
			),
			expectedSugges: []string{
				"Check your command line arguments",
				"Run with --help for usage information",
			},
		},
		{
			name: "processing file read error",
			err: gibidiutils.NewStructuredError(
				gibidiutils.ErrorTypeProcessing,
				gibidiutils.CodeProcessingFileRead,
				"read error",
				"corrupted.dat",
				nil,
			),
			expectedSugges: []string{
				"Check file permissions",
				"Verify the file is not corrupted",
			},
		},
		{
			name: "IO file write error",
			err: gibidiutils.NewStructuredError(
				gibidiutils.ErrorTypeIO,
				gibidiutils.CodeIOFileWrite,
				"write failed",
				"/output.txt",
				nil,
			),
			expectedSugges: []string{
				"Check file/directory permissions",
				"Verify available disk space",
			},
		},
		{
			name: "unknown error type",
			err: gibidiutils.NewStructuredError(
				gibidiutils.ErrorTypeUnknown,
				"UNKNOWN",
				"unknown error",
				"",
				nil,
			),
			expectedSugges: []string{
				"Check your command line arguments",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			ui := &UIManager{
				enableColors: false,
				output:       buf,
			}
			prev := color.NoColor
			color.NoColor = true
			t.Cleanup(func() { color.NoColor = prev })

			ef := &ErrorFormatter{ui: ui}
			ef.provideSuggestions(tt.err)

			output := buf.String()
			for _, suggestion := range tt.expectedSugges {
				assert.Contains(t, output, suggestion)
			}
		})
	}
}

func TestProvideFileSystemSuggestions(t *testing.T) {
	tests := []struct {
		name           string
		err            *gibidiutils.StructuredError
		expectedSugges []string
	}{
		{
			name: "permission denied",
			err: gibidiutils.NewStructuredError(
				gibidiutils.ErrorTypeFileSystem,
				gibidiutils.CodeFSPermission,
				"permission denied",
				"/root/secret",
				nil,
			),
			expectedSugges: []string{
				"Check file/directory permissions",
				"Verify the path is correct",
			},
		},
		{
			name: "path resolution error",
			err: gibidiutils.NewStructuredError(
				gibidiutils.ErrorTypeFileSystem,
				gibidiutils.CodeFSPathResolution,
				"path error",
				"../../../etc",
				nil,
			),
			expectedSugges: []string{
				"Use an absolute path instead of relative",
			},
		},
		{
			name: "file not found",
			err: gibidiutils.NewStructuredError(
				gibidiutils.ErrorTypeFileSystem,
				gibidiutils.CodeFSNotFound,
				"not found",
				"/missing.txt",
				nil,
			),
			expectedSugges: []string{
				"Check if the file/directory exists: /missing.txt",
			},
		},
		{
			name: "default filesystem error",
			err: gibidiutils.NewStructuredError(
				gibidiutils.ErrorTypeFileSystem,
				"OTHER_FS_ERROR",
				"other error",
				"/some/path",
				nil,
			),
			expectedSugges: []string{
				"Check file/directory permissions",
				"Verify the path is correct",
				"Path: /some/path",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			ui := &UIManager{
				enableColors: false,
				output:       buf,
			}

			ef := &ErrorFormatter{ui: ui}
			ef.provideFileSystemSuggestions(tt.err)

			output := buf.String()
			for _, suggestion := range tt.expectedSugges {
				assert.Contains(t, output, suggestion)
			}
		})
	}
}

func TestProvideValidationSuggestions(t *testing.T) {
	tests := []struct {
		name           string
		err            *gibidiutils.StructuredError
		expectedSugges []string
	}{
		{
			name: "format validation",
			err: gibidiutils.NewStructuredError(
				gibidiutils.ErrorTypeValidation,
				gibidiutils.CodeValidationFormat,
				"invalid format",
				"",
				nil,
			),
			expectedSugges: []string{
				"Use a supported format: markdown, json, yaml",
				"Example: -format markdown",
			},
		},
		{
			name: "path validation",
			err: gibidiutils.NewStructuredError(
				gibidiutils.ErrorTypeValidation,
				gibidiutils.CodeValidationPath,
				"invalid path",
				"",
				nil,
			),
			expectedSugges: []string{
				"Check your command line arguments",
				"Run with --help for usage information",
			},
		},
		{
			name: "size validation",
			err: gibidiutils.NewStructuredError(
				gibidiutils.ErrorTypeValidation,
				gibidiutils.CodeValidationSize,
				"size error",
				"",
				nil,
			),
			expectedSugges: []string{
				"Increase file size limit in config.yaml",
				"Use smaller files or exclude large files",
			},
		},
		{
			name: "required validation",
			err: gibidiutils.NewStructuredError(
				gibidiutils.ErrorTypeValidation,
				gibidiutils.CodeValidationRequired,
				"required",
				"",
				nil,
			),
			expectedSugges: []string{
				"Check your command line arguments",
				"Run with --help for usage information",
			},
		},
		{
			name: "default validation",
			err: gibidiutils.NewStructuredError(
				gibidiutils.ErrorTypeValidation,
				"OTHER_VALIDATION",
				"other",
				"",
				nil,
			),
			expectedSugges: []string{
				"Check your command line arguments",
				"Run with --help for usage information",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			ui := &UIManager{
				enableColors: false,
				output:       buf,
			}

			ef := &ErrorFormatter{ui: ui}
			ef.provideValidationSuggestions(tt.err)

			output := buf.String()
			for _, suggestion := range tt.expectedSugges {
				assert.Contains(t, output, suggestion)
			}
		})
	}
}

func TestProvideProcessingSuggestions(t *testing.T) {
	tests := []struct {
		name           string
		err            *gibidiutils.StructuredError
		expectedSugges []string
	}{
		{
			name: "file read error",
			err: gibidiutils.NewStructuredError(
				gibidiutils.ErrorTypeProcessing,
				gibidiutils.CodeProcessingFileRead,
				"read error",
				"",
				nil,
			),
			expectedSugges: []string{
				"Check file permissions",
				"Verify the file is not corrupted",
			},
		},
		{
			name: "collection error",
			err: gibidiutils.NewStructuredError(
				gibidiutils.ErrorTypeProcessing,
				gibidiutils.CodeProcessingCollection,
				"collection error",
				"",
				nil,
			),
			expectedSugges: []string{
				"Check if the source directory exists and is readable",
				"Verify directory permissions",
			},
		},
		{
			name: "encoding error",
			err: gibidiutils.NewStructuredError(
				gibidiutils.ErrorTypeProcessing,
				gibidiutils.CodeProcessingEncode,
				"encoding error",
				"",
				nil,
			),
			expectedSugges: []string{
				"Try reducing concurrency: -concurrency 1",
				"Check available system resources",
			},
		},
		{
			name: "default processing",
			err: gibidiutils.NewStructuredError(
				gibidiutils.ErrorTypeProcessing,
				"OTHER",
				"other error",
				"",
				nil,
			),
			expectedSugges: []string{
				"Try reducing concurrency: -concurrency 1",
				"Check available system resources",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			ui := &UIManager{
				enableColors: false,
				output:       buf,
			}

			ef := &ErrorFormatter{ui: ui}
			ef.provideProcessingSuggestions(tt.err)

			output := buf.String()
			for _, suggestion := range tt.expectedSugges {
				assert.Contains(t, output, suggestion)
			}
		})
	}
}

func TestProvideIOSuggestions(t *testing.T) {
	tests := []struct {
		name           string
		err            *gibidiutils.StructuredError
		expectedSugges []string
	}{
		{
			name: "file create error",
			err: gibidiutils.NewStructuredError(
				gibidiutils.ErrorTypeIO,
				gibidiutils.CodeIOFileCreate,
				"create error",
				"",
				nil,
			),
			expectedSugges: []string{
				"Check if the destination directory exists",
				"Verify write permissions for the output file",
				"Ensure sufficient disk space",
			},
		},
		{
			name: "file write error",
			err: gibidiutils.NewStructuredError(
				gibidiutils.ErrorTypeIO,
				gibidiutils.CodeIOFileWrite,
				"write error",
				"",
				nil,
			),
			expectedSugges: []string{
				"Check file/directory permissions",
				"Verify available disk space",
			},
		},
		{
			name: "encoding error",
			err: gibidiutils.NewStructuredError(
				gibidiutils.ErrorTypeIO,
				gibidiutils.CodeIOEncoding,
				"encoding error",
				"",
				nil,
			),
			expectedSugges: []string{
				"Check file/directory permissions",
				"Verify available disk space",
			},
		},
		{
			name: "default IO error",
			err: gibidiutils.NewStructuredError(
				gibidiutils.ErrorTypeIO,
				"OTHER",
				"other error",
				"",
				nil,
			),
			expectedSugges: []string{
				"Check file/directory permissions",
				"Verify available disk space",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			ui := &UIManager{
				enableColors: false,
				output:       buf,
			}

			ef := &ErrorFormatter{ui: ui}
			ef.provideIOSuggestions(tt.err)

			output := buf.String()
			for _, suggestion := range tt.expectedSugges {
				assert.Contains(t, output, suggestion)
			}
		})
	}
}

func TestProvideGenericSuggestions(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		expectedSugges []string
	}{
		{
			name: "permission error",
			err:  errors.New("permission denied accessing file"),
			expectedSugges: []string{
				"Check file/directory permissions",
				"Try running with appropriate privileges",
			},
		},
		{
			name: "not found error",
			err:  errors.New("no such file or directory"),
			expectedSugges: []string{
				"Verify the file/directory path is correct",
				"Check if the file exists",
			},
		},
		{
			name: "memory error",
			err:  errors.New("out of memory"),
			expectedSugges: []string{
				"Check your command line arguments",
				"Run with --help for usage information",
				"Try with -concurrency 1 to reduce resource usage",
			},
		},
		{
			name: "timeout error",
			err:  errors.New("operation timed out"),
			expectedSugges: []string{
				"Check your command line arguments",
				"Run with --help for usage information",
				"Try with -concurrency 1 to reduce resource usage",
			},
		},
		{
			name: "connection error",
			err:  errors.New("connection refused"),
			expectedSugges: []string{
				"Check your command line arguments",
				"Run with --help for usage information",
				"Try with -concurrency 1 to reduce resource usage",
			},
		},
		{
			name: "default error",
			err:  errors.New("unknown error occurred"),
			expectedSugges: []string{
				"Check your command line arguments",
				"Run with --help for usage information",
				"Try with -concurrency 1 to reduce resource usage",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			ui := &UIManager{
				enableColors: false,
				output:       buf,
			}

			ef := &ErrorFormatter{ui: ui}
			ef.provideGenericSuggestions(tt.err)

			output := buf.String()
			for _, suggestion := range tt.expectedSugges {
				assert.Contains(t, output, suggestion)
			}
		})
	}
}

func TestMissingSourceError(t *testing.T) {
	err := &MissingSourceError{}

	assert.Equal(t, "source directory is required", err.Error())
}

func TestNewMissingSourceErrorType(t *testing.T) {
	err := NewMissingSourceError()

	assert.NotNil(t, err)
	assert.Equal(t, "source directory is required", err.Error())

	var msErr *MissingSourceError
	ok := errors.As(err, &msErr)
	assert.True(t, ok)
	assert.NotNil(t, msErr)
}

// Test error formatting with colors enabled
func TestFormatErrorWithColors(t *testing.T) {
	buf := &bytes.Buffer{}
	ui := &UIManager{
		enableColors: true,
		output:       buf,
	}
	prev := color.NoColor
	color.NoColor = false
	t.Cleanup(func() { color.NoColor = prev })

	ef := NewErrorFormatter(ui)
	err := gibidiutils.NewStructuredError(
		gibidiutils.ErrorTypeValidation,
		gibidiutils.CodeValidationFormat,
		"invalid format",
		"",
		nil,
	)

	ef.FormatError(err)

	output := buf.String()
	// When colors are enabled, some output may go directly to stdout
	// Check for suggestions that are captured in the buffer
	assert.Contains(t, output, "Use a supported format: markdown, json, yaml")
	assert.Contains(t, output, "Example: -format markdown")
}

// Test wrapped error handling
func TestFormatWrappedError(t *testing.T) {
	buf := &bytes.Buffer{}
	ui := &UIManager{
		enableColors: false,
		output:       buf,
	}

	ef := NewErrorFormatter(ui)

	innerErr := errors.New("inner error")
	wrappedErr := gibidiutils.WrapError(
		innerErr,
		gibidiutils.ErrorTypeProcessing,
		gibidiutils.CodeProcessingFileRead,
		"wrapper message",
	)

	ef.FormatError(wrappedErr)

	output := buf.String()
	assert.Contains(t, output, "wrapper message")
}

// Test all suggestion paths get called
func TestSuggestionPathCoverage(t *testing.T) {
	buf := &bytes.Buffer{}
	ui := &UIManager{
		enableColors: false,
		output:       buf,
	}
	ef := &ErrorFormatter{ui: ui}

	// Test all error types
	errorTypes := []gibidiutils.ErrorType{
		gibidiutils.ErrorTypeFileSystem,
		gibidiutils.ErrorTypeValidation,
		gibidiutils.ErrorTypeProcessing,
		gibidiutils.ErrorTypeIO,
		gibidiutils.ErrorTypeConfiguration,
		gibidiutils.ErrorTypeUnknown,
	}

	for _, errType := range errorTypes {
		t.Run(errType.String(), func(t *testing.T) {
			buf.Reset()
			err := gibidiutils.NewStructuredError(
				errType,
				"TEST_CODE",
				"test error",
				"/test/path",
				nil,
			)
			ef.provideSuggestions(err)

			output := buf.String()
			// Should have some suggestion output
			assert.NotEmpty(t, output)
		})
	}
}

// Test suggestion helper functions with various inputs
func TestSuggestHelpers(t *testing.T) {
	tests := []struct {
		name     string
		testFunc func(*ErrorFormatter)
	}{
		{
			name: "suggestFileAccess",
			testFunc: func(ef *ErrorFormatter) {
				ef.suggestFileAccess("/root/file")
			},
		},
		{
			name: "suggestPathResolution",
			testFunc: func(ef *ErrorFormatter) {
				ef.suggestPathResolution("../../../etc")
			},
		},
		{
			name: "suggestFileNotFound",
			testFunc: func(ef *ErrorFormatter) {
				ef.suggestFileNotFound("/missing")
			},
		},
		{
			name: "suggestFileSystemGeneral",
			testFunc: func(ef *ErrorFormatter) {
				ef.suggestFileSystemGeneral("/path")
			},
		},
		{
			name: "provideDefaultSuggestions",
			testFunc: func(ef *ErrorFormatter) {
				ef.provideDefaultSuggestions()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			ui := &UIManager{
				enableColors: false,
				output:       buf,
			}
			ef := &ErrorFormatter{ui: ui}

			tt.testFunc(ef)

			output := buf.String()
			// Each should produce some output
			assert.NotEmpty(t, output)
			// Should contain bullet point
			assert.Contains(t, output, gibidiutils.IconBullet)
		})
	}
}

// Test edge cases in error message analysis
func TestGenericSuggestionsEdgeCases(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{"empty message", errors.New("")},
		{"very long message", errors.New(strings.Repeat("error ", 100))},
		{"special characters", errors.New("error!@#$%^&*()")},
		{"newlines", errors.New("error\nwith\nnewlines")},
		{"unicode", errors.New("error with 中文 characters")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			ui := &UIManager{
				enableColors: false,
				output:       buf,
			}
			ef := &ErrorFormatter{ui: ui}

			// Should not panic
			ef.provideGenericSuggestions(tt.err)

			output := buf.String()
			// Should have some output
			assert.NotEmpty(t, output)
		})
	}
}
