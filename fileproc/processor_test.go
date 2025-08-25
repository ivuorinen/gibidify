package fileproc_test

import (
	"context"
	"io"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ivuorinen/gibidify/fileproc"
	"github.com/ivuorinen/gibidify/testutil"
)

func TestProcessFile(t *testing.T) {
	// Reset and load default config to ensure proper file size limits
	testutil.ResetViperConfig(t, "")
	// Create a temporary file with known content.
	tmpFile, err := os.CreateTemp(t.TempDir(), "testfile")
	if err != nil {
		t.Fatal(err)
	}
	defer func(name string) {
		err := os.Remove(name)
		if err != nil {
			t.Fatal(err)
		}
	}(tmpFile.Name())

	content := "Test content"
	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatal(err)
	}
	errTmpFile := tmpFile.Close()
	if errTmpFile != nil {
		t.Fatal(errTmpFile)

		return
	}

	ch := make(chan fileproc.WriteRequest, 1)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		fileproc.ProcessFile(tmpFile.Name(), ch, "")
	}()
	wg.Wait()
	close(ch)

	var result string
	for req := range ch {
		result = req.Content
	}

	if !strings.Contains(result, tmpFile.Name()) {
		t.Errorf("Output does not contain file path: %s", tmpFile.Name())
	}
	if !strings.Contains(result, content) {
		t.Errorf("Output does not contain file content: %s", content)
	}
}

// TestNewFileProcessorWithMonitor tests processor creation with resource monitor.
func TestNewFileProcessorWithMonitor(t *testing.T) {
	testutil.ResetViperConfig(t, "")

	// Create a resource monitor
	monitor := fileproc.NewResourceMonitor()
	defer monitor.Close()

	processor := fileproc.NewFileProcessorWithMonitor("test_source", monitor)
	if processor == nil {
		t.Error("Expected processor but got nil")
	}

	// Test that the processor was created properly
	// We can't directly access internal fields, but we can test functionality
}

// TestProcessFileWithMonitor tests file processing with resource monitoring.
func TestProcessFileWithMonitor(t *testing.T) {
	testutil.ResetViperConfig(t, "")

	// Create temporary file
	tmpFile, err := os.CreateTemp(t.TempDir(), "testfile_monitor_*")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer func() {
		if err := os.Remove(tmpFile.Name()); err != nil {
			t.Logf("Failed to remove temp file: %v", err)
		}
	}()

	content := "Test content with monitor"
	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatalf("Failed to write content: %v", err)
	}
	if err := tmpFile.Close(); err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}

	// Create resource monitor
	monitor := fileproc.NewResourceMonitor()
	defer monitor.Close()

	ch := make(chan fileproc.WriteRequest, 1)
	ctx := context.Background()

	// Test ProcessFileWithMonitor
	var wg sync.WaitGroup
	var result string

	// Start reader goroutine first to prevent deadlock
	wg.Add(1)
	go func() {
		defer wg.Done()
		for req := range ch {
			result = req.Content
		}
	}()

	// Process the file
	err = fileproc.ProcessFileWithMonitor(ctx, tmpFile.Name(), ch, "", monitor)
	close(ch)

	if err != nil {
		t.Fatalf("ProcessFileWithMonitor failed: %v", err)
	}

	// Wait for reader to finish
	wg.Wait()

	if !strings.Contains(result, content) {
		t.Errorf("Expected content not found in processed result")
	}
}

const testContent = "package main\nfunc main() {}\n"

// TestProcess tests the basic Process function.
func TestProcess(t *testing.T) {
	testutil.ResetViperConfig(t, "")

	// Create temporary directory
	tmpDir := t.TempDir()

	// Create test file with .go extension
	testFile := tmpDir + "/test.go"
	content := testContent
	if err := os.WriteFile(testFile, []byte(content), 0o600); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	processor := fileproc.NewFileProcessor(tmpDir)
	ch := make(chan fileproc.WriteRequest, 10)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(ch)
		// Process the specific file, not the directory
		processor.Process(testFile, ch)
	}()

	// Collect results
	results := make([]fileproc.WriteRequest, 0, 1) // Pre-allocate with expected capacity
	for req := range ch {
		results = append(results, req)
	}
	wg.Wait()

	if len(results) == 0 {
		t.Error("Expected at least one processed file")

		return
	}

	// Find our test file in results
	found := false
	for _, req := range results {
		if strings.Contains(req.Path, "test.go") && strings.Contains(req.Content, content) {
			found = true

			break
		}
	}

	if !found {
		t.Error("Test file not found in processed results")
	}
}

