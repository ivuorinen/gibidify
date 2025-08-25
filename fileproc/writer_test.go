package fileproc_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/ivuorinen/gibidify/fileproc"
)

func TestStartWriter_Formats(t *testing.T) {
	// Define table-driven test cases
	tests := []struct {
		name        string
		format      string
		expectError bool
	}{
		{"JSON format", "json", false},
		{"YAML format", "yaml", false},
		{"Markdown format", "markdown", false},
		{"Invalid format", "invalid", true},
	}

	for _, tc := range tests {
		t.Run(
			tc.name, func(t *testing.T) {
				data := runWriterTest(t, tc.format)
				if tc.expectError {
					verifyErrorOutput(t, data)
				} else {
					verifyValidOutput(t, data, tc.format)
					verifyPrefixSuffix(t, data)
				}
			},
		)
	}
}

// runWriterTest executes the writer with the given format and returns the output data.
func runWriterTest(t *testing.T, format string) []byte {
	t.Helper()
	outFile, err := os.CreateTemp(t.TempDir(), "gibidify_test_output")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer func() {
		if closeErr := outFile.Close(); closeErr != nil {
			t.Errorf("close temp file: %v", closeErr)
		}
		if removeErr := os.Remove(outFile.Name()); removeErr != nil {
			t.Errorf("remove temp file: %v", removeErr)
		}
	}()

	// Prepare channels
	writeCh := make(chan fileproc.WriteRequest, 2)
	doneCh := make(chan struct{})

	// Write a couple of sample requests
	writeCh <- fileproc.WriteRequest{Path: "sample.go", Content: "package main"}
	writeCh <- fileproc.WriteRequest{Path: "example.py", Content: "def foo(): pass"}
	close(writeCh)

	// Start the writer
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		fileproc.StartWriter(outFile, writeCh, doneCh, format, "PREFIX", "SUFFIX")
	}()

	// Wait until writer signals completion
	wg.Wait()
	<-doneCh // make sure all writes finished

	// Read output
	data, err := os.ReadFile(outFile.Name())
	if err != nil {
		t.Fatalf("Error reading output file: %v", err)
	}

	return data
}

// verifyErrorOutput checks that error cases produce no output.
func verifyErrorOutput(t *testing.T, data []byte) {
	t.Helper()
	if len(data) != 0 {
		t.Errorf("Expected no output for invalid format, got:\n%s", data)
	}
}

// verifyValidOutput checks format-specific output validity.
func verifyValidOutput(t *testing.T, data []byte, format string) {
	t.Helper()
	content := string(data)
	switch format {
	case "json":
		var outStruct fileproc.OutputData
		if err := json.Unmarshal(data, &outStruct); err != nil {
			t.Errorf("JSON unmarshal failed: %v", err)
		}
	case "yaml":
		var outStruct fileproc.OutputData
		if err := yaml.Unmarshal(data, &outStruct); err != nil {
			t.Errorf("YAML unmarshal failed: %v", err)
		}
	case "markdown":
		if !strings.Contains(content, "```") {
			t.Error("Expected markdown code fences not found")
		}
	}
}

// verifyPrefixSuffix checks that output contains expected prefix and suffix.
func verifyPrefixSuffix(t *testing.T, data []byte) {
	t.Helper()
	content := string(data)
	if !strings.Contains(content, "PREFIX") {
		t.Errorf("Missing prefix in output: %s", data)
	}
	if !strings.Contains(content, "SUFFIX") {
		t.Errorf("Missing suffix in output: %s", data)
	}
}

// verifyPrefixSuffixWith checks that output contains expected custom prefix and suffix.
func verifyPrefixSuffixWith(t *testing.T, data []byte, expectedPrefix, expectedSuffix string) {
	t.Helper()
	content := string(data)
	if !strings.Contains(content, expectedPrefix) {
		t.Errorf("Missing prefix '%s' in output: %s", expectedPrefix, data)
	}
	if !strings.Contains(content, expectedSuffix) {
		t.Errorf("Missing suffix '%s' in output: %s", expectedSuffix, data)
	}
}

