// Package config handles application configuration management.
package config

import (
	"strings"

	"github.com/spf13/viper"
)

// GetFileSizeLimit returns the file size limit from configuration.
func GetFileSizeLimit() int64 {
	return viper.GetInt64("fileSizeLimit")
}

// GetIgnoredDirectories returns the list of directories to ignore.
func GetIgnoredDirectories() []string {
	return viper.GetStringSlice("ignoreDirectories")
}

// GetMaxConcurrency returns the maximum concurrency level.
func GetMaxConcurrency() int {
	return viper.GetInt("maxConcurrency")
}

// GetSupportedFormats returns the list of supported output formats.
func GetSupportedFormats() []string {
	return viper.GetStringSlice("supportedFormats")
}

// GetFilePatterns returns the list of file patterns.
func GetFilePatterns() []string {
	return viper.GetStringSlice("filePatterns")
}

// IsValidFormat checks if the given format is valid.
func IsValidFormat(format string) bool {
	format = strings.ToLower(strings.TrimSpace(format))
	supportedFormats := map[string]bool{
		"json":     true,
		"yaml":     true,
		"markdown": true,
	}

	return supportedFormats[format]
}

// GetFileTypesEnabled returns whether file types are enabled.
func GetFileTypesEnabled() bool {
	return viper.GetBool("fileTypes.enabled")
}

// GetCustomImageExtensions returns custom image extensions.
func GetCustomImageExtensions() []string {
	return viper.GetStringSlice("fileTypes.customImageExtensions")
}

// GetCustomBinaryExtensions returns custom binary extensions.
func GetCustomBinaryExtensions() []string {
	return viper.GetStringSlice("fileTypes.customBinaryExtensions")
}

// GetCustomLanguages returns custom language mappings.
func GetCustomLanguages() map[string]string {
	return viper.GetStringMapString("fileTypes.customLanguages")
}

// GetDisabledImageExtensions returns disabled image extensions.
func GetDisabledImageExtensions() []string {
	return viper.GetStringSlice("fileTypes.disabledImageExtensions")
}

// GetDisabledBinaryExtensions returns disabled binary extensions.
func GetDisabledBinaryExtensions() []string {
	return viper.GetStringSlice("fileTypes.disabledBinaryExtensions")
}

// GetDisabledLanguageExtensions returns disabled language extensions.
func GetDisabledLanguageExtensions() []string {
	return viper.GetStringSlice("fileTypes.disabledLanguageExtensions")
}

// Backpressure getters

// GetBackpressureEnabled returns whether backpressure is enabled.
func GetBackpressureEnabled() bool {
	return viper.GetBool("backpressure.enabled")
}

// GetMaxPendingFiles returns the maximum pending files.
func GetMaxPendingFiles() int {
	return viper.GetInt("backpressure.maxPendingFiles")
}

// GetMaxPendingWrites returns the maximum pending writes.
func GetMaxPendingWrites() int {
	return viper.GetInt("backpressure.maxPendingWrites")
}

// GetMaxMemoryUsage returns the maximum memory usage.
func GetMaxMemoryUsage() int64 {
	return viper.GetInt64("backpressure.maxMemoryUsage")
}

// GetMemoryCheckInterval returns the memory check interval.
func GetMemoryCheckInterval() int {
	return viper.GetInt("backpressure.memoryCheckInterval")
}

// Resource limits getters

// GetResourceLimitsEnabled returns whether resource limits are enabled.
func GetResourceLimitsEnabled() bool {
	return viper.GetBool("resourceLimits.enabled")
}

// GetMaxFiles returns the maximum number of files.
func GetMaxFiles() int {
	return viper.GetInt("resourceLimits.maxFiles")
}

// GetMaxTotalSize returns the maximum total size.
func GetMaxTotalSize() int64 {
	return viper.GetInt64("resourceLimits.maxTotalSize")
}

// GetFileProcessingTimeoutSec returns the file processing timeout in seconds.
func GetFileProcessingTimeoutSec() int {
	return viper.GetInt("resourceLimits.fileProcessingTimeoutSec")
}

// GetOverallTimeoutSec returns the overall timeout in seconds.
func GetOverallTimeoutSec() int {
	return viper.GetInt("resourceLimits.overallTimeoutSec")
}

// GetMaxConcurrentReads returns the maximum concurrent reads.
func GetMaxConcurrentReads() int {
	return viper.GetInt("resourceLimits.maxConcurrentReads")
}

// GetRateLimitFilesPerSec returns the rate limit files per second.
func GetRateLimitFilesPerSec() int {
	return viper.GetInt("resourceLimits.rateLimitFilesPerSec")
}

// GetHardMemoryLimitMB returns the hard memory limit in MB.
func GetHardMemoryLimitMB() int {
	return viper.GetInt("resourceLimits.hardMemoryLimitMB")
}

// GetEnableGracefulDegradation returns whether graceful degradation is enabled.
func GetEnableGracefulDegradation() bool {
	return viper.GetBool("resourceLimits.enableGracefulDegradation")
}

