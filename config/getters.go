// Package config handles application configuration management.
package config

import (
	"strings"

	"github.com/spf13/viper"

	"github.com/ivuorinen/gibidify/shared"
)

// FileSizeLimit returns the file size limit from configuration.
// Default: ConfigFileSizeLimitDefault (5MB).
func FileSizeLimit() int64 {
	return viper.GetInt64(shared.ConfigKeyFileSizeLimit)
}

// IgnoredDirectories returns the list of directories to ignore.
// Default: ConfigIgnoredDirectoriesDefault.
func IgnoredDirectories() []string {
	return viper.GetStringSlice(shared.ConfigKeyIgnoreDirectories)
}

// MaxConcurrency returns the maximum concurrency level.
// Returns 0 if not set (caller should determine appropriate default).
func MaxConcurrency() int {
	return viper.GetInt(shared.ConfigKeyMaxConcurrency)
}

// SupportedFormats returns the list of supported output formats.
// Returns empty slice if not set.
func SupportedFormats() []string {
	return viper.GetStringSlice(shared.ConfigKeySupportedFormats)
}

// FilePatterns returns the list of file patterns.
// Returns empty slice if not set.
func FilePatterns() []string {
	return viper.GetStringSlice(shared.ConfigKeyFilePatterns)
}

// IsValidFormat checks if the given format is valid.
func IsValidFormat(format string) bool {
	format = strings.ToLower(strings.TrimSpace(format))
	supportedFormats := map[string]bool{
		shared.FormatJSON:     true,
		shared.FormatYAML:     true,
		shared.FormatMarkdown: true,
	}

	return supportedFormats[format]
}

// FileTypesEnabled returns whether file types are enabled.
// Default: ConfigFileTypesEnabledDefault (true).
func FileTypesEnabled() bool {
	return viper.GetBool(shared.ConfigKeyFileTypesEnabled)
}

// CustomImageExtensions returns custom image extensions.
// Default: ConfigCustomImageExtensionsDefault (empty).
func CustomImageExtensions() []string {
	return viper.GetStringSlice(shared.ConfigKeyFileTypesCustomImageExtensions)
}

// CustomBinaryExtensions returns custom binary extensions.
// Default: ConfigCustomBinaryExtensionsDefault (empty).
func CustomBinaryExtensions() []string {
	return viper.GetStringSlice(shared.ConfigKeyFileTypesCustomBinaryExtensions)
}

// CustomLanguages returns custom language mappings.
// Default: ConfigCustomLanguagesDefault (empty).
func CustomLanguages() map[string]string {
	return viper.GetStringMapString(shared.ConfigKeyFileTypesCustomLanguages)
}

// DisabledImageExtensions returns disabled image extensions.
// Default: ConfigDisabledImageExtensionsDefault (empty).
func DisabledImageExtensions() []string {
	return viper.GetStringSlice(shared.ConfigKeyFileTypesDisabledImageExtensions)
}

// DisabledBinaryExtensions returns disabled binary extensions.
// Default: ConfigDisabledBinaryExtensionsDefault (empty).
func DisabledBinaryExtensions() []string {
	return viper.GetStringSlice(shared.ConfigKeyFileTypesDisabledBinaryExtensions)
}

// DisabledLanguageExtensions returns disabled language extensions.
// Default: ConfigDisabledLanguageExtensionsDefault (empty).
func DisabledLanguageExtensions() []string {
	return viper.GetStringSlice(shared.ConfigKeyFileTypesDisabledLanguageExts)
}

// Backpressure getters

// BackpressureEnabled returns whether backpressure is enabled.
// Default: ConfigBackpressureEnabledDefault (true).
func BackpressureEnabled() bool {
	return viper.GetBool(shared.ConfigKeyBackpressureEnabled)
}

// MaxPendingFiles returns the maximum pending files.
// Default: ConfigMaxPendingFilesDefault (1000).
func MaxPendingFiles() int {
	return viper.GetInt(shared.ConfigKeyBackpressureMaxPendingFiles)
}

// MaxPendingWrites returns the maximum pending writes.
// Default: ConfigMaxPendingWritesDefault (100).
func MaxPendingWrites() int {
	return viper.GetInt(shared.ConfigKeyBackpressureMaxPendingWrites)
}

