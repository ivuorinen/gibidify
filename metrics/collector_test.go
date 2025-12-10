package metrics

import (
	"errors"
	"fmt"
	"math"
	"sync"
	"testing"
	"time"

	"github.com/ivuorinen/gibidify/shared"
)

func TestNewCollector(t *testing.T) {
	collector := NewCollector()

	if collector == nil {
		t.Fatal("NewCollector returned nil")
	}

	if collector.formatCounts == nil {
		t.Error("formatCounts map not initialized")
	}

	if collector.errorCounts == nil {
		t.Error("errorCounts map not initialized")
	}

	if collector.phaseTimings == nil {
		t.Error("phaseTimings map not initialized")
	}

	maxInt := shared.MetricsMaxInt64
	if collector.smallestFile != maxInt {
		t.Errorf("smallestFile not initialized correctly, got %d, want %d", collector.smallestFile, maxInt)
	}
}

func TestRecordFileProcessedSuccess(t *testing.T) {
	collector := NewCollector()

	result := FileProcessingResult{
		FilePath:       shared.TestPathTestFileGo,
		FileSize:       1024,
		Format:         "go",
		ProcessingTime: 10 * time.Millisecond,
		Success:        true,
	}

	collector.RecordFileProcessed(result)

	metrics := collector.CurrentMetrics()

	if metrics.TotalFiles != 1 {
		t.Errorf(shared.TestFmtExpectedTotalFiles, metrics.TotalFiles)
	}

	if metrics.ProcessedFiles != 1 {
		t.Errorf("Expected ProcessedFiles=1, got %d", metrics.ProcessedFiles)
	}

	if metrics.ProcessedSize != 1024 {
		t.Errorf("Expected ProcessedSize=1024, got %d", metrics.ProcessedSize)
	}

	if metrics.FormatCounts["go"] != 1 {
		t.Errorf("Expected go format count=1, got %d", metrics.FormatCounts["go"])
	}

	if metrics.LargestFile != 1024 {
		t.Errorf("Expected LargestFile=1024, got %d", metrics.LargestFile)
	}

	if metrics.SmallestFile != 1024 {
		t.Errorf("Expected SmallestFile=1024, got %d", metrics.SmallestFile)
	}
}

func TestRecordFileProcessedError(t *testing.T) {
	collector := NewCollector()

	result := FileProcessingResult{
		FilePath:       "/test/error.txt",
		FileSize:       512,
		Format:         "txt",
		ProcessingTime: 5 * time.Millisecond,
		Success:        false,
		Error:          errors.New(shared.TestErrTestErrorMsg),
	}

	collector.RecordFileProcessed(result)

	metrics := collector.CurrentMetrics()

	if metrics.TotalFiles != 1 {
		t.Errorf(shared.TestFmtExpectedTotalFiles, metrics.TotalFiles)
	}

	if metrics.ErrorFiles != 1 {
		t.Errorf("Expected ErrorFiles=1, got %d", metrics.ErrorFiles)
	}

	if metrics.ProcessedFiles != 0 {
		t.Errorf("Expected ProcessedFiles=0, got %d", metrics.ProcessedFiles)
	}

	if metrics.ErrorCounts[shared.TestErrTestErrorMsg] != 1 {
		t.Errorf("Expected error count=1, got %d", metrics.ErrorCounts[shared.TestErrTestErrorMsg])
	}
}

func TestRecordFileProcessedSkipped(t *testing.T) {
	collector := NewCollector()

	result := FileProcessingResult{
		FilePath:   "/test/skipped.bin",
		FileSize:   256,
		Success:    false,
		Skipped:    true,
		SkipReason: "binary file",
	}

	collector.RecordFileProcessed(result)

	metrics := collector.CurrentMetrics()

	if metrics.TotalFiles != 1 {
		t.Errorf(shared.TestFmtExpectedTotalFiles, metrics.TotalFiles)
	}

	if metrics.SkippedFiles != 1 {
		t.Errorf("Expected SkippedFiles=1, got %d", metrics.SkippedFiles)
	}

	if metrics.ProcessedFiles != 0 {
		t.Errorf("Expected ProcessedFiles=0, got %d", metrics.ProcessedFiles)
	}
}

func TestRecordPhaseTime(t *testing.T) {
	collector := NewCollector()

	collector.RecordPhaseTime(shared.MetricsPhaseCollection, 100*time.Millisecond)
	collector.RecordPhaseTime(shared.MetricsPhaseProcessing, 200*time.Millisecond)
	collector.RecordPhaseTime(shared.MetricsPhaseCollection, 50*time.Millisecond) // Add to existing

	metrics := collector.CurrentMetrics()

	if metrics.PhaseTimings[shared.MetricsPhaseCollection] != 150*time.Millisecond {
		t.Errorf("Expected collection phase time=150ms, got %v", metrics.PhaseTimings[shared.MetricsPhaseCollection])
	}

	if metrics.PhaseTimings[shared.MetricsPhaseProcessing] != 200*time.Millisecond {
		t.Errorf("Expected processing phase time=200ms, got %v", metrics.PhaseTimings[shared.MetricsPhaseProcessing])
	}
}

