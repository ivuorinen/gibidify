// Package fileproc provides resource monitoring and limit enforcement for security.
package fileproc

import (
	"context"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/ivuorinen/gibidify/config"
	"github.com/ivuorinen/gibidify/utils"
)

// ResourceMonitor monitors resource usage and enforces limits to prevent DoS attacks.
type ResourceMonitor struct {
	enabled               bool
	maxFiles              int
	maxTotalSize          int64
	fileProcessingTimeout time.Duration
	overallTimeout        time.Duration
	maxConcurrentReads    int
	rateLimitFilesPerSec  int
	hardMemoryLimitMB     int
	enableGracefulDegr    bool
	enableResourceMon     bool

	// Current state tracking
	filesProcessed       int64
	totalSizeProcessed   int64
	concurrentReads      int64
	startTime            time.Time
	lastRateLimitCheck   time.Time
	hardMemoryLimitBytes int64

	// Rate limiting
	rateLimiter   *time.Ticker
	rateLimitChan chan struct{}

	// Synchronization
	mu                     sync.RWMutex
	violationLogged        map[string]bool
	degradationActive      bool
	emergencyStopRequested bool
}

// ResourceMetrics holds comprehensive resource usage metrics.
type ResourceMetrics struct {
	FilesProcessed      int64         `json:"files_processed"`
	TotalSizeProcessed  int64         `json:"total_size_processed"`
	ConcurrentReads     int64         `json:"concurrent_reads"`
	ProcessingDuration  time.Duration `json:"processing_duration"`
	AverageFileSize     float64       `json:"average_file_size"`
	ProcessingRate      float64       `json:"processing_rate_files_per_sec"`
	MemoryUsageMB       int64         `json:"memory_usage_mb"`
	MaxMemoryUsageMB    int64         `json:"max_memory_usage_mb"`
	ViolationsDetected  []string      `json:"violations_detected"`
	DegradationActive   bool          `json:"degradation_active"`
	EmergencyStopActive bool          `json:"emergency_stop_active"`
	LastUpdated         time.Time     `json:"last_updated"`
}

// ResourceViolation represents a detected resource limit violation.
type ResourceViolation struct {
	Type      string                 `json:"type"`
	Message   string                 `json:"message"`
	Current   interface{}            `json:"current"`
	Limit     interface{}            `json:"limit"`
	Timestamp time.Time              `json:"timestamp"`
	Context   map[string]interface{} `json:"context"`
}

// NewResourceMonitor creates a new resource monitor with configuration.
func NewResourceMonitor() *ResourceMonitor {
	rm := &ResourceMonitor{
		enabled:               config.GetResourceLimitsEnabled(),
		maxFiles:              config.GetMaxFiles(),
		maxTotalSize:          config.GetMaxTotalSize(),
		fileProcessingTimeout: time.Duration(config.GetFileProcessingTimeoutSec()) * time.Second,
		overallTimeout:        time.Duration(config.GetOverallTimeoutSec()) * time.Second,
		maxConcurrentReads:    config.GetMaxConcurrentReads(),
		rateLimitFilesPerSec:  config.GetRateLimitFilesPerSec(),
		hardMemoryLimitMB:     config.GetHardMemoryLimitMB(),
		enableGracefulDegr:    config.GetEnableGracefulDegradation(),
		enableResourceMon:     config.GetEnableResourceMonitoring(),
		startTime:             time.Now(),
		lastRateLimitCheck:    time.Now(),
		violationLogged:       make(map[string]bool),
		hardMemoryLimitBytes:  int64(config.GetHardMemoryLimitMB()) * 1024 * 1024,
	}

	// Initialize rate limiter if rate limiting is enabled
	if rm.enabled && rm.rateLimitFilesPerSec > 0 {
		interval := time.Second / time.Duration(rm.rateLimitFilesPerSec)
		rm.rateLimiter = time.NewTicker(interval)
		rm.rateLimitChan = make(chan struct{}, rm.rateLimitFilesPerSec)

		// Pre-fill the rate limit channel
		for i := 0; i < rm.rateLimitFilesPerSec; i++ {
			select {
			case rm.rateLimitChan <- struct{}{}:
			default:
				goto rateLimitFull
			}
		}
	rateLimitFull:

		// Start rate limiter refill goroutine
		go rm.rateLimiterRefill()
	}

	return rm
}