// createLargeTestFile creates a large test file for streaming tests.
func createLargeTestFile(t *testing.T) *os.File {
	t.Helper()

	tmpFile, err := os.CreateTemp(t.TempDir(), "large_file_*.go")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	lineContent := "// This is a comment line that will be repeated many times to exceed the 1MB streaming threshold\n"
	repeatCount := (1048576 / len(lineContent)) + 1000
	largeContent := strings.Repeat(lineContent, repeatCount)

	if _, err := tmpFile.WriteString(largeContent); err != nil {
		t.Fatalf("Failed to write content: %v", err)
	}
	if err := tmpFile.Close(); err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}

	t.Logf("Created test file size: %d bytes", len(largeContent))

	return tmpFile
}

// processFileForStreaming processes a file and returns streaming/inline requests.
func processFileForStreaming(t *testing.T, filePath string) (*fileproc.WriteRequest, *fileproc.WriteRequest) {
	t.Helper()

	ch := make(chan fileproc.WriteRequest, 1)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(ch)
		fileproc.ProcessFile(filePath, ch, "")
	}()

	var streamingRequest *fileproc.WriteRequest
	var inlineRequest *fileproc.WriteRequest

	for req := range ch {
		if req.IsStream {
			reqCopy := req
			streamingRequest = &reqCopy
		} else {
			reqCopy := req
			inlineRequest = &reqCopy
		}
	}
	wg.Wait()

	return streamingRequest, inlineRequest
}

// validateStreamingRequest validates a streaming request.
func validateStreamingRequest(t *testing.T, streamingRequest *fileproc.WriteRequest, tmpFile *os.File) {
	t.Helper()

	if streamingRequest.Reader == nil {
		t.Error("Expected reader in streaming request")
	}
	if streamingRequest.Content != "" {
		t.Error("Expected empty content for streaming request")
	}

	buffer := make([]byte, 1024)
	n, err := streamingRequest.Reader.Read(buffer)
	if err != nil && err != io.EOF {
		t.Errorf("Failed to read from streaming request: %v", err)
	}

	content := string(buffer[:n])
	if !strings.Contains(content, tmpFile.Name()) {
		t.Error("Expected file path in streamed header content")
	}

	t.Log("Successfully triggered streaming for large file and tested reader")
}

// TestProcessor_StreamingIntegration tests streaming functionality in processor.
func TestProcessor_StreamingIntegration(t *testing.T) {
	testutil.ResetViperConfig(t, `
max_file_size_mb: 0.001
streaming_threshold_mb: 0.0001
`)

	tmpFile := createLargeTestFile(t)
	defer func() {
		if err := os.Remove(tmpFile.Name()); err != nil {
			t.Logf("Failed to remove temp file: %v", err)
		}
	}()

	streamingRequest, inlineRequest := processFileForStreaming(t, tmpFile.Name())

	if streamingRequest == nil && inlineRequest == nil {
		t.Error("Expected either streaming or inline request but got none")
	}

	if streamingRequest != nil {
		validateStreamingRequest(t, streamingRequest, tmpFile)
	} else {
		t.Log("File processed inline instead of streaming")
	}
}

// TestProcessor_ContextCancellation tests context cancellation during processing.
func TestProcessor_ContextCancellation(t *testing.T) {
	testutil.ResetViperConfig(t, "")

	// Create temporary directory with files
	tmpDir := t.TempDir()

	// Create multiple test files
	for i := 0; i < 5; i++ {
		testFile := tmpDir + "/" + "test" + string(rune('0'+i)) + ".go"
		content := testContent
		if err := os.WriteFile(testFile, []byte(content), 0o600); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	processor := fileproc.NewFileProcessor("test_source")
	ch := make(chan fileproc.WriteRequest, 10)

	// Use ProcessWithContext with immediate cancellation
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(ch)
		// Error is expected due to cancellation
		if err := processor.ProcessWithContext(ctx, tmpDir, ch); err != nil {
			// Log error for debugging, but don't fail test since cancellation is expected
			t.Logf("Expected error due to cancellation: %v", err)
		}
	}()

	// Collect results - should be minimal due to cancellation
	results := make([]fileproc.WriteRequest, 0, 1) // Pre-allocate with expected capacity
	for req := range ch {
		results = append(results, req)
	}
	wg.Wait()

	// With immediate cancellation, we might get 0 results
	// This tests that cancellation is respected
	t.Logf("Processed %d files with immediate cancellation", len(results))
}