func TestConcurrencyTracking(t *testing.T) {
	collector := NewCollector()

	// Initial concurrency should be 0
	metrics := collector.CurrentMetrics()
	if metrics.CurrentConcurrency != 0 {
		t.Errorf("Expected initial concurrency=0, got %d", metrics.CurrentConcurrency)
	}

	// Increment concurrency
	collector.IncrementConcurrency()
	collector.IncrementConcurrency()

	metrics = collector.CurrentMetrics()
	if metrics.CurrentConcurrency != 2 {
		t.Errorf("Expected concurrency=2, got %d", metrics.CurrentConcurrency)
	}

	// Decrement concurrency
	collector.DecrementConcurrency()

	metrics = collector.CurrentMetrics()
	if metrics.CurrentConcurrency != 1 {
		t.Errorf("Expected concurrency=1, got %d", metrics.CurrentConcurrency)
	}
}

func TestFileSizeTracking(t *testing.T) {
	collector := NewCollector()

	files := []FileProcessingResult{
		{FilePath: "small.txt", FileSize: 100, Success: true, Format: "txt"},
		{FilePath: "large.txt", FileSize: 5000, Success: true, Format: "txt"},
		{FilePath: "medium.txt", FileSize: 1000, Success: true, Format: "txt"},
	}

	for _, file := range files {
		collector.RecordFileProcessed(file)
	}

	metrics := collector.CurrentMetrics()

	if metrics.LargestFile != 5000 {
		t.Errorf("Expected LargestFile=5000, got %d", metrics.LargestFile)
	}

	if metrics.SmallestFile != 100 {
		t.Errorf("Expected SmallestFile=100, got %d", metrics.SmallestFile)
	}

	expectedAvg := float64(6100) / 3 // (100 + 5000 + 1000) / 3
	if math.Abs(metrics.AverageFileSize-expectedAvg) > 0.1 {
		t.Errorf("Expected AverageFileSize=%.1f, got %.1f", expectedAvg, metrics.AverageFileSize)
	}
}

func TestConcurrentAccess(t *testing.T) {
	collector := NewCollector()

	// Test concurrent file processing
	var wg sync.WaitGroup
	numGoroutines := 10
	filesPerGoroutine := 100

	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < filesPerGoroutine; j++ {
				result := FileProcessingResult{
					FilePath: fmt.Sprintf("/test/file_%d_%d.go", id, j),
					FileSize: int64(j + 1),
					Success:  true,
					Format:   "go",
				}
				collector.RecordFileProcessed(result)
			}
		}(i)
	}

	wg.Wait()

	metrics := collector.CurrentMetrics()
	expectedTotal := int64(numGoroutines * filesPerGoroutine)

	if metrics.TotalFiles != expectedTotal {
		t.Errorf("Expected TotalFiles=%d, got %d", expectedTotal, metrics.TotalFiles)
	}

	if metrics.ProcessedFiles != expectedTotal {
		t.Errorf("Expected ProcessedFiles=%d, got %d", expectedTotal, metrics.ProcessedFiles)
	}

	if metrics.FormatCounts["go"] != expectedTotal {
		t.Errorf("Expected go format count=%d, got %d", expectedTotal, metrics.FormatCounts["go"])
	}
}

func TestFinishAndGetFinalMetrics(t *testing.T) {
	collector := NewCollector()

	// Process some files
	result := FileProcessingResult{
		FilePath: shared.TestPathTestFileGo,
		FileSize: 1024,
		Success:  true,
		Format:   "go",
	}
	collector.RecordFileProcessed(result)

	collector.Finish()

	finalMetrics := collector.FinalMetrics()

	if finalMetrics.EndTime.IsZero() {
		t.Error("EndTime should be set after Finish()")
	}

	if finalMetrics.ProcessingTime < 0 {
		t.Error("ProcessingTime should be >= 0 after Finish()")
	}

	if finalMetrics.ProcessedFiles != 1 {
		t.Errorf("Expected ProcessedFiles=1, got %d", finalMetrics.ProcessedFiles)
	}
}

func TestGenerateReport(t *testing.T) {
	collector := NewCollector()

	// Add some test data
	files := []FileProcessingResult{
		{FilePath: "file1.go", FileSize: 1000, Success: true, Format: "go"},
		{FilePath: "file2.js", FileSize: 2000, Success: true, Format: "js"},
		{FilePath: "file3.go", FileSize: 500, Success: false, Error: errors.New("syntax error")},
	}

	for _, file := range files {
		collector.RecordFileProcessed(file)
	}

	collector.RecordPhaseTime(shared.MetricsPhaseCollection, 100*time.Millisecond)
	collector.RecordPhaseTime(shared.MetricsPhaseProcessing, 200*time.Millisecond)

	// Call Finish to mirror production usage where GenerateReport is called after processing completes
	collector.Finish()

	report := collector.GenerateReport()

	if report.Summary.TotalFiles != 3 {
		t.Errorf("Expected Summary.TotalFiles=3, got %d", report.Summary.TotalFiles)
	}

	if report.FormatBreakdown["go"].Count != 1 {
		t.Errorf("Expected go format count=1, got %d", report.FormatBreakdown["go"].Count)
	}

	if report.FormatBreakdown["js"].Count != 1 {
		t.Errorf("Expected js format count=1, got %d", report.FormatBreakdown["js"].Count)
	}

	if len(report.ErrorBreakdown) != 1 {
		t.Errorf("Expected 1 error type, got %d", len(report.ErrorBreakdown))
	}

	if len(report.PhaseBreakdown) != 2 {
		t.Errorf("Expected 2 phases, got %d", len(report.PhaseBreakdown))
	}

	if len(report.Recommendations) == 0 {
		t.Error("Expected some recommendations")
	}
}

