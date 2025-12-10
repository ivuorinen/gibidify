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

	// ConfigFileProcessingTimeoutSecDefault is the default timeout for individual file processing (30 seconds).
	ConfigFileProcessingTimeoutSecDefault = 30
	// ConfigFileProcessingTimeoutSecMin is the minimum allowed file processing timeout (1 second).
	ConfigFileProcessingTimeoutSecMin = 1
	// ConfigFileProcessingTimeoutSecMax is the maximum allowed file processing timeout (300 seconds).
	ConfigFileProcessingTimeoutSecMax = 300

	// ConfigOverallTimeoutSecDefault is the default timeout for overall processing (3600 seconds = 1 hour).
	ConfigOverallTimeoutSecDefault = 3600
	// ConfigOverallTimeoutSecMin is the minimum allowed overall timeout (10 seconds).
	ConfigOverallTimeoutSecMin = 10
	// ConfigOverallTimeoutSecMax is the maximum allowed overall timeout (86400 seconds = 24 hours).
	ConfigOverallTimeoutSecMax = 86400

	// ConfigMaxConcurrentReadsDefault is the default maximum concurrent file reading operations.
	ConfigMaxConcurrentReadsDefault = 10
	// ConfigMaxConcurrentReadsMin is the minimum allowed concurrent reads.
	ConfigMaxConcurrentReadsMin = 1
	// ConfigMaxConcurrentReadsMax is the maximum allowed concurrent reads.
	ConfigMaxConcurrentReadsMax = 100

	// ConfigRateLimitFilesPerSecDefault is the default rate limit for file processing (0 = disabled).
	ConfigRateLimitFilesPerSecDefault = 0
	// ConfigRateLimitFilesPerSecMin is the minimum rate limit.
	ConfigRateLimitFilesPerSecMin = 0
	// ConfigRateLimitFilesPerSecMax is the maximum rate limit.
	ConfigRateLimitFilesPerSecMax = 10000

	// ConfigHardMemoryLimitMBDefault is the default hard memory limit (512MB).
	ConfigHardMemoryLimitMBDefault = 512
	// ConfigHardMemoryLimitMBMin is the minimum hard memory limit (64MB).
	ConfigHardMemoryLimitMBMin = 64
	// ConfigHardMemoryLimitMBMax is the maximum hard memory limit (8192MB = 8GB).
	ConfigHardMemoryLimitMBMax = 8192

	// ConfigMaxPendingFilesDefault is the default maximum files in file channel buffer.
	ConfigMaxPendingFilesDefault = 1000
	// ConfigMaxPendingWritesDefault is the default maximum writes in write channel buffer.
	ConfigMaxPendingWritesDefault = 100
	// ConfigMaxMemoryUsageDefault is the default maximum memory usage (100MB).
	ConfigMaxMemoryUsageDefault = 100 * BytesPerMB
	// ConfigMemoryCheckIntervalDefault is the default memory check interval (every 1000 files).
	ConfigMemoryCheckIntervalDefault = 1000

	// ConfigMaxConcurrencyDefault is the default maximum concurrency (high enough for typical systems).
	ConfigMaxConcurrencyDefault = 32

	// FileTypeRegistryMaxCacheSize is the default maximum cache size for file type registry.
	FileTypeRegistryMaxCacheSize = 500

	// ConfigMarkdownHeaderLevelDefault is the default header level for file sections.
	ConfigMarkdownHeaderLevelDefault = 0
	// ConfigMarkdownMaxLineLengthDefault is the default maximum line length (0 = unlimited).
	ConfigMarkdownMaxLineLengthDefault = 0
)

