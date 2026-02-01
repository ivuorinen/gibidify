package fileproc_test

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/spf13/viper"

	"github.com/ivuorinen/gibidify/config"
	"github.com/ivuorinen/gibidify/fileproc"
	"github.com/ivuorinen/gibidify/shared"
	"github.com/ivuorinen/gibidify/testutil"
)

// writeTempConfig creates a temporary config file with the given YAML content
// and returns the directory path containing the config file.
func writeTempConfig(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(content), 0o600); err != nil {
		t.Fatalf("Failed to create temp config: %v", err)
	}
	return dir
}

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
	}

	ch := make(chan fileproc.WriteRequest, 1)
	var wg sync.WaitGroup
	wg.Go(func() {
		defer close(ch)
		fileproc.ProcessFile(tmpFile.Name(), ch, "")
	})

	var result string
	for req := range ch {
		result = req.Content
	}
	wg.Wait()

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

	// Exercise the processor to verify monitor integration
	tmpFile, err := os.CreateTemp(t.TempDir(), "monitor_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString("test content"); err != nil {
		t.Fatal(err)
	}
	if err := tmpFile.Close(); err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	writeCh := make(chan fileproc.WriteRequest, 1)

	var wg sync.WaitGroup
	wg.Go(func() {
		defer close(writeCh)
		if err := processor.ProcessWithContext(ctx, tmpFile.Name(), writeCh); err != nil {
			t.Errorf("ProcessWithContext failed: %v", err)
		}
	})

	// Drain channel first to avoid deadlock if producer sends multiple requests
	requestCount := 0
	for range writeCh {
		requestCount++
	}

	// Wait for goroutine to finish after channel is drained
	wg.Wait()

	if requestCount == 0 {
		t.Error("Expected at least one write request from processor")
	}
}

// TestProcessFileWithMonitor tests file processing with resource monitoring.
func TestProcessFileWithMonitor(t *testing.T) {
	testutil.ResetViperConfig(t, "")

	// Create temporary file
	tmpFile, err := os.CreateTemp(t.TempDir(), "testfile_monitor_*")
	if err != nil {
		t.Fatalf(shared.TestMsgFailedToCreateFile, err)
	}
	defer func() {
		if err := os.Remove(tmpFile.Name()); err != nil {
			t.Logf("Failed to remove temp file: %v", err)
		}
	}()

	content := "Test content with monitor"
	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatalf(shared.TestMsgFailedToWriteContent, err)
	}
	if err := tmpFile.Close(); err != nil {
		t.Fatalf(shared.TestMsgFailedToCloseFile, err)
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
	wg.Go(func() {
		for req := range ch {
			result = req.Content
		}
	})

	// Process the file
	err = fileproc.ProcessFileWithMonitor(ctx, tmpFile.Name(), ch, "", monitor)
	close(ch)

	if err != nil {
		t.Fatalf("ProcessFileWithMonitor failed: %v", err)
	}

	// Wait for reader to finish
	wg.Wait()

	if !strings.Contains(result, content) {
		t.Error("Expected content not found in processed result")
	}
}

const testContent = "package main\nfunc main() {}\n"

// TestProcess tests the basic Process function.
func TestProcess(t *testing.T) {
	testutil.ResetViperConfig(t, "")

	// Create temporary directory
	tmpDir := t.TempDir()

	// Create test file with .go extension
	testFile := filepath.Join(tmpDir, "test.go")
	content := testContent
	if err := os.WriteFile(testFile, []byte(content), 0o600); err != nil {
		t.Fatalf(shared.TestMsgFailedToCreateTestFile, err)
	}

	processor := fileproc.NewFileProcessor(tmpDir)
	ch := make(chan fileproc.WriteRequest, 10)

	var wg sync.WaitGroup
	wg.Go(func() {
		defer close(ch)
		// Process the specific file, not the directory
		processor.Process(testFile, ch)
	})

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
		if strings.Contains(req.Path, shared.TestFileGo) && strings.Contains(req.Content, content) {
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
		t.Fatalf(shared.TestMsgFailedToCreateFile, err)
	}

	lineContent := "// Repeated comment line to exceed streaming threshold\n"
	repeatCount := (1048576 / len(lineContent)) + 1000
	largeContent := strings.Repeat(lineContent, repeatCount)

	if _, err := tmpFile.WriteString(largeContent); err != nil {
		t.Fatalf(shared.TestMsgFailedToWriteContent, err)
	}
	if err := tmpFile.Close(); err != nil {
		t.Fatalf(shared.TestMsgFailedToCloseFile, err)
	}

	t.Logf("Created test file size: %d bytes", len(largeContent))

	return tmpFile
}

