// Package shared provides common constants used across the gibidify application.
package shared

// Byte Conversion Constants
const (
	// BytesPerKB is the number of bytes in a kilobyte (1024).
	BytesPerKB = 1024
	// BytesPerMB is the number of bytes in a megabyte (1024 * 1024).
	BytesPerMB = 1024 * BytesPerKB
	// BytesPerGB is the number of bytes in a gigabyte (1024 * 1024 * 1024).
	BytesPerGB = 1024 * BytesPerMB
)

// Configuration Default Values - Numeric Constants
const (
	// ConfigFileSizeLimitDefault is the default maximum file size (5MB).
	ConfigFileSizeLimitDefault = 5 * BytesPerMB
	// ConfigFileSizeLimitMin is the minimum allowed file size limit (1KB).
	ConfigFileSizeLimitMin = BytesPerKB
	// ConfigFileSizeLimitMax is the maximum allowed file size limit (100MB).
	ConfigFileSizeLimitMax = 100 * BytesPerMB

	// ConfigMaxFilesDefault is the default maximum number of files to process.
	ConfigMaxFilesDefault = 10000
	// ConfigMaxFilesMin is the minimum allowed file count limit.
	ConfigMaxFilesMin = 1
	// ConfigMaxFilesMax is the maximum allowed file count limit.
	ConfigMaxFilesMax = 1000000

	// ConfigMaxTotalSizeDefault is the default maximum total size of files (1GB).
	ConfigMaxTotalSizeDefault = BytesPerGB
	// ConfigMaxTotalSizeMin is the minimum allowed total size limit (1MB).
	ConfigMaxTotalSizeMin = BytesPerMB
	// ConfigMaxTotalSizeMax is the maximum allowed total size limit (100GB).
	ConfigMaxTotalSizeMax = 100 * BytesPerGB

	// ConfigMaxConcurrencyDefault is the default maximum concurrency (high enough for typical systems).
	ConfigMaxConcurrencyDefault = 32
)

// Configuration Default Values - Boolean Constants
const (
	// ConfigFileTypesEnabledDefault is the default state for file type detection.
	ConfigFileTypesEnabledDefault = true
)

// Configuration Keys - Viper Path Constants
const (
	// ConfigKeyFileSizeLimit is the config key for file size limit.
	ConfigKeyFileSizeLimit = "fileSizeLimit"
	// ConfigKeyMaxConcurrency is the config key for max concurrency.
	ConfigKeyMaxConcurrency = "maxConcurrency"
	// ConfigKeyIgnoreDirectories is the config key for ignored directories.
	ConfigKeyIgnoreDirectories = "ignoreDirectories"

	// ConfigKeyFileTypesEnabled is the config key for fileTypes.enabled.
	ConfigKeyFileTypesEnabled = "fileTypes.enabled"
	// ConfigKeyFileTypesCustomImageExtensions is the config key for fileTypes.customImageExtensions.
	ConfigKeyFileTypesCustomImageExtensions = "fileTypes.customImageExtensions"
	// ConfigKeyFileTypesCustomBinaryExtensions is the config key for fileTypes.customBinaryExtensions.
	ConfigKeyFileTypesCustomBinaryExtensions = "fileTypes.customBinaryExtensions"
	// ConfigKeyFileTypesCustomLanguages is the config key for fileTypes.customLanguages.
	ConfigKeyFileTypesCustomLanguages = "fileTypes.customLanguages"
	// ConfigKeyFileTypesDisabledImageExtensions is the config key for fileTypes.disabledImageExtensions.
	ConfigKeyFileTypesDisabledImageExtensions = "fileTypes.disabledImageExtensions"
	// ConfigKeyFileTypesDisabledBinaryExtensions is the config key for fileTypes.disabledBinaryExtensions.
	ConfigKeyFileTypesDisabledBinaryExtensions = "fileTypes.disabledBinaryExtensions"
	// ConfigKeyFileTypesDisabledLanguageExts is the config key for fileTypes.disabledLanguageExtensions.
	ConfigKeyFileTypesDisabledLanguageExts = "fileTypes.disabledLanguageExtensions"

	// ConfigKeyResourceLimitsMaxFiles is the config key for resourceLimits.maxFiles.
	ConfigKeyResourceLimitsMaxFiles = "resourceLimits.maxFiles"
	// ConfigKeyResourceLimitsMaxTotalSize is the config key for resourceLimits.maxTotalSize.
	ConfigKeyResourceLimitsMaxTotalSize = "resourceLimits.maxTotalSize"
)

