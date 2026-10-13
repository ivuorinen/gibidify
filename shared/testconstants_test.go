package shared

// Test fixture constants used only by this package's tests. The shared fixture
// set lives in testutil, which these internal tests cannot import (testutil
// imports shared).
const (
	// TestContentEmpty is empty content test string.
	TestContentEmpty = "empty content"
	// TestContentHelloWorld is hello world test string.
	TestContentHelloWorld = "hello world"
	// TestContentTest is generic test content string.
	TestContentTest = "test content"
	// TestErrAccessDenied is an access denied error message.
	TestErrAccessDenied = "access denied"
	// TestErrDiskFull is a disk full error message.
	TestErrDiskFull = "disk full"
	// TestFileOutputTXT is an output.txt test file name.
	TestFileOutputTXT = "output.txt"
	// TestFmtExpectedCode is format string for code assertions.
	TestFmtExpectedCode = "Expected Code %q, got %q"
	// TestFmtExpectedContent is format string for content assertions.
	TestFmtExpectedContent = "Expected content %q, got %q"
	// TestFmtExpectedErrorTypeIO is format string for error type IO assertions.
	TestFmtExpectedErrorTypeIO = "Expected ErrorTypeIO, got %v"
	// TestFmtExpectedFilePath is format string for file path assertions.
	TestFmtExpectedFilePath = "Expected FilePath %q, got %q"
	// TestFmtExpectedMessage is format string for message assertions.
	TestFmtExpectedMessage = "Expected Message %q, got %q"
	// TestFmtExpectedType is format string for type assertions.
	TestFmtExpectedType = "Expected Type %v, got %v"
	// TestLoggerDebugMsg is a debug level test message.
	TestLoggerDebugMsg = "debug message"
	// TestLoggerInfoMsg is an info level test message.
	TestLoggerInfoMsg = "info message"
	// TestLoggerWarnMsg is a warn level test message.
	TestLoggerWarnMsg = "warn message"
	// TestMsgEmptyPath is used for empty path test cases.
	TestMsgEmptyPath = "empty path"
	// TestMsgFailedToCreateTestDir is used for test directory creation failures.
	TestMsgFailedToCreateTestDir = "Failed to create test directory: %v"
	// TestPathTestEmptyTXT is a test empty.txt path.
	TestPathTestEmptyTXT = "/test/empty.txt"
	// TestPathTestFileTXT is a test file.txt path.
	TestPathTestFileTXT = "/test/file.txt"
	// TestPathTmpNetworkData is a temp network data path.
	TestPathTmpNetworkData = "/tmp/network.data"
)