// processFileForStreaming processes a file and returns streaming/inline requests.
func processFileForStreaming(t *testing.T, filePath string) (streamingReq, inlineReq *fileproc.WriteRequest) {
	t.Helper()

	ch := make(chan fileproc.WriteRequest, 1)

	var wg sync.WaitGroup
	wg.Go(func() {
		defer close(ch)
		fileproc.ProcessFile(filePath, ch, "")
	})

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

// TestProcessorStreamingIntegration tests streaming functionality in processor.
func TestProcessorStreamingIntegration(t *testing.T) {
	configDir := writeTempConfig(t, `
max_file_size_mb: 0.001
streaming_threshold_mb: 0.0001
`)
	testutil.ResetViperConfig(t, configDir)

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

// TestProcessorContextCancellation tests context cancellation during processing.
func TestProcessorContextCancellation(t *testing.T) {
	testutil.ResetViperConfig(t, "")

	// Create temporary directory with files
	tmpDir := t.TempDir()

	// Create multiple test files
	for i := 0; i < 5; i++ {
		testFile := filepath.Join(tmpDir, fmt.Sprintf("test%d.go", i))
		content := testContent
		if err := os.WriteFile(testFile, []byte(content), 0o600); err != nil {
			t.Fatalf(shared.TestMsgFailedToCreateTestFile, err)
		}
	}

	processor := fileproc.NewFileProcessor("test_source")
	ch := make(chan fileproc.WriteRequest, 10)

	// Use ProcessWithContext with immediate cancellation
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	var wg sync.WaitGroup
	wg.Go(func() {
		defer close(ch)
		// Error is expected due to cancellation
		if err := processor.ProcessWithContext(ctx, tmpDir, ch); err != nil {
			// Log error for debugging, but don't fail test since cancellation is expected
			t.Logf("Expected error due to cancellation: %v", err)
		}
	})

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

// TestProcessorValidationEdgeCases tests edge cases in file validation.
func TestProcessorValidationEdgeCases(t *testing.T) {
	configDir := writeTempConfig(t, `
max_file_size_mb: 0.001  # 1KB limit for testing
`)
	testutil.ResetViperConfig(t, configDir)

	tmpDir := t.TempDir()

	// Test case 1: Non-existent file
	nonExistentFile := filepath.Join(tmpDir, "does-not-exist.go")
	processor := fileproc.NewFileProcessor(tmpDir)
	ch := make(chan fileproc.WriteRequest, 1)

	var wg sync.WaitGroup
	wg.Go(func() {
		defer close(ch)
		processor.Process(nonExistentFile, ch)
	})

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
	largeFile := filepath.Join(tmpDir, "large.go")
	largeContent := strings.Repeat("// Large file content\n", 100) // > 1KB
	if err := os.WriteFile(largeFile, []byte(largeContent), 0o600); err != nil {
		t.Fatalf("Failed to create large file: %v", err)
	}

	ch2 := make(chan fileproc.WriteRequest, 1)
	wg.Go(func() {
		defer close(ch2)
		processor.Process(largeFile, ch2)
	})

	results2 := make([]fileproc.WriteRequest, 0)
	for req := range ch2 {
		results2 = append(results2, req)
	}
	wg.Wait()

	// Should get results because even large files are processed (just different strategy)
	t.Logf("Large file processing results: %d", len(results2))
}

// TestProcessorContextCancellationDuringValidation tests context cancellation during file validation.
func TestProcessorContextCancellationDuringValidation(t *testing.T) {
	testutil.ResetViperConfig(t, "")

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")
	content := testContent
	if err := os.WriteFile(testFile, []byte(content), 0o600); err != nil {
		t.Fatalf(shared.TestMsgFailedToCreateTestFile, err)
	}

	processor := fileproc.NewFileProcessor(tmpDir)

	// Create context that we'll cancel during processing
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	// Let context expire
	time.Sleep(1 * time.Millisecond)

	ch := make(chan fileproc.WriteRequest, 1)
	var wg sync.WaitGroup
	wg.Go(func() {
		defer close(ch)
		if err := processor.ProcessWithContext(ctx, testFile, ch); err != nil {
			t.Logf("ProcessWithContext error (may be expected): %v", err)
		}
	})

	results := make([]fileproc.WriteRequest, 0)
	for req := range ch {
		results = append(results, req)
	}
	wg.Wait()

	// Should get no results due to context cancellation
	t.Logf("Results with canceled context: %d", len(results))
}

// TestProcessorInMemoryProcessingEdgeCases tests edge cases in in-memory processing.
func TestProcessorInMemoryProcessingEdgeCases(t *testing.T) {
	testutil.ResetViperConfig(t, "")

	tmpDir := t.TempDir()

	// Test with empty file
	emptyFile := filepath.Join(tmpDir, "empty.go")
	if err := os.WriteFile(emptyFile, []byte(""), 0o600); err != nil {
		t.Fatalf("Failed to create empty file: %v", err)
	}

	processor := fileproc.NewFileProcessor(tmpDir)
	ch := make(chan fileproc.WriteRequest, 1)

	var wg sync.WaitGroup
	wg.Go(func() {
		defer close(ch)
		processor.Process(emptyFile, ch)
	})

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

// TestProcessorStreamingEdgeCases tests edge cases in streaming processing.
func TestProcessorStreamingEdgeCases(t *testing.T) {
	testutil.ResetViperConfig(t, "")

	tmpDir := t.TempDir()

	// Create a file larger than streaming threshold but test error conditions
	largeFile := filepath.Join(tmpDir, "large_stream.go")
	largeContent := strings.Repeat("// Large streaming file content line\n", 50000) // > 1MB
	if err := os.WriteFile(largeFile, []byte(largeContent), 0o600); err != nil {
		t.Fatalf("Failed to create large file: %v", err)
	}

	processor := fileproc.NewFileProcessor(tmpDir)

	// Test with context that gets canceled during streaming
	ctx, cancel := context.WithCancel(context.Background())
	ch := make(chan fileproc.WriteRequest, 1)

	var wg sync.WaitGroup
	wg.Go(func() {
		defer close(ch)

		// Start processing
		// Error is expected due to cancellation
		if err := processor.ProcessWithContext(ctx, largeFile, ch); err != nil {
			// Log error for debugging, but don't fail test since cancellation is expected
			t.Logf("Expected error due to cancellation: %v", err)
		}
	})

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

// Benchmarks for processor hot paths

// BenchmarkProcessFileInline benchmarks inline file processing for small files.
func BenchmarkProcessFileInline(b *testing.B) {
	// Initialize config for file processing
	viper.Reset()
	config.LoadConfig()

	// Create a small test file
	tmpFile, err := os.CreateTemp(b.TempDir(), "bench_inline_*.go")
	if err != nil {
		b.Fatalf(shared.TestMsgFailedToCreateFile, err)
	}

	content := strings.Repeat("// Inline benchmark content\n", 100) // ~2.6KB
	if _, err := tmpFile.WriteString(content); err != nil {
		b.Fatalf(shared.TestMsgFailedToWriteContent, err)
	}
	if err := tmpFile.Close(); err != nil {
		b.Fatalf(shared.TestMsgFailedToCloseFile, err)
	}

	b.ResetTimer()
	for b.Loop() {
		ch := make(chan fileproc.WriteRequest, 1)
		var wg sync.WaitGroup
		wg.Go(func() {
			defer close(ch)
			fileproc.ProcessFile(tmpFile.Name(), ch, "")
		})
		for req := range ch {
			_ = req // Drain channel
		}
		wg.Wait()
	}
}

// BenchmarkProcessFileStreaming benchmarks streaming file processing for large files.
func BenchmarkProcessFileStreaming(b *testing.B) {
	// Initialize config for file processing
	viper.Reset()
	config.LoadConfig()

	// Create a large test file that triggers streaming
	tmpFile, err := os.CreateTemp(b.TempDir(), "bench_streaming_*.go")
	if err != nil {
		b.Fatalf(shared.TestMsgFailedToCreateFile, err)
	}

	// Create content larger than streaming threshold (1MB)
	lineContent := "// Streaming benchmark content line that will be repeated\n"
	repeatCount := (1048576 / len(lineContent)) + 1000
	content := strings.Repeat(lineContent, repeatCount)

	if _, err := tmpFile.WriteString(content); err != nil {
		b.Fatalf(shared.TestMsgFailedToWriteContent, err)
	}
	if err := tmpFile.Close(); err != nil {
		b.Fatalf(shared.TestMsgFailedToCloseFile, err)
	}

	b.ResetTimer()
	for b.Loop() {
		ch := make(chan fileproc.WriteRequest, 1)
		var wg sync.WaitGroup
		wg.Go(func() {
			defer close(ch)
			fileproc.ProcessFile(tmpFile.Name(), ch, "")
		})
		for req := range ch {
			// If streaming, read some content to exercise the reader
			if req.IsStream && req.Reader != nil {
				buffer := make([]byte, 4096)
				for {
					_, err := req.Reader.Read(buffer)
					if err != nil {
						break
					}
				}
			}
		}
		wg.Wait()
	}
}

// BenchmarkProcessorWithContext benchmarks ProcessWithContext for a single file.
func BenchmarkProcessorWithContext(b *testing.B) {
	tmpDir := b.TempDir()
	testFile := filepath.Join(tmpDir, "bench_context.go")
	content := strings.Repeat("// Benchmark file content\n", 50)
	if err := os.WriteFile(testFile, []byte(content), 0o600); err != nil {
		b.Fatalf(shared.TestMsgFailedToCreateTestFile, err)
	}

	processor := fileproc.NewFileProcessor(tmpDir)
	ctx := context.Background()

	b.ResetTimer()
	for b.Loop() {
		ch := make(chan fileproc.WriteRequest, 1)
		var wg sync.WaitGroup
		wg.Go(func() {
			defer close(ch)
			_ = processor.ProcessWithContext(ctx, testFile, ch)
		})
		for req := range ch {
			_ = req // Drain channel
		}
		wg.Wait()
	}
}

// BenchmarkProcessorWithMonitor benchmarks processing with resource monitoring.
func BenchmarkProcessorWithMonitor(b *testing.B) {
	tmpDir := b.TempDir()
	testFile := filepath.Join(tmpDir, "bench_monitor.go")
	content := strings.Repeat("// Benchmark file content with monitor\n", 50)
	if err := os.WriteFile(testFile, []byte(content), 0o600); err != nil {
		b.Fatalf(shared.TestMsgFailedToCreateTestFile, err)
	}

	monitor := fileproc.NewResourceMonitor()
	defer monitor.Close()

	processor := fileproc.NewFileProcessorWithMonitor(tmpDir, monitor)
	ctx := context.Background()

	b.ResetTimer()
	for b.Loop() {
		ch := make(chan fileproc.WriteRequest, 1)
		var wg sync.WaitGroup
		wg.Go(func() {
			defer close(ch)
			_ = processor.ProcessWithContext(ctx, testFile, ch)
		})
		for req := range ch {
			_ = req // Drain channel
		}
		wg.Wait()
	}
}

// BenchmarkProcessorConcurrent benchmarks concurrent file processing.
func BenchmarkProcessorConcurrent(b *testing.B) {
	tmpDir := b.TempDir()

	// Create multiple test files
	testFiles := make([]string, 10)
	for i := 0; i < 10; i++ {
		testFiles[i] = filepath.Join(tmpDir, fmt.Sprintf("bench_concurrent_%d.go", i))
		content := strings.Repeat(fmt.Sprintf("// Concurrent file %d content\n", i), 50)
		if err := os.WriteFile(testFiles[i], []byte(content), 0o600); err != nil {
			b.Fatalf(shared.TestMsgFailedToCreateTestFile, err)
		}
	}

	processor := fileproc.NewFileProcessor(tmpDir)
	ctx := context.Background()
	fileCount := len(testFiles)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			testFile := testFiles[i%fileCount]
			ch := make(chan fileproc.WriteRequest, 1)
			var wg sync.WaitGroup
			wg.Go(func() {
				defer close(ch)
				_ = processor.ProcessWithContext(ctx, testFile, ch)
			})
			for req := range ch {
				_ = req // Drain channel
			}
			wg.Wait()
			i++
		}
	})
}
