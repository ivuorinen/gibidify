package utils

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"
)

// mockCloser is a mock reader that implements io.Closer
type mockCloser struct {
	io.Reader
	closed    bool
	closeErr  error
	closeChan chan bool
}

func (m *mockCloser) Close() error {
	m.closed = true
	if m.closeChan != nil {
		m.closeChan <- true
	}
	return m.closeErr
}

func TestSafeCloseReader(t *testing.T) {
	tests := []struct {
		name     string
		reader   io.Reader
		path     string
		closeErr error
		wantLog  bool
	}{
		{
			name:     "close successful",
			reader:   &mockCloser{Reader: strings.NewReader("test"), closeErr: nil},
			path:     "test.txt",
			closeErr: nil,
			wantLog:  false,
		},
		{
			name:     "close with error",
			reader:   &mockCloser{Reader: strings.NewReader("test"), closeErr: errors.New("close failed")},
			path:     "test.txt",
			closeErr: errors.New("close failed"),
			wantLog:  true,
		},
		{
			name:     "reader without closer",
			reader:   strings.NewReader("test"),
			path:     "test.txt",
			closeErr: nil,
			wantLog:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Call SafeCloseReader
			SafeCloseReader(tt.reader, tt.path)

			// Check if closer was called
			if mc, ok := tt.reader.(*mockCloser); ok {
				if !mc.closed {
					t.Error("Expected Close() to be called on mockCloser")
				}
			}
		})
	}
}

func TestWriteWithErrorWrap(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		errorMsg    string
		filePath    string
		writeErr    error
		wantErr     bool
		wantErrType ErrorType
	}{
		{
			name:     "successful write",
			content:  "test content",
			errorMsg: "write failed",
			filePath: "test.txt",
			writeErr: nil,
			wantErr:  false,
		},
		{
			name:        "write error with file path",
			content:     "test content",
			errorMsg:    "write failed",
			filePath:    "test.txt",
			writeErr:    errors.New("disk full"),
			wantErr:     true,
			wantErrType: ErrorTypeIO,
		},
		{
			name:        "write error without file path",
			content:     "test content",
			errorMsg:    "write failed",
			filePath:    "",
			writeErr:    errors.New("disk full"),
			wantErr:     true,
			wantErrType: ErrorTypeIO,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			var writer io.Writer = &buf

			// Use error writer if we want to simulate an error
			if tt.writeErr != nil {
				writer = &errorWriter{err: tt.writeErr}
			}

			err := WriteWithErrorWrap(writer, tt.content, tt.errorMsg, tt.filePath)

			if (err != nil) != tt.wantErr {
				t.Errorf("WriteWithErrorWrap() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				if structErr, ok := err.(*StructuredError); ok {
					if structErr.Type != tt.wantErrType {
						t.Errorf("Expected error type %v, got %v", tt.wantErrType, structErr.Type)
					}
				}
			} else {
				if buf.String() != tt.content {
					t.Errorf("Expected content %q, got %q", tt.content, buf.String())
				}
			}
		})
	}
}

func TestStreamContent(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		chunkSize    int
		filePath     string
		processChunk func([]byte) []byte
		wantOutput   string
		wantErr      bool
	}{
		{
			name:         "stream without processing",
			input:        "hello world",
			chunkSize:    1024,
			filePath:     "test.txt",
			processChunk: nil,
			wantOutput:   "hello world",
			wantErr:      false,
		},
		{
			name:      "stream with uppercase processing",
			input:     "hello world",
			chunkSize: 1024,
			filePath:  "test.txt",
			processChunk: func(b []byte) []byte {
				return bytes.ToUpper(b)
			},
			wantOutput: "HELLO WORLD",
			wantErr:    false,
		},
		{
			name:         "stream in small chunks",
			input:        "hello world",
			chunkSize:    2,
			filePath:     "test.txt",
			processChunk: nil,
			wantOutput:   "hello world",
			wantErr:      false,
		},
		{
			name:      "stream with prefix processing",
			input:     "test",
			chunkSize: 1024,
			filePath:  "test.txt",
			processChunk: func(b []byte) []byte {
				return append([]byte("PREFIX:"), b...)
			},
			wantOutput: "PREFIX:test",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.input)
			var writer bytes.Buffer

			err := StreamContent(reader, &writer, tt.chunkSize, tt.filePath, tt.processChunk)

			if (err != nil) != tt.wantErr {
				t.Errorf("StreamContent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && writer.String() != tt.wantOutput {
				t.Errorf("StreamContent() output = %q, want %q", writer.String(), tt.wantOutput)
			}
		})
	}
}

