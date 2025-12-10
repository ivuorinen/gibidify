package shared

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"
)

// Mock test objects - local to avoid import cycles.

// mockCloser implements io.ReadCloser with configurable close error.
type mockCloser struct {
	closeError error
	closed     bool
}

func (m *mockCloser) Read(_ []byte) (n int, err error) {
	return 0, io.EOF
}

func (m *mockCloser) Close() error {
	m.closed = true

	return m.closeError
}

// mockReader implements io.Reader that returns EOF.
type mockReader struct{}

func (m *mockReader) Read(_ []byte) (n int, err error) {
	return 0, io.EOF
}

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

func TestSafeCloseReader(t *testing.T) {
	tests := []struct {
		name         string
		reader       io.Reader
		path         string
		expectClosed bool
		expectError  bool
		closeError   error
	}{
		{
			name:         "closer reader success",
			reader:       &mockCloser{},
			path:         TestPathBase,
			expectClosed: true,
			expectError:  false,
		},
		{
			name:         "closer reader with error",
			reader:       &mockCloser{closeError: errors.New("close failed")},
			path:         TestPathBase,
			expectClosed: true,
			expectError:  true,
			closeError:   errors.New("close failed"),
		},
		{
			name:         "non-closer reader",
			reader:       &mockReader{},
			path:         TestPathBase,
			expectClosed: false,
			expectError:  false,
		},
		{
			name:         "closer reader with empty path",
			reader:       &mockCloser{},
			path:         "",
			expectClosed: true,
			expectError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				// Capture the reader if it's a mockCloser
				var closerMock *mockCloser
				if closer, ok := tt.reader.(*mockCloser); ok {
					closerMock = closer
				}

				// Call SafeCloseReader (should not panic)
				SafeCloseReader(tt.reader, tt.path)

				// Verify expectations
				if closerMock != nil {
					if closerMock.closed != tt.expectClosed {
						t.Errorf("Expected closed=%v, got %v", tt.expectClosed, closerMock.closed)
					}
				}
				// Note: Error logging is tested indirectly through no panic
			},
		)
	}
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

// validateStreamError validates error expectations for stream operations.
func validateStreamError(t *testing.T, err error, errContains, filePath string) {
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
		return
	}

	if structErr.Type != ErrorTypeIO {
		t.Errorf(TestFmtExpectedErrorTypeIO, structErr.Type)
	}
	if filePath != "" && structErr.FilePath != filePath {
		t.Errorf(TestFmtExpectedFilePath, filePath, structErr.FilePath)
	}
}

