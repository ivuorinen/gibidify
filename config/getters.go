// Package config handles application configuration management.
package config

import (
	"strings"

	"github.com/ivuorinen/gibidify/shared"
)

// FileSizeLimit returns the file size limit from configuration.
// Default: ConfigFileSizeLimitDefault (5MB).
func FileSizeLimit() int64 {
	return GetInt64(shared.ConfigKeyFileSizeLimit)
}

// IgnoredDirectories returns the list of directories to ignore.
// Default: ConfigIgnoredDirectoriesDefault.
func IgnoredDirectories() []string {
	return GetStringSlice(shared.ConfigKeyIgnoreDirectories)
}

// MaxConcurrency returns the maximum concurrency level.
// Returns 0 if not set (caller should determine appropriate default).
func MaxConcurrency() int {
	return GetInt(shared.ConfigKeyMaxConcurrency)
}

// SupportedFormats returns the list of supported output formats.
// Returns empty slice if not set.
func SupportedFormats() []string {
	return GetStringSlice(shared.ConfigKeySupportedFormats)
}

// FilePatterns returns the list of file patterns.
// Returns empty slice if not set.
func FilePatterns() []string {
	return GetStringSlice(shared.ConfigKeyFilePatterns)
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
	return GetBool(shared.ConfigKeyFileTypesEnabled)
}

// CustomImageExtensions returns custom image extensions.
// Default: ConfigCustomImageExtensionsDefault (empty).
func CustomImageExtensions() []string {
	return GetStringSlice(shared.ConfigKeyFileTypesCustomImageExtensions)
}

// CustomBinaryExtensions returns custom binary extensions.
// Default: ConfigCustomBinaryExtensionsDefault (empty).
func CustomBinaryExtensions() []string {
	return GetStringSlice(shared.ConfigKeyFileTypesCustomBinaryExtensions)
}

// CustomLanguages returns custom language mappings.
// Default: ConfigCustomLanguagesDefault (empty).
func CustomLanguages() map[string]string {
	return GetStringMapString(shared.ConfigKeyFileTypesCustomLanguages)
}

// DisabledImageExtensions returns disabled image extensions.
// Default: ConfigDisabledImageExtensionsDefault (empty).
func DisabledImageExtensions() []string {
	return GetStringSlice(shared.ConfigKeyFileTypesDisabledImageExtensions)
}

// DisabledBinaryExtensions returns disabled binary extensions.
// Default: ConfigDisabledBinaryExtensionsDefault (empty).
func DisabledBinaryExtensions() []string {
	return GetStringSlice(shared.ConfigKeyFileTypesDisabledBinaryExtensions)
}

// DisabledLanguageExtensions returns disabled language extensions.
// Default: ConfigDisabledLanguageExtensionsDefault (empty).
func DisabledLanguageExtensions() []string {
	return GetStringSlice(shared.ConfigKeyFileTypesDisabledLanguageExts)
}

// Resource limit getters (size caps)

// MaxFiles returns the maximum number of files.
// Default: ConfigMaxFilesDefault (10000).
func MaxFiles() int {
	return GetInt(shared.ConfigKeyResourceLimitsMaxFiles)
}

// MaxTotalSize returns the maximum total size.
// Default: ConfigMaxTotalSizeDefault (1GB).
func MaxTotalSize() int64 {
	return GetInt64(shared.ConfigKeyResourceLimitsMaxTotalSize)
}
