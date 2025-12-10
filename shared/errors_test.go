package shared

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"
)

// captureLogOutput captures logger output for testing.
func captureLogOutput(f func()) string {
	var buf bytes.Buffer
	logger := GetLogger()
	logger.SetOutput(&buf)
	defer logger.SetOutput(io.Discard) // Set to discard to avoid test output noise
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
			err:       errors.New(TestErrDiskFull),
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
		t.Run(
			tt.name, func(t *testing.T) {
				output := captureLogOutput(
					func() {
						LogError(tt.operation, tt.err, tt.args...)
					},
				)

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
			},
		)
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
		t.Run(
			tt.name, func(t *testing.T) {
				output := captureLogOutput(
					func() {
						LogErrorf(tt.err, tt.format, tt.args...)
					},
				)

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
			},
		)
	}
}

func TestLogErrorConcurrency(_ *testing.T) {
	// Test that LogError is safe for concurrent use
	done := make(chan bool)
	for i := range 10 {
		go func(n int) {
			LogError("concurrent operation", fmt.Errorf("error %d", n))
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for range 10 {
		<-done
	}
}

func TestLogErrorfConcurrency(_ *testing.T) {
	// Test that LogErrorf is safe for concurrent use
	done := make(chan bool)
	for i := range 10 {
		go func(n int) {
			LogErrorf(fmt.Errorf("error %d", n), "concurrent operation %d", n)
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for range 10 {
		<-done
	}
}

// BenchmarkLogError benchmarks the LogError function.
func BenchmarkLogError(b *testing.B) {
	err := errors.New("benchmark error")
	// Disable output during benchmark
	logger := GetLogger()
	logger.SetOutput(io.Discard)
	defer logger.SetOutput(io.Discard)

	for b.Loop() {
		LogError("benchmark operation", err)
	}
}

// BenchmarkLogErrorf benchmarks the LogErrorf function.
func BenchmarkLogErrorf(b *testing.B) {
	err := errors.New("benchmark error")
	// Disable output during benchmark
	logger := GetLogger()
	logger.SetOutput(io.Discard)
	defer logger.SetOutput(io.Discard)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		LogErrorf(err, "benchmark operation %d", i)
	}
}

// BenchmarkLogErrorNil benchmarks LogError with nil error (no-op case).
func BenchmarkLogErrorNil(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		LogError("benchmark operation", nil)
	}
}

func TestErrorTypeString(t *testing.T) {
	tests := []struct {
		name     string
		errType  ErrorType
		expected string
	}{
		{
			name:     "CLI error type",
			errType:  ErrorTypeCLI,
			expected: "CLI",
		},
		{
			name:     "FileSystem error type",
			errType:  ErrorTypeFileSystem,
			expected: "FileSystem",
		},
		{
			name:     "Processing error type",
			errType:  ErrorTypeProcessing,
			expected: "Processing",
		},
		{
			name:     "Configuration error type",
			errType:  ErrorTypeConfiguration,
			expected: "Configuration",
		},
		{
			name:     "IO error type",
			errType:  ErrorTypeIO,
			expected: "IO",
		},
		{
			name:     "Validation error type",
			errType:  ErrorTypeValidation,
			expected: "Validation",
		},
		{
			name:     "Unknown error type",
			errType:  ErrorTypeUnknown,
			expected: "Unknown",
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				result := tt.errType.String()
				if result != tt.expected {
					t.Errorf("ErrorType.String() = %q, want %q", result, tt.expected)
				}
			},
		)
	}
}

func TestStructuredErrorError(t *testing.T) {
	tests := []struct {
		name     string
		err      *StructuredError
		expected string
	}{
		{
			name: "error without cause",
			err: &StructuredError{
				Type:    ErrorTypeFileSystem,
				Code:    "ACCESS_DENIED",
				Message: "permission denied",
			},
			expected: "FileSystem [ACCESS_DENIED]: permission denied",
		},
		{
			name: "error with cause",
			err: &StructuredError{
				Type:    ErrorTypeIO,
				Code:    "WRITE_FAILED",
				Message: "unable to write file",
				Cause:   errors.New(TestErrDiskFull),
			},
			expected: "IO [WRITE_FAILED]: unable to write file: disk full",
		},
		{
			name: "error with empty message",
			err: &StructuredError{
				Type:    ErrorTypeValidation,
				Code:    "INVALID_FORMAT",
				Message: "",
			},
			expected: "Validation [INVALID_FORMAT]: ",
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				result := tt.err.Error()
				if result != tt.expected {
					t.Errorf("StructuredError.Error() = %q, want %q", result, tt.expected)
				}
			},
		)
	}
}

