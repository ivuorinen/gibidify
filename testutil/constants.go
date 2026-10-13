// Package testutil provides common testing utilities and helper functions.
package testutil

import "errors"

// Test fixture constants shared by the test suites of several packages. They
// live here, in a test-support package, rather than in the production `shared`
// package, so fixture strings are not compiled into the gibidify binary and
// renaming one cannot change production behaviour.
const (
	// TestCLIFlagConcurrency is the -concurrency flag.
	TestCLIFlagConcurrency = "-concurrency"
	// TestCLIFlagDestination is the -destination flag.
	TestCLIFlagDestination = "-destination"
	// TestCLIFlagFormat is the -format flag.
	TestCLIFlagFormat = "-format"
	// TestCLIFlagNoUI is the -no-ui flag.
	TestCLIFlagNoUI = "-no-ui"
	// TestCLIFlagSource is the -source flag.
	TestCLIFlagSource = "-source"
	// TestCaseDifferentErrorTypes is the name for different error types test cases.
	TestCaseDifferentErrorTypes = "different error types"
	// TestCaseEmptyOperationName is the name for empty operation test cases.
	TestCaseEmptyOperationName = "empty operation name"
	// TestCaseFunctionAvailability is the name for function availability test cases.
	TestCaseFunctionAvailability = "function availability"
	// TestCaseMessageTest is the name for message test cases.
	TestCaseMessageTest = "message test"
	// TestCaseSuccessCases is the name for success test cases.
	TestCaseSuccessCases = "success cases"
	// TestCaseTestOperation is the name for test operation cases.
	TestCaseTestOperation = "test operation"
	// TestContent is common test file content.
	TestContent = "Hello World"
	// TestContentDocumentation is documentation test string.
	TestContentDocumentation = "# Documentation"
	// TestContentPackageHandlers is package handlers test string.
	TestContentPackageHandlers = "package handlers"
	// TestDirPermission is the default directory permission for test directories.
	TestDirPermission = 0o750
	// TestErrCannotAccessFile is an error message for file access errors.
	TestErrCannotAccessFile = "cannot access file"
	// TestErrEmptyFilePath is error message for empty file paths.
	TestErrEmptyFilePath = "empty file path"
	// TestErrTestErrorMsg is a generic test error message string.
	TestErrTestErrorMsg = "test error"
	// TestExtensionCustom is a custom extension for testing.
	TestExtensionCustom = ".custom"
	// TestExtensionSpecial is a special extension for testing.
	TestExtensionSpecial = ".SPECIAL"
	// TestExtensionValid is a valid extension for testing custom extensions.
	TestExtensionValid = ".valid"
	// TestFile1 is a common test filename.
	TestFile1 = "file1.txt"
	// TestFile2 is a common test filename.
	TestFile2 = "file2.txt"
	// TestFileBinaryDLL is a DLL test file name.
	TestFileBinaryDLL = "binary.dll"
	// TestFileConfigJSON is a config.json test file name.
	TestFileConfigJSON = "config.json"
	// TestFileDLL is a DLL test file name.
	TestFileDLL = "test.dll"
	// TestFileEXE is an EXE test file name.
	TestFileEXE = "test.exe"
	// TestFileGo is a common Go test file name.
	TestFileGo = "test.go"
	// TestFileHelperGo is a helper.go test file name.
	TestFileHelperGo = "helper.go"
	// TestFileImageJPG is a JPG test file name.
	TestFileImageJPG = "image.jpg"
	// TestFileJPG is a JPG test file name.
	TestFileJPG = "test.jpg"
	// TestFileMainGo is a main.go test file name.
	TestFileMainGo = "main.go"
	// TestFilePNG is a PNG test file name.
	TestFilePNG = "test.png"
	// TestFilePermission is the default file permission for test files.
	TestFilePermission = 0o600
	// TestFilePy is a Python test file name.
	TestFilePy = "test.py"
	// TestFileReadmeMD is a README.md test file name.
	TestFileReadmeMD = "README.md"
	// TestFileScriptPy is a Python script test file name.
	TestFileScriptPy = "script.py"
	// TestFileSharedGo is a common shared Go file name.
	TestFileSharedGo = "shared.go"
	// TestFileTXT is a common test file name.
	TestFileTXT = "test.txt"
	// TestFileValid is a test file with .valid extension.
	TestFileValid = "test.valid"
	// TestFileWebP is a WebP test file name.
	TestFileWebP = "test.webp"
	// TestFmtDirectoryShouldExist is format string for directory existence assertions.
	TestFmtDirectoryShouldExist = "Directory %s should exist: %v"
	// TestFmtExpectedContent is format string for content assertions.
	TestFmtExpectedContent = "Expected content %q, got %q"
	// TestFmtExpectedCount is format string for count assertions.
	TestFmtExpectedCount = "Expected %d %s, got %d"
	// TestFmtExpectedFilePath is format string for file path assertions.
	TestFmtExpectedFilePath = "Expected FilePath %q, got %q"
	// TestFmtPathShouldBeDirectory is format string for directory type assertions.
	TestFmtPathShouldBeDirectory = "Path %s should be a directory"
	// TestKeyName is a common test key name.
	TestKeyName = "test.key"
	// TestMD is a common markdown test file.
	TestMD = "test.md"
	// TestMsgErrorShouldContain is used to check if error message contains expected text.
	TestMsgErrorShouldContain = "Error should contain %q, got: %v"
	// TestMsgExpectedError is used when an error was expected but none occurred.
	TestMsgExpectedError = "Expected error but got none"
	// TestMsgExpectedExtensionWithoutDot is error message for extension validation.
	TestMsgExpectedExtensionWithoutDot = "Expected extension without dot to not work"
	// TestMsgFailedToCloseFile is used for file close failures.
	TestMsgFailedToCloseFile = "Failed to close temp file: %v"
	// TestMsgFailedToCreateFile is used for file creation failures.
	TestMsgFailedToCreateFile = "Failed to create temp file: %v"
	// TestMsgFailedToCreateTestFile is used for test file creation failures.
	TestMsgFailedToCreateTestFile = "Failed to create test file: %v"
	// TestMsgFailedToRemoveTempFile is used for temp file removal failures.
	TestMsgFailedToRemoveTempFile = "Failed to remove temp file: %v"
	// TestMsgFailedToWriteContent is used for content write failures.
	TestMsgFailedToWriteContent = "Failed to write content: %v"
	// TestMsgOperationFailed is used when an operation fails.
	TestMsgOperationFailed = "Operation %s failed: %v"
	// TestMsgOperationNoError is used when an operation expected error but got none.
	TestMsgOperationNoError = "Operation %s expected error but got none"
	// TestMsgOutputMissingSubstring is used when output doesn't contain expected text.
	TestMsgOutputMissingSubstring = "Output missing expected substring: %q\nFull output:\n%s"
	// TestMsgUnexpectedError is used when an unexpected error occurred.
	TestMsgUnexpectedError = "Unexpected error: %v"
	// TestOpParsingFlags is used in error messages for flag parsing operations.
	TestOpParsingFlags = "parsing flags"
	// TestOpValidatingConcurrency is used for concurrency validation.
	TestOpValidatingConcurrency = "validating concurrency"
	// TestOutputMD is a common output markdown filename.
	TestOutputMD = "output.md"
	// TestOutputMarkdown is a common test output markdown file path.
	TestOutputMarkdown = "/test/output.md"
	// TestPathBase is a base test path.
	TestPathBase = "/test/path"
	// TestProgressMessage is a common progress message.
	TestProgressMessage = "Processing files"
	// TestSharedGoContent is content for shared.go test files.
	TestSharedGoContent = "package main\n\nfunc Helper() {}"
	// TestSourcePath is a common test source directory path.
	TestSourcePath = "/test/source"
	// TestSuggestCheckArguments suggests checking command line arguments.
	TestSuggestCheckArguments = "Check your command line arguments"
	// TestSuggestCheckExists suggests checking if path exists.
	TestSuggestCheckExists = "Check if the path exists:"
	// TestSuggestCheckFileExists suggests checking if file/directory exists.
	TestSuggestCheckFileExists = "Check if the file/directory exists:"
	// TestSuggestCheckPermissions suggests checking file permissions.
	TestSuggestCheckPermissions = "Check file/directory permissions"
	// TestSuggestUseAbsolutePath suggests using absolute paths.
	TestSuggestUseAbsolutePath = "Use an absolute path instead of relative"
	// TestSuggestVerifyPath suggests verifying the path.
	TestSuggestVerifyPath = "Verify the path is correct"
	// TestSuggestionsPlain is the plain suggestions header without emoji.
	TestSuggestionsPlain = "Suggestions:"
	// TestSuggestionsWarning is the warning-style suggestions header.
	TestSuggestionsWarning = "⚠ Suggestions:"
	// TestTerminalXterm256 is a common terminal type for testing.
	TestTerminalXterm256 = "xterm-256color"
	// LiteralPackageMain is the string literal "package main" used in test files.
	LiteralPackageMain = "package main"
)

// ErrTestError is a generic test error.
var ErrTestError = errors.New(TestErrTestErrorMsg)
