// Package metrics provides performance monitoring and reporting capabilities.
package metrics

import (
	"math"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/ivuorinen/gibidify/shared"
)

// NewCollector creates a new metrics collector.
func NewCollector() *Collector {
	now := time.Now()

	return &Collector{
		startTime:    now,
		lastUpdate:   now,
		formatCounts: make(map[string]int64),
		errorCounts:  make(map[string]int64),
		phaseTimings: make(map[string]time.Duration),
		smallestFile: math.MaxInt64, // Initialize to max value to properly track minimum
	}
}

// RecordFileProcessed records the successful processing of a file.
func (c *Collector) RecordFileProcessed(result FileProcessingResult) {
	atomic.AddInt64(&c.totalFiles, 1)

	c.updateFileStatusCounters(result)
	atomic.AddInt64(&c.totalSize, result.FileSize)
	c.updateFormatAndErrorCounts(result)
}

// updateFileStatusCounters updates counters based on file processing result.
func (c *Collector) updateFileStatusCounters(result FileProcessingResult) {
	switch {
	case result.Success:
		atomic.AddInt64(&c.processedFiles, 1)
		atomic.AddInt64(&c.processedSize, result.FileSize)
		c.updateFileSizeExtremes(result.FileSize)
	case result.Skipped:
		atomic.AddInt64(&c.skippedFiles, 1)
	default:
		atomic.AddInt64(&c.errorFiles, 1)
	}
}

// updateFileSizeExtremes updates the largest and smallest file size atomically.
func (c *Collector) updateFileSizeExtremes(fileSize int64) {
	// Update the largest file atomically
	for {
		current := atomic.LoadInt64(&c.largestFile)
		if fileSize <= current {
			break
		}
		if atomic.CompareAndSwapInt64(&c.largestFile, current, fileSize) {
			break
		}
	}

	// Update the smallest file atomically
	for {
		current := atomic.LoadInt64(&c.smallestFile)
		if fileSize >= current {
			break
		}
		if atomic.CompareAndSwapInt64(&c.smallestFile, current, fileSize) {
			break
		}
	}
}

// updateFormatAndErrorCounts updates format and error counts with mutex protection.
func (c *Collector) updateFormatAndErrorCounts(result FileProcessingResult) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if result.Format != "" {
		c.formatCounts[result.Format]++
	}
	if result.Error != nil {
		errorType := c.simplifyErrorType(result.Error)
		c.errorCounts[errorType]++
	}
	c.lastUpdate = time.Now()
}

// simplifyErrorType simplifies error messages for better aggregation.
func (c *Collector) simplifyErrorType(err error) string {
	errorType := err.Error()
	// Simplify error types for better aggregation
	if len(errorType) > 50 {
		errorType = errorType[:50] + "..."
	}

	return errorType
}

// RecordPhaseTime records the time spent in a processing phase.
func (c *Collector) RecordPhaseTime(phase string, duration time.Duration) {
	c.mu.Lock()
	c.phaseTimings[phase] += duration
	c.mu.Unlock()
}

// IncrementConcurrency increments the current concurrency counter.
func (c *Collector) IncrementConcurrency() {
	newVal := atomic.AddInt32(&c.concurrency, 1)

	// Update peak concurrency if current is higher
	for {
		peak := atomic.LoadInt32(&c.peakConcurrency)
		if newVal <= peak || atomic.CompareAndSwapInt32(&c.peakConcurrency, peak, newVal) {
			break
		}
	}
}

// DecrementConcurrency decrements the current concurrency counter.
// Prevents negative values if calls are imbalanced.
func (c *Collector) DecrementConcurrency() {
	for {
		cur := atomic.LoadInt32(&c.concurrency)
		if cur == 0 {
			return
		}
		if atomic.CompareAndSwapInt32(&c.concurrency, cur, cur-1) {
			return
		}
	}
}