func TestStructuredErrorUnwrap(t *testing.T) {
	originalErr := errors.New("original error")

	tests := []struct {
		name     string
		err      *StructuredError
		expected error
	}{
		{
			name: "error with cause",
			err: &StructuredError{
				Type:  ErrorTypeIO,
				Code:  "READ_FAILED",
				Cause: originalErr,
			},
			expected: originalErr,
		},
		{
			name: "error without cause",
			err: &StructuredError{
				Type: ErrorTypeValidation,
				Code: "INVALID_INPUT",
			},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				result := tt.err.Unwrap()
				if !errors.Is(result, tt.expected) {
					t.Errorf("StructuredError.Unwrap() = %v, want %v", result, tt.expected)
				}
			},
		)
	}
}

func TestStructuredErrorWithContext(t *testing.T) {
	err := &StructuredError{
		Type:    ErrorTypeProcessing,
		Code:    "PROCESSING_FAILED",
		Message: "processing error",
	}

	// Test adding context to error without existing context
	result := err.WithContext("key1", "value1")

	// Should return the same error instance
	if !errors.Is(result, err) {
		t.Error("WithContext() should return the same error instance")
	}

	// Check that context was added
	if len(err.Context) != 1 {
		t.Errorf("Expected context length 1, got %d", len(err.Context))
	}

	if err.Context["key1"] != "value1" {
		t.Errorf("Expected context key1=value1, got %v", err.Context["key1"])
	}

	// Test adding more context
	err = err.WithContext("key2", 42)

	if len(err.Context) != 2 {
		t.Errorf("Expected context length 2, got %d", len(err.Context))
	}

	if err.Context["key2"] != 42 {
		t.Errorf("Expected context key2=42, got %v", err.Context["key2"])
	}
}

func TestStructuredErrorWithFilePath(t *testing.T) {
	err := &StructuredError{
		Type:    ErrorTypeFileSystem,
		Code:    "FILE_NOT_FOUND",
		Message: "file not found",
	}

	filePath := "/path/to/file.txt"
	result := err.WithFilePath(filePath)

	// Should return the same error instance
	if !errors.Is(result, err) {
		t.Error("WithFilePath() should return the same error instance")
	}

	// Check that file path was set
	if err.FilePath != filePath {
		t.Errorf(TestFmtExpectedFilePath, filePath, err.FilePath)
	}

	// Test overwriting existing file path
	newPath := "/another/path.txt"
	err = err.WithFilePath(newPath)

	if err.FilePath != newPath {
		t.Errorf(TestFmtExpectedFilePath, newPath, err.FilePath)
	}
}

func TestStructuredErrorWithLine(t *testing.T) {
	err := &StructuredError{
		Type:    ErrorTypeValidation,
		Code:    "SYNTAX_ERROR",
		Message: "syntax error",
	}

	lineNum := 42
	result := err.WithLine(lineNum)

	// Should return the same error instance
	if !errors.Is(result, err) {
		t.Error("WithLine() should return the same error instance")
	}

	// Check that line number was set
	if err.Line != lineNum {
		t.Errorf(TestFmtExpectedLine, lineNum, err.Line)
	}

	// Test overwriting existing line number
	newLine := 100
	err = err.WithLine(newLine)

	if err.Line != newLine {
		t.Errorf(TestFmtExpectedLine, newLine, err.Line)
	}
}

