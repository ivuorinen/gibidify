package shared

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

// Mock test objects - local to avoid import cycles.

// mockWriter implements io.Writer with configurable write error.
type mockWriter struct {
	writeError error
	written    []byte
}

func (m *mockWriter) Write(p []byte) (n int, err error) {
	if m.writeError != nil {
		return 0, m.writeError
	}
	m.written = append(m.written, p...)

	return len(p), nil
}

// validateWriteError validates error expectations for write operations.
func validateWriteError(t *testing.T, err error, errContains, filePath string) {
	t.Helper()

	if err == nil {
		t.Error("Expected error, got nil")

		return
	}

	if errContains != "" && !strings.Contains(err.Error(), errContains) {
		t.Errorf("Error should contain %q, got: %v", errContains, err.Error())
	}

	var structErr *StructuredError
	if !errors.As(err, &structErr) {
		t.Error("Expected StructuredError")

		return
	}

	if structErr.Type != ErrorTypeIO {
		t.Errorf(TestFmtExpectedErrorTypeIO, structErr.Type)
	}
	if structErr.Code != CodeIOWrite {
		t.Errorf("Expected CodeIOWrite, got %v", structErr.Code)
	}
	if filePath != "" && structErr.FilePath != filePath {
		t.Errorf(TestFmtExpectedFilePath, filePath, structErr.FilePath)
	}
}

func TestWriteWithErrorWrap(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		errorMsg    string
		filePath    string
		writeError  error
		wantErr     bool
		errContains string
	}{
		{
			name:       "successful write",
			content:    TestContentTest,
			errorMsg:   "write failed",
			filePath:   TestPathTestFileTXT,
			writeError: nil,
			wantErr:    false,
		},
		{
			name:        "write error with file path",
			content:     TestContentTest,
			errorMsg:    "custom error message",
			filePath:    TestPathTestFileTXT,
			writeError:  errors.New(TestErrDiskFull),
			wantErr:     true,
			errContains: "custom error message",
		},
		{
			name:        "write error without file path",
			content:     TestContentTest,
			errorMsg:    "write operation failed",
			filePath:    "",
			writeError:  errors.New("network error"),
			wantErr:     true,
			errContains: "write operation failed",
		},
		{
			name:       TestContentEmpty,
			content:    "",
			errorMsg:   "empty write",
			filePath:   TestPathTestEmptyTXT,
			writeError: nil,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				writer := &mockWriter{writeError: tt.writeError}
				err := WriteWithErrorWrap(writer, tt.content, tt.errorMsg, tt.filePath)

				if tt.wantErr {
					validateWriteError(t, err, tt.errContains, tt.filePath)

					return
				}

				if err != nil {
					t.Errorf("WriteWithErrorWrap() unexpected error: %v", err)
				}
				if string(writer.written) != tt.content {
					t.Errorf(TestFmtExpectedContent, tt.content, string(writer.written))
				}
			},
		)
	}
}

func TestEscapeForJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple string",
			input:    TestContentHelloWorld,
			expected: TestContentHelloWorld,
		},
		{
			name:     "string with quotes",
			input:    `hello "quoted" world`,
			expected: `hello \"quoted\" world`,
		},
		{
			name:     "string with newlines",
			input:    "line 1\nline 2\nline 3",
			expected: "line 1\\nline 2\\nline 3",
		},
		{
			name:     "string with tabs",
			input:    "col1\tcol2\tcol3",
			expected: "col1\\tcol2\\tcol3",
		},
		{
			name:     "string with backslashes",
			input:    `path\to\file`,
			expected: `path\\to\\file`,
		},
		{
			name:     "string with unicode",
			input:    "Hello 世界 🌍",
			expected: "Hello 世界 🌍",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "control characters",
			input:    "\x00\x01\x1f",
			expected: "\\u0000\\u0001\\u001f",
		},
		{
			name:     "mixed special characters",
			input:    "Line 1\n\t\"Quoted\"\r\nLine 2\\",
			expected: "Line 1\\n\\t\\\"Quoted\\\"\\r\\nLine 2\\\\",
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				result := EscapeForJSON(tt.input)
				if result != tt.expected {
					t.Errorf("EscapeForJSON() = %q, want %q", result, tt.expected)
				}
			},
		)
	}
}

