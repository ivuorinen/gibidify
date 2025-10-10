package cli

// Test constants to avoid duplication in test files.
// These constants are used across multiple test files in the cli package.
const (
	// Error messages
	testErrFileNotFound     = "file not found"
	testErrPermissionDenied = "permission denied"
	testErrInvalidFormat    = "invalid format"
	testErrOther            = "other error"
	testErrEncoding         = "encoding error"
	testErrSourceRequired   = "source directory is required"
	testErrPathTraversal    = "path traversal attempt detected"
	testPathTraversalPath   = "../../../etc/passwd"

	// Suggestion messages
	testSuggestionsHeader   = "Suggestions:"
	testSuggestCheckPerms   = "Check file/directory permissions"
	testSuggestVerifyPath   = "Verify the path is correct"
	testSuggestFormat       = "Use a supported format: markdown, json, yaml"
	testSuggestFormatEx     = "Example: -format markdown"
	testSuggestCheckArgs    = "Check your command line arguments"
	testSuggestHelp         = "Run with --help for usage information"
	testSuggestDiskSpace    = "Verify available disk space"
	testSuggestReduceConcur = "Try with -concurrency 1 to reduce resource usage"

	// UI test strings
	testWithColors    = "with colors"
	testWithoutColors = "without colors"
	testProcessingMsg = "Processing files"

	// Flag names
	testFlagSource      = "-source"
	testFlagConcurrency = "-concurrency"

	// Test file paths
	testFilePath1 = "/test/file1.go"
	testFilePath2 = "/test/file2.go"

	// Output markers
	testErrorSuffix = " Error"
)