// Configuration Default Values - Boolean Constants
const (
	// ConfigFileTypesEnabledDefault is the default state for file type detection.
	ConfigFileTypesEnabledDefault = true

	// ConfigBackpressureEnabledDefault is the default state for backpressure.
	ConfigBackpressureEnabledDefault = true

	// ConfigResourceLimitsEnabledDefault is the default state for resource limits.
	ConfigResourceLimitsEnabledDefault = true
	// ConfigEnableGracefulDegradationDefault is the default state for graceful degradation.
	ConfigEnableGracefulDegradationDefault = true
	// ConfigEnableResourceMonitoringDefault is the default state for resource monitoring.
	ConfigEnableResourceMonitoringDefault = true

	// ConfigMetadataIncludeStatsDefault is the default for including stats in metadata.
	ConfigMetadataIncludeStatsDefault = false
	// ConfigMetadataIncludeTimestampDefault is the default for including timestamp.
	ConfigMetadataIncludeTimestampDefault = false
	// ConfigMetadataIncludeFileCountDefault is the default for including file count.
	ConfigMetadataIncludeFileCountDefault = false
	// ConfigMetadataIncludeSourcePathDefault is the default for including source path.
	ConfigMetadataIncludeSourcePathDefault = false
	// ConfigMetadataIncludeFileTypesDefault is the default for including file types.
	ConfigMetadataIncludeFileTypesDefault = false
	// ConfigMetadataIncludeProcessingTimeDefault is the default for including processing time.
	ConfigMetadataIncludeProcessingTimeDefault = false
	// ConfigMetadataIncludeTotalSizeDefault is the default for including total size.
	ConfigMetadataIncludeTotalSizeDefault = false
	// ConfigMetadataIncludeMetricsDefault is the default for including metrics.
	ConfigMetadataIncludeMetricsDefault = false

	// ConfigMarkdownUseCodeBlocksDefault is the default for using code blocks.
	ConfigMarkdownUseCodeBlocksDefault = false
	// ConfigMarkdownIncludeLanguageDefault is the default for including language in code blocks.
	ConfigMarkdownIncludeLanguageDefault = false
	// ConfigMarkdownTableOfContentsDefault is the default for table of contents.
	ConfigMarkdownTableOfContentsDefault = false
	// ConfigMarkdownUseCollapsibleDefault is the default for collapsible sections.
	ConfigMarkdownUseCollapsibleDefault = false
	// ConfigMarkdownSyntaxHighlightingDefault is the default for syntax highlighting.
	ConfigMarkdownSyntaxHighlightingDefault = false
	// ConfigMarkdownLineNumbersDefault is the default for line numbers.
	ConfigMarkdownLineNumbersDefault = false
	// ConfigMarkdownFoldLongFilesDefault is the default for folding long files.
	ConfigMarkdownFoldLongFilesDefault = false
)

// Configuration Default Values - String Constants
const (
	// ConfigOutputTemplateDefault is the default output template (empty = use built-in).
	ConfigOutputTemplateDefault = ""
	// ConfigMarkdownCustomCSSDefault is the default custom CSS.
	ConfigMarkdownCustomCSSDefault = ""
	// ConfigCustomHeaderDefault is the default custom header template.
	ConfigCustomHeaderDefault = ""
	// ConfigCustomFooterDefault is the default custom footer template.
	ConfigCustomFooterDefault = ""
	// ConfigCustomFileHeaderDefault is the default custom file header template.
	ConfigCustomFileHeaderDefault = ""
	// ConfigCustomFileFooterDefault is the default custom file footer template.
	ConfigCustomFileFooterDefault = ""
)

