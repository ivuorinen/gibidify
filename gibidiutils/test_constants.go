package gibidiutils

// Test constants to avoid duplication in test files.
// These constants are used across multiple test files in the gibidiutils package.
const (
	// Error messages

	testErrFileNotFound  = "file not found"
	testErrWriteFailed   = "write failed"
	testErrInvalidFormat = "invalid format"

	// Path validation messages

	testEmptyPath             = "empty path"
	testPathTraversal         = "path traversal"
	testPathTraversalAttempt  = "path traversal attempt"
	testPathTraversalDetected = "path traversal attempt detected"
)
