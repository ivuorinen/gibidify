package metrics

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestNewReporter(t *testing.T) {
	collector := NewCollector()
	reporter := NewReporter(collector, true, true)

	if reporter == nil {
		t.Fatal("NewReporter returned nil")
	}

	if reporter.collector != collector {
		t.Error("Reporter collector not set correctly")
	}

	if !reporter.verbose {
		t.Error("Verbose flag not set correctly")
	}

	if !reporter.colors {
		t.Error("Colors flag not set correctly")
	}
}

func TestReportProgress_Basic(t *testing.T) {
	collector := NewCollector()
	reporter := NewReporter(collector, false, false)

	// Add some test data
	result := FileProcessingResult{
		FilePath: "/test/file.go",
		FileSize: 1024,
		Success:  true,
		Format:   "go",
	}
	collector.RecordFileProcessed(result)

	// Wait to ensure FilesPerSecond calculation
	time.Sleep(10 * time.Millisecond)

	progress := reporter.ReportProgress()

	if !strings.Contains(progress, "Processed: 1 files") {
		t.Errorf("Expected progress to contain processed files count, got: %s", progress)
	}

	if !strings.Contains(progress, "files/sec") {
		t.Errorf("Expected progress to contain files/sec, got: %s", progress)
	}
}

func TestReportProgress_WithErrors(t *testing.T) {
	collector := NewCollector()
	reporter := NewReporter(collector, false, false)

	// Add successful file
	successResult := FileProcessingResult{
		FilePath: "/test/success.go",
		FileSize: 1024,
		Success:  true,
		Format:   "go",
	}
	collector.RecordFileProcessed(successResult)

	// Add error file
	errorResult := FileProcessingResult{
		FilePath: "/test/error.go",
		FileSize: 512,
		Success:  false,
		Error:    errors.New("syntax error"),
	}
	collector.RecordFileProcessed(errorResult)

	progress := reporter.ReportProgress()

	if !strings.Contains(progress, "Processed: 1 files") {
		t.Errorf("Expected progress to contain processed files count, got: %s", progress)
	}

	if !strings.Contains(progress, "Errors: 1") {
		t.Errorf("Expected progress to contain error count, got: %s", progress)
	}
}

func TestReportProgress_WithSkipped(t *testing.T) {
	collector := NewCollector()
	reporter := NewReporter(collector, false, false)

	// Add successful file
	successResult := FileProcessingResult{
		FilePath: "/test/success.go",
		FileSize: 1024,
		Success:  true,
		Format:   "go",
	}
	collector.RecordFileProcessed(successResult)

	// Add skipped file
	skippedResult := FileProcessingResult{
		FilePath:   "/test/binary.exe",
		FileSize:   2048,
		Success:    false,
		Skipped:    true,
		SkipReason: "binary file",
	}
	collector.RecordFileProcessed(skippedResult)

	progress := reporter.ReportProgress()

	if !strings.Contains(progress, "Skipped: 1") {
		t.Errorf("Expected progress to contain skipped count, got: %s", progress)
	}
}

func TestReportProgress_Verbose(t *testing.T) {
	collector := NewCollector()
	reporter := NewReporter(collector, true, false)

	// Add test data
	files := []FileProcessingResult{
		{FilePath: "/test/file1.go", FileSize: 1000, Success: true, Format: "go"},
		{FilePath: "/test/file2.js", FileSize: 2000, Success: true, Format: "js"},
		{FilePath: "/test/file3.py", FileSize: 1500, Success: true, Format: "py"},
	}

	for _, file := range files {
		collector.RecordFileProcessed(file)
	}

	collector.RecordPhaseTime(PhaseCollection, 50*time.Millisecond)
	collector.RecordPhaseTime(PhaseProcessing, 100*time.Millisecond)

	progress := reporter.ReportProgress()

	// Check for verbose content
	if !strings.Contains(progress, "=== Processing Statistics ===") {
		t.Error("Expected verbose header not found")
	}

	if !strings.Contains(progress, "Format Breakdown:") {
		t.Error("Expected format breakdown not found")
	}

	if !strings.Contains(progress, "go: 1 files") {
		t.Error("Expected go format count not found")
	}

	if !strings.Contains(progress, "Memory - Current:") {
		t.Error("Expected memory information not found")
	}

	if !strings.Contains(progress, "Concurrency - Current:") {
		t.Error("Expected concurrency information not found")
	}
}

func TestReportFinal_Basic(t *testing.T) {
	collector := NewCollector()
	reporter := NewReporter(collector, false, false)

	// Add test data
	files := []FileProcessingResult{
		{FilePath: "/test/file1.go", FileSize: 1000, Success: true, Format: "go"},
		{FilePath: "/test/file2.js", FileSize: 2000, Success: true, Format: "js"},
		{FilePath: "/test/error.py", FileSize: 500, Success: false, Error: errors.New("syntax error")},
	}

	for _, file := range files {
		collector.RecordFileProcessed(file)
	}

	collector.Finish()
	final := reporter.ReportFinal()

	if !strings.Contains(final, "=== Processing Complete ===") {
		t.Error("Expected completion header not found")
	}

	if !strings.Contains(final, "Total Files: 3") {
		t.Error("Expected total files count not found")
	}

	if !strings.Contains(final, "Processed: 2") {
		t.Error("Expected processed files count not found")
	}

	if !strings.Contains(final, "Errors: 1") {
		t.Error("Expected error count not found")
	}
}

