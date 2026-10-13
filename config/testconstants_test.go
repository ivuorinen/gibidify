package config

// Test fixture constants for this package's internal tests, which cannot
// import testutil (testutil imports config).
const (
	// TestExtensionCustom is a custom extension for testing.
	TestExtensionCustom = ".custom"
	// TestFmtExpectedCount is format string for count assertions.
	TestFmtExpectedCount = "Expected %d %s, got %d"
)