// GetEnableResourceMonitoring returns whether resource monitoring is enabled.
func GetEnableResourceMonitoring() bool {
	return viper.GetBool("resourceLimits.enableResourceMonitoring")
}

// Template system getters

// GetOutputTemplate returns the selected output template name.
func GetOutputTemplate() string {
	return viper.GetString("output.template")
}

// GetTemplateMetadataIncludeStats returns whether to include stats in metadata.
func GetTemplateMetadataIncludeStats() bool {
	return viper.GetBool("output.metadata.includeStats")
}

// GetTemplateMetadataIncludeTimestamp returns whether to include timestamp in metadata.
func GetTemplateMetadataIncludeTimestamp() bool {
	return viper.GetBool("output.metadata.includeTimestamp")
}

// GetTemplateMetadataIncludeFileCount returns whether to include file count in metadata.
func GetTemplateMetadataIncludeFileCount() bool {
	return viper.GetBool("output.metadata.includeFileCount")
}

// GetTemplateMetadataIncludeSourcePath returns whether to include source path in metadata.
func GetTemplateMetadataIncludeSourcePath() bool {
	return viper.GetBool("output.metadata.includeSourcePath")
}

// GetTemplateMetadataIncludeFileTypes returns whether to include file types in metadata.
func GetTemplateMetadataIncludeFileTypes() bool {
	return viper.GetBool("output.metadata.includeFileTypes")
}

// GetTemplateMetadataIncludeProcessingTime returns whether to include processing time in metadata.
func GetTemplateMetadataIncludeProcessingTime() bool {
	return viper.GetBool("output.metadata.includeProcessingTime")
}

// GetTemplateMetadataIncludeTotalSize returns whether to include total size in metadata.
func GetTemplateMetadataIncludeTotalSize() bool {
	return viper.GetBool("output.metadata.includeTotalSize")
}

// GetTemplateMetadataIncludeMetrics returns whether to include metrics in metadata.
func GetTemplateMetadataIncludeMetrics() bool {
	return viper.GetBool("output.metadata.includeMetrics")
}

// GetTemplateMarkdownUseCodeBlocks returns whether to use code blocks in markdown.
func GetTemplateMarkdownUseCodeBlocks() bool {
	return viper.GetBool("output.markdown.useCodeBlocks")
}

// GetTemplateMarkdownIncludeLanguage returns whether to include language in code blocks.
func GetTemplateMarkdownIncludeLanguage() bool {
	return viper.GetBool("output.markdown.includeLanguage")
}

// GetTemplateMarkdownHeaderLevel returns the header level for file sections.
func GetTemplateMarkdownHeaderLevel() int {
	return viper.GetInt("output.markdown.headerLevel")
}

// GetTemplateMarkdownTableOfContents returns whether to include table of contents.
func GetTemplateMarkdownTableOfContents() bool {
	return viper.GetBool("output.markdown.tableOfContents")
}

// GetTemplateMarkdownUseCollapsible returns whether to use collapsible sections.
func GetTemplateMarkdownUseCollapsible() bool {
	return viper.GetBool("output.markdown.useCollapsible")
}

// GetTemplateMarkdownSyntaxHighlighting returns whether to enable syntax highlighting.
func GetTemplateMarkdownSyntaxHighlighting() bool {
	return viper.GetBool("output.markdown.syntaxHighlighting")
}

// GetTemplateMarkdownLineNumbers returns whether to include line numbers.
func GetTemplateMarkdownLineNumbers() bool {
	return viper.GetBool("output.markdown.lineNumbers")
}

// GetTemplateMarkdownFoldLongFiles returns whether to fold long files.
func GetTemplateMarkdownFoldLongFiles() bool {
	return viper.GetBool("output.markdown.foldLongFiles")
}

// GetTemplateMarkdownMaxLineLength returns the maximum line length.
func GetTemplateMarkdownMaxLineLength() int {
	return viper.GetInt("output.markdown.maxLineLength")
}

// GetTemplateCustomCSS returns custom CSS for markdown output.
func GetTemplateCustomCSS() string {
	return viper.GetString("output.markdown.customCSS")
}

// GetTemplateCustomHeader returns custom header template.
func GetTemplateCustomHeader() string {
	return viper.GetString("output.custom.header")
}

// GetTemplateCustomFooter returns custom footer template.
func GetTemplateCustomFooter() string {
	return viper.GetString("output.custom.footer")
}

// GetTemplateCustomFileHeader returns custom file header template.
func GetTemplateCustomFileHeader() string {
	return viper.GetString("output.custom.fileHeader")
}

// GetTemplateCustomFileFooter returns custom file footer template.
func GetTemplateCustomFileFooter() string {
	return viper.GetString("output.custom.fileFooter")
}

// GetTemplateVariables returns custom template variables.
func GetTemplateVariables() map[string]string {
	return viper.GetStringMapString("output.variables")
}