// Configuration Collections - Slice and Map Variables
var (
	// ConfigIgnoredDirectoriesDefault is the default list of directories to ignore.
	ConfigIgnoredDirectoriesDefault = []string{
		"vendor", "node_modules", ".git", "dist", "build", "target",
		"bower_components", "cache", "tmp",
	}

	// ConfigCustomImageExtensionsDefault is the default list of custom image extensions.
	ConfigCustomImageExtensionsDefault = []string{}

	// ConfigCustomBinaryExtensionsDefault is the default list of custom binary extensions.
	ConfigCustomBinaryExtensionsDefault = []string{}

	// ConfigDisabledImageExtensionsDefault is the default list of disabled image extensions.
	ConfigDisabledImageExtensionsDefault = []string{}

	// ConfigDisabledBinaryExtensionsDefault is the default list of disabled binary extensions.
	ConfigDisabledBinaryExtensionsDefault = []string{}

	// ConfigDisabledLanguageExtensionsDefault is the default list of disabled language extensions.
	ConfigDisabledLanguageExtensionsDefault = []string{}

	// ConfigCustomLanguagesDefault is the default custom language mappings.
	ConfigCustomLanguagesDefault = map[string]string{}
)

// Test Paths and Files
const (
	// TestSourcePath is a common test source directory path.
	TestSourcePath = "/test/source"
	// TestOutputMarkdown is a common test output markdown file path.
	TestOutputMarkdown = "/test/output.md"
	// TestFile1 is a common test filename.
	TestFile1 = "file1.txt"
	// TestFile2 is a common test filename.
	TestFile2 = "file2.txt"
	// TestOutputMD is a common output markdown filename.
	TestOutputMD = "output.md"
	// TestMD is a common markdown test file.
	TestMD = "test.md"
)

// Test CLI Flags
const (
	// TestCLIFlagSource is the -source flag.
	TestCLIFlagSource = "-source"
	// TestCLIFlagDestination is the -destination flag.
	TestCLIFlagDestination = "-destination"
	// TestCLIFlagFormat is the -format flag.
	TestCLIFlagFormat = "-format"
	// TestCLIFlagNoUI is the -no-ui flag.
	TestCLIFlagNoUI = "-no-ui"
	// TestCLIFlagConcurrency is the -concurrency flag.
	TestCLIFlagConcurrency = "-concurrency"
)

// Test Content Strings
const (
	// TestContent is common test file content.
	TestContent = "Hello World"
	// TestSharedGoContent is content for shared.go test files.
	TestSharedGoContent = "package main\n\nfunc Helper() {}"
	// TestContentTest is generic test content string.
	TestContentTest = "test content"
	// TestContentEmpty is empty content test string.
	TestContentEmpty = "empty content"
	// TestContentHelloWorld is hello world test string.
	TestContentHelloWorld = "hello world"
	// TestContentDocumentation is documentation test string.
	TestContentDocumentation = "# Documentation"
	// TestContentPackageHandlers is package handlers test string.
	TestContentPackageHandlers = "package handlers"
)

// Test Error Messages
const (
	// TestMsgExpectedError is used when an error was expected but none occurred.
	TestMsgExpectedError = "Expected error but got none"
	// TestMsgErrorShouldContain is used to check if error message contains expected text.
	TestMsgErrorShouldContain = "Error should contain %q, got: %v"
	// TestMsgUnexpectedError is used when an unexpected error occurred.
	TestMsgUnexpectedError = "Unexpected error: %v"
	// TestMsgFailedToCreateFile is used for file creation failures.
	TestMsgFailedToCreateFile = "Failed to create temp file: %v"
	// TestMsgFailedToRemoveTempFile is used for temp file removal failures.
	TestMsgFailedToRemoveTempFile = "Failed to remove temp file: %v"
	// TestMsgOutputMissingSubstring is used when output doesn't contain expected text.
	TestMsgOutputMissingSubstring = "Output missing expected substring: %q\nFull output:\n%s"
	// TestMsgOperationFailed is used when an operation fails.
	TestMsgOperationFailed = "Operation %s failed: %v"
	// TestMsgOperationNoError is used when an operation expected error but got none.
	TestMsgOperationNoError = "Operation %s expected error but got none"
	// TestMsgTimeoutWriterCompletion is used for writer timeout errors.
	TestMsgTimeoutWriterCompletion = "timeout waiting for writer completion (doneCh)"
	// TestMsgFailedToCreateTestDir is used for test directory creation failures.
	TestMsgFailedToCreateTestDir = "Failed to create test directory: %v"
	// TestMsgFailedToCreateTestFile is used for test file creation failures.
	TestMsgFailedToCreateTestFile = "Failed to create test file: %v"
	// TestMsgFailedToWriteContent is used for content write failures.
	TestMsgFailedToWriteContent = "Failed to write content: %v"
	// TestMsgFailedToCloseFile is used for file close failures.
	TestMsgFailedToCloseFile = "Failed to close temp file: %v"
)