// ValidateFileProcessing checks if a file can be processed based on resource limits.
func (rm *ResourceMonitor) ValidateFileProcessing(filePath string, fileSize int64) error {
	if !rm.enabled {
		return nil
	}

	rm.mu.RLock()
	defer rm.mu.RUnlock()

	// Check if emergency stop is active
	if rm.emergencyStopRequested {
		return utils.NewStructuredError(
			utils.ErrorTypeValidation,
			utils.CodeResourceLimitMemory,
			"processing stopped due to emergency memory condition",
			filePath,
			map[string]interface{}{
				"emergency_stop_active": true,
			},
		)
	}

	// Check file count limit
	currentFiles := atomic.LoadInt64(&rm.filesProcessed)
	if int(currentFiles) >= rm.maxFiles {
		return utils.NewStructuredError(
			utils.ErrorTypeValidation,
			utils.CodeResourceLimitFiles,
			"maximum file count limit exceeded",
			filePath,
			map[string]interface{}{
				"current_files": currentFiles,
				"max_files":     rm.maxFiles,
			},
		)
	}

	// Check total size limit
	currentTotalSize := atomic.LoadInt64(&rm.totalSizeProcessed)
	if currentTotalSize+fileSize > rm.maxTotalSize {
		return utils.NewStructuredError(
			utils.ErrorTypeValidation,
			utils.CodeResourceLimitTotalSize,
			"maximum total size limit would be exceeded",
			filePath,
			map[string]interface{}{
				"current_total_size": currentTotalSize,
				"file_size":          fileSize,
				"max_total_size":     rm.maxTotalSize,
			},
		)
	}

	// Check overall timeout
	if time.Since(rm.startTime) > rm.overallTimeout {
		return utils.NewStructuredError(
			utils.ErrorTypeValidation,
			utils.CodeResourceLimitTimeout,
			"overall processing timeout exceeded",
			filePath,
			map[string]interface{}{
				"processing_duration": time.Since(rm.startTime),
				"overall_timeout":     rm.overallTimeout,
			},
		)
	}

	return nil
}

// AcquireReadSlot attempts to acquire a slot for concurrent file reading.
func (rm *ResourceMonitor) AcquireReadSlot(ctx context.Context) error {
	if !rm.enabled {
		return nil
	}

	// Wait for available read slot
	for {
		currentReads := atomic.LoadInt64(&rm.concurrentReads)
		if currentReads < int64(rm.maxConcurrentReads) {
			if atomic.CompareAndSwapInt64(&rm.concurrentReads, currentReads, currentReads+1) {
				break
			}
			// CAS failed, retry
			continue
		}

		// Wait and retry
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Millisecond):
			// Continue loop
		}
	}

	return nil
}

// ReleaseReadSlot releases a concurrent reading slot.
func (rm *ResourceMonitor) ReleaseReadSlot() {
	if rm.enabled {
		atomic.AddInt64(&rm.concurrentReads, -1)
	}
}

// WaitForRateLimit waits for rate limiting if enabled.
func (rm *ResourceMonitor) WaitForRateLimit(ctx context.Context) error {
	if !rm.enabled || rm.rateLimitFilesPerSec <= 0 {
		return nil
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-rm.rateLimitChan:
		return nil
	case <-time.After(time.Second): // Fallback timeout
		logrus.Warn("Rate limiting timeout exceeded, continuing without rate limit")
		return nil
	}
}