// validateStructuredErrorBasics validates basic structured error fields.
func validateStructuredErrorBasics(
	t *testing.T,
	err *StructuredError,
	errorType ErrorType,
	code, message, filePath string,
) {
	t.Helper()

	if err.Type != errorType {
		t.Errorf(TestFmtExpectedType, errorType, err.Type)
	}
	if err.Code != code {
		t.Errorf(TestFmtExpectedCode, code, err.Code)
	}
	if err.Message != message {
		t.Errorf(TestFmtExpectedMessage, message, err.Message)
	}
	if err.FilePath != filePath {
		t.Errorf(TestFmtExpectedFilePath, filePath, err.FilePath)
	}
}

// validateStructuredErrorContext validates context fields.
func validateStructuredErrorContext(t *testing.T, err *StructuredError, expectedContext map[string]any) {
	t.Helper()

	if expectedContext == nil {
		if len(err.Context) != 0 {
			t.Errorf("Expected empty context, got %v", err.Context)
		}

		return
	}

	if len(err.Context) != len(expectedContext) {
		t.Errorf("Expected context length %d, got %d", len(expectedContext), len(err.Context))
	}

	for k, v := range expectedContext {
		if err.Context[k] != v {
			t.Errorf("Expected context[%q] = %v, got %v", k, v, err.Context[k])
		}
	}
}

func TestNewStructuredError(t *testing.T) {
	tests := []struct {
		name      string
		errorType ErrorType
		code      string
		message   string
		filePath  string
		context   map[string]any
	}{
		{
			name:      "basic structured error",
			errorType: ErrorTypeFileSystem,
			code:      "ACCESS_DENIED",
			message:   TestErrAccessDenied,
			filePath:  "/test/file.txt",
			context:   nil,
		},
		{
			name:      "error with context",
			errorType: ErrorTypeValidation,
			code:      "INVALID_FORMAT",
			message:   "invalid format",
			filePath:  "",
			context: map[string]any{
				"expected": "json",
				"got":      "xml",
			},
		},
		{
			name:      "error with all fields",
			errorType: ErrorTypeIO,
			code:      "WRITE_FAILED",
			message:   "write failed",
			filePath:  "/output/file.txt",
			context: map[string]any{
				"bytes_written": 1024,
				"total_size":    2048,
			},
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				err := NewStructuredError(tt.errorType, tt.code, tt.message, tt.filePath, tt.context)
				validateStructuredErrorBasics(t, err, tt.errorType, tt.code, tt.message, tt.filePath)
				validateStructuredErrorContext(t, err, tt.context)
			},
		)
	}
}

func TestNewStructuredErrorf(t *testing.T) {
	tests := []struct {
		name        string
		errorType   ErrorType
		code        string
		format      string
		args        []any
		expectedMsg string
	}{
		{
			name:        "formatted error without args",
			errorType:   ErrorTypeProcessing,
			code:        "PROCESSING_FAILED",
			format:      TestErrProcessingFailed,
			args:        nil,
			expectedMsg: TestErrProcessingFailed,
		},
		{
			name:        "formatted error with args",
			errorType:   ErrorTypeValidation,
			code:        "INVALID_VALUE",
			format:      "invalid value %q, expected between %d and %d",
			args:        []any{"150", 0, 100},
			expectedMsg: "invalid value \"150\", expected between 0 and 100",
		},
		{
			name:        "formatted error with multiple types",
			errorType:   ErrorTypeIO,
			code:        "READ_ERROR",
			format:      "failed to read %d bytes from %s",
			args:        []any{1024, "/tmp/file.txt"},
			expectedMsg: "failed to read 1024 bytes from /tmp/file.txt",
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				err := NewStructuredErrorf(tt.errorType, tt.code, tt.format, tt.args...)

				if err.Type != tt.errorType {
					t.Errorf(TestFmtExpectedType, tt.errorType, err.Type)
				}
				if err.Code != tt.code {
					t.Errorf(TestFmtExpectedCode, tt.code, err.Code)
				}
				if err.Message != tt.expectedMsg {
					t.Errorf(TestFmtExpectedMessage, tt.expectedMsg, err.Message)
				}
			},
		)
	}
}