// TestStartWriter_StreamingFormats tests streaming functionality in all writers.
func TestStartWriter_StreamingFormats(t *testing.T) {
	tests := []struct {
		name     string
		format   string
		content  string
		fileSize int64
	}{
		{"JSON streaming", "json", strings.Repeat("line\n", 1000), 5000},
		{"YAML streaming", "yaml", strings.Repeat("data: value\n", 1000), 13000},
		{"Markdown streaming", "markdown", strings.Repeat("# Header\nContent\n", 1000), 19000},
	}

	for _, tc := range tests {
		t.Run(
			tc.name, func(t *testing.T) {
				data := runStreamingWriterTest(t, tc.format, tc.content)

				// Verify output is not empty
				if len(data) == 0 {
					t.Error("Expected streaming output but got empty result")
				}

				// Format-specific validation
				verifyValidOutput(t, data, tc.format)
				verifyPrefixSuffixWith(t, data, "STREAM_PREFIX", "STREAM_SUFFIX")

				// Verify content was written
				content := string(data)
				if !strings.Contains(content, "stream_test.txt") {
					t.Error("Expected file path in streaming output")
				}
			},
		)
	}
}

// runStreamingWriterTest executes the writer with streaming content.
func runStreamingWriterTest(t *testing.T, format, content string) []byte {
	t.Helper()

	// Create temp file with content for streaming
	contentFile, err := os.CreateTemp(t.TempDir(), "content_*.txt")
	if err != nil {
		t.Fatalf("Failed to create content file: %v", err)
	}
	defer func() {
		if err := os.Remove(contentFile.Name()); err != nil {
			t.Logf("Failed to remove content file: %v", err)
		}
	}()

	if _, err := contentFile.WriteString(content); err != nil {
		t.Fatalf("Failed to write content file: %v", err)
	}
	if err := contentFile.Close(); err != nil {
		t.Fatalf("Failed to close content file: %v", err)
	}

	// Create output file
	outFile, err := os.CreateTemp(t.TempDir(), "gibidify_stream_test_output")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer func() {
		if closeErr := outFile.Close(); closeErr != nil {
			t.Errorf("close temp file: %v", closeErr)
		}
		if removeErr := os.Remove(outFile.Name()); removeErr != nil {
			t.Errorf("remove temp file: %v", removeErr)
		}
	}()

	// Prepare channels with streaming request
	writeCh := make(chan fileproc.WriteRequest, 1)
	doneCh := make(chan struct{})

	// Create reader for streaming
	reader, err := os.Open(contentFile.Name())
	if err != nil {
		t.Fatalf("Failed to open content file for reading: %v", err)
	}
	defer func() {
		if err := reader.Close(); err != nil {
			t.Logf("Failed to close reader: %v", err)
		}
	}()

	// Write streaming request
	writeCh <- fileproc.WriteRequest{
		Path:     "stream_test.txt",
		Content:  "", // Empty for streaming
		IsStream: true,
		Reader:   reader,
	}
	close(writeCh)

	// Start the writer
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		fileproc.StartWriter(outFile, writeCh, doneCh, format, "STREAM_PREFIX", "STREAM_SUFFIX")
	}()

	// Wait until writer signals completion
	wg.Wait()
	<-doneCh

	// Read output
	data, err := os.ReadFile(outFile.Name())
	if err != nil {
		t.Fatalf("Error reading output file: %v", err)
	}

	return data
}

// setupReadOnlyFile creates a read-only file for error testing.
func setupReadOnlyFile(t *testing.T) (*os.File, chan fileproc.WriteRequest, chan struct{}) {
	t.Helper()

	outFile, err := os.CreateTemp(t.TempDir(), "readonly_*")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	if err := outFile.Chmod(0o444); err != nil {
		t.Fatalf("Failed to make file read-only: %v", err)
	}

	writeCh := make(chan fileproc.WriteRequest, 1)
	doneCh := make(chan struct{})

	writeCh <- fileproc.WriteRequest{
		Path:    "test.go",
		Content: "package main",
	}
	close(writeCh)

	return outFile, writeCh, doneCh
}

// setupStreamingError creates a streaming request with a failing reader.
func setupStreamingError(t *testing.T) (*os.File, chan fileproc.WriteRequest, chan struct{}) {
	t.Helper()

	outFile, err := os.CreateTemp(t.TempDir(), "yaml_stream_*")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	writeCh := make(chan fileproc.WriteRequest, 1)
	doneCh := make(chan struct{})

	pr, pw := io.Pipe()
	err = pw.CloseWithError(errors.New("simulated stream error"))
	if err != nil {
		return nil, nil, nil
	}

	writeCh <- fileproc.WriteRequest{
		Path:     "stream_fail.yaml",
		Content:  "", // Empty for streaming
		IsStream: true,
		Reader:   pr,
	}
	close(writeCh)

	return outFile, writeCh, doneCh
}