// Configuration Keys - Viper Path Constants
const (
	// ConfigKeyFileSizeLimit is the config key for file size limit.
	ConfigKeyFileSizeLimit = "fileSizeLimit"
	// ConfigKeyMaxConcurrency is the config key for max concurrency.
	ConfigKeyMaxConcurrency = "maxConcurrency"
	// ConfigKeySupportedFormats is the config key for supported formats.
	ConfigKeySupportedFormats = "supportedFormats"
	// ConfigKeyFilePatterns is the config key for file patterns.
	ConfigKeyFilePatterns = "filePatterns"
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

	// ConfigKeyBackpressureEnabled is the config key for backpressure.enabled.
	ConfigKeyBackpressureEnabled = "backpressure.enabled"
	// ConfigKeyBackpressureMaxPendingFiles is the config key for backpressure.maxPendingFiles.
	ConfigKeyBackpressureMaxPendingFiles = "backpressure.maxPendingFiles"
	// ConfigKeyBackpressureMaxPendingWrites is the config key for backpressure.maxPendingWrites.
	ConfigKeyBackpressureMaxPendingWrites = "backpressure.maxPendingWrites"
	// ConfigKeyBackpressureMaxMemoryUsage is the config key for backpressure.maxMemoryUsage.
	ConfigKeyBackpressureMaxMemoryUsage = "backpressure.maxMemoryUsage"
	// ConfigKeyBackpressureMemoryCheckInt is the config key for backpressure.memoryCheckInterval.
	ConfigKeyBackpressureMemoryCheckInt = "backpressure.memoryCheckInterval"

	// ConfigKeyResourceLimitsEnabled is the config key for resourceLimits.enabled.
	ConfigKeyResourceLimitsEnabled = "resourceLimits.enabled"
	// ConfigKeyResourceLimitsMaxFiles is the config key for resourceLimits.maxFiles.
	ConfigKeyResourceLimitsMaxFiles = "resourceLimits.maxFiles"
	// ConfigKeyResourceLimitsMaxTotalSize is the config key for resourceLimits.maxTotalSize.
	ConfigKeyResourceLimitsMaxTotalSize = "resourceLimits.maxTotalSize"
	// ConfigKeyResourceLimitsFileProcessingTO is the config key for resourceLimits.fileProcessingTimeoutSec.
	ConfigKeyResourceLimitsFileProcessingTO = "resourceLimits.fileProcessingTimeoutSec"
	// ConfigKeyResourceLimitsOverallTO is the config key for resourceLimits.overallTimeoutSec.
	ConfigKeyResourceLimitsOverallTO = "resourceLimits.overallTimeoutSec"
	// ConfigKeyResourceLimitsMaxConcurrentReads is the config key for resourceLimits.maxConcurrentReads.
	ConfigKeyResourceLimitsMaxConcurrentReads = "resourceLimits.maxConcurrentReads"
	// ConfigKeyResourceLimitsRateLimitFilesPerSec is the config key for resourceLimits.rateLimitFilesPerSec.
	ConfigKeyResourceLimitsRateLimitFilesPerSec = "resourceLimits.rateLimitFilesPerSec"
	// ConfigKeyResourceLimitsHardMemoryLimitMB is the config key for resourceLimits.hardMemoryLimitMB.
	ConfigKeyResourceLimitsHardMemoryLimitMB = "resourceLimits.hardMemoryLimitMB"
	// ConfigKeyResourceLimitsEnableGracefulDeg is the config key for resourceLimits.enableGracefulDegradation.
	ConfigKeyResourceLimitsEnableGracefulDeg = "resourceLimits.enableGracefulDegradation"
	// ConfigKeyResourceLimitsEnableMonitoring is the config key for resourceLimits.enableResourceMonitoring.
	ConfigKeyResourceLimitsEnableMonitoring = "resourceLimits.enableResourceMonitoring"

	// ConfigKeyOutputTemplate is the config key for output.template.
	ConfigKeyOutputTemplate = "output.template"
	// ConfigKeyOutputMarkdownHeaderLevel is the config key for output.markdown.headerLevel.
	ConfigKeyOutputMarkdownHeaderLevel = "output.markdown.headerLevel"
	// ConfigKeyOutputMarkdownMaxLineLen is the config key for output.markdown.maxLineLength.
	ConfigKeyOutputMarkdownMaxLineLen = "output.markdown.maxLineLength"
	// ConfigKeyOutputMarkdownCustomCSS is the config key for output.markdown.customCSS.
	ConfigKeyOutputMarkdownCustomCSS = "output.markdown.customCSS"
	// ConfigKeyOutputCustomHeader is the config key for output.custom.header.
	ConfigKeyOutputCustomHeader = "output.custom.header"
	// ConfigKeyOutputCustomFooter is the config key for output.custom.footer.
	ConfigKeyOutputCustomFooter = "output.custom.footer"
	// ConfigKeyOutputCustomFileHeader is the config key for output.custom.fileHeader.
	ConfigKeyOutputCustomFileHeader = "output.custom.fileHeader"
	// ConfigKeyOutputCustomFileFooter is the config key for output.custom.fileFooter.
	ConfigKeyOutputCustomFileFooter = "output.custom.fileFooter"
	// ConfigKeyOutputVariables is the config key for output.variables.
	ConfigKeyOutputVariables = "output.variables"
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

	// ConfigTemplateVariablesDefault is the default template variables.
	ConfigTemplateVariablesDefault = map[string]string{}

	// ConfigSupportedFormatsDefault is the default list of supported output formats.
	ConfigSupportedFormatsDefault = []string{"json", "yaml", "markdown"}

	// ConfigFilePatternsDefault is the default list of file patterns (empty = all files).
	ConfigFilePatternsDefault = []string{}
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
	// TestFile1Name is test1.txt used in benchmark tests.
	TestFile1Name = "test1.txt"
	// TestFile2Name is test2.txt used in benchmark tests.
	TestFile2Name = "test2.txt"
	// TestFile3Name is test3.md used in benchmark tests.
	TestFile3Name = "test3.md"
	// TestFile1Go is a common Go test file path.
	TestFile1Go = "/test/file.go"
	// TestFile1GoAlt is an alternative Go test file path.
	TestFile1GoAlt = "/test/file1.go"
	// TestFile2JS is a common JavaScript test file path.
	TestFile2JS = "/test/file2.js"
	// TestErrorPy is a Python test file path for error scenarios.
	TestErrorPy = "/test/error.py"
	// TestNetworkData is a network data file path for testing.
	TestNetworkData = "/tmp/network.data"
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
	// TestConcurrencyList is a common concurrency list for benchmarks.
	TestConcurrencyList = "1,2,4,8"
	// TestFormatList is a common format list for tests.
	TestFormatList = "json,yaml,markdown"
	// TestSharedGoContent is content for shared.go test files.
	TestSharedGoContent = "package main\n\nfunc Helper() {}"
	// TestSafeConversion is used in safe conversion tests.
	TestSafeConversion = "safe conversion"
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
	// TestMsgFailedToClose is used for file close failures.
	TestMsgFailedToClose = "Failed to close pipe writer: %v"
	// TestMsgFailedToCreateFile is used for file creation failures.
	TestMsgFailedToCreateFile = "Failed to create temp file: %v"
	// TestMsgFailedToRemoveTempFile is used for temp file removal failures.
	TestMsgFailedToRemoveTempFile = "Failed to remove temp file: %v"
	// TestMsgFailedToReadOutput is used for output read failures.
	TestMsgFailedToReadOutput = "Failed to read captured output: %v"
	// TestMsgFailedToCreateTempDir is used for temp directory creation failures.
	TestMsgFailedToCreateTempDir = "Failed to create temp dir: %v"
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
	// TestMsgNewEngineFailed is used when template engine creation fails.
	TestMsgNewEngineFailed = "NewEngine failed: %v"
	// TestMsgRenderFileContentFailed is used when rendering file content fails.
	TestMsgRenderFileContentFailed = "RenderFileContent failed: %v"
	// TestMsgFailedToCreatePipe is used for pipe creation failures.
	TestMsgFailedToCreatePipe = "Failed to create pipe: %v"
	// TestMsgFailedToWriteContent is used for content write failures.
	TestMsgFailedToWriteContent = "Failed to write content: %v"
	// TestMsgFailedToCloseFile is used for file close failures.
	TestMsgFailedToCloseFile = "Failed to close temp file: %v"
	// TestFileStreamTest is a stream test filename.
	TestFileStreamTest = "stream_test.txt"
)

