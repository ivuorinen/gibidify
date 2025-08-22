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
	if rm.rateLimiter != nil {
		rm.rateLimiter.Stop()
	}
}