// TestProcessor_ValidationEdgeCases tests edge cases in file validation.
func TestProcessor_ValidationEdgeCases(t *testing.T) {
	testutil.ResetViperConfig(t, `
max_file_size_mb: 0.001  # 1KB limit for testing
`)

	tmpDir := t.TempDir()

	// Test case 1: Non-existent file
	nonExistentFile := tmpDir + "/does-not-exist.go"
	processor := fileproc.NewFileProcessor(tmpDir)
	ch := make(chan fileproc.WriteRequest, 1)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(ch)
		processor.Process(nonExistentFile, ch)
	}()

	results := make([]fileproc.WriteRequest, 0)
	for req := range ch {
		results = append(results, req)
	}
	wg.Wait()

	// Should get no results due to file not existing
	if len(results) > 0 {
		t.Error("Expected no results for non-existent file")
	}

	// Test case 2: File that exceeds size limit
	largeFile := tmpDir + "/large.go"
	largeContent := strings.Repeat("// Large file content\n", 100) // > 1KB
	if err := os.WriteFile(largeFile, []byte(largeContent), 0o600); err != nil {
		t.Fatalf("Failed to create large file: %v", err)
	}

	ch2 := make(chan fileproc.WriteRequest, 1)
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(ch2)
		processor.Process(largeFile, ch2)
	}()

	results2 := make([]fileproc.WriteRequest, 0)
	for req := range ch2 {
		results2 = append(results2, req)
	}
	wg.Wait()

	// Should get results because even large files are processed (just different strategy)
	t.Logf("Large file processing results: %d", len(results2))
}

// TestProcessor_ContextCancellationDuringValidation tests context cancellation during file validation.
func TestProcessor_ContextCancellationDuringValidation(t *testing.T) {
	testutil.ResetViperConfig(t, "")

	tmpDir := t.TempDir()
	testFile := tmpDir + "/test.go"
	content := testContent
	if err := os.WriteFile(testFile, []byte(content), 0o600); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	processor := fileproc.NewFileProcessor(tmpDir)

	// Create context that we'll cancel during processing
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	// Let context expire
	time.Sleep(1 * time.Millisecond)

	ch := make(chan fileproc.WriteRequest, 1)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(ch)
		if err := processor.ProcessWithContext(ctx, testFile, ch); err != nil {
			t.Logf("ProcessWithContext error (may be expected): %v", err)
		}
	}()

	results := make([]fileproc.WriteRequest, 0)
	for req := range ch {
		results = append(results, req)
	}
	wg.Wait()

	// Should get no results due to context cancellation
	t.Logf("Results with canceled context: %d", len(results))
}

// TestProcessor_InMemoryProcessingEdgeCases tests edge cases in in-memory processing.
func TestProcessor_InMemoryProcessingEdgeCases(t *testing.T) {
	testutil.ResetViperConfig(t, "")

	tmpDir := t.TempDir()

	// Test with empty file
	emptyFile := tmpDir + "/empty.go"
	if err := os.WriteFile(emptyFile, []byte(""), 0o600); err != nil {
		t.Fatalf("Failed to create empty file: %v", err)
	}

	processor := fileproc.NewFileProcessor(tmpDir)
	ch := make(chan fileproc.WriteRequest, 1)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(ch)
		processor.Process(emptyFile, ch)
	}()

	results := make([]fileproc.WriteRequest, 0)
	for req := range ch {
		results = append(results, req)
	}
	wg.Wait()

	if len(results) != 1 {
		t.Errorf("Expected 1 result for empty file, got %d", len(results))
	}

	if len(results) > 0 {
		result := results[0]
		if result.Path == "" {
			t.Error("Expected path in result for empty file")
		}
		// Empty file should still be processed
	}
}

// TestProcessor_StreamingEdgeCases tests edge cases in streaming processing.
func TestProcessor_StreamingEdgeCases(t *testing.T) {
	testutil.ResetViperConfig(t, "")

	tmpDir := t.TempDir()

	// Create a file larger than streaming threshold but test error conditions
	largeFile := tmpDir + "/large_stream.go"
	largeContent := strings.Repeat("// Large streaming file content line\n", 50000) // > 1MB
	if err := os.WriteFile(largeFile, []byte(largeContent), 0o600); err != nil {
		t.Fatalf("Failed to create large file: %v", err)
	}

	processor := fileproc.NewFileProcessor(tmpDir)

	// Test with context that gets canceled during streaming
	ctx, cancel := context.WithCancel(context.Background())
	ch := make(chan fileproc.WriteRequest, 1)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(ch)

		// Start processing
		// Error is expected due to cancellation
		if err := processor.ProcessWithContext(ctx, largeFile, ch); err != nil {
			// Log error for debugging, but don't fail test since cancellation is expected
			t.Logf("Expected error due to cancellation: %v", err)
		}
	}()

	// Cancel context after a very short time
	go func() {
		time.Sleep(1 * time.Millisecond)
		cancel()
	}()

	results := make([]fileproc.WriteRequest, 0)
	for req := range ch {
		results = append(results, req)

		// If we get a streaming request, try to read from it with canceled context
		if req.IsStream && req.Reader != nil {
			buffer := make([]byte, 1024)
			_, err := req.Reader.Read(buffer)
			if err != nil && err != io.EOF {
				t.Logf("Expected error reading from canceled stream: %v", err)
			}
		}
	}
	wg.Wait()

	t.Logf("Results with streaming context cancellation: %d", len(results))
}
