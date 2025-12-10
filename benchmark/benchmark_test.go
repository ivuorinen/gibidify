package benchmark

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/ivuorinen/gibidify/shared"
)

// capturedOutput captures stdout output from a function call.
func capturedOutput(t *testing.T, fn func()) string {
	t.Helper()
	original := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf(shared.TestMsgFailedToCreatePipe, err)
	}
	defer r.Close()
	defer func() { os.Stdout = original }()
	os.Stdout = w

	fn()

	if err := w.Close(); err != nil {
		t.Logf(shared.TestMsgFailedToClose, err)
	}

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf(shared.TestMsgFailedToReadOutput, err)
	}

	return buf.String()
}

// verifyOutputContains checks if output contains all expected strings.
func verifyOutputContains(t *testing.T, testName, output string, expected []string) {
	t.Helper()
	for _, check := range expected {
		if !strings.Contains(output, check) {
			t.Errorf("Test %s: output missing expected content: %q\nFull output:\n%s", testName, check, output)
		}
	}
}

// TestFileCollectionBenchmark tests the file collection benchmark.
func TestFileCollectionBenchmark(t *testing.T) {
	result, err := FileCollectionBenchmark("", 10)
	if err != nil {
		t.Fatalf("FileCollectionBenchmark failed: %v", err)
	}

	if result.Name != "FileCollection" {
		t.Errorf("Expected name 'FileCollection', got %s", result.Name)
	}

	// Debug information
	t.Logf("Files processed: %d", result.FilesProcessed)
	t.Logf("Duration: %v", result.Duration)
	t.Logf("Bytes processed: %d", result.BytesProcessed)

	if result.FilesProcessed <= 0 {
		t.Errorf(shared.TestFmtExpectedFilesProcessed, result.FilesProcessed)
	}

	if result.Duration <= 0 {
		t.Errorf("Expected duration > 0, got %v", result.Duration)
	}
}

// TestFileProcessingBenchmark tests the file processing benchmark.
func TestFileProcessingBenchmark(t *testing.T) {
	result, err := FileProcessingBenchmark("", "json", 2)
	if err != nil {
		t.Fatalf("FileProcessingBenchmark failed: %v", err)
	}

	if result.FilesProcessed <= 0 {
		t.Errorf(shared.TestFmtExpectedFilesProcessed, result.FilesProcessed)
	}

	if result.Duration <= 0 {
		t.Errorf("Expected duration > 0, got %v", result.Duration)
	}
}

// TestConcurrencyBenchmark tests the concurrency benchmark.
func TestConcurrencyBenchmark(t *testing.T) {
	concurrencyLevels := []int{1, 2}
	suite, err := ConcurrencyBenchmark("", "json", concurrencyLevels)
	if err != nil {
		t.Fatalf("ConcurrencyBenchmark failed: %v", err)
	}

	if suite.Name != "ConcurrencyBenchmark" {
		t.Errorf("Expected name 'ConcurrencyBenchmark', got %s", suite.Name)
	}

	if len(suite.Results) != len(concurrencyLevels) {
		t.Errorf(shared.TestFmtExpectedResults, len(concurrencyLevels), len(suite.Results))
	}

	for i, result := range suite.Results {
		if result.FilesProcessed <= 0 {
			t.Errorf("Result %d: "+shared.TestFmtExpectedFilesProcessed, i, result.FilesProcessed)
		}
	}
}

// TestFormatBenchmark tests the format benchmark.
func TestFormatBenchmark(t *testing.T) {
	formats := []string{"json", "yaml"}
	suite, err := FormatBenchmark("", formats)
	if err != nil {
		t.Fatalf("FormatBenchmark failed: %v", err)
	}

	if suite.Name != "FormatBenchmark" {
		t.Errorf("Expected name 'FormatBenchmark', got %s", suite.Name)
	}

	if len(suite.Results) != len(formats) {
		t.Errorf(shared.TestFmtExpectedResults, len(formats), len(suite.Results))
	}

	for i, result := range suite.Results {
		if result.FilesProcessed <= 0 {
			t.Errorf("Result %d: "+shared.TestFmtExpectedFilesProcessed, i, result.FilesProcessed)
		}
	}
}

// TestCreateBenchmarkFiles tests the benchmark file creation.
func TestCreateBenchmarkFiles(t *testing.T) {
	tempDir, cleanup, err := createBenchmarkFiles(5)
	if err != nil {
		t.Fatalf("createBenchmarkFiles failed: %v", err)
	}
	defer cleanup()

	if tempDir == "" {
		t.Error("Expected non-empty temp directory")
	}

	// Verify files were created
	// This is tested indirectly through the benchmark functions
}

