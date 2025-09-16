// Package fileproc handles file processing, collection, and output formatting.
package fileproc

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

// Close cleans up the resource monitor.
func (rm *ResourceMonitor) Close() {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	// Prevent multiple closes
	if rm.closed {
		return
	}
	rm.closed = true

	// Signal goroutines to stop
	if rm.done != nil {
		close(rm.done)
	}

	// Stop the ticker
	if rm.rateLimiter != nil {
		rm.rateLimiter.Stop()
	}
}