// setupSpecialCharacters creates requests with special characters.
func setupSpecialCharacters(t *testing.T) (*os.File, chan fileproc.WriteRequest, chan struct{}) {
	t.Helper()

	outFile, err := os.CreateTemp(t.TempDir(), "markdown_special_*")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	writeCh := make(chan fileproc.WriteRequest, 2)
	doneCh := make(chan struct{})

	writeCh <- fileproc.WriteRequest{
		Path:    "special\ncharacters.md",
		Content: "Content with\x00null bytes and\ttabs",
	}

	writeCh <- fileproc.WriteRequest{
		Path:    "empty.md",
		Content: "",
	}
	close(writeCh)

	return outFile, writeCh, doneCh
}

// runErrorHandlingTest runs a single error handling test.
func runErrorHandlingTest(
	t *testing.T,
	outFile *os.File,
	writeCh chan fileproc.WriteRequest,
	doneCh chan struct{},
	format string,
) {
	t.Helper()

	defer func() {
		if err := os.Remove(outFile.Name()); err != nil {
			t.Logf("Failed to remove temp file: %v", err)
		}
	}()
	defer func() {
		if err := outFile.Close(); err != nil {
			t.Logf("Failed to close temp file: %v", err)
		}
	}()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		fileproc.StartWriter(outFile, writeCh, doneCh, format, "PREFIX", "SUFFIX")
	}()

	wg.Wait()
	<-doneCh
}

// TestStartWriter_ErrorHandling tests error scenarios in writers.
func TestStartWriter_ErrorHandling(t *testing.T) {
	tests := []struct {
		name        string
		format      string
		setupError  func(t *testing.T) (*os.File, chan fileproc.WriteRequest, chan struct{})
		expectError bool
	}{
		{
			name:        "JSON writer with read-only file",
			format:      "json",
			setupError:  setupReadOnlyFile,
			expectError: true,
		},
		{
			name:        "YAML writer with streaming error",
			format:      "yaml",
			setupError:  setupStreamingError,
			expectError: true,
		},
		{
			name:        "Markdown writer with special characters",
			format:      "markdown",
			setupError:  setupSpecialCharacters,
			expectError: false,
		},
	}

	for _, tc := range tests {
		t.Run(
			tc.name, func(t *testing.T) {
				outFile, writeCh, doneCh := tc.setupError(t)
				runErrorHandlingTest(t, outFile, writeCh, doneCh, tc.format)
			},
		)
	}
}

// setupCloseTest sets up files and channels for close testing.
func setupCloseTest(t *testing.T) (*os.File, chan fileproc.WriteRequest, chan struct{}) {
	t.Helper()

	outFile, err := os.CreateTemp(t.TempDir(), "close_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	writeCh := make(chan fileproc.WriteRequest, 5)
	doneCh := make(chan struct{})

	for i := 0; i < 5; i++ {
		writeCh <- fileproc.WriteRequest{
			Path:    fmt.Sprintf("file%d.txt", i),
			Content: fmt.Sprintf("Content %d", i),
		}
	}
	close(writeCh)

	return outFile, writeCh, doneCh
}

// runCloseTest executes writer and validates output.
func runCloseTest(
	t *testing.T,
	outFile *os.File,
	writeCh chan fileproc.WriteRequest,
	doneCh chan struct{},
	format string,
) {
	t.Helper()

	defer func() {
		if err := os.Remove(outFile.Name()); err != nil {
			t.Logf("Failed to remove temp file: %v", err)
		}
	}()
	defer func() {
		if err := outFile.Close(); err != nil {
			t.Logf("Failed to close temp file: %v", err)
		}
	}()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		fileproc.StartWriter(outFile, writeCh, doneCh, format, "TEST_PREFIX", "TEST_SUFFIX")
	}()

	wg.Wait()
	<-doneCh

	data, err := os.ReadFile(outFile.Name())
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	if len(data) == 0 {
		t.Error("Expected non-empty output file")
	}

	content := string(data)
	if !strings.Contains(content, "TEST_PREFIX") {
		t.Error("Expected prefix in output")
	}
	if !strings.Contains(content, "TEST_SUFFIX") {
		t.Error("Expected suffix in output")
	}
}

// TestStartWriter_WriterCloseErrors tests error handling during writer close operations.
func TestStartWriter_WriterCloseErrors(t *testing.T) {
	tests := []struct {
		name   string
		format string
	}{
		{"JSON close handling", "json"},
		{"YAML close handling", "yaml"},
		{"Markdown close handling", "markdown"},
	}

	for _, tc := range tests {
		t.Run(
			tc.name, func(t *testing.T) {
				outFile, writeCh, doneCh := setupCloseTest(t)
				runCloseTest(t, outFile, writeCh, doneCh, tc.format)
			},
		)
	}
}
