// Package cli provides command-line interface functionality for gibidify.
package cli

import (
	"strings"

	"github.com/ivuorinen/gibidify/config"
	"github.com/ivuorinen/gibidify/shared"
)

// logFinalStats logs back-pressure, resource usage, and processing statistics.
func (p *Processor) logFinalStats() {
	p.logBackpressureStats()
	p.logResourceStats()
	p.finalizeAndReportMetrics()
	p.logVerboseStats()
	if p.resourceMonitor != nil {
		p.resourceMonitor.Close()
	}
}

// logBackpressureStats logs back-pressure statistics.
func (p *Processor) logBackpressureStats() {
	// Check backpressure is non-nil before dereferencing
	if p.backpressure == nil {
		return
	}

	logger := shared.GetLogger()
	backpressureStats := p.backpressure.Stats()
	if backpressureStats.Enabled {
		logger.Infof(
			"Back-pressure stats: processed=%d files, memory=%dMB/%dMB",
			backpressureStats.FilesProcessed,
			backpressureStats.CurrentMemoryUsage/int64(shared.BytesPerMB),
			backpressureStats.MaxMemoryUsage/int64(shared.BytesPerMB),
		)
	}
}

// logResourceStats logs resource monitoring statistics.
func (p *Processor) logResourceStats() {
	// Check resource monitoring is enabled and monitor is non-nil before dereferencing
	if !config.ResourceLimitsEnabled() {
		return
	}

	if p.resourceMonitor == nil {
		return
	}

	logger := shared.GetLogger()
	resourceStats := p.resourceMonitor.Metrics()

	logger.Infof(
		"Resource stats: processed=%d files, totalSize=%dMB, avgFileSize=%.2fKB, rate=%.2f files/sec",
		resourceStats.FilesProcessed, resourceStats.TotalSizeProcessed/int64(shared.BytesPerMB),
		resourceStats.AverageFileSize/float64(shared.BytesPerKB), resourceStats.ProcessingRate,
	)

	if len(resourceStats.ViolationsDetected) > 0 {
		logger.Warnf("Resource violations detected: %v", resourceStats.ViolationsDetected)
	}

	if resourceStats.DegradationActive {
		logger.Warnf("Processing completed with degradation mode active")
	}

	if resourceStats.EmergencyStopActive {
		logger.Errorf("Processing completed with emergency stop active")
	}
}

// finalizeAndReportMetrics finalizes metrics collection and displays the final report.
func (p *Processor) finalizeAndReportMetrics() {
	if p.metricsCollector != nil {
		p.metricsCollector.Finish()
	}

	if p.metricsReporter != nil {
		finalReport := p.metricsReporter.ReportFinal()
		if finalReport != "" && p.ui != nil {
			// Use UI manager to respect NoUI flag - remove trailing newline if present
			p.ui.PrintInfo("%s", strings.TrimSuffix(finalReport, "\n"))
		}
	}
}

// logVerboseStats logs detailed structured statistics when verbose mode is enabled.
func (p *Processor) logVerboseStats() {
	if !p.flags.Verbose || p.metricsCollector == nil {
		return
	}

	logger := shared.GetLogger()
	report := p.metricsCollector.GenerateReport()
	fields := map[string]any{
		"total_files":      report.Summary.TotalFiles,
		"processed_files":  report.Summary.ProcessedFiles,
		"skipped_files":    report.Summary.SkippedFiles,
		"error_files":      report.Summary.ErrorFiles,
		"processing_time":  report.Summary.ProcessingTime,
		"files_per_second": report.Summary.FilesPerSecond,
		"bytes_per_second": report.Summary.BytesPerSecond,
		"memory_usage_mb":  report.Summary.CurrentMemoryMB,
	}
	logger.WithFields(fields).Info("Processing completed with comprehensive metrics")
}