func TestEscapeForJSON(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "simple string",
			input: "hello world",
			want:  "hello world",
		},
		{
			name:  "string with quotes",
			input: `say "hello"`,
			want:  `say \"hello\"`,
		},
		{
			name:  "string with backslash",
			input: `path\to\file`,
			want:  `path\\to\\file`,
		},
		{
			name:  "string with newline",
			input: "line1\nline2",
			want:  `line1\nline2`,
		},
		{
			name:  "string with tab",
			input: "col1\tcol2",
			want:  `col1\tcol2`,
		},
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
		{
			name:  "string with unicode",
			input: "hello 世界",
			want:  "hello 世界",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EscapeForJSON(tt.input)
			if got != tt.want {
				t.Errorf("EscapeForJSON() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestEscapeForYAML(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "simple string",
			input: "hello",
			want:  "hello",
		},
		{
			name:  "string with space",
			input: "hello world",
			want:  `"hello world"`,
		},
		{
			name:  "string with colon",
			input: "key:value",
			want:  `"key:value"`,
		},
		{
			name:  "string starting with dash",
			input: "-item",
			want:  `"-item"`,
		},
		{
			name:  "boolean true",
			input: "true",
			want:  `"true"`,
		},
		{
			name:  "boolean false",
			input: "false",
			want:  `"false"`,
		},
		{
			name:  "null",
			input: "null",
			want:  `"null"`,
		},
		{
			name:  "tilde",
			input: "~",
			want:  `"~"`,
		},
		{
			name:  "empty string",
			input: "",
			want:  `""`,
		},
		{
			name:  "string with quotes",
			input: `say "hello"`,
			want:  `"say \"hello\""`,
		},
		{
			name:  "string with backslash",
			input: `path\file`,
			want:  `"path\\file"`,
		},
		{
			name:  "string with newline",
			input: "line1\nline2",
			want:  "\"line1\nline2\"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EscapeForYAML(tt.input)
			if got != tt.want {
				t.Errorf("EscapeForYAML() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestStreamLines(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		filePath      string
		lineProcessor func(string) string
		wantOutput    string
		wantErr       bool
	}{
		{
			name:          "stream without processing",
			input:         "line1\nline2\nline3",
			filePath:      "test.txt",
			lineProcessor: nil,
			wantOutput:    "line1\nline2\nline3\n",
			wantErr:       false,
		},
		{
			name:     "stream with uppercase processing",
			input:    "line1\nline2",
			filePath: "test.txt",
			lineProcessor: func(s string) string {
				return strings.ToUpper(s)
			},
			wantOutput: "LINE1\nLINE2\n",
			wantErr:    false,
		},
		{
			name:     "stream with prefix processing",
			input:    "item1\nitem2",
			filePath: "test.txt",
			lineProcessor: func(s string) string {
				return "- " + s
			},
			wantOutput: "- item1\n- item2\n",
			wantErr:    false,
		},
		{
			name:          "single line without newline",
			input:         "singleline",
			filePath:      "test.txt",
			lineProcessor: nil,
			wantOutput:    "singleline\n",
			wantErr:       false,
		},
		{
			name:          "empty input",
			input:         "",
			filePath:      "test.txt",
			lineProcessor: nil,
			wantOutput:    "",
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.input)
			var writer bytes.Buffer

			err := StreamLines(reader, &writer, tt.filePath, tt.lineProcessor)

			if (err != nil) != tt.wantErr {
				t.Errorf("StreamLines() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && writer.String() != tt.wantOutput {
				t.Errorf("StreamLines() output = %q, want %q", writer.String(), tt.wantOutput)
			}
		})
	}
}

// errorWriter is a mock writer that always returns an error
type errorWriter struct {
	err error
}

func (ew *errorWriter) Write(p []byte) (n int, err error) {
	return 0, ew.err
}

// errorReader is a mock reader that returns an error
type errorReader struct {
	err error
}

func (er *errorReader) Read(p []byte) (n int, err error) {
	return 0, er.err
}

func TestStreamContentWithReadError(t *testing.T) {
	reader := &errorReader{err: errors.New("read error")}
	var writer bytes.Buffer

	err := StreamContent(reader, &writer, 1024, "test.txt", nil)

	if err == nil {
		t.Error("StreamContent() expected error, got nil")
	}

	if structErr, ok := err.(*StructuredError); ok {
		if structErr.Type != ErrorTypeIO {
			t.Errorf("Expected error type %v, got %v", ErrorTypeIO, structErr.Type)
		}
	}
}

func TestStreamContentWithWriteError(t *testing.T) {
	reader := strings.NewReader("test content")
	writer := &errorWriter{err: errors.New("write error")}

	err := StreamContent(reader, writer, 1024, "test.txt", nil)

	if err == nil {
		t.Error("StreamContent() expected error, got nil")
	}

	if structErr, ok := err.(*StructuredError); ok {
		if structErr.Type != ErrorTypeIO {
			t.Errorf("Expected error type %v, got %v", ErrorTypeIO, structErr.Type)
		}
	}
}

func TestStreamLinesWithReadError(t *testing.T) {
	reader := &errorReader{err: errors.New("read error")}
	var writer bytes.Buffer

	err := StreamLines(reader, &writer, "test.txt", nil)

	if err == nil {
		t.Error("StreamLines() expected error, got nil")
	}

	if structErr, ok := err.(*StructuredError); ok {
		if structErr.Type != ErrorTypeIO {
			t.Errorf("Expected error type %v, got %v", ErrorTypeIO, structErr.Type)
		}
	}
}

func TestStreamLinesWithWriteError(t *testing.T) {
	reader := strings.NewReader("line1\nline2")
	writer := &errorWriter{err: errors.New("write error")}

	err := StreamLines(reader, writer, "test.txt", nil)

	if err == nil {
		t.Error("StreamLines() expected error, got nil")
	}

	if structErr, ok := err.(*StructuredError); ok {
		if structErr.Type != ErrorTypeIO {
			t.Errorf("Expected error type %v, got %v", ErrorTypeIO, structErr.Type)
		}
	}
}
