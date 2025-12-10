package fileproc_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/ivuorinen/gibidify/fileproc"
	"github.com/ivuorinen/gibidify/shared"
)

func TestStartWriterFormats(t *testing.T) {
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
		t.Fatalf(shared.TestMsgFailedToCreateFile, err)
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
	writeCh <- fileproc.WriteRequest{Path: "sample.go", Content: shared.LiteralPackageMain}
	writeCh <- fileproc.WriteRequest{Path: "example.py", Content: "def foo(): pass"}
	close(writeCh)

	// Start the writer
	var wg sync.WaitGroup
	wg.Go(func() {
		fileproc.StartWriter(outFile, writeCh, doneCh, format, "PREFIX", "SUFFIX")
	})

	// Wait until writer signals completion
	wg.Wait()
	select {
	case <-doneCh: // make sure all writes finished
	case <-time.After(3 * time.Second):
		t.Fatal(shared.TestMsgTimeoutWriterCompletion)
	}

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
	default:
		// Unknown format - basic validation that we have content
		if len(content) == 0 {
			t.Errorf("Unexpected format %s with empty content", format)
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
func TestStartWriterStreamingFormats(t *testing.T) {
	tests := []struct {
		name    string
		format  string
		content string
	}{
		{"JSON streaming", "json", strings.Repeat("line\n", 1000)},
		{"YAML streaming", "yaml", strings.Repeat("data: value\n", 1000)},
		{"Markdown streaming", "markdown", strings.Repeat("# Header\nContent\n", 1000)},
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
				if !strings.Contains(content, shared.TestFileStreamTest) {
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
		t.Fatalf(shared.TestMsgFailedToCreateFile, err)
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
		Path:     shared.TestFileStreamTest,
		Content:  "", // Empty for streaming
		IsStream: true,
		Reader:   reader,
	}
	close(writeCh)

	// Start the writer
	var wg sync.WaitGroup
	wg.Go(func() {
		fileproc.StartWriter(outFile, writeCh, doneCh, format, "STREAM_PREFIX", "STREAM_SUFFIX")
	})

	// Wait until writer signals completion
	wg.Wait()
	select {
	case <-doneCh:
	case <-time.After(3 * time.Second):
		t.Fatal(shared.TestMsgTimeoutWriterCompletion)
	}

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

	outPath := filepath.Join(t.TempDir(), "readonly_out")
	outFile, err := os.Create(outPath)
	if err != nil {
		t.Fatalf(shared.TestMsgFailedToCreateFile, err)
	}

	// Close writable FD and reopen as read-only so writes will fail
	_ = outFile.Close()
	outFile, err = os.OpenFile(outPath, os.O_RDONLY, 0)
	if err != nil {
		t.Fatalf("Failed to reopen as read-only: %v", err)
	}

	writeCh := make(chan fileproc.WriteRequest, 1)
	doneCh := make(chan struct{})

	writeCh <- fileproc.WriteRequest{
		Path:    shared.TestFileGo,
		Content: shared.LiteralPackageMain,
	}
	close(writeCh)

	return outFile, writeCh, doneCh
}

// setupStreamingError creates a streaming request with a failing reader.
func setupStreamingError(t *testing.T) (*os.File, chan fileproc.WriteRequest, chan struct{}) {
	t.Helper()

	outFile, err := os.CreateTemp(t.TempDir(), "yaml_stream_*")
	if err != nil {
		t.Fatalf(shared.TestMsgFailedToCreateFile, err)
	}

	writeCh := make(chan fileproc.WriteRequest, 1)
	doneCh := make(chan struct{})

	pr, pw := io.Pipe()
	if err := pw.CloseWithError(errors.New("simulated stream error")); err != nil {
		t.Fatalf("failed to set pipe error: %v", err)
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
		t.Fatalf(shared.TestMsgFailedToCreateFile, err)
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
	expectEmpty bool,
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
	wg.Go(func() {
		fileproc.StartWriter(outFile, writeCh, doneCh, format, "PREFIX", "SUFFIX")
	})

	wg.Wait()

	// Wait for doneCh with timeout to prevent test hangs
	select {
	case <-doneCh:
	case <-time.After(3 * time.Second):
		t.Fatal(shared.TestMsgTimeoutWriterCompletion)
	}

	// Read output file and verify based on expectation
	data, err := os.ReadFile(outFile.Name())
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	if expectEmpty && len(data) != 0 {
		t.Errorf("expected empty output on error, got %d bytes", len(data))
	}
	if !expectEmpty && len(data) == 0 {
		t.Error("expected non-empty output, got empty")
	}
}

// TestStartWriter_ErrorHandling tests error scenarios in writers.
func TestStartWriterErrorHandling(t *testing.T) {
	tests := []struct {
		name              string
		format            string
		setupError        func(t *testing.T) (*os.File, chan fileproc.WriteRequest, chan struct{})
		expectEmptyOutput bool
	}{
		{
			name:              "JSON writer with read-only file",
			format:            "json",
			setupError:        setupReadOnlyFile,
			expectEmptyOutput: true,
		},
		{
			name:              "YAML writer with streaming error",
			format:            "yaml",
			setupError:        setupStreamingError,
			expectEmptyOutput: false, // Partial writes are acceptable before streaming errors
		},
		{
			name:              "Markdown writer with special characters",
			format:            "markdown",
			setupError:        setupSpecialCharacters,
			expectEmptyOutput: false,
		},
	}

	for _, tc := range tests {
		t.Run(
			tc.name, func(t *testing.T) {
				outFile, writeCh, doneCh := tc.setupError(t)
				runErrorHandlingTest(t, outFile, writeCh, doneCh, tc.format, tc.expectEmptyOutput)
			},
		)
	}
}

// setupCloseTest sets up files and channels for close testing.
func setupCloseTest(t *testing.T) (*os.File, chan fileproc.WriteRequest, chan struct{}) {
	t.Helper()

	outFile, err := os.CreateTemp(t.TempDir(), "close_test_*")
	if err != nil {
		t.Fatalf(shared.TestMsgFailedToCreateFile, err)
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
	wg.Go(func() {
		fileproc.StartWriter(outFile, writeCh, doneCh, format, "TEST_PREFIX", "TEST_SUFFIX")
	})

	wg.Wait()
	select {
	case <-doneCh:
	case <-time.After(3 * time.Second):
		t.Fatal(shared.TestMsgTimeoutWriterCompletion)
	}

	data, err := os.ReadFile(outFile.Name())
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	if len(data) == 0 {
		t.Error("Expected non-empty output file")
	}

	verifyPrefixSuffixWith(t, data, "TEST_PREFIX", "TEST_SUFFIX")
}

// TestStartWriter_WriterCloseErrors tests error handling during writer close operations.
func TestStartWriterWriterCloseErrors(t *testing.T) {
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

// Benchmarks for writer performance

// BenchmarkStartWriter benchmarks basic writer operations across formats.
func BenchmarkStartWriter(b *testing.B) {
	formats := []string{"json", "yaml", "markdown"}

	for _, format := range formats {
		b.Run(format, func(b *testing.B) {
			for b.Loop() {
				outFile, err := os.CreateTemp(b.TempDir(), "bench_output_*")
				if err != nil {
					b.Fatalf("Failed to create temp file: %v", err)
				}

				writeCh := make(chan fileproc.WriteRequest, 2)
				doneCh := make(chan struct{})

				writeCh <- fileproc.WriteRequest{Path: "sample.go", Content: shared.LiteralPackageMain}
				writeCh <- fileproc.WriteRequest{Path: "example.py", Content: "def foo(): pass"}
				close(writeCh)

				fileproc.StartWriter(outFile, writeCh, doneCh, format, "PREFIX", "SUFFIX")
				<-doneCh

				_ = outFile.Close()
			}
		})
	}
}

// benchStreamingIteration runs a single streaming benchmark iteration.
func benchStreamingIteration(b *testing.B, format, content string) {
	b.Helper()

	contentFile := createBenchContentFile(b, content)
	defer func() { _ = os.Remove(contentFile) }()

	reader, err := os.Open(contentFile)
	if err != nil {
		b.Fatalf("Failed to open content file: %v", err)
	}
	defer func() { _ = reader.Close() }()

	outFile, err := os.CreateTemp(b.TempDir(), "bench_stream_output_*")
	if err != nil {
		b.Fatalf("Failed to create output file: %v", err)
	}
	defer func() { _ = outFile.Close() }()

	writeCh := make(chan fileproc.WriteRequest, 1)
	doneCh := make(chan struct{})

	writeCh <- fileproc.WriteRequest{
		Path:     shared.TestFileStreamTest,
		Content:  "",
		IsStream: true,
		Reader:   reader,
	}
	close(writeCh)

	fileproc.StartWriter(outFile, writeCh, doneCh, format, "PREFIX", "SUFFIX")
	<-doneCh
}

// createBenchContentFile creates a temp file with content for benchmarks.
func createBenchContentFile(b *testing.B, content string) string {
	b.Helper()

	contentFile, err := os.CreateTemp(b.TempDir(), "content_*")
	if err != nil {
		b.Fatalf("Failed to create content file: %v", err)
	}
	if _, err := contentFile.WriteString(content); err != nil {
		b.Fatalf("Failed to write content: %v", err)
	}
	if err := contentFile.Close(); err != nil {
		b.Fatalf("Failed to close content file: %v", err)
	}

	return contentFile.Name()
}

// BenchmarkStartWriterStreaming benchmarks streaming writer operations across formats.
func BenchmarkStartWriterStreaming(b *testing.B) {
	formats := []string{"json", "yaml", "markdown"}
	content := strings.Repeat("line content\n", 1000)

	for _, format := range formats {
		b.Run(format, func(b *testing.B) {
			for b.Loop() {
				benchStreamingIteration(b, format, content)
			}
		})
	}
}
