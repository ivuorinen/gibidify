package utils

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
)

// captureLogOutput captures logrus output for testing
func captureLogOutput(f func()) string {
	var buf bytes.Buffer
	logrus.SetOutput(&buf)
	defer logrus.SetOutput(logrus.StandardLogger().Out)
	f()
	return buf.String()
}

func TestLogError(t *testing.T) {
	tests := []struct {
		name      string
		operation string
		err       error
		args      []any
		wantLog   string
		wantEmpty bool
	}{
		{
			name:      "nil error should not log",
			operation: "test operation",
			err:       nil,
			args:      nil,
			wantEmpty: true,
		},
		{
			name:      "basic error logging",
			operation: "failed to read file",
			err:       errors.New("permission denied"),
			args:      nil,
			wantLog:   "failed to read file: permission denied",
		},
		{
			name:      "error with formatting args",
			operation: "failed to process file %s",
			err:       errors.New("file too large"),
			args:      []any{"test.txt"},
			wantLog:   "failed to process file test.txt: file too large",
		},
		{
			name:      "error with multiple formatting args",
			operation: "failed to copy from %s to %s",
			err:       errors.New("disk full"),
			args:      []any{"source.txt", "dest.txt"},
			wantLog:   "failed to copy from source.txt to dest.txt: disk full",
		},
		{
			name:      "wrapped error",
			operation: "database operation failed",
			err:       fmt.Errorf("connection error: %w", errors.New("timeout")),
			args:      nil,
			wantLog:   "database operation failed: connection error: timeout",
		},
		{
			name:      "empty operation string",
			operation: "",
			err:       errors.New("some error"),
			args:      nil,
			wantLog:   ": some error",
		},
		{
			name:      "operation with percentage sign",
			operation: "processing 50% complete",
			err:       errors.New("interrupted"),
			args:      nil,
			wantLog:   "processing 50% complete: interrupted",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := captureLogOutput(func() {
				LogError(tt.operation, tt.err, tt.args...)
			})

			if tt.wantEmpty {
				if output != "" {
					t.Errorf("LogError() logged output when error was nil: %q", output)
				}
				return
			}

			if !strings.Contains(output, tt.wantLog) {
				t.Errorf("LogError() output = %q, want to contain %q", output, tt.wantLog)
			}

			// Verify it's logged at ERROR level
			if !strings.Contains(output, "level=error") {
				t.Errorf("LogError() should log at ERROR level, got: %q", output)
			}
		})
	}
}

func TestLogErrorf(t *testing.T) {
	tests := []struct {
		name      string
		err       error
		format    string
		args      []any
		wantLog   string
		wantEmpty bool
	}{
		{
			name:      "nil error should not log",
			err:       nil,
			format:    "operation %s failed",
			args:      []any{"test"},
			wantEmpty: true,
		},
		{
			name:    "basic formatted error",
			err:     errors.New("not found"),
			format:  "file %s not found",
			args:    []any{"config.yaml"},
			wantLog: "file config.yaml not found: not found",
		},
		{
			name:    "multiple format arguments",
			err:     errors.New("invalid range"),
			format:  "value %d is not between %d and %d",
			args:    []any{150, 0, 100},
			wantLog: "value 150 is not between 0 and 100: invalid range",
		},
		{
			name:    "no format arguments",
			err:     errors.New("generic error"),
			format:  "operation failed",
			args:    nil,
			wantLog: "operation failed: generic error",
		},
		{
			name:    "format with different types",
			err:     errors.New("type mismatch"),
			format:  "expected %s but got %d",
			args:    []any{"string", 42},
			wantLog: "expected string but got 42: type mismatch",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := captureLogOutput(func() {
				LogErrorf(tt.err, tt.format, tt.args...)
			})

			if tt.wantEmpty {
				if output != "" {
					t.Errorf("LogErrorf() logged output when error was nil: %q", output)
				}
				return
			}

			if !strings.Contains(output, tt.wantLog) {
				t.Errorf("LogErrorf() output = %q, want to contain %q", output, tt.wantLog)
			}

			// Verify it's logged at ERROR level
			if !strings.Contains(output, "level=error") {
				t.Errorf("LogErrorf() should log at ERROR level, got: %q", output)
			}
		})
	}
}

func TestLogErrorConcurrency(t *testing.T) {
	// Test that LogError is safe for concurrent use
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(n int) {
			LogError("concurrent operation", fmt.Errorf("error %d", n))
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestLogErrorfConcurrency(t *testing.T) {
	// Test that LogErrorf is safe for concurrent use
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(n int) {
			LogErrorf(fmt.Errorf("error %d", n), "concurrent operation %d", n)
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}

// BenchmarkLogError benchmarks the LogError function
func BenchmarkLogError(b *testing.B) {
	err := errors.New("benchmark error")
	// Disable output during benchmark
	logrus.SetOutput(bytes.NewBuffer(nil))
	defer logrus.SetOutput(logrus.StandardLogger().Out)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		LogError("benchmark operation", err)
	}
}

// BenchmarkLogErrorf benchmarks the LogErrorf function
func BenchmarkLogErrorf(b *testing.B) {
	err := errors.New("benchmark error")
	// Disable output during benchmark
	logrus.SetOutput(bytes.NewBuffer(nil))
	defer logrus.SetOutput(logrus.StandardLogger().Out)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		LogErrorf(err, "benchmark operation %d", i)
	}
}

// BenchmarkLogErrorNil benchmarks LogError with nil error (no-op case)
func BenchmarkLogErrorNil(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		LogError("benchmark operation", nil)
	}
}