// Test UI Strings
const (
	// TestSuggestionsPlain is the plain suggestions header without emoji.
	TestSuggestionsPlain = "Suggestions:"
	// TestSuggestionsWarning is the warning-style suggestions header.
	TestSuggestionsWarning = "⚠ Suggestions:"
	// TestSuggestionsIcon is the icon-style suggestions header.
	TestSuggestionsIcon = "💡 Suggestions:"
	// TestOutputErrorMarker is the error output marker.
	TestOutputErrorMarker = "❌ Error:"
	// TestOutputSuccessMarker is the success output marker.
	TestOutputSuccessMarker = "✓ Success:"
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
	// TestErrSyntaxError is a syntax error message.
	TestErrSyntaxError = "syntax error"
	// TestErrDiskFull is a disk full error message.
	TestErrDiskFull = "disk full"
	// TestErrAccessDenied is an access denied error message.
	TestErrAccessDenied = "access denied"
	// TestErrProcessingFailed is a processing failed error message.
	TestErrProcessingFailed = "processing failed"
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
	// TestPathTestFileGo is a test file.go path.
	TestPathTestFileGo = "/test/file.go"
	// TestPathTestFileTXT is a test file.txt path.
	TestPathTestFileTXT = "/test/file.txt"
	// TestPathTestErrorGo is a test error.go path.
	TestPathTestErrorGo = "/test/error.go"
	// TestPathTestFile1Go is a test file1.go path.
	TestPathTestFile1Go = "/test/file1.go"
	// TestPathTestFile2JS is a test file2.js path.
	TestPathTestFile2JS = "/test/file2.js"
	// TestPathTestErrorPy is a test error.py path.
	TestPathTestErrorPy = "/test/error.py"
	// TestPathTestEmptyTXT is a test empty.txt path.
	TestPathTestEmptyTXT = "/test/empty.txt"
	// TestPathTestProject is a test project path.
	TestPathTestProject = "/test/project"
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
	// TestFileJSON is a JSON test file name.
	TestFileJSON = "test.json"
	// TestFileConfigJSON is a config.json test file name.
	TestFileConfigJSON = "config.json"
	// TestFileReadmeMD is a README.md test file name.
	TestFileReadmeMD = "README.md"
	// TestFileOutputTXT is an output.txt test file name.
	TestFileOutputTXT = "output.txt"
	// TestFileConfigYAML is a config.yaml test file name.
	TestFileConfigYAML = "config.yaml"
	// TestFileGoExt is a file.go test file name.
	TestFileGoExt = "file.go"
)