// CheckHardMemoryLimit checks if hard memory limit is exceeded and takes action.
func (rm *ResourceMonitor) CheckHardMemoryLimit() error {
	if !rm.enabled || rm.hardMemoryLimitMB <= 0 {
		return nil
	}

	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	currentMemory := int64(m.Alloc)

	if currentMemory > rm.hardMemoryLimitBytes {
		rm.mu.Lock()
		defer rm.mu.Unlock()

		// Log violation if not already logged
		violationKey := "hard_memory_limit"
		if !rm.violationLogged[violationKey] {
			logrus.Errorf("Hard memory limit exceeded: %dMB > %dMB",
				currentMemory/1024/1024, rm.hardMemoryLimitMB)
			rm.violationLogged[violationKey] = true
		}

		if rm.enableGracefulDegr {
			// Force garbage collection
			runtime.GC()

			// Check again after GC
			runtime.ReadMemStats(&m)
			currentMemory = int64(m.Alloc)

			if currentMemory > rm.hardMemoryLimitBytes {
				// Still over limit, activate emergency stop
				rm.emergencyStopRequested = true
				return utils.NewStructuredError(
					utils.ErrorTypeValidation,
					utils.CodeResourceLimitMemory,
					"hard memory limit exceeded, emergency stop activated",
					"",
					map[string]interface{}{
						"current_memory_mb": currentMemory / 1024 / 1024,
						"limit_mb":          rm.hardMemoryLimitMB,
						"emergency_stop":    true,
					},
				)
			} else {
				// Memory freed by GC, continue with degradation
				rm.degradationActive = true
				logrus.Info("Memory freed by garbage collection, continuing with degradation mode")
			}
		} else {
			// No graceful degradation, hard stop
			return utils.NewStructuredError(
				utils.ErrorTypeValidation,
				utils.CodeResourceLimitMemory,
				"hard memory limit exceeded",
				"",
				map[string]interface{}{
					"current_memory_mb": currentMemory / 1024 / 1024,
					"limit_mb":          rm.hardMemoryLimitMB,
				},
			)
		}
	}

	return nil
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
		MemoryUsageMB:       int64(m.Alloc) / 1024 / 1024,
		MaxMemoryUsageMB:    int64(rm.hardMemoryLimitMB),
		ViolationsDetected:  violations,
		DegradationActive:   rm.degradationActive,
		EmergencyStopActive: rm.emergencyStopRequested,
		LastUpdated:         time.Now(),
	}
}

// IsEmergencyStopActive returns whether emergency stop is active.
func (rm *ResourceMonitor) IsEmergencyStopActive() bool {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	return rm.emergencyStopRequested
}

// IsDegradationActive returns whether degradation mode is active.
func (rm *ResourceMonitor) IsDegradationActive() bool {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	return rm.degradationActive
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

// Close cleans up the resource monitor.
func (rm *ResourceMonitor) Close() {
	if rm.rateLimiter != nil {
		rm.rateLimiter.Stop()
	}
}

// rateLimiterRefill refills the rate limiting channel periodically.
func (rm *ResourceMonitor) rateLimiterRefill() {
	for range rm.rateLimiter.C {
		select {
		case rm.rateLimitChan <- struct{}{}:
		default:
			// Channel is full, skip
		}
	}
}

// CreateFileProcessingContext creates a context with file processing timeout.
func (rm *ResourceMonitor) CreateFileProcessingContext(parent context.Context) (context.Context, context.CancelFunc) {
	if !rm.enabled || rm.fileProcessingTimeout <= 0 {
		return parent, func() {}
	}
	return context.WithTimeout(parent, rm.fileProcessingTimeout)
}

// CreateOverallProcessingContext creates a context with overall processing timeout.
func (rm *ResourceMonitor) CreateOverallProcessingContext(parent context.Context) (context.Context, context.CancelFunc) {
	if !rm.enabled || rm.overallTimeout <= 0 {
		return parent, func() {}
	}
	return context.WithTimeout(parent, rm.overallTimeout)
}
