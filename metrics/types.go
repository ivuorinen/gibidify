// Package metrics provides comprehensive processing statistics and profiling capabilities.
package metrics

import (
	"sync"
	"time"
)

// ProcessingMetrics provides comprehensive processing statistics.
type ProcessingMetrics struct {
	// File processing metrics
	TotalFiles     int64     `json:"total_files"`
	ProcessedFiles int64     `json:"processed_files"`
	SkippedFiles   int64     `json:"skipped_files"`
	ErrorFiles     int64     `json:"error_files"`
	LastUpdated    time.Time `json:"last_updated"`

	// Size metrics
	TotalSize       int64   `json:"total_size_bytes"`
	ProcessedSize   int64   `json:"processed_size_bytes"`
	AverageFileSize float64 `json:"average_file_size_bytes"`
	LargestFile     int64   `json:"largest_file_bytes"`
	SmallestFile    int64   `json:"smallest_file_bytes"`

	// Performance metrics
	StartTime      time.Time     `json:"start_time"`
	EndTime        time.Time     `json:"end_time,omitempty"`
	ProcessingTime time.Duration `json:"processing_duration"`
	FilesPerSecond float64       `json:"files_per_second"`
	BytesPerSecond float64       `json:"bytes_per_second"`

	// Memory and resource metrics
	PeakMemoryMB    int64 `json:"peak_memory_mb"`
	CurrentMemoryMB int64 `json:"current_memory_mb"`
	GoroutineCount  int   `json:"goroutine_count"`

	// Format specific metrics
	FormatCounts map[string]int64 `json:"format_counts"`
	ErrorCounts  map[string]int64 `json:"error_counts"`

	// Concurrency metrics
	MaxConcurrency     int   `json:"max_concurrency"`
	CurrentConcurrency int32 `json:"current_concurrency"`

	// Phase timings
	PhaseTimings map[string]time.Duration `json:"phase_timings"`
}

// Collector collects and manages processing metrics.
type Collector struct {
	metrics    ProcessingMetrics
	mu         sync.RWMutex
	startTime  time.Time
	lastUpdate time.Time

	// Atomic counters for high-concurrency access
	totalFiles     int64
	processedFiles int64
	skippedFiles   int64
	errorFiles     int64
	totalSize      int64
	processedSize  int64
	largestFile    int64
	smallestFile   int64 // Using max int64 as initial value to track minimum

	// Concurrency tracking
	concurrency     int32
	peakConcurrency int32

	// Format and error tracking with mutex protection
	formatCounts map[string]int64
	errorCounts  map[string]int64

	// Phase timing tracking
	phaseTimings map[string]time.Duration
}

// FileProcessingResult represents the result of processing a single file.
type FileProcessingResult struct {
	FilePath       string        `json:"file_path"`
	FileSize       int64         `json:"file_size"`
	Format         string        `json:"format"`
	ProcessingTime time.Duration `json:"processing_time"`
	Success        bool          `json:"success"`
	Error          error         `json:"error,omitempty"`
	Skipped        bool          `json:"skipped"`
	SkipReason     string        `json:"skip_reason,omitempty"`
}

// ProfileReport represents a comprehensive profiling report.
type ProfileReport struct {
	Summary          ProcessingMetrics        `json:"summary"`
	TopLargestFiles  []FileInfo               `json:"top_largest_files"`
	TopSlowestFiles  []FileInfo               `json:"top_slowest_files"`
	FormatBreakdown  map[string]FormatMetrics `json:"format_breakdown"`
	ErrorBreakdown   map[string]int64         `json:"error_breakdown"`
	HourlyStats      []HourlyProcessingStats  `json:"hourly_stats,omitempty"`
	PhaseBreakdown   map[string]PhaseMetrics  `json:"phase_breakdown"`
	PerformanceIndex float64                  `json:"performance_index"`
	Recommendations  []string                 `json:"recommendations"`
}

// FileInfo represents information about a processed file.
type FileInfo struct {
	Path           string        `json:"path"`
	Size           int64         `json:"size"`
	ProcessingTime time.Duration `json:"processing_time"`
	Format         string        `json:"format"`
}

// FormatMetrics represents metrics for a specific file format.
type FormatMetrics struct {
	Count                 int64         `json:"count"`
	TotalSize             int64         `json:"total_size"`
	AverageSize           float64       `json:"average_size"`
	TotalProcessingTime   time.Duration `json:"total_processing_time"`
	AverageProcessingTime time.Duration `json:"average_processing_time"`
}

// HourlyProcessingStats represents processing statistics for an hour.
type HourlyProcessingStats struct {
	Hour           time.Time `json:"hour"`
	FilesProcessed int64     `json:"files_processed"`
	BytesProcessed int64     `json:"bytes_processed"`
	AverageRate    float64   `json:"average_rate"`
}

// PhaseMetrics represents timing metrics for processing phases.
type PhaseMetrics struct {
	TotalTime   time.Duration `json:"total_time"`
	Count       int64         `json:"count"`
	AverageTime time.Duration `json:"average_time"`
	Percentage  float64       `json:"percentage_of_total"`
}