func TestEscapeForYAML(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple string no quotes needed",
			input:    "hello",
			expected: "hello",
		},
		{
			name:     "string with spaces needs quotes",
			input:    TestContentHelloWorld,
			expected: `"hello world"`,
		},
		{
			name:     "string with colon needs quotes",
			input:    "key:value",
			expected: `"key:value"`,
		},
		{
			name:     "string starting with dash",
			input:    "-value",
			expected: `"-value"`,
		},
		{
			name:     "string starting with question mark",
			input:    "?value",
			expected: `"?value"`,
		},
		{
			name:     "string starting with colon",
			input:    ":value",
			expected: `":value"`,
		},
		{
			name:     "boolean true",
			input:    "true",
			expected: `"true"`,
		},
		{
			name:     "boolean false",
			input:    "false",
			expected: `"false"`,
		},
		{
			name:     "null value",
			input:    "null",
			expected: `"null"`,
		},
		{
			name:     "tilde null",
			input:    "~",
			expected: `"~"`,
		},
		{
			name:     "empty string",
			input:    "",
			expected: `""`,
		},
		{
			name:     "string with newlines",
			input:    "line1\nline2",
			expected: "\"line1\nline2\"",
		},
		{
			name:     "string with tabs",
			input:    "col1\tcol2",
			expected: "\"col1\tcol2\"",
		},
		{
			name:     "string with brackets",
			input:    "[list]",
			expected: `"[list]"`,
		},
		{
			name:     "string with braces",
			input:    "{object}",
			expected: `"{object}"`,
		},
		{
			name:     "string with pipe",
			input:    "value|other",
			expected: `"value|other"`,
		},
		{
			name:     "string with greater than",
			input:    "value>other",
			expected: `"value>other"`,
		},
		{
			name:     "string with quotes and backslashes",
			input:    `path\to"file"`,
			expected: `"path\\to\"file\""`,
		},
		{
			name:     "normal identifier",
			input:    "normalValue123",
			expected: "normalValue123",
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				result := EscapeForYAML(tt.input)
				if result != tt.expected {
					t.Errorf("EscapeForYAML() = %q, want %q", result, tt.expected)
				}
			},
		)
	}
}

// Benchmark tests for performance-critical functions.
func BenchmarkEscapeForJSON(b *testing.B) {
	testString := `This is a "test string" with various characters: \n\t\r and some unicode: 世界`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = EscapeForJSON(testString)
	}
}

func BenchmarkEscapeForYAML(b *testing.B) {
	testString := `This is a test string with: spaces, "quotes", and special chars -?:`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = EscapeForYAML(testString)
	}
}

// TestCheckContextCancellation tests the CheckContextCancellation function.
func TestCheckContextCancellation(t *testing.T) {
	tests := []struct {
		name          string
		setupContext  func() context.Context
		operation     string
		expectError   bool
		errorContains string
	}{
		{
			name:         "active context",
			setupContext: context.Background,
			operation:    "test operation",
			expectError:  false,
		},
		{
			name: "canceled context",
			setupContext: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel() // Cancel immediately
				return ctx
			},
			operation:     "test operation",
			expectError:   true,
			errorContains: "test operation canceled",
		},
		{
			name: "timeout context",
			setupContext: func() context.Context {
				ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
				defer cancel()
				// Wait for timeout
				time.Sleep(1 * time.Millisecond)
				return ctx
			},
			operation:     "timeout operation",
			expectError:   true,
			errorContains: "timeout operation canceled",
		},
		{
			name: "context with deadline exceeded",
			setupContext: func() context.Context {
				ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-1*time.Hour))
				defer cancel()
				return ctx
			},
			operation:     "deadline operation",
			expectError:   true,
			errorContains: "deadline operation canceled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.setupContext()
			err := CheckContextCancellation(ctx, tt.operation)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for %s, but got none", tt.name)
				} else if !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("Error %q should contain %q", err.Error(), tt.errorContains)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for %s: %v", tt.name, err)
				}
			}
		})
	}
}

// assertNoError is a helper that fails the test if err is not nil.
func assertNoError(t *testing.T, err error, msg string) {
	t.Helper()
	if err != nil {
		t.Errorf("%s: %v", msg, err)
	}
}

// assertError is a helper that fails the test if err is nil.
func assertError(t *testing.T, err error, msg string) {
	t.Helper()
	if err == nil {
		t.Error(msg)
	}
}

// TestContextCancellationIntegration tests integration scenarios.
func TestContextCancellationIntegration(t *testing.T) {
	t.Run("chained context checks", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())

		// First check should pass
		err := CheckContextCancellation(ctx, "first check")
		assertNoError(t, err, "First check should pass")

		// Cancel context
		cancel()

		// Second check should fail
		err = CheckContextCancellation(ctx, "second check")
		assertError(t, err, "Second check should fail after cancellation")
	})

	t.Run("context cancellation propagation", func(t *testing.T) {
		// Test that context cancellation properly propagates through nested calls
		parentCtx, parentCancel := context.WithCancel(context.Background())
		childCtx, childCancel := context.WithCancel(parentCtx)

		defer parentCancel()
		defer childCancel()

		// Both contexts should be active initially
		err := CheckContextCancellation(parentCtx, "parent")
		assertNoError(t, err, "Parent context should be active")

		err = CheckContextCancellation(childCtx, "child")
		assertNoError(t, err, "Child context should be active")

		// Cancel parent - child should also be canceled
		parentCancel()

		err = CheckContextCancellation(childCtx, "child after parent cancel")
		assertError(t, err, "Child context should be canceled when parent is canceled")
	})
}
