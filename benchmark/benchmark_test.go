package benchmark

import (
	"runtime"
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
