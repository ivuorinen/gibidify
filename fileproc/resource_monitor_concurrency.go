// Package fileproc handles file processing, collection, and output formatting.
package fileproc

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"
)

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
			return fmt.Errorf("context canceled while waiting for read slot: %w", ctx.Err())
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

// CreateFileProcessingContext creates a context with file processing timeout.
func (rm *ResourceMonitor) CreateFileProcessingContext(parent context.Context) (context.Context, context.CancelFunc) {
	if !rm.enabled || rm.fileProcessingTimeout <= 0 {
		// No-op cancel function - monitoring disabled or no timeout configured
		return parent, func() {}
	}

	return context.WithTimeout(parent, rm.fileProcessingTimeout)
}

// CreateOverallProcessingContext creates a context with overall processing timeout.
func (rm *ResourceMonitor) CreateOverallProcessingContext(parent context.Context) (
	context.Context,
	context.CancelFunc,
) {
	if !rm.enabled || rm.overallTimeout <= 0 {
		// No-op cancel function - monitoring disabled or no timeout configured
		return parent, func() {}
	}

	return context.WithTimeout(parent, rm.overallTimeout)
}