// BenchmarkFileCollection benchmarks the file collection process.
func BenchmarkFileCollection(b *testing.B) {
	for i := 0; i < b.N; i++ {
		result, err := FileCollectionBenchmark("", 50)
		if err != nil {
			b.Fatalf("FileCollectionBenchmark failed: %v", err)
		}
		if result.FilesProcessed <= 0 {
			b.Errorf(shared.TestFmtExpectedFilesProcessed, result.FilesProcessed)
		}
	}
}

// BenchmarkFileProcessing benchmarks the file processing pipeline.
func BenchmarkFileProcessing(b *testing.B) {
	for i := 0; i < b.N; i++ {
		result, err := FileProcessingBenchmark("", "json", runtime.NumCPU())
		if err != nil {
			b.Fatalf("FileProcessingBenchmark failed: %v", err)
		}
		if result.FilesProcessed <= 0 {
			b.Errorf(shared.TestFmtExpectedFilesProcessed, result.FilesProcessed)
		}
	}
}

// BenchmarkConcurrency benchmarks different concurrency levels.
func BenchmarkConcurrency(b *testing.B) {
	concurrencyLevels := []int{1, 2, 4}

	for i := 0; i < b.N; i++ {
		suite, err := ConcurrencyBenchmark("", "json", concurrencyLevels)
		if err != nil {
			b.Fatalf("ConcurrencyBenchmark failed: %v", err)
		}
		if len(suite.Results) != len(concurrencyLevels) {
			b.Errorf(shared.TestFmtExpectedResults, len(concurrencyLevels), len(suite.Results))
		}
	}
}

// BenchmarkFormats benchmarks different output formats.
func BenchmarkFormats(b *testing.B) {
	formats := []string{"json", "yaml", "markdown"}

	for i := 0; i < b.N; i++ {
		suite, err := FormatBenchmark("", formats)
		if err != nil {
			b.Fatalf("FormatBenchmark failed: %v", err)
		}
		if len(suite.Results) != len(formats) {
			b.Errorf(shared.TestFmtExpectedResults, len(formats), len(suite.Results))
		}
	}
}

// TestPrintResult tests the PrintResult function.
func TestPrintResult(t *testing.T) {
	// Create a test result
	result := &Result{
		Name:           "Test Benchmark",
		Duration:       1 * time.Second,
		FilesProcessed: 100,
		BytesProcessed: 2048000, // ~2MB for easy calculation
	}

	// Capture stdout
	original := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf(shared.TestMsgFailedToCreatePipe, err)
	}
	defer r.Close()
	defer func() { os.Stdout = original }()
	os.Stdout = w

	// Call PrintResult
	PrintResult(result)

	// Close writer and read captured output
	if err := w.Close(); err != nil {
		t.Logf(shared.TestMsgFailedToClose, err)
	}

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf(shared.TestMsgFailedToReadOutput, err)
	}
	output := buf.String()

	// Verify expected content
	expectedContents := []string{
		"=== Test Benchmark ===",
		"Duration: 1s",
		"Files Processed: 100",
		"Bytes Processed: 2048000",
		"1.95 MB", // 2048000 / 1024 / 1024 ≈ 1.95
	}

	for _, expected := range expectedContents {
		if !strings.Contains(output, expected) {
			t.Errorf("PrintResult output missing expected content: %q\nFull output:\n%s", expected, output)
		}
	}
}

// TestPrintSuite tests the PrintSuite function.
func TestPrintSuite(t *testing.T) {
	// Create a test suite with multiple results
	suite := &Suite{
		Name: "Test Suite",
		Results: []Result{
			{
				Name:           "Benchmark 1",
				Duration:       500 * time.Millisecond,
				FilesProcessed: 50,
				BytesProcessed: 1024000, // 1MB
			},
			{
				Name:           "Benchmark 2",
				Duration:       750 * time.Millisecond,
				FilesProcessed: 75,
				BytesProcessed: 1536000, // 1.5MB
			},
		},
	}

	// Capture stdout
	original := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf(shared.TestMsgFailedToCreatePipe, err)
	}
	defer r.Close()
	defer func() { os.Stdout = original }()
	os.Stdout = w

	// Call PrintSuite
	PrintSuite(suite)

	// Close writer and read captured output
	if err := w.Close(); err != nil {
		t.Logf(shared.TestMsgFailedToClose, err)
	}

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf(shared.TestMsgFailedToReadOutput, err)
	}
	output := buf.String()

	// Verify expected content
	expectedContents := []string{
		"=== Test Suite ===",
		"=== Benchmark 1 ===",
		"Duration: 500ms",
		"Files Processed: 50",
		"=== Benchmark 2 ===",
		"Duration: 750ms",
		"Files Processed: 75",
	}

	for _, expected := range expectedContents {
		if !strings.Contains(output, expected) {
			t.Errorf("PrintSuite output missing expected content: %q\nFull output:\n%s", expected, output)
		}
	}

	// Verify both results are printed
	benchmark1Count := strings.Count(output, "=== Benchmark 1 ===")
	benchmark2Count := strings.Count(output, "=== Benchmark 2 ===")

	if benchmark1Count != 1 {
		t.Errorf("Expected exactly 1 occurrence of 'Benchmark 1', got %d", benchmark1Count)
	}
	if benchmark2Count != 1 {
		t.Errorf("Expected exactly 1 occurrence of 'Benchmark 2', got %d", benchmark2Count)
	}
}

