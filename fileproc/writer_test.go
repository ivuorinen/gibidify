package fileproc_test

import (
	"encoding/json"
	"fmt"
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

// setupReadOnlyFile creates a read-only file for error testing.
func setupReadOnlyFile(t *testing.T) (*os.File, chan fileproc.WriteRequest, chan struct{}) {
	t.Helper()

	outPath := filepath.Join(t.TempDir(), "readonly_out")
	outFile, err := os.Create(outPath) //nolint:gosec // G304: path constructed from t.TempDir()
	if err != nil {
		t.Fatalf(shared.TestMsgFailedToCreateFile, err)
	}

	// Close writable FD and reopen as read-only so writes will fail
	_ = outFile.Close()
	outFile, err = os.OpenFile(outPath, os.O_RDONLY, 0) //nolint:gosec // G304: path constructed from t.TempDir()
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

// TestStartWriterErrorHandling tests error scenarios in writers.
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

// TestStartWriterWriterCloseErrors tests error handling during writer close operations.
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
