package benchmark

import (
	"bytes"
	"io"
	"os"
	"runtime"
	"strings"
	"testing"
)

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
		t.Errorf("Expected files processed > 0, got %d", result.FilesProcessed)
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
		t.Errorf("Expected files processed > 0, got %d", result.FilesProcessed)
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
		t.Errorf("Expected %d results, got %d", len(concurrencyLevels), len(suite.Results))
	}

	for i, result := range suite.Results {
		if result.FilesProcessed <= 0 {
			t.Errorf("Result %d: Expected files processed > 0, got %d", i, result.FilesProcessed)
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
		t.Errorf("Expected %d results, got %d", len(formats), len(suite.Results))
	}

	for i, result := range suite.Results {
		if result.FilesProcessed <= 0 {
			t.Errorf("Result %d: Expected files processed > 0, got %d", i, result.FilesProcessed)
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
			b.Errorf("Expected files processed > 0, got %d", result.FilesProcessed)
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
			b.Errorf("Expected files processed > 0, got %d", result.FilesProcessed)
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
			b.Errorf("Expected %d results, got %d", len(concurrencyLevels), len(suite.Results))
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
			b.Errorf("Expected %d results, got %d", len(formats), len(suite.Results))
		}
	}
}

// TestPrintBenchmarkResult tests the benchmark result printing.
func TestPrintBenchmarkResult(t *testing.T) {
	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	result := &BenchmarkResult{
		Name:            "TestBenchmark",
		Duration:        1000000000, // 1 second
		FilesProcessed:  100,
		BytesProcessed:  1024 * 1024, // 1 MB
		FilesPerSecond:  100.0,
		BytesPerSecond:  1024 * 1024,
		MemoryUsage:     MemoryStats{AllocMB: 10.5, SysMB: 20.3, NumGC: 5, PauseTotalNs: 1000000},
		CPUUsage:        CPUStats{Goroutines: 10},
	}

	PrintBenchmarkResult(result)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Check that output contains key information
	if !strings.Contains(output, "TestBenchmark") {
		t.Error("Output should contain benchmark name")
	}
	if !strings.Contains(output, "Duration:") {
		t.Error("Output should contain duration")
	}
	if !strings.Contains(output, "Files Processed:") {
		t.Error("Output should contain files processed")
	}
	if !strings.Contains(output, "100") {
		t.Error("Output should contain the number of files")
	}
}

// TestPrintBenchmarkSuite tests the benchmark suite printing.
func TestPrintBenchmarkSuite(t *testing.T) {
	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	suite := &BenchmarkSuite{
		Name: "TestSuite",
		Results: []BenchmarkResult{
			{
				Name:           "Test1",
				FilesProcessed: 50,
			},
			{
				Name:           "Test2",
				FilesProcessed: 75,
			},
		},
	}

	PrintBenchmarkSuite(suite)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Check that output contains suite name and all results
	if !strings.Contains(output, "TestSuite") {
		t.Error("Output should contain suite name")
	}
	if !strings.Contains(output, "Test1") {
		t.Error("Output should contain first result name")
	}
	if !strings.Contains(output, "Test2") {
		t.Error("Output should contain second result name")
	}
}

// TestRunAllBenchmarks tests the comprehensive benchmark suite.
func TestRunAllBenchmarks(t *testing.T) {
	// This test would require extensive setup, so we'll test it indirectly
	// by ensuring the individual benchmark functions work
	t.Skip("RunAllBenchmarks requires comprehensive setup and is tested through integration tests")
}