// MaxMemoryUsage returns the maximum memory usage.
// Default: ConfigMaxMemoryUsageDefault (100MB).
func MaxMemoryUsage() int64 {
	return viper.GetInt64(shared.ConfigKeyBackpressureMaxMemoryUsage)
}

// MemoryCheckInterval returns the memory check interval.
// Default: ConfigMemoryCheckIntervalDefault (1000 files).
func MemoryCheckInterval() int {
	return viper.GetInt(shared.ConfigKeyBackpressureMemoryCheckInt)
}

// Resource limits getters

// ResourceLimitsEnabled returns whether resource limits are enabled.
// Default: ConfigResourceLimitsEnabledDefault (true).
func ResourceLimitsEnabled() bool {
	return viper.GetBool(shared.ConfigKeyResourceLimitsEnabled)
}

// MaxFiles returns the maximum number of files.
// Default: ConfigMaxFilesDefault (10000).
func MaxFiles() int {
	return viper.GetInt(shared.ConfigKeyResourceLimitsMaxFiles)
}

// MaxTotalSize returns the maximum total size.
// Default: ConfigMaxTotalSizeDefault (1GB).
func MaxTotalSize() int64 {
	return viper.GetInt64(shared.ConfigKeyResourceLimitsMaxTotalSize)
}

// FileProcessingTimeoutSec returns the file processing timeout in seconds.
// Default: ConfigFileProcessingTimeoutSecDefault (30 seconds).
func FileProcessingTimeoutSec() int {
	return viper.GetInt(shared.ConfigKeyResourceLimitsFileProcessingTO)
}

// OverallTimeoutSec returns the overall timeout in seconds.
// Default: ConfigOverallTimeoutSecDefault (3600 seconds).
func OverallTimeoutSec() int {
	return viper.GetInt(shared.ConfigKeyResourceLimitsOverallTO)
}

// MaxConcurrentReads returns the maximum concurrent reads.
// Default: ConfigMaxConcurrentReadsDefault (10).
func MaxConcurrentReads() int {
	return viper.GetInt(shared.ConfigKeyResourceLimitsMaxConcurrentReads)
}

// RateLimitFilesPerSec returns the rate limit files per second.
// Default: ConfigRateLimitFilesPerSecDefault (0 = disabled).
func RateLimitFilesPerSec() int {
	return viper.GetInt(shared.ConfigKeyResourceLimitsRateLimitFilesPerSec)
}

// HardMemoryLimitMB returns the hard memory limit in MB.
// Default: ConfigHardMemoryLimitMBDefault (512MB).
func HardMemoryLimitMB() int {
	return viper.GetInt(shared.ConfigKeyResourceLimitsHardMemoryLimitMB)
}

// EnableGracefulDegradation returns whether graceful degradation is enabled.
// Default: ConfigEnableGracefulDegradationDefault (true).
func EnableGracefulDegradation() bool {
	return viper.GetBool(shared.ConfigKeyResourceLimitsEnableGracefulDeg)
}

// EnableResourceMonitoring returns whether resource monitoring is enabled.
// Default: ConfigEnableResourceMonitoringDefault (true).
func EnableResourceMonitoring() bool {
	return viper.GetBool(shared.ConfigKeyResourceLimitsEnableMonitoring)
}

// Template system getters

// OutputTemplate returns the selected output template name.
// Default: ConfigOutputTemplateDefault (empty string).
func OutputTemplate() string {
	return viper.GetString(shared.ConfigKeyOutputTemplate)
}

// metadataBool is a helper for metadata boolean configuration values.
// All metadata flags default to false.
func metadataBool(key string) bool {
	return viper.GetBool("output.metadata." + key)
}

// TemplateMetadataIncludeStats returns whether to include stats in metadata.
func TemplateMetadataIncludeStats() bool {
	return metadataBool("includeStats")
}

// TemplateMetadataIncludeTimestamp returns whether to include timestamp in metadata.
func TemplateMetadataIncludeTimestamp() bool {
	return metadataBool("includeTimestamp")
}

// TemplateMetadataIncludeFileCount returns whether to include file count in metadata.
func TemplateMetadataIncludeFileCount() bool {
	return metadataBool("includeFileCount")
}

// TemplateMetadataIncludeSourcePath returns whether to include source path in metadata.
func TemplateMetadataIncludeSourcePath() bool {
	return metadataBool("includeSourcePath")
}