func TestStreamContent(t *testing.T) {
	tests := []struct {
		name            string
		content         string
		chunkSize       int
		filePath        string
		writeError      error
		processChunk    func([]byte) []byte
		wantErr         bool
		expectedContent string
		errContains     string
	}{
		{
			name:            "successful streaming",
			content:         "hello world test content",
			chunkSize:       8,
			filePath:        TestPathTestFileTXT,
			expectedContent: "hello world test content",
		},
		{
			name:            "streaming with chunk processor",
			content:         "abc def ghi",
			chunkSize:       4,
			filePath:        TestPathTestFileTXT,
			processChunk:    bytes.ToUpper,
			expectedContent: "ABC DEF GHI",
		},
		{
			name:        "write error during streaming",
			content:     TestContentTest,
			chunkSize:   4,
			filePath:    TestPathTestFileTXT,
			writeError:  errors.New(TestErrDiskFull),
			wantErr:     true,
			errContains: "failed to write content chunk",
		},
		{
			name:            TestContentEmpty,
			content:         "",
			chunkSize:       1024,
			filePath:        TestPathTestEmptyTXT,
			expectedContent: "",
		},
		{
			name:            "large chunk size",
			content:         "small content",
			chunkSize:       1024,
			filePath:        TestPathTestFileTXT,
			expectedContent: "small content",
		},
		{
			name:            "nil processor function",
			content:         "unchanged content",
			chunkSize:       8,
			filePath:        TestPathTestFileTXT,
			processChunk:    nil,
			expectedContent: "unchanged content",
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				reader := strings.NewReader(tt.content)
				writer := &mockWriter{writeError: tt.writeError}
				err := StreamContent(reader, writer, tt.chunkSize, tt.filePath, tt.processChunk)

				if tt.wantErr {
					validateStreamError(t, err, tt.errContains, tt.filePath)

					return
				}

				if err != nil {
					t.Errorf("StreamContent() unexpected error: %v", err)
				}
				if string(writer.written) != tt.expectedContent {
					t.Errorf(TestFmtExpectedContent, tt.expectedContent, string(writer.written))
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

func TestStreamLines(t *testing.T) {
	tests := []struct {
		name            string
		content         string
		filePath        string
		readError       bool
		writeError      error
		lineProcessor   func(string) string
		wantErr         bool
		expectedContent string
		errContains     string
	}{
		{
			name:            "successful line streaming",
			content:         "line1\nline2\nline3",
			filePath:        TestPathTestFileTXT,
			expectedContent: "line1\nline2\nline3\n",
		},
		{
			name:            "line streaming with processor",
			content:         "hello\nworld",
			filePath:        TestPathTestFileTXT,
			lineProcessor:   strings.ToUpper,
			expectedContent: "HELLO\nWORLD\n",
		},
		{
			name:            TestContentEmpty,
			content:         "",
			filePath:        TestPathTestEmptyTXT,
			expectedContent: "",
		},
		{
			name:            "single line no newline",
			content:         "single line",
			filePath:        TestPathTestFileTXT,
			expectedContent: "single line\n",
		},
		{
			name:            "content ending with newline",
			content:         "line1\nline2\n",
			filePath:        TestPathTestFileTXT,
			expectedContent: "line1\nline2\n",
		},
		{
			name:        "write error during processing",
			content:     "line1\nline2",
			filePath:    TestPathTestFileTXT,
			writeError:  errors.New(TestErrDiskFull),
			wantErr:     true,
			errContains: "failed to write processed line",
		},
		{
			name:            "nil line processor",
			content:         "unchanged\ncontent",
			filePath:        TestPathTestFileTXT,
			lineProcessor:   nil,
			expectedContent: "unchanged\ncontent\n",
		},
		{
			name:            "multiple empty lines",
			content:         "\n\n\n",
			filePath:        TestPathTestFileTXT,
			expectedContent: "\n\n\n",
		},
		{
			name:     "line processor with special characters",
			content:  "hello\t world\ntest\rline",
			filePath: TestPathTestFileTXT,
			lineProcessor: func(line string) string {
				return strings.ReplaceAll(line, "\t", " ")
			},
			expectedContent: "hello  world\ntest\rline\n",
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				reader := strings.NewReader(tt.content)
				writer := &mockWriter{writeError: tt.writeError}
				err := StreamLines(reader, writer, tt.filePath, tt.lineProcessor)

				if tt.wantErr {
					validateStreamError(t, err, tt.errContains, tt.filePath)

					return
				}

				if err != nil {
					t.Errorf("StreamLines() unexpected error: %v", err)
				}
				if string(writer.written) != tt.expectedContent {
					t.Errorf(TestFmtExpectedContent, tt.expectedContent, string(writer.written))
				}
			},
		)
	}
}

// validateWriteProcessedChunkResult validates the result of writeProcessedChunk operation.
func validateWriteProcessedChunkResult(t *testing.T, writer *mockWriter, err error, wantErr bool, expected string) {
	t.Helper()

	if wantErr {
		if err == nil {
			t.Error("writeProcessedChunk() expected error, got nil")
		}
		return
	}

	if err != nil {
		t.Errorf("writeProcessedChunk() unexpected error: %v", err)
		return
	}

	if string(writer.written) != expected {
		t.Errorf("Expected %q, got %q", expected, string(writer.written))
	}
}

// Test helper functions indirectly through their usage.
func TestWriteProcessedChunk(t *testing.T) {
	tests := []struct {
		name         string
		chunk        []byte
		filePath     string
		processChunk func([]byte) []byte
		writeError   error
		wantErr      bool
		expected     string
	}{
		{
			name:         "successful chunk processing",
			chunk:        []byte("hello"),
			filePath:     TestPathTestFileTXT,
			processChunk: bytes.ToUpper,
			expected:     "HELLO",
		},
		{
			name:         "no processor",
			chunk:        []byte("unchanged"),
			filePath:     TestPathTestFileTXT,
			processChunk: nil,
			expected:     "unchanged",
		},
		{
			name:       "write error",
			chunk:      []byte("test"),
			filePath:   TestPathTestFileTXT,
			writeError: errors.New("write failed"),
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				writer := &mockWriter{writeError: tt.writeError}
				err := writeProcessedChunk(writer, tt.chunk, tt.filePath, tt.processChunk)
				validateWriteProcessedChunkResult(t, writer, err, tt.wantErr, tt.expected)
			},
		)
	}
}

// testWrapErrorFunc is a helper function to test error wrapping functions without duplication.
func testWrapErrorFunc(
	t *testing.T,
	wrapFunc func(error, string) error,
	expectedCode string,
	expectedMessage string,
	testName string,
) {
	t.Helper()

	originalErr := errors.New("original " + testName + " error")
	filePath := TestPathTestFileTXT

	wrappedErr := wrapFunc(originalErr, filePath)

	// Should return a StructuredError
	var structErr *StructuredError
	if !errors.As(wrappedErr, &structErr) {
		t.Fatal("Expected StructuredError")
	}

	// Verify error properties
	if structErr.Type != ErrorTypeIO {
		t.Errorf(TestFmtExpectedErrorTypeIO, structErr.Type)
	}
	if structErr.Code != expectedCode {
		t.Errorf("Expected %v, got %v", expectedCode, structErr.Code)
	}
	if structErr.FilePath != filePath {
		t.Errorf(TestFmtExpectedFilePath, filePath, structErr.FilePath)
	}
	if structErr.Message != expectedMessage {
		t.Errorf("Expected message %q, got %q", expectedMessage, structErr.Message)
	}

	// Test with empty file path
	wrappedErrEmpty := wrapFunc(originalErr, "")
	var structErrEmpty *StructuredError
	if errors.As(wrappedErrEmpty, &structErrEmpty) && structErrEmpty.FilePath != "" {
		t.Errorf("Expected empty FilePath, got %q", structErrEmpty.FilePath)
	}
}

func TestWrapWriteError(t *testing.T) {
	testWrapErrorFunc(t, wrapWriteError, CodeIOWrite, "failed to write content chunk", "write")
}

func TestWrapReadError(t *testing.T) {
	testWrapErrorFunc(t, wrapReadError, CodeIORead, "failed to read content chunk", "read")
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

func BenchmarkStreamContent(b *testing.B) {
	content := strings.Repeat("This is test content that will be streamed in chunks.\n", 1000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reader := strings.NewReader(content)
		writer := &bytes.Buffer{}
		_ = StreamContent(reader, writer, 1024, TestPathTestFileTXT, nil) // nolint:errcheck // benchmark test
	}
}

// Integration test.
func TestWriterIntegration(t *testing.T) {
	// Test a complete workflow using multiple writer utilities
	content := `Line 1 with "quotes"
Line 2 with special chars: {}[]
Line 3 with unicode: 世界`

	// Test JSON escaping in content
	var jsonBuf bytes.Buffer
	processedContent := EscapeForJSON(content)
	err := WriteWithErrorWrap(
		&jsonBuf,
		fmt.Sprintf(`{"content":"%s"}`, processedContent),
		"JSON write failed",
		"/test/file.json",
	)
	if err != nil {
		t.Fatalf("JSON integration failed: %v", err)
	}

	// Test YAML escaping and line streaming
	var yamlBuf bytes.Buffer
	reader := strings.NewReader(content)
	err = StreamLines(
		reader, &yamlBuf, "/test/file.yaml", func(line string) string {
			return "content: " + EscapeForYAML(line)
		},
	)
	if err != nil {
		t.Fatalf("YAML integration failed: %v", err)
	}

	// Verify both outputs contain expected content
	jsonOutput := jsonBuf.String()
	yamlOutput := yamlBuf.String()

	if !strings.Contains(jsonOutput, `\"quotes\"`) {
		t.Error("JSON output should contain escaped quotes")
	}
	if !strings.Contains(yamlOutput, `"Line 2 with special chars: {}[]"`) {
		t.Error("YAML output should contain quoted special characters line")
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
			name: "active context",
			setupContext: func() context.Context {
				return context.Background()
			},
			operation:   "test operation",
			expectError: false,
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

// TestWithContextCheck tests the WithContextCheck function.
func TestWithContextCheck(t *testing.T) {
	tests := []struct {
		name          string
		setupContext  func() context.Context
		operation     string
		fn            func() error
		expectError   bool
		errorContains string
	}{
		{
			name: "active context with successful operation",
			setupContext: func() context.Context {
				return context.Background()
			},
			operation: "successful operation",
			fn: func() error {
				return nil
			},
			expectError: false,
		},
		{
			name: "active context with failing operation",
			setupContext: func() context.Context {
				return context.Background()
			},
			operation: "failing operation",
			fn: func() error {
				return errors.New("operation failed")
			},
			expectError:   true,
			errorContains: "operation failed",
		},
		{
			name: "canceled context before operation",
			setupContext: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel() // Cancel immediately
				return ctx
			},
			operation: "canceled operation",
			fn: func() error {
				t.Error("Function should not be called with canceled context")
				return nil
			},
			expectError:   true,
			errorContains: "canceled operation canceled",
		},
		{
			name: "timeout context before operation",
			setupContext: func() context.Context {
				ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
				defer cancel()
				// Wait for timeout
				time.Sleep(1 * time.Millisecond)
				return ctx
			},
			operation: "timeout operation",
			fn: func() error {
				t.Error("Function should not be called with timed out context")
				return nil
			},
			expectError:   true,
			errorContains: "timeout operation canceled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.setupContext()
			err := WithContextCheck(ctx, tt.operation, tt.fn)

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
	t.Run("multiple operations with context check", func(t *testing.T) {
		ctx := context.Background()

		// First operation should succeed
		err := WithContextCheck(ctx, "operation 1", func() error {
			return nil
		})
		assertNoError(t, err, "First operation failed")

		// Second operation should also succeed
		err = WithContextCheck(ctx, "operation 2", func() error {
			return nil
		})
		assertNoError(t, err, "Second operation failed")
	})

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

		// Third operation should also fail
		err = WithContextCheck(ctx, "third operation", func() error {
			t.Error("Function should not be called")
			return nil
		})
		assertError(t, err, "Third operation should fail after cancellation")
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