// CurrentMetrics returns the current metrics snapshot.
func (c *Collector) CurrentMetrics() ProcessingMetrics {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	now := time.Now()
	processingTime := now.Sub(c.startTime)

	totalFiles := atomic.LoadInt64(&c.totalFiles)
	processedFiles := atomic.LoadInt64(&c.processedFiles)
	processedSize := atomic.LoadInt64(&c.processedSize)

	var avgFileSize float64
	if processedFiles > 0 {
		avgFileSize = float64(processedSize) / float64(processedFiles)
	}

	var filesPerSec, bytesPerSec float64
	if processingTime.Seconds() > 0 {
		filesPerSec = float64(processedFiles) / processingTime.Seconds()
		bytesPerSec = float64(processedSize) / processingTime.Seconds()
	}

	smallestFile := atomic.LoadInt64(&c.smallestFile)
	if smallestFile == math.MaxInt64 {
		smallestFile = 0 // No files processed yet
	}

	// Copy maps to avoid race conditions
	formatCounts := make(map[string]int64)
	for k, v := range c.formatCounts {
		formatCounts[k] = v
	}

	errorCounts := make(map[string]int64)
	for k, v := range c.errorCounts {
		errorCounts[k] = v
	}

	phaseTimings := make(map[string]time.Duration)
	for k, v := range c.phaseTimings {
		phaseTimings[k] = v
	}

	return ProcessingMetrics{
		TotalFiles:         totalFiles,
		ProcessedFiles:     processedFiles,
		SkippedFiles:       atomic.LoadInt64(&c.skippedFiles),
		ErrorFiles:         atomic.LoadInt64(&c.errorFiles),
		LastUpdated:        c.lastUpdate,
		TotalSize:          atomic.LoadInt64(&c.totalSize),
		ProcessedSize:      processedSize,
		AverageFileSize:    avgFileSize,
		LargestFile:        atomic.LoadInt64(&c.largestFile),
		SmallestFile:       smallestFile,
		StartTime:          c.startTime,
		ProcessingTime:     processingTime,
		FilesPerSecond:     filesPerSec,
		BytesPerSecond:     bytesPerSec,
		PeakMemoryMB:       shared.BytesToMB(m.Sys),
		CurrentMemoryMB:    shared.BytesToMB(m.Alloc),
		GoroutineCount:     runtime.NumGoroutine(),
		FormatCounts:       formatCounts,
		ErrorCounts:        errorCounts,
		MaxConcurrency:     int(atomic.LoadInt32(&c.peakConcurrency)),
		CurrentConcurrency: atomic.LoadInt32(&c.concurrency),
		PhaseTimings:       phaseTimings,
	}
}

// Finish marks the end of processing and records final metrics.
func (c *Collector) Finish() {
	// Get current metrics first (which will acquire its own lock)
	currentMetrics := c.CurrentMetrics()

	// Then update final metrics with lock
	c.mu.Lock()
	defer c.mu.Unlock()

	c.metrics = currentMetrics
	c.metrics.EndTime = time.Now()
	c.metrics.ProcessingTime = c.metrics.EndTime.Sub(c.startTime)
}

// FinalMetrics returns the final metrics after processing is complete.
func (c *Collector) FinalMetrics() ProcessingMetrics {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.metrics
}