// Test UI Strings
const (
	// TestSuggestionsPlain is the plain suggestions header without emoji.
	TestSuggestionsPlain = "Suggestions:"
	// TestSuggestionsWarning is the warning-style suggestions header.
	TestSuggestionsWarning = "⚠ Suggestions:"
	// TestSuggestCheckPermissions suggests checking file permissions.
	TestSuggestCheckPermissions = "Check file/directory permissions"
	// TestSuggestCheckArguments suggests checking command line arguments.
	TestSuggestCheckArguments = "Check your command line arguments"
	// TestSuggestVerifyPath suggests verifying the path.
	TestSuggestVerifyPath = "Verify the path is correct"
	// TestSuggestCheckExists suggests checking if path exists.
	TestSuggestCheckExists = "Check if the path exists:"
	// TestSuggestCheckFileExists suggests checking if file/directory exists.
	TestSuggestCheckFileExists = "Check if the file/directory exists:"
	// TestSuggestUseAbsolutePath suggests using absolute paths.
	TestSuggestUseAbsolutePath = "Use an absolute path instead of relative"
)

// Test Error Strings and Categories
const (
	// TestErrEmptyFilePath is error message for empty file paths.
	TestErrEmptyFilePath = "empty file path"
	// TestErrTestErrorMsg is a generic test error message string.
	TestErrTestErrorMsg = "test error"
	// TestErrDiskFull is a disk full error message.
	TestErrDiskFull = "disk full"
	// TestErrAccessDenied is an access denied error message.
	TestErrAccessDenied = "access denied"
	// TestErrCannotAccessFile is an error message for file access errors.
	TestErrCannotAccessFile = "cannot access file"
)

// Test Terminal and UI Strings
const (
	// TestTerminalXterm256 is a common terminal type for testing.
	TestTerminalXterm256 = "xterm-256color"
	// TestProgressMessage is a common progress message.
	TestProgressMessage = "Processing files"
)

// Test Logger Messages
const (
	// TestLoggerDebugMsg is a debug level test message.
	TestLoggerDebugMsg = "debug message"
	// TestLoggerInfoMsg is an info level test message.
	TestLoggerInfoMsg = "info message"
	// TestLoggerWarnMsg is a warn level test message.
	TestLoggerWarnMsg = "warn message"
)

// Test Assertion Case Names
const (
	// TestCaseSuccessCases is the name for success test cases.
	TestCaseSuccessCases = "success cases"
	// TestCaseEmptyOperationName is the name for empty operation test cases.
	TestCaseEmptyOperationName = "empty operation name"
	// TestCaseDifferentErrorTypes is the name for different error types test cases.
	TestCaseDifferentErrorTypes = "different error types"
	// TestCaseFunctionAvailability is the name for function availability test cases.
	TestCaseFunctionAvailability = "function availability"
	// TestCaseMessageTest is the name for message test cases.
	TestCaseMessageTest = "message test"
	// TestCaseTestOperation is the name for test operation cases.
	TestCaseTestOperation = "test operation"
)

// Test File Extensions and Special Names
const (
	// TestExtensionSpecial is a special extension for testing.
	TestExtensionSpecial = ".SPECIAL"
	// TestExtensionValid is a valid extension for testing custom extensions.
	TestExtensionValid = ".valid"
	// TestExtensionCustom is a custom extension for testing.
	TestExtensionCustom = ".custom"
)

