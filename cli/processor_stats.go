package cli

import (
	"github.com/sirupsen/logrus"

	"github.com/ivuorinen/gibidify/config"
)

// logFinalStats logs the final back-pressure and resource monitoring statistics.
func (p *Processor) logFinalStats() {
	// Log back-pressure stats
	backpressureStats := p.backpressure.GetStats()
	if backpressureStats.Enabled {
		logrus.Infof("Back-pressure stats: processed=%d files, memory=%dMB/%dMB",
			backpressureStats.FilesProcessed, backpressureStats.CurrentMemoryUsage/1024/1024, backpressureStats.MaxMemoryUsage/1024/1024)
	}

	// Log resource monitoring stats
	resourceStats := p.resourceMonitor.GetMetrics()
	if config.GetResourceLimitsEnabled() {
		logrus.Infof("Resource stats: processed=%d files, totalSize=%dMB, avgFileSize=%.2fKB, rate=%.2f files/sec",
			resourceStats.FilesProcessed, resourceStats.TotalSizeProcessed/1024/1024,
			resourceStats.AverageFileSize/1024, resourceStats.ProcessingRate)

		if len(resourceStats.ViolationsDetected) > 0 {
			logrus.Warnf("Resource violations detected: %v", resourceStats.ViolationsDetected)
		}

		if resourceStats.DegradationActive {
			logrus.Warnf("Processing completed with degradation mode active")
		}

		if resourceStats.EmergencyStopActive {
			logrus.Errorf("Processing completed with emergency stop active")
		}
	}

	// Clean up resource monitor
	p.resourceMonitor.Close()
}