func TestReportFinal_Verbose(t *testing.T) {
	collector := NewCollector()
	reporter := NewReporter(collector, true, false)

	// Add comprehensive test data
	files := []FileProcessingResult{
		{FilePath: "/test/file1.go", FileSize: 1000, Success: true, Format: "go"},
		{FilePath: "/test/file2.go", FileSize: 2000, Success: true, Format: "go"},
		{FilePath: "/test/file3.js", FileSize: 1500, Success: true, Format: "js"},
		{FilePath: "/test/error.py", FileSize: 500, Success: false, Error: errors.New("syntax error")},
		{FilePath: "/test/skip.bin", FileSize: 3000, Success: false, Skipped: true, SkipReason: "binary"},
	}

	for _, file := range files {
		collector.RecordFileProcessed(file)
	}

	collector.RecordPhaseTime(PhaseCollection, 50*time.Millisecond)
	collector.RecordPhaseTime(PhaseProcessing, 150*time.Millisecond)
	collector.RecordPhaseTime(PhaseWriting, 25*time.Millisecond)

	collector.Finish()
	final := reporter.ReportFinal()

	// Check comprehensive report sections
	if !strings.Contains(final, "=== Comprehensive Processing Report ===") {
		t.Error("Expected comprehensive header not found")
	}

	if !strings.Contains(final, "SUMMARY:") {
		t.Error("Expected summary section not found")
	}

	if !strings.Contains(final, "FORMAT BREAKDOWN:") {
		t.Error("Expected format breakdown section not found")
	}

	if !strings.Contains(final, "PHASE BREAKDOWN:") {
		t.Error("Expected phase breakdown section not found")
	}

	if !strings.Contains(final, "ERROR BREAKDOWN:") {
		t.Error("Expected error breakdown section not found")
	}

	if !strings.Contains(final, "RESOURCE USAGE:") {
		t.Error("Expected resource usage section not found")
	}

	if !strings.Contains(final, "FILE SIZE STATISTICS:") {
		t.Error("Expected file size statistics section not found")
	}

	if !strings.Contains(final, "RECOMMENDATIONS:") {
		t.Error("Expected recommendations section not found")
	}

	// Check specific values
	if !strings.Contains(final, "go: 2 files") {
		t.Error("Expected go format count not found")
	}

	if !strings.Contains(final, "js: 1 files") {
		t.Error("Expected js format count not found")
	}

	if !strings.Contains(final, "syntax error: 1 occurrences") {
		t.Error("Expected error count not found")
	}
}

func TestFormatBytes(t *testing.T) {
	collector := NewCollector()
	reporter := NewReporter(collector, false, false)

	testCases := []struct {
		bytes    int64
		expected string
	}{
		{0, "0B"},
		{512, "512B"},
		{1024, "1.0KB"},
		{1536, "1.5KB"},
		{1024 * 1024, "1.0MB"},
		{1024 * 1024 * 1024, "1.0GB"},
		{5 * 1024 * 1024, "5.0MB"},
	}

	for _, tc := range testCases {
		t.Run(tc.expected, func(t *testing.T) {
			result := reporter.formatBytes(tc.bytes)
			if result != tc.expected {
				t.Errorf("formatBytes(%d) = %s, want %s", tc.bytes, result, tc.expected)
			}
		})
	}
}

func TestGetQuickStats(t *testing.T) {
	collector := NewCollector()
	reporter := NewReporter(collector, false, false)

	// Add test data
	files := []FileProcessingResult{
		{FilePath: "/test/file1.go", FileSize: 1000, Success: true, Format: "go"},
		{FilePath: "/test/file2.js", FileSize: 2000, Success: true, Format: "js"},
		{FilePath: "/test/error.py", FileSize: 500, Success: false, Error: errors.New("test error")},
	}

	for _, file := range files {
		collector.RecordFileProcessed(file)
	}

	// Wait to ensure rate calculation
	time.Sleep(10 * time.Millisecond)

	stats := reporter.GetQuickStats()

	if !strings.Contains(stats, "2/3 files") {
		t.Errorf("Expected processed/total files, got: %s", stats)
	}

	if !strings.Contains(stats, "/s)") {
		t.Errorf("Expected rate information, got: %s", stats)
	}

	if !strings.Contains(stats, "1 errors") {
		t.Errorf("Expected error count, got: %s", stats)
	}
}

func TestGetQuickStats_WithColors(t *testing.T) {
	collector := NewCollector()
	reporter := NewReporter(collector, false, true)

	// Add error file
	errorResult := FileProcessingResult{
		FilePath: "/test/error.go",
		FileSize: 512,
		Success:  false,
		Error:    errors.New("test error"),
	}
	collector.RecordFileProcessed(errorResult)

	stats := reporter.GetQuickStats()

	// Should contain ANSI color codes for errors
	if !strings.Contains(stats, "\033[31m") {
		t.Errorf("Expected color codes for errors, got: %s", stats)
	}

	if !strings.Contains(stats, "\033[0m") {
		t.Errorf("Expected color reset code, got: %s", stats)
	}
}

func TestReporter_EmptyData(t *testing.T) {
	collector := NewCollector()
	reporter := NewReporter(collector, false, false)

	// Test with no data
	progress := reporter.ReportProgress()
	if !strings.Contains(progress, "Processed: 0 files") {
		t.Errorf("Expected empty progress report, got: %s", progress)
	}

	final := reporter.ReportFinal()
	if !strings.Contains(final, "Total Files: 0") {
		t.Errorf("Expected empty final report, got: %s", final)
	}

	stats := reporter.GetQuickStats()
	if !strings.Contains(stats, "0/0 files") {
		t.Errorf("Expected empty stats, got: %s", stats)
	}
}