// Test Paths
const (
	// TestPathBase is a base test path.
	TestPathBase = "/test/path"
	// TestPathTestFileTXT is a test file.txt path.
	TestPathTestFileTXT = "/test/file.txt"
	// TestPathTestEmptyTXT is a test empty.txt path.
	TestPathTestEmptyTXT = "/test/empty.txt"
	// TestPathTmpNetworkData is a temp network data path.
	TestPathTmpNetworkData = "/tmp/network.data"
	// TestPathEtcPasswdTraversal is a path traversal test path.
	TestPathEtcPasswdTraversal = "../../../etc/passwd" // #nosec G101 -- test constant, not credentials
)

// Test File Names
const (
	// TestFileTXT is a common test file name.
	TestFileTXT = "test.txt"
	// TestFileGo is a common Go test file name.
	TestFileGo = "test.go"
	// TestFileSharedGo is a common shared Go file name.
	TestFileSharedGo = "shared.go"
	// TestFilePNG is a PNG test file name.
	TestFilePNG = "test.png"
	// TestFileJPG is a JPG test file name.
	TestFileJPG = "test.jpg"
	// TestFileEXE is an EXE test file name.
	TestFileEXE = "test.exe"
	// TestFileDLL is a DLL test file name.
	TestFileDLL = "test.dll"
	// TestFilePy is a Python test file name.
	TestFilePy = "test.py"
	// TestFileValid is a test file with .valid extension.
	TestFileValid = "test.valid"
	// TestFileWebP is a WebP test file name.
	TestFileWebP = "test.webp"
	// TestFileImageJPG is a JPG test file name.
	TestFileImageJPG = "image.jpg"
	// TestFileBinaryDLL is a DLL test file name.
	TestFileBinaryDLL = "binary.dll"
	// TestFileScriptPy is a Python script test file name.
	TestFileScriptPy = "script.py"
	// TestFileMainGo is a main.go test file name.
	TestFileMainGo = "main.go"
	// TestFileHelperGo is a helper.go test file name.
	TestFileHelperGo = "helper.go"
	// TestFileConfigJSON is a config.json test file name.
	TestFileConfigJSON = "config.json"
	// TestFileReadmeMD is a README.md test file name.
	TestFileReadmeMD = "README.md"
	// TestFileOutputTXT is an output.txt test file name.
	TestFileOutputTXT = "output.txt"
	// TestFileConfigYAML is a config.yaml test file name.
	TestFileConfigYAML = "config.yaml"
)

// Test Validation and Operation Strings
const (
	// TestOpParsingFlags is used in error messages for flag parsing operations.
	TestOpParsingFlags = "parsing flags"
	// TestOpValidatingConcurrency is used for concurrency validation.
	TestOpValidatingConcurrency = "validating concurrency"
	// TestKeyName is a common test key name.
	TestKeyName = "test.key"
	// TestMsgExpectedExtensionWithoutDot is error message for extension validation.
	TestMsgExpectedExtensionWithoutDot = "Expected extension without dot to not work"
	// TestMsgSourcePath is the validation message for source path.
	TestMsgSourcePath = "source path"
	// TestMsgEmptyPath is used for empty path test cases.
	TestMsgEmptyPath = "empty path"
	// TestMsgPathTraversalAttempt is used for path traversal detection tests.
	TestMsgPathTraversalAttempt = "path traversal attempt detected"
)

// Test Structured Error Format Strings
const (
	// TestFmtExpectedFilePath is format string for file path assertions.
	TestFmtExpectedFilePath = "Expected FilePath %q, got %q"
	// TestFmtExpectedType is format string for type assertions.
	TestFmtExpectedType = "Expected Type %v, got %v"
	// TestFmtExpectedCode is format string for code assertions.
	TestFmtExpectedCode = "Expected Code %q, got %q"
	// TestFmtExpectedMessage is format string for message assertions.
	TestFmtExpectedMessage = "Expected Message %q, got %q"
	// TestFmtExpectedCount is format string for count assertions.
	TestFmtExpectedCount = "Expected %d %s, got %d"
	// TestFmtExpectedContent is format string for content assertions.
	TestFmtExpectedContent = "Expected content %q, got %q"
	// TestFmtExpectedErrorTypeIO is format string for error type IO assertions.
	TestFmtExpectedErrorTypeIO = "Expected ErrorTypeIO, got %v"
	// TestFmtDirectoryShouldExist is format string for directory existence assertions.
	TestFmtDirectoryShouldExist = "Directory %s should exist: %v"
	// TestFmtPathShouldBeDirectory is format string for directory type assertions.
	TestFmtPathShouldBeDirectory = "Path %s should be a directory"
)

