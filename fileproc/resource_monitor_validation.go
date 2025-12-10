// Package fileproc handles file processing, collection, and output formatting.
package fileproc

import (
	"runtime"
	"sync/atomic"
	"time"

	"github.com/ivuorinen/gibidify/shared"
)

// ValidateFileProcessing checks if a file can be processed based on resource limits.
func (rm *ResourceMonitor) ValidateFileProcessing(filePath string, fileSize int64) error {
	if !rm.enabled {
		return nil
	}

	rm.mu.RLock()
	defer rm.mu.RUnlock()

	// Check if emergency stop is active
	if rm.emergencyStopRequested {
		return shared.NewStructuredError(
			shared.ErrorTypeValidation,
			shared.CodeResourceLimitMemory,
			"processing stopped due to emergency memory condition",
			filePath,
			map[string]any{
				"emergency_stop_active": true,
			},
		)
	}

	// Check file count limit
	currentFiles := atomic.LoadInt64(&rm.filesProcessed)
	if int(currentFiles) >= rm.maxFiles {
		return shared.NewStructuredError(
			shared.ErrorTypeValidation,
			shared.CodeResourceLimitFiles,
			"maximum file count limit exceeded",
			filePath,
			map[string]any{
				"current_files": currentFiles,
				"max_files":     rm.maxFiles,
			},
		)
	}

	// Check total size limit
	currentTotalSize := atomic.LoadInt64(&rm.totalSizeProcessed)
	if currentTotalSize+fileSize > rm.maxTotalSize {
		return shared.NewStructuredError(
			shared.ErrorTypeValidation,
			shared.CodeResourceLimitTotalSize,
			"maximum total size limit would be exceeded",
			filePath,
			map[string]any{
				"current_total_size": currentTotalSize,
				"file_size":          fileSize,
				"max_total_size":     rm.maxTotalSize,
			},
		)
	}

	// Check overall timeout
	if time.Since(rm.startTime) > rm.overallTimeout {
		return shared.NewStructuredError(
			shared.ErrorTypeValidation,
			shared.CodeResourceLimitTimeout,
			"overall processing timeout exceeded",
			filePath,
			map[string]any{
				"processing_duration": time.Since(rm.startTime),
				"overall_timeout":     rm.overallTimeout,
			},
		)
	}

	return nil
}

// CheckHardMemoryLimit checks if hard memory limit is exceeded and takes action.
func (rm *ResourceMonitor) CheckHardMemoryLimit() error {
	if !rm.enabled || rm.hardMemoryLimitMB <= 0 {
		return nil
	}

	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	currentMemory := shared.SafeUint64ToInt64WithDefault(m.Alloc, 0)

	if currentMemory <= rm.hardMemoryLimitBytes {
		return nil
	}

	return rm.handleMemoryLimitExceeded(currentMemory)
}

// handleMemoryLimitExceeded handles the case when hard memory limit is exceeded.
func (rm *ResourceMonitor) handleMemoryLimitExceeded(currentMemory int64) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	rm.logMemoryViolation(currentMemory)

	if !rm.enableGracefulDegr {
		return rm.createHardMemoryLimitError(currentMemory, false)
	}

	return rm.tryGracefulRecovery(currentMemory)
}

// logMemoryViolation logs memory limit violation if not already logged.
func (rm *ResourceMonitor) logMemoryViolation(currentMemory int64) {
	violationKey := "hard_memory_limit"

	// Ensure map is initialized
	if rm.violationLogged == nil {
		rm.violationLogged = make(map[string]bool)
	}

	if rm.violationLogged[violationKey] {
		return
	}

	logger := shared.GetLogger()
	logger.Errorf("Hard memory limit exceeded: %dMB > %dMB",
		currentMemory/int64(shared.BytesPerMB), rm.hardMemoryLimitMB)
	rm.violationLogged[violationKey] = true
}

// tryGracefulRecovery attempts graceful recovery by forcing GC.
func (rm *ResourceMonitor) tryGracefulRecovery(_ int64) error {
	// Force garbage collection
	runtime.GC()

	// Check again after GC
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	newMemory := shared.SafeUint64ToInt64WithDefault(m.Alloc, 0)

	if newMemory > rm.hardMemoryLimitBytes {
		// Still over limit, activate emergency stop
		rm.emergencyStopRequested = true

		return rm.createHardMemoryLimitError(newMemory, true)
	}

	// Memory freed by GC, continue with degradation
	rm.degradationActive = true
	logger := shared.GetLogger()
	logger.Info("Memory freed by garbage collection, continuing with degradation mode")

	return nil
}

// createHardMemoryLimitError creates a structured error for memory limit exceeded.
func (rm *ResourceMonitor) createHardMemoryLimitError(currentMemory int64, emergencyStop bool) error {
	message := "hard memory limit exceeded"
	if emergencyStop {
		message = "hard memory limit exceeded, emergency stop activated"
	}

	context := map[string]any{
		"current_memory_mb": currentMemory / int64(shared.BytesPerMB),
		"limit_mb":          rm.hardMemoryLimitMB,
	}
	if emergencyStop {
		context["emergency_stop"] = true
	}

	return shared.NewStructuredError(
		shared.ErrorTypeValidation,
		shared.CodeResourceLimitMemory,
		message,
		"",
		context,
	)
}
