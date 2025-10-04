package fileproc

import (
	"math"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/sirupsen/logrus"
)

// safeUint64ToInt64 safely converts uint64 to int64, returning max int64 on overflow
func safeUint64ToInt64(val uint64) int64 {
	if val > uint64(math.MaxInt64) {
		return math.MaxInt64 // Return max int64 on overflow
	}
	return int64(val)
}

// RecordFileProcessed records that a file has been successfully processed.
func (rm *ResourceMonitor) RecordFileProcessed(fileSize int64) {
	if rm.enabled {
		atomic.AddInt64(&rm.filesProcessed, 1)
		atomic.AddInt64(&rm.totalSizeProcessed, fileSize)
	}
}

// GetMetrics returns current resource usage metrics.
func (rm *ResourceMonitor) GetMetrics() ResourceMetrics {
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
		ProcessingDuration:  duration,
		AverageFileSize:     avgFileSize,
		ProcessingRate:      processingRate,
		MemoryUsageMB:       safeUint64ToInt64(m.Alloc) / 1024 / 1024,
		MaxMemoryUsageMB:    int64(rm.hardMemoryLimitMB),
		ViolationsDetected:  violations,
		DegradationActive:   rm.degradationActive,
		EmergencyStopActive: rm.emergencyStopRequested,
		LastUpdated:         time.Now(),
	}
}

// LogResourceInfo logs current resource limit configuration.
func (rm *ResourceMonitor) LogResourceInfo() {
	if rm.enabled {
		logrus.Infof("Resource limits enabled: maxFiles=%d, maxTotalSize=%dMB, fileTimeout=%ds, overallTimeout=%ds",
			rm.maxFiles, rm.maxTotalSize/1024/1024, int(rm.fileProcessingTimeout.Seconds()), int(rm.overallTimeout.Seconds()))
		logrus.Infof("Resource limits: maxConcurrentReads=%d, rateLimitFPS=%d, hardMemoryMB=%d",
			rm.maxConcurrentReads, rm.rateLimitFilesPerSec, rm.hardMemoryLimitMB)
		logrus.Infof("Resource features: gracefulDegradation=%v, monitoring=%v",
			rm.enableGracefulDegr, rm.enableResourceMon)
	} else {
		logrus.Info("Resource limits disabled")
	}
}