// CLI Error Messages
const (
	// CLIMsgErrorFormat is the error message format.
	CLIMsgErrorFormat = "Error: %s"
	// CLIMsgSuggestions is the suggestions header.
	CLIMsgSuggestions = "Suggestions:"
	// CLIMsgCheckFilePermissions suggests checking file permissions.
	CLIMsgCheckFilePermissions = "  • Check file/directory permissions\n"
	// CLIMsgCheckCommandLineArgs suggests checking command line arguments.
	CLIMsgCheckCommandLineArgs = "  • Check your command line arguments\n"
	// CLIMsgRunWithHelp suggests running with help flag.
	CLIMsgRunWithHelp = "  • Run with --help for usage information\n"
)

// CLI Processing Messages
const (
	// CLIMsgFoundFilesToProcess is the message format when files are found to process.
	CLIMsgFoundFilesToProcess = "Found %d files to process"
	// CLIMsgFileProcessingWorker is the worker identifier for file processing.
	CLIMsgFileProcessingWorker = "file processing worker"
)

// CLI UI Constants
const (
	// UIProgressBarChar is the character used for progress bar display.
	UIProgressBarChar = "█"
)

// Error Format Strings
const (
	// ErrorFmtWithCause is the format string for errors with cause information.
	ErrorFmtWithCause = "%s: %v"
	// LogLevelWarningAlias is an alias for the warning log level used in validation.
	LogLevelWarningAlias = "warning"
)

// File Processing Error Messages
const (
	// FileProcessingMsgFailedToProcess is the error message format for processing failures.
	FileProcessingMsgFailedToProcess = "Failed to process file: %s"
	// FileProcessingMsgSizeExceeds is the error message when file size exceeds limit.
	FileProcessingMsgSizeExceeds = "file size (%d bytes) exceeds limit (%d bytes)"
)

// ============================================================================
// YAML WRITER FORMATS
// ============================================================================

const (
	// YAMLFmtFileEntry is the format string for YAML file entries.
	YAMLFmtFileEntry = "  - path: %s\n    language: %s\n    content: |\n"
)

// ============================================================================
// YAML/STRING LITERAL VALUES
// ============================================================================

const (
	// LiteralTrue is the string literal "true" used in YAML/env comparisons.
	LiteralTrue = "true"
	// LiteralFalse is the string literal "false" used in YAML/env comparisons.
	LiteralFalse = "false"
	// LiteralNull is the string literal "null" used in YAML comparisons.
	LiteralNull = "null"
	// LiteralPackageMain is the string literal "package main" used in test files.
	LiteralPackageMain = "package main"
)

// Test File Permissions
const (
	// TestFilePermission is the default file permission for test files.
	TestFilePermission = 0o644
	// TestDirPermission is the default directory permission for test directories.
	TestDirPermission = 0o755
)

// Log Level Constants
const (
	// LogLevelDebug logs all messages including debug information.
	LogLevelDebug LogLevel = "debug"
	// LogLevelInfo logs info, warning, and error messages.
	LogLevelInfo LogLevel = "info"
	// LogLevelWarn logs warning and error messages only.
	LogLevelWarn LogLevel = "warn"
	// LogLevelError logs error messages only.
	LogLevelError LogLevel = "error"
)

// ============================================================================
// FORMAT CONSTANTS
// ============================================================================

const (
	// FormatJSON is the JSON format identifier.
	FormatJSON = "json"
	// FormatYAML is the YAML format identifier.
	FormatYAML = "yaml"
	// FormatMarkdown is the Markdown format identifier.
	FormatMarkdown = "markdown"
)

// ============================================================================
// CLI ARGUMENT NAMES
// ============================================================================

const (
	// CLIArgSource is the source argument name.
	CLIArgSource = "source"
	// CLIArgFormat is the format argument name.
	CLIArgFormat = "format"
	// CLIArgConcurrency is the concurrency argument name.
	CLIArgConcurrency = "concurrency"
)

// ============================================================================
// APPLICATION CONSTANTS
// ============================================================================

const (
	// AppName is the application name.
	AppName = "gibidify"
)
