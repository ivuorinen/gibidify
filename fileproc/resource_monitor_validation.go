package fileproc

import (
	"math"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/ivuorinen/gibidify/utils"
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

// CheckHardMemoryLimit checks if hard memory limit is exceeded and takes action.
func (rm *ResourceMonitor) CheckHardMemoryLimit() error {
	if !rm.enabled || rm.hardMemoryLimitMB <= 0 {
		return nil
	}

	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	// Check for overflow before converting uint64 to int64
	if m.Alloc > uint64(math.MaxInt64) {
		return utils.NewStructuredError(utils.ErrorTypeValidation, utils.CodeValidationSize, "memory allocation exceeds int64 maximum", "", nil)
	}
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
			// Check for overflow before converting uint64 to int64
			if m.Alloc > uint64(math.MaxInt64) {
				rm.emergencyStopRequested = true
				return utils.NewStructuredError(utils.ErrorTypeValidation, utils.CodeValidationSize, "memory allocation exceeds int64 maximum after GC", "", nil)
			}
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