func TestReset(t *testing.T) {
	collector := NewCollector()

	// Add some data
	result := FileProcessingResult{
		FilePath: shared.TestPathTestFileGo,
		FileSize: 1024,
		Success:  true,
		Format:   "go",
	}
	collector.RecordFileProcessed(result)
	collector.RecordPhaseTime(shared.MetricsPhaseCollection, 100*time.Millisecond)

	// Verify data exists
	metrics := collector.CurrentMetrics()
	if metrics.TotalFiles == 0 {
		t.Error("Expected data before reset")
	}

	// Reset
	collector.Reset()

	// Verify reset
	metrics = collector.CurrentMetrics()
	if metrics.TotalFiles != 0 {
		t.Errorf("Expected TotalFiles=0 after reset, got %d", metrics.TotalFiles)
	}

	if metrics.ProcessedFiles != 0 {
		t.Errorf("Expected ProcessedFiles=0 after reset, got %d", metrics.ProcessedFiles)
	}

	if len(metrics.FormatCounts) != 0 {
		t.Errorf("Expected empty FormatCounts after reset, got %d entries", len(metrics.FormatCounts))
	}

	if len(metrics.PhaseTimings) != 0 {
		t.Errorf("Expected empty PhaseTimings after reset, got %d entries", len(metrics.PhaseTimings))
	}
}

// Benchmarks for collector hot paths

func BenchmarkCollectorRecordFileProcessed(b *testing.B) {
	collector := NewCollector()
	result := FileProcessingResult{
		FilePath:       shared.TestPathTestFileGo,
		FileSize:       1024,
		Format:         "go",
		ProcessingTime: 10 * time.Millisecond,
		Success:        true,
	}

	for b.Loop() {
		collector.RecordFileProcessed(result)
	}
}

func BenchmarkCollectorRecordFileProcessedConcurrent(b *testing.B) {
	collector := NewCollector()
	result := FileProcessingResult{
		FilePath:       shared.TestPathTestFileGo,
		FileSize:       1024,
		Format:         "go",
		ProcessingTime: 10 * time.Millisecond,
		Success:        true,
	}

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			collector.RecordFileProcessed(result)
		}
	})
}

func BenchmarkCollectorCurrentMetrics(b *testing.B) {
	collector := NewCollector()

	// Add some baseline data
	for i := 0; i < 100; i++ {
		result := FileProcessingResult{
			FilePath: fmt.Sprintf("/test/file%d.go", i),
			FileSize: int64(i * 100),
			Format:   "go",
			Success:  true,
		}
		collector.RecordFileProcessed(result)
	}

	b.ResetTimer()
	for b.Loop() {
		_ = collector.CurrentMetrics()
	}
}

func BenchmarkCollectorGenerateReport(b *testing.B) {
	benchmarks := []struct {
		name  string
		files int
	}{
		{"10files", 10},
		{"100files", 100},
		{"1000files", 1000},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			collector := NewCollector()

			// Add test data
			formats := []string{"go", "js", "py", "ts", "rs", "java", "cpp", "rb"}
			for i := 0; i < bm.files; i++ {
				var result FileProcessingResult
				if i%10 == 0 {
					result = FileProcessingResult{
						FilePath: fmt.Sprintf("/test/error%d.go", i),
						FileSize: 500,
						Success:  false,
						Error:    errors.New(shared.TestErrTestErrorMsg),
					}
				} else {
					result = FileProcessingResult{
						FilePath: fmt.Sprintf("/test/file%d.go", i),
						FileSize: int64(i * 100),
						Format:   formats[i%len(formats)],
						Success:  true,
					}
				}
				collector.RecordFileProcessed(result)
			}

			collector.RecordPhaseTime(shared.MetricsPhaseCollection, 50*time.Millisecond)
			collector.RecordPhaseTime(shared.MetricsPhaseProcessing, 150*time.Millisecond)
			collector.Finish()

			b.ResetTimer()
			for b.Loop() {
				_ = collector.GenerateReport()
			}
		})
	}
}

func BenchmarkCollectorConcurrencyTracking(b *testing.B) {
	collector := NewCollector()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			collector.IncrementConcurrency()
			collector.DecrementConcurrency()
		}
	})
}