// TestPrintResult_EdgeCases tests edge cases for PrintResult.
func TestPrintResultEdgeCases(t *testing.T) {
	tests := []struct {
		name   string
		result *Result
		checks []string
	}{
		{
			name: "zero values",
			result: &Result{
				Name:           "Zero Benchmark",
				Duration:       0,
				FilesProcessed: 0,
				BytesProcessed: 0,
			},
			checks: []string{
				"=== Zero Benchmark ===",
				"Duration: 0s",
				"Files Processed: 0",
				"Bytes Processed: 0",
				"0.00 MB",
			},
		},
		{
			name: "large values",
			result: &Result{
				Name:           "Large Benchmark",
				Duration:       1 * time.Hour,
				FilesProcessed: 1000000,
				BytesProcessed: 1073741824, // 1GB
			},
			checks: []string{
				"=== Large Benchmark ===",
				"Duration: 1h0m0s",
				"Files Processed: 1000000",
				"Bytes Processed: 1073741824",
				"1024.00 MB",
			},
		},
		{
			name: "empty name",
			result: &Result{
				Name:           "",
				Duration:       100 * time.Millisecond,
				FilesProcessed: 10,
				BytesProcessed: 1024,
			},
			checks: []string{
				"===  ===", // Empty name between === markers
				"Duration: 100ms",
				"Files Processed: 10",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.result
			output := capturedOutput(t, func() { PrintResult(result) })
			verifyOutputContains(t, tt.name, output, tt.checks)
		})
	}
}

// TestPrintSuite_EdgeCases tests edge cases for PrintSuite.
func TestPrintSuiteEdgeCases(t *testing.T) {
	tests := []struct {
		name   string
		suite  *Suite
		checks []string
	}{
		{
			name: "empty suite",
			suite: &Suite{
				Name:    "Empty Suite",
				Results: []Result{},
			},
			checks: []string{
				"=== Empty Suite ===",
			},
		},
		{
			name: "suite with empty name",
			suite: &Suite{
				Name: "",
				Results: []Result{
					{
						Name:           "Single Benchmark",
						Duration:       200 * time.Millisecond,
						FilesProcessed: 20,
						BytesProcessed: 2048,
					},
				},
			},
			checks: []string{
				"===  ===", // Empty name
				"=== Single Benchmark ===",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			suite := tt.suite
			output := capturedOutput(t, func() { PrintSuite(suite) })
			verifyOutputContains(t, tt.name, output, tt.checks)
		})
	}
}

// TestRunAllBenchmarks tests the RunAllBenchmarks function.
func TestRunAllBenchmarks(t *testing.T) {
	// Create a temporary directory with some test files
	srcDir := t.TempDir()

	// Create a few test files
	testFiles := []struct {
		name    string
		content string
	}{
		{shared.TestFileMainGo, "package main\nfunc main() {}"},
		{shared.TestFile2Name, "Hello World"},
		{shared.TestFile3Name, "# Test Markdown"},
	}

	for _, file := range testFiles {
		filePath := filepath.Join(srcDir, file.name)
		err := os.WriteFile(filePath, []byte(file.content), 0o644)
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", file.name, err)
		}
	}

	// Capture stdout to verify output
	original := os.Stdout
	r, w, pipeErr := os.Pipe()
	if pipeErr != nil {
		t.Fatalf(shared.TestMsgFailedToCreatePipe, pipeErr)
	}
	defer func() { os.Stdout = original }()
	os.Stdout = w

	// Call RunAllBenchmarks
	err := RunAllBenchmarks(srcDir)

	// Close writer and read captured output
	if closeErr := w.Close(); closeErr != nil {
		t.Logf(shared.TestMsgFailedToClose, closeErr)
	}

	var buf bytes.Buffer
	if _, copyErr := io.Copy(&buf, r); copyErr != nil {
		t.Fatalf(shared.TestMsgFailedToReadOutput, copyErr)
	}
	output := buf.String()

	// Check for error
	if err != nil {
		t.Errorf("RunAllBenchmarks failed: %v", err)
	}

	// Verify expected output content
	expectedContents := []string{
		"Running gibidify benchmark suite...",
		"Running file collection benchmark...",
		"Running format benchmarks...",
		"Running concurrency benchmarks...",
	}

	for _, expected := range expectedContents {
		if !strings.Contains(output, expected) {
			t.Errorf("RunAllBenchmarks output missing expected content: %q\nFull output:\n%s", expected, output)
		}
	}

	// The function should not panic and should complete successfully
	t.Log("RunAllBenchmarks completed successfully with output captured")
}
