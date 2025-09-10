package metrics

import (
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"
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

	if collector.smallestFile != maxInt64 {
		t.Errorf("smallestFile not initialized correctly, got %d, want %d", collector.smallestFile, maxInt64)
	}
}

func TestRecordFileProcessed_Success(t *testing.T) {
	collector := NewCollector()

	result := FileProcessingResult{
		FilePath:       "/test/file.go",
		FileSize:       1024,
		Format:         "go",
		ProcessingTime: 10 * time.Millisecond,
		Success:        true,
	}

	collector.RecordFileProcessed(result)

	metrics := collector.GetCurrentMetrics()

	if metrics.TotalFiles != 1 {
		t.Errorf("Expected TotalFiles=1, got %d", metrics.TotalFiles)
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

func TestRecordFileProcessed_Error(t *testing.T) {
	collector := NewCollector()

	result := FileProcessingResult{
		FilePath:       "/test/error.txt",
		FileSize:       512,
		Format:         "txt",
		ProcessingTime: 5 * time.Millisecond,
		Success:        false,
		Error:          errors.New("test error"),
	}

	collector.RecordFileProcessed(result)

	metrics := collector.GetCurrentMetrics()

	if metrics.TotalFiles != 1 {
		t.Errorf("Expected TotalFiles=1, got %d", metrics.TotalFiles)
	}

	if metrics.ErrorFiles != 1 {
		t.Errorf("Expected ErrorFiles=1, got %d", metrics.ErrorFiles)
	}

	if metrics.ProcessedFiles != 0 {
		t.Errorf("Expected ProcessedFiles=0, got %d", metrics.ProcessedFiles)
	}

	if metrics.ErrorCounts["test error"] != 1 {
		t.Errorf("Expected error count=1, got %d", metrics.ErrorCounts["test error"])
	}
}

func TestRecordFileProcessed_Skipped(t *testing.T) {
	collector := NewCollector()

	result := FileProcessingResult{
		FilePath:   "/test/skipped.bin",
		FileSize:   256,
		Success:    false,
		Skipped:    true,
		SkipReason: "binary file",
	}

	collector.RecordFileProcessed(result)

	metrics := collector.GetCurrentMetrics()

	if metrics.TotalFiles != 1 {
		t.Errorf("Expected TotalFiles=1, got %d", metrics.TotalFiles)
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

	collector.RecordPhaseTime(PhaseCollection, 100*time.Millisecond)
	collector.RecordPhaseTime(PhaseProcessing, 200*time.Millisecond)
	collector.RecordPhaseTime(PhaseCollection, 50*time.Millisecond) // Add to existing

	metrics := collector.GetCurrentMetrics()

	if metrics.PhaseTimings[PhaseCollection] != 150*time.Millisecond {
		t.Errorf("Expected collection phase time=150ms, got %v", metrics.PhaseTimings[PhaseCollection])
	}

	if metrics.PhaseTimings[PhaseProcessing] != 200*time.Millisecond {
		t.Errorf("Expected processing phase time=200ms, got %v", metrics.PhaseTimings[PhaseProcessing])
	}
}

func TestConcurrencyTracking(t *testing.T) {
	collector := NewCollector()

	// Initial concurrency should be 0
	metrics := collector.GetCurrentMetrics()
	if metrics.CurrentConcurrency != 0 {
		t.Errorf("Expected initial concurrency=0, got %d", metrics.CurrentConcurrency)
	}

	// Increment concurrency
	collector.IncrementConcurrency()
	collector.IncrementConcurrency()

	metrics = collector.GetCurrentMetrics()
	if metrics.CurrentConcurrency != 2 {
		t.Errorf("Expected concurrency=2, got %d", metrics.CurrentConcurrency)
	}

	// Decrement concurrency
	collector.DecrementConcurrency()

	metrics = collector.GetCurrentMetrics()
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

	metrics := collector.GetCurrentMetrics()

	if metrics.LargestFile != 5000 {
		t.Errorf("Expected LargestFile=5000, got %d", metrics.LargestFile)
	}

	if metrics.SmallestFile != 100 {
		t.Errorf("Expected SmallestFile=100, got %d", metrics.SmallestFile)
	}

	expectedAvg := float64(6100) / 3 // (100 + 5000 + 1000) / 3
	if abs(metrics.AverageFileSize-expectedAvg) > 0.1 {
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

	metrics := collector.GetCurrentMetrics()
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
		FilePath: "/test/file.go",
		FileSize: 1024,
		Success:  true,
		Format:   "go",
	}
	collector.RecordFileProcessed(result)

	// Wait a bit to ensure processing time > 0
	time.Sleep(10 * time.Millisecond)

	collector.Finish()

	finalMetrics := collector.GetFinalMetrics()

	if finalMetrics.EndTime.IsZero() {
		t.Error("EndTime should be set after Finish()")
	}

	if finalMetrics.ProcessingTime <= 0 {
		t.Error("ProcessingTime should be > 0 after Finish()")
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

	collector.RecordPhaseTime(PhaseCollection, 100*time.Millisecond)
	collector.RecordPhaseTime(PhaseProcessing, 200*time.Millisecond)

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
		FilePath: "/test/file.go",
		FileSize: 1024,
		Success:  true,
		Format:   "go",
	}
	collector.RecordFileProcessed(result)
	collector.RecordPhaseTime(PhaseCollection, 100*time.Millisecond)

	// Verify data exists
	metrics := collector.GetCurrentMetrics()
	if metrics.TotalFiles == 0 {
		t.Error("Expected data before reset")
	}

	// Reset
	collector.Reset()

	// Verify reset
	metrics = collector.GetCurrentMetrics()
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

// Helper function for floating point comparison.
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}

	return x
}