// GenerateReport generates a comprehensive profiling report.
func (c *Collector) GenerateReport() ProfileReport {
	metrics := c.CurrentMetrics()

	// Generate format breakdown
	formatBreakdown := make(map[string]FormatMetrics)
	for format, count := range metrics.FormatCounts {
		// For now, we don't have detailed per-format timing data
		// This could be enhanced in the future
		formatBreakdown[format] = FormatMetrics{
			Count:                 count,
			TotalSize:             0, // Would need to track this separately
			AverageSize:           0,
			TotalProcessingTime:   0,
			AverageProcessingTime: 0,
		}
	}

	// Generate phase breakdown
	phaseBreakdown := make(map[string]PhaseMetrics)
	totalPhaseTime := time.Duration(0)
	for _, duration := range metrics.PhaseTimings {
		totalPhaseTime += duration
	}

	for phase, duration := range metrics.PhaseTimings {
		percentage := float64(0)
		if totalPhaseTime > 0 {
			percentage = float64(duration) / float64(totalPhaseTime) * 100
		}

		phaseBreakdown[phase] = PhaseMetrics{
			TotalTime:   duration,
			Count:       1, // For now, we track total time per phase
			AverageTime: duration,
			Percentage:  percentage,
		}
	}

	// Calculate performance index (files per second normalized)
	performanceIndex := metrics.FilesPerSecond
	if performanceIndex > shared.MetricsPerformanceIndexCap {
		performanceIndex = shared.MetricsPerformanceIndexCap // Cap for reasonable indexing
	}

	// Generate recommendations
	recommendations := c.generateRecommendations(metrics)

	return ProfileReport{
		Summary:          metrics,
		TopLargestFiles:  []FileInfo{}, // Would need separate tracking
		TopSlowestFiles:  []FileInfo{}, // Would need separate tracking
		FormatBreakdown:  formatBreakdown,
		ErrorBreakdown:   metrics.ErrorCounts,
		PhaseBreakdown:   phaseBreakdown,
		PerformanceIndex: performanceIndex,
		Recommendations:  recommendations,
	}
}

// generateRecommendations generates performance recommendations based on metrics.
func (c *Collector) generateRecommendations(metrics ProcessingMetrics) []string {
	var recommendations []string

	// Memory usage recommendations
	if metrics.CurrentMemoryMB > 500 {
		recommendations = append(recommendations, "Consider reducing memory usage - current usage is high (>500MB)")
	}

	// Processing rate recommendations
	if metrics.FilesPerSecond < 10 && metrics.ProcessedFiles > 100 {
		recommendations = append(recommendations,
			"Processing rate is low (<10 files/sec) - consider optimizing file I/O")
	}

	// Error rate recommendations
	if metrics.TotalFiles > 0 {
		errorRate := float64(metrics.ErrorFiles) / float64(metrics.TotalFiles) * 100
		if errorRate > 5 {
			recommendations = append(recommendations, "High error rate (>5%) detected - review error logs")
		}
	}

	// Concurrency recommendations
	halfMaxConcurrency := shared.SafeIntToInt32WithDefault(metrics.MaxConcurrency/2, 1)
	if halfMaxConcurrency > 0 && metrics.CurrentConcurrency < halfMaxConcurrency {
		recommendations = append(recommendations,
			"Low concurrency utilization - consider increasing concurrent processing")
	}

	// Large file recommendations
	const largeSizeThreshold = 50 * shared.BytesPerMB // 50MB
	if metrics.LargestFile > largeSizeThreshold {
		recommendations = append(
			recommendations,
			"Very large files detected (>50MB) - consider streaming processing for large files",
		)
	}

	return recommendations
}

// Reset resets all metrics to initial state.
func (c *Collector) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	c.startTime = now
	c.lastUpdate = now

	atomic.StoreInt64(&c.totalFiles, 0)
	atomic.StoreInt64(&c.processedFiles, 0)
	atomic.StoreInt64(&c.skippedFiles, 0)
	atomic.StoreInt64(&c.errorFiles, 0)
	atomic.StoreInt64(&c.totalSize, 0)
	atomic.StoreInt64(&c.processedSize, 0)
	atomic.StoreInt64(&c.largestFile, 0)
	atomic.StoreInt64(&c.smallestFile, math.MaxInt64)
	atomic.StoreInt32(&c.concurrency, 0)

	c.formatCounts = make(map[string]int64)
	c.errorCounts = make(map[string]int64)
	c.metrics = ProcessingMetrics{} // Clear final snapshot
	c.phaseTimings = make(map[string]time.Duration)
}