// Test Validation and Operation Strings
const (
	// TestOpParsingFlags is used in error messages for flag parsing operations.
	TestOpParsingFlags = "parsing flags"
	// TestOpValidatingConcurrency is used for concurrency validation.
	TestOpValidatingConcurrency = "validating concurrency"
	// TestMsgInvalidConcurrencyLevel is error message for invalid concurrency.
	TestMsgInvalidConcurrencyLevel = "invalid concurrency level"
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
	// TestCfgResourceLimitsEnabled is the config key for resource limits enabled.
	TestCfgResourceLimitsEnabled = "resourceLimits.enabled"
)

// Test Structured Error Format Strings
const (
	// TestFmtExpectedFilePath is format string for file path assertions.
	TestFmtExpectedFilePath = "Expected FilePath %q, got %q"
	// TestFmtExpectedLine is format string for line number assertions.
	TestFmtExpectedLine = "Expected Line %d, got %d"
	// TestFmtExpectedType is format string for type assertions.
	TestFmtExpectedType = "Expected Type %v, got %v"
	// TestFmtExpectedCode is format string for code assertions.
	TestFmtExpectedCode = "Expected Code %q, got %q"
	// TestFmtExpectedMessage is format string for message assertions.
	TestFmtExpectedMessage = "Expected Message %q, got %q"
	// TestFmtExpectedCount is format string for count assertions.
	TestFmtExpectedCount = "Expected %d %s, got %d"
	// TestFmtExpectedGot is generic format string for assertions.
	TestFmtExpectedGot = "%s returned: %v (type: %T)"
	// TestFmtExpectedFilesProcessed is format string for files processed assertion.
	TestFmtExpectedFilesProcessed = "Expected files processed > 0, got %d"
	// TestFmtExpectedResults is format string for results count assertion.
	TestFmtExpectedResults = "Expected %d results, got %d"
	// TestFmtExpectedTotalFiles is format string for total files assertion.
	TestFmtExpectedTotalFiles = "Expected TotalFiles=1, got %d"
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

// File Processing Constants
const (
	// FileProcessingStreamChunkSize is the size of chunks when streaming large files (64KB).
	FileProcessingStreamChunkSize = 64 * BytesPerKB
	// FileProcessingStreamThreshold is the file size above which we use streaming (1MB).
	FileProcessingStreamThreshold = BytesPerMB
	// FileProcessingMaxMemoryBuffer is the maximum memory to use for buffering content (10MB).
	FileProcessingMaxMemoryBuffer = 10 * BytesPerMB
)

// File Processing Error Messages
const (
	// FileProcessingMsgFailedToProcess is the error message format for processing failures.
	FileProcessingMsgFailedToProcess = "Failed to process file: %s"
	// FileProcessingMsgSizeExceeds is the error message when file size exceeds limit.
	FileProcessingMsgSizeExceeds = "file size (%d bytes) exceeds limit (%d bytes)"
)

// Metrics Constants
const (
	// MetricsPhaseCollection represents the collection phase.
	MetricsPhaseCollection = "collection"
	// MetricsPhaseProcessing represents the processing phase.
	MetricsPhaseProcessing = "processing"
	// MetricsPhaseWriting represents the writing phase.
	MetricsPhaseWriting = "writing"
	// MetricsPhaseFinalize represents the finalize phase.
	MetricsPhaseFinalize = "finalize"
	// MetricsMaxInt64 is the maximum int64 value for initial smallest file tracking.
	MetricsMaxInt64 = int64(^uint64(0) >> 1)
	// MetricsPerformanceIndexCap is the maximum performance index value for reasonable indexing.
	MetricsPerformanceIndexCap = 1000
)

// Metrics Format Strings
const (
	// MetricsFmtProcessingTime is the format string for processing time display.
	MetricsFmtProcessingTime = "Processing Time: %v\n"
	// MetricsFmtFileCount is the format string for file count display.
	MetricsFmtFileCount = "  %s: %d files\n"
	// MetricsFmtBytesShort is the format string for bytes without suffix.
	MetricsFmtBytesShort = "%dB"
	// MetricsFmtBytesHuman is the format string for human-readable bytes.
	MetricsFmtBytesHuman = "%.1f%cB"
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

// ============================================================================
// TEMPLATE CONSTANTS
// ============================================================================

const (
	// TemplateFmtTimestamp is the Go time format for timestamps in templates.
	TemplateFmtTimestamp = "2006-01-02 15:04:05"
)

// ============================================================================
// BENCHMARK CONSTANTS
// ============================================================================

const (
	// BenchmarkDefaultFileCount is the default number of files to create for benchmarks.
	BenchmarkDefaultFileCount = 100
	// BenchmarkDefaultIterations is the default number of iterations for benchmarks.
	BenchmarkDefaultIterations = 1000
)

// ============================================================================
// BENCHMARK MESSAGES
// ============================================================================

const (
	// BenchmarkMsgFailedToCreateFiles is the error message when benchmark file creation fails.
	BenchmarkMsgFailedToCreateFiles = "failed to create benchmark files"
	// BenchmarkMsgCollectionFailed is the error message when collection benchmark fails.
	BenchmarkMsgCollectionFailed = "benchmark file collection failed"
	// BenchmarkMsgRunningCollection is the status message when running collection benchmark.
	BenchmarkMsgRunningCollection = "Running file collection benchmark..."
	// BenchmarkMsgFileCollectionFailed is the error message when file collection benchmark fails.
	BenchmarkMsgFileCollectionFailed = "file collection benchmark failed"
	// BenchmarkMsgConcurrencyFailed is the error message when concurrency benchmark fails.
	BenchmarkMsgConcurrencyFailed = "concurrency benchmark failed"
	// BenchmarkMsgFormatFailed is the error message when format benchmark fails.
	BenchmarkMsgFormatFailed = "format benchmark failed"
	// BenchmarkFmtSectionHeader is the format string for benchmark section headers.
	BenchmarkFmtSectionHeader = "=== %s ===\n"
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
	// CLIArgAll is the all benchmarks argument value.
	CLIArgAll = "all"
)

// ============================================================================
// APPLICATION CONSTANTS
// ============================================================================

const (
	// AppName is the application name.
	AppName = "gibidify"
)
