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