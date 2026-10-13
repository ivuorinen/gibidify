// Package shared provides common constants used across the gibidify application.
// Test fixtures live in the testutil package, not here, so they are not
// compiled into the binary.
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
)

// Configuration Default Values - Boolean Constants
const (
	// ConfigFileTypesEnabledDefault is the default state for file type detection.
	ConfigFileTypesEnabledDefault = true
)

// Configuration Keys - dot-separated paths into config.yaml
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
	// ConfigIgnoredDirectoriesDefault is the default list of directories to
	// ignore. VCS metadata (.git, .hg, .svn) is skipped unconditionally by the
	// walker and is not part of this replaceable list.
	ConfigIgnoredDirectoriesDefault = []string{
		"vendor", "node_modules", "dist", "build", "target",
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
)

// Output file permissions
const (
	// OutputFilePermission is the mode of a newly created output file. The
	// output holds source code, so it is readable by its owner only.
	OutputFilePermission = 0o600
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
	// FileProcessingMsgSizeExceeds is the error message when file size exceeds limit.
	FileProcessingMsgSizeExceeds = "file size (%d bytes) exceeds limit (%d bytes)"
)

// ============================================================================
// STRING LITERAL VALUES
// ============================================================================

const (
	// LiteralTrue is the string literal "true" used in env comparisons.
	LiteralTrue = "true"
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
