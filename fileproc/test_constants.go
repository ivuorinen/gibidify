package fileproc

// Test constants to avoid duplication in test files.
// These constants are used across multiple test files in the fileproc package.
const (
	// Backpressure configuration keys
	testBackpressureEnabled     = "backpressure.enabled"
	testBackpressureMaxMemory   = "backpressure.maxMemoryUsage"
	testBackpressureMemoryCheck = "backpressure.memoryCheckInterval"
	testBackpressureMaxFiles    = "backpressure.maxPendingFiles"
	testBackpressureMaxWrites   = "backpressure.maxPendingWrites"
)