// TestStructuredErrorMethods tests methods on StructuredError.
func TestStructuredErrorMethods(t *testing.T) {
	baseErr := errors.New("base error")
	structErr := WrapError(baseErr, ErrorTypeIO, CodeIORead, "read failed")

	// Test Unwrap
	if structErr.Unwrap() != baseErr {
		t.Errorf("Expected Unwrap to return base error")
	}

	// Test WithContext
	structErr = structErr.WithContext("key", "value")
	if structErr.Context["key"] != "value" {
		t.Errorf("Expected context key to be set")
	}

	// Test WithFilePath
	structErr = structErr.WithFilePath("/test/file.txt")
	if structErr.FilePath != "/test/file.txt" {
		t.Errorf("Expected file path to be set")
	}

	// Test WithLine
	structErr = structErr.WithLine(42)
	if structErr.Line != 42 {
		t.Errorf("Expected line number to be set")
	}
}

// TestNewStructuredErrorf tests creating formatted structured errors.
func TestNewStructuredErrorf(t *testing.T) {
	structErr := NewStructuredErrorf(ErrorTypeValidation, CodeValidationFormat, "validation failed for %s", "field1")

	if structErr.Type != ErrorTypeValidation {
		t.Errorf("Expected error type %v, got %v", ErrorTypeValidation, structErr.Type)
	}

	if structErr.Code != CodeValidationFormat {
		t.Errorf("Expected error code %v, got %v", CodeValidationFormat, structErr.Code)
	}

	if !strings.Contains(structErr.Message, "field1") {
		t.Errorf("Expected message to contain formatted value")
	}
}

// TestWrapErrorf tests wrapping errors with formatted messages.
func TestWrapErrorf(t *testing.T) {
	baseErr := errors.New("base error")
	wrappedErr := WrapErrorf(baseErr, ErrorTypeFileSystem, CodeFSPathResolution, "failed to process %s", "file.txt")

	if wrappedErr.Type != ErrorTypeFileSystem {
		t.Errorf("Expected error type %v, got %v", ErrorTypeFileSystem, wrappedErr.Type)
	}

	if !strings.Contains(wrappedErr.Message, "file.txt") {
		t.Errorf("Expected message to contain formatted value")
	}

	if wrappedErr.Cause != baseErr {
		t.Errorf("Expected cause error to be set")
	}
}

// TestErrorConstructors tests the various error constructor functions.
func TestErrorConstructors(t *testing.T) {
	tests := []struct {
		name        string
		constructor func() *StructuredError
		wantType    ErrorType
	}{
		{
			name: "NewCLIMissingSourceError",
			constructor: func() *StructuredError {
				return NewCLIMissingSourceError()
			},
			wantType: ErrorTypeCLI,
		},
		{
			name: "NewFileSystemError",
			constructor: func() *StructuredError {
				return NewFileSystemError(CodeFSNotFound, "test error")
			},
			wantType: ErrorTypeFileSystem,
		},
		{
			name: "NewProcessingError",
			constructor: func() *StructuredError {
				return NewProcessingError(CodeProcessingFileRead, "test error")
			},
			wantType: ErrorTypeProcessing,
		},
		{
			name: "NewIOError",
			constructor: func() *StructuredError {
				return NewIOError(CodeIORead, "test error")
			},
			wantType: ErrorTypeIO,
		},
		{
			name: "NewValidationError",
			constructor: func() *StructuredError {
				return NewValidationError(CodeValidationFormat, "test error")
			},
			wantType: ErrorTypeValidation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.constructor()
			if err.Type != tt.wantType {
				t.Errorf("Expected error type %v, got %v", tt.wantType, err.Type)
			}
		})
	}
}

// TestErrorString tests the String method on error types.
func TestErrorString(t *testing.T) {
	tests := []struct {
		name       string
		errorType  ErrorType
		wantString string
	}{
		{"IO error", ErrorTypeIO, "IO"},
		{"FileSystem error", ErrorTypeFileSystem, "FileSystem"},
		{"Validation error", ErrorTypeValidation, "Validation"},
		{"Processing error", ErrorTypeProcessing, "Processing"},
		{"CLI error", ErrorTypeCLI, "CLI"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.errorType.String()
			if got != tt.wantString {
				t.Errorf("ErrorType.String() = %q, want %q", got, tt.wantString)
			}
		})
	}
}

// TestStructuredErrorError tests the Error() method.
func TestStructuredErrorError(t *testing.T) {
	baseErr := errors.New("base error")
	structErr := WrapError(baseErr, ErrorTypeIO, CodeIORead, "read failed")
	structErr = structErr.WithFilePath("/test/file.txt").WithLine(10)

	errMsg := structErr.Error()

	// Check that the error message contains key information
	if !strings.Contains(errMsg, "read failed") {
		t.Error("Error message should contain the message")
	}
	if !strings.Contains(errMsg, "base error") {
		t.Error("Error message should contain the base error")
	}
}