// validateWrapErrorResult validates wrap error results.
func validateWrapErrorResult(
	t *testing.T,
	result *StructuredError,
	originalErr error,
	errorType ErrorType,
	code, message string,
) {
	t.Helper()

	if result.Type != errorType {
		t.Errorf(TestFmtExpectedType, errorType, result.Type)
	}
	if result.Code != code {
		t.Errorf(TestFmtExpectedCode, code, result.Code)
	}
	if result.Message != message {
		t.Errorf(TestFmtExpectedMessage, message, result.Message)
	}
	if !errors.Is(result.Cause, originalErr) {
		t.Errorf("Expected Cause %v, got %v", originalErr, result.Cause)
	}

	if originalErr != nil && !errors.Is(result, originalErr) {
		t.Error("Expected errors.Is to return true for wrapped error")
	}
}

func TestWrapError(t *testing.T) {
	originalErr := errors.New("original error")

	tests := []struct {
		name      string
		err       error
		errorType ErrorType
		code      string
		message   string
	}{
		{
			name:      "wrap simple error",
			err:       originalErr,
			errorType: ErrorTypeFileSystem,
			code:      "ACCESS_DENIED",
			message:   TestErrAccessDenied,
		},
		{
			name:      "wrap nil error",
			err:       nil,
			errorType: ErrorTypeValidation,
			code:      "INVALID_INPUT",
			message:   "invalid input",
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				result := WrapError(tt.err, tt.errorType, tt.code, tt.message)
				validateWrapErrorResult(t, result, tt.err, tt.errorType, tt.code, tt.message)
			},
		)
	}
}

func TestWrapErrorf(t *testing.T) {
	originalErr := errors.New(TestErrDiskFull)

	tests := []struct {
		name        string
		err         error
		errorType   ErrorType
		code        string
		format      string
		args        []any
		expectedMsg string
	}{
		{
			name:        "wrap with formatted message",
			err:         originalErr,
			errorType:   ErrorTypeIO,
			code:        "WRITE_FAILED",
			format:      "failed to write %d bytes to %s",
			args:        []any{1024, "/tmp/output.txt"},
			expectedMsg: "failed to write 1024 bytes to /tmp/output.txt",
		},
		{
			name:        "wrap without args",
			err:         originalErr,
			errorType:   ErrorTypeProcessing,
			code:        "PROCESSING_ERROR",
			format:      TestErrProcessingFailed,
			args:        nil,
			expectedMsg: TestErrProcessingFailed,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				result := WrapErrorf(tt.err, tt.errorType, tt.code, tt.format, tt.args...)

				if result.Type != tt.errorType {
					t.Errorf(TestFmtExpectedType, tt.errorType, result.Type)
				}
				if result.Code != tt.code {
					t.Errorf(TestFmtExpectedCode, tt.code, result.Code)
				}
				if result.Message != tt.expectedMsg {
					t.Errorf(TestFmtExpectedMessage, tt.expectedMsg, result.Message)
				}
				if !errors.Is(result.Cause, tt.err) {
					t.Errorf("Expected Cause %v, got %v", tt.err, result.Cause)
				}
			},
		)
	}
}

// validatePredefinedError validates predefined error constructor results.
func validatePredefinedError(t *testing.T, err *StructuredError, expectedType ErrorType, name, code, message string) {
	t.Helper()

	if err.Type != expectedType {
		t.Errorf(TestFmtExpectedType, expectedType, err.Type)
	}

	if name != "NewMissingSourceError" {
		if err.Code != code {
			t.Errorf(TestFmtExpectedCode, code, err.Code)
		}
		if err.Message != message {
			t.Errorf(TestFmtExpectedMessage, message, err.Message)
		}
	}
}