// TemplateMetadataIncludeFileTypes returns whether to include file types in metadata.
func TemplateMetadataIncludeFileTypes() bool {
	return metadataBool("includeFileTypes")
}

// TemplateMetadataIncludeProcessingTime returns whether to include processing time in metadata.
func TemplateMetadataIncludeProcessingTime() bool {
	return metadataBool("includeProcessingTime")
}

// TemplateMetadataIncludeTotalSize returns whether to include total size in metadata.
func TemplateMetadataIncludeTotalSize() bool {
	return metadataBool("includeTotalSize")
}

// TemplateMetadataIncludeMetrics returns whether to include metrics in metadata.
func TemplateMetadataIncludeMetrics() bool {
	return metadataBool("includeMetrics")
}

// markdownBool is a helper for markdown boolean configuration values.
// All markdown flags default to false.
func markdownBool(key string) bool {
	return viper.GetBool("output.markdown." + key)
}

// TemplateMarkdownUseCodeBlocks returns whether to use code blocks in markdown.
func TemplateMarkdownUseCodeBlocks() bool {
	return markdownBool("useCodeBlocks")
}

// TemplateMarkdownIncludeLanguage returns whether to include language in code blocks.
func TemplateMarkdownIncludeLanguage() bool {
	return markdownBool("includeLanguage")
}

// TemplateMarkdownHeaderLevel returns the header level for file sections.
// Default: ConfigMarkdownHeaderLevelDefault (0).
func TemplateMarkdownHeaderLevel() int {
	return viper.GetInt(shared.ConfigKeyOutputMarkdownHeaderLevel)
}

// TemplateMarkdownTableOfContents returns whether to include table of contents.
func TemplateMarkdownTableOfContents() bool {
	return markdownBool("tableOfContents")
}

// TemplateMarkdownUseCollapsible returns whether to use collapsible sections.
func TemplateMarkdownUseCollapsible() bool {
	return markdownBool("useCollapsible")
}

// TemplateMarkdownSyntaxHighlighting returns whether to enable syntax highlighting.
func TemplateMarkdownSyntaxHighlighting() bool {
	return markdownBool("syntaxHighlighting")
}

// TemplateMarkdownLineNumbers returns whether to include line numbers.
func TemplateMarkdownLineNumbers() bool {
	return markdownBool("lineNumbers")
}

// TemplateMarkdownFoldLongFiles returns whether to fold long files.
func TemplateMarkdownFoldLongFiles() bool {
	return markdownBool("foldLongFiles")
}

// TemplateMarkdownMaxLineLength returns the maximum line length.
// Default: ConfigMarkdownMaxLineLengthDefault (0 = unlimited).
func TemplateMarkdownMaxLineLength() int {
	return viper.GetInt(shared.ConfigKeyOutputMarkdownMaxLineLen)
}

// TemplateCustomCSS returns custom CSS for markdown output.
// Default: ConfigMarkdownCustomCSSDefault (empty string).
func TemplateCustomCSS() string {
	return viper.GetString(shared.ConfigKeyOutputMarkdownCustomCSS)
}

// TemplateCustomHeader returns custom header template.
// Default: ConfigCustomHeaderDefault (empty string).
func TemplateCustomHeader() string {
	return viper.GetString(shared.ConfigKeyOutputCustomHeader)
}

// TemplateCustomFooter returns custom footer template.
// Default: ConfigCustomFooterDefault (empty string).
func TemplateCustomFooter() string {
	return viper.GetString(shared.ConfigKeyOutputCustomFooter)
}

// TemplateCustomFileHeader returns custom file header template.
// Default: ConfigCustomFileHeaderDefault (empty string).
func TemplateCustomFileHeader() string {
	return viper.GetString(shared.ConfigKeyOutputCustomFileHeader)
}

// TemplateCustomFileFooter returns custom file footer template.
// Default: ConfigCustomFileFooterDefault (empty string).
func TemplateCustomFileFooter() string {
	return viper.GetString(shared.ConfigKeyOutputCustomFileFooter)
}

// TemplateVariables returns custom template variables.
// Default: ConfigTemplateVariablesDefault (empty map).
func TemplateVariables() map[string]string {
	return viper.GetStringMapString(shared.ConfigKeyOutputVariables)
}
