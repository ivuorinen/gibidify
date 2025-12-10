// Package fileproc handles file processing, collection, and output formatting.
package fileproc

import (
	"runtime"
	"sync/atomic"
	"time"

	"github.com/ivuorinen/gibidify/shared"
)

// RecordFileProcessed records that a file has been successfully processed.
func (rm *ResourceMonitor) RecordFileProcessed(fileSize int64) {
	if rm.enabled {
		atomic.AddInt64(&rm.filesProcessed, 1)
		atomic.AddInt64(&rm.totalSizeProcessed, fileSize)
	}
}

// Metrics returns current resource usage metrics.
func (rm *ResourceMonitor) Metrics() ResourceMetrics {
	if !rm.enableResourceMon {
		return ResourceMetrics{}
	}

	rm.mu.RLock()
	defer rm.mu.RUnlock()

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	filesProcessed := atomic.LoadInt64(&rm.filesProcessed)
	totalSize := atomic.LoadInt64(&rm.totalSizeProcessed)
	duration := time.Since(rm.startTime)

	avgFileSize := float64(0)
	if filesProcessed > 0 {
		avgFileSize = float64(totalSize) / float64(filesProcessed)
	}

	processingRate := float64(0)
	if duration.Seconds() > 0 {
		processingRate = float64(filesProcessed) / duration.Seconds()
	}

	// Collect violations
	violations := make([]string, 0, len(rm.violationLogged))
	for violation := range rm.violationLogged {
		violations = append(violations, violation)
	}

	return ResourceMetrics{
		FilesProcessed:      filesProcessed,
		TotalSizeProcessed:  totalSize,
		ConcurrentReads:     atomic.LoadInt64(&rm.concurrentReads),
		MaxConcurrentReads:  int64(rm.maxConcurrentReads),
		ProcessingDuration:  duration,
		AverageFileSize:     avgFileSize,
		ProcessingRate:      processingRate,
		MemoryUsageMB:       shared.BytesToMB(m.Alloc),
		MaxMemoryUsageMB:    int64(rm.hardMemoryLimitMB),
		ViolationsDetected:  violations,
		DegradationActive:   rm.degradationActive,
		EmergencyStopActive: rm.emergencyStopRequested,
		LastUpdated:         time.Now(),
	}
}

// LogResourceInfo logs current resource limit configuration.
func (rm *ResourceMonitor) LogResourceInfo() {
	logger := shared.GetLogger()
	if rm.enabled {
		logger.Infof("Resource limits enabled: maxFiles=%d, maxTotalSize=%dMB, fileTimeout=%ds, overallTimeout=%ds",
			rm.maxFiles, rm.maxTotalSize/int64(shared.BytesPerMB), int(rm.fileProcessingTimeout.Seconds()),
			int(rm.overallTimeout.Seconds()))
		logger.Infof("Resource limits: maxConcurrentReads=%d, rateLimitFPS=%d, hardMemoryMB=%d",
			rm.maxConcurrentReads, rm.rateLimitFilesPerSec, rm.hardMemoryLimitMB)
		logger.Infof("Resource features: gracefulDegradation=%v, monitoring=%v",
			rm.enableGracefulDegr, rm.enableResourceMon)
	} else {
		logger.Info("Resource limits disabled")
	}
}