func TestPredefinedErrorConstructors(t *testing.T) {
	tests := []struct {
		name         string
		constructor  func(string, string) *StructuredError
		expectedType ErrorType
	}{
		{
			name:         "NewMissingSourceError",
			constructor:  func(_, _ string) *StructuredError { return NewMissingSourceError() },
			expectedType: ErrorTypeCLI,
		},
		{
			name:         "NewFileSystemError",
			constructor:  NewFileSystemError,
			expectedType: ErrorTypeFileSystem,
		},
		{
			name:         "NewProcessingError",
			constructor:  NewProcessingError,
			expectedType: ErrorTypeProcessing,
		},
		{
			name:         "NewIOError",
			constructor:  NewIOError,
			expectedType: ErrorTypeIO,
		},
		{
			name:         "NewValidationError",
			constructor:  NewValidationError,
			expectedType: ErrorTypeValidation,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				code := "TEST_CODE"
				message := "test message"

				var err *StructuredError
				if tt.name == "NewMissingSourceError" {
					err = NewMissingSourceError()
				} else {
					err = tt.constructor(code, message)
				}

				validatePredefinedError(t, err, tt.expectedType, tt.name, code, message)
			},
		)
	}
}

func TestStructuredErrorIntegration(t *testing.T) {
	// Test a complete structured error workflow
	originalErr := errors.New("connection timeout")

	// Create and modify error through chaining
	err := WrapError(originalErr, ErrorTypeIO, "READ_TIMEOUT", "failed to read from network").
		WithFilePath(TestPathTmpNetworkData).
		WithLine(42).
		WithContext("host", "example.com").
		WithContext("port", 8080)

	// Test error interface implementation
	errorMsg := err.Error()
	expectedMsg := "IO [READ_TIMEOUT]: failed to read from network: connection timeout"
	if errorMsg != expectedMsg {
		t.Errorf("Expected error message %q, got %q", expectedMsg, errorMsg)
	}

	// Test unwrapping
	if !errors.Is(err, originalErr) {
		t.Error("Expected errors.Is to return true for wrapped error")
	}

	// Test properties
	if err.FilePath != TestPathTmpNetworkData {
		t.Errorf(TestFmtExpectedFilePath, TestPathTmpNetworkData, err.FilePath)
	}
	if err.Line != 42 {
		t.Errorf(TestFmtExpectedLine, 42, err.Line)
	}
	if len(err.Context) != 2 {
		t.Errorf("Expected context length 2, got %d", len(err.Context))
	}
	if err.Context["host"] != "example.com" {
		t.Errorf("Expected context host=example.com, got %v", err.Context["host"])
	}
	if err.Context["port"] != 8080 {
		t.Errorf("Expected context port=8080, got %v", err.Context["port"])
	}
}

func TestErrorTypeConstants(t *testing.T) {
	// Test that all error type constants are properly defined
	types := []ErrorType{
		ErrorTypeCLI,
		ErrorTypeFileSystem,
		ErrorTypeProcessing,
		ErrorTypeConfiguration,
		ErrorTypeIO,
		ErrorTypeValidation,
		ErrorTypeUnknown,
	}

	// Ensure all types have unique string representations
	seen := make(map[string]bool)
	for _, errType := range types {
		str := errType.String()
		if seen[str] {
			t.Errorf("Duplicate string representation: %q", str)
		}
		seen[str] = true

		if str == "" {
			t.Errorf("Empty string representation for error type %v", errType)
		}
	}
}

// Benchmark tests for StructuredError operations.
func BenchmarkNewStructuredError(b *testing.B) {
	context := map[string]any{
		"key1": "value1",
		"key2": 42,
	}

	for b.Loop() {
		_ = NewStructuredError( // nolint:errcheck // benchmark test
			ErrorTypeFileSystem,
			"ACCESS_DENIED",
			TestErrAccessDenied,
			"/test/file.txt",
			context,
		)
	}
}

func BenchmarkStructuredErrorError(b *testing.B) {
	err := NewStructuredError(ErrorTypeIO, "WRITE_FAILED", "write operation failed", "/tmp/file.txt", nil)

	for b.Loop() {
		_ = err.Error()
	}
}

func BenchmarkStructuredErrorWithContext(b *testing.B) {
	err := NewStructuredError(ErrorTypeProcessing, "PROC_FAILED", TestErrProcessingFailed, "", nil)

	for i := 0; b.Loop(); i++ {
		_ = err.WithContext(fmt.Sprintf("key%d", i), fmt.Sprintf("value%d", i)) // nolint:errcheck // benchmark test
	}
}
