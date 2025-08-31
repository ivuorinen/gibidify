package cli

import (
	"github.com/ivuorinen/gibidify/config"
	"github.com/ivuorinen/gibidify/utils"
)

// logFinalStats logs the final back-pressure, resource monitoring, and comprehensive processing statistics.
func (p *Processor) logFinalStats() {
	logger := utils.GetLogger()

	// Log back-pressure stats
	backpressureStats := p.backpressure.GetStats()
	if backpressureStats.Enabled {
		logger.Infof(
			"Back-pressure stats: processed=%d files, memory=%dMB/%dMB",
			backpressureStats.FilesProcessed,
			backpressureStats.CurrentMemoryUsage/1024/1024,
			backpressureStats.MaxMemoryUsage/1024/1024,
		)
	}

	// Log resource monitoring stats
	resourceStats := p.resourceMonitor.GetMetrics()
	if config.GetResourceLimitsEnabled() {
		logger.Infof(
			"Resource stats: processed=%d files, totalSize=%dMB, avgFileSize=%.2fKB, rate=%.2f files/sec",
			resourceStats.FilesProcessed, resourceStats.TotalSizeProcessed/1024/1024,
			resourceStats.AverageFileSize/1024, resourceStats.ProcessingRate,
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

	// Finalize and report comprehensive metrics
	if p.metricsCollector != nil {
		p.metricsCollector.Finish()
	}

	// Display final metrics report via UI manager
	if p.metricsReporter != nil {
		finalReport := p.metricsReporter.ReportFinal()
		if finalReport != "" {
			// Use UI manager to respect NoUI flag - remove trailing newline if present
			reportText := finalReport
			if len(reportText) > 0 && reportText[len(reportText)-1] == '\n' {
				reportText = reportText[:len(reportText)-1]
			}
			p.ui.PrintInfo("%s", reportText)
		}
	}

	// Log for structured logging if verbose
	if p.flags.Verbose {
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

	// Clean up resource monitor
	p.resourceMonitor.Close()
}
