// Package config handles application configuration management.
package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"

	"github.com/ivuorinen/gibidify/shared"
)

// ValidateConfig validates the loaded configuration.
func ValidateConfig() error {
	var validationErrors []string

	// Validate basic settings
	validationErrors = append(validationErrors, validateBasicSettings()...)
	validationErrors = append(validationErrors, validateFileTypeSettings()...)
	validationErrors = append(validationErrors, validateBackpressureSettings()...)
	validationErrors = append(validationErrors, validateResourceLimitSettings()...)

	if len(validationErrors) > 0 {
		return shared.NewStructuredError(
			shared.ErrorTypeConfiguration,
			shared.CodeConfigValidation,
			"configuration validation failed: "+strings.Join(validationErrors, "; "),
			"",
			map[string]any{"validation_errors": validationErrors},
		)
	}

	return nil
}

// validateBasicSettings validates basic configuration settings.
func validateBasicSettings() []string {
	var validationErrors []string

	validationErrors = append(validationErrors, validateFileSizeLimit()...)
	validationErrors = append(validationErrors, validateIgnoreDirectories()...)
	validationErrors = append(validationErrors, validateSupportedFormats()...)
	validationErrors = append(validationErrors, validateConcurrencySettings()...)
	validationErrors = append(validationErrors, validateFilePatterns()...)

	return validationErrors
}

// validateFileSizeLimit validates the file size limit setting.
func validateFileSizeLimit() []string {
	var validationErrors []string

	fileSizeLimit := viper.GetInt64(shared.ConfigKeyFileSizeLimit)
	if fileSizeLimit < shared.ConfigFileSizeLimitMin {
		validationErrors = append(
			validationErrors,
			fmt.Sprintf("fileSizeLimit (%d) is below minimum (%d)", fileSizeLimit, shared.ConfigFileSizeLimitMin),
		)
	}
	if fileSizeLimit > shared.ConfigFileSizeLimitMax {
		validationErrors = append(
			validationErrors,
			fmt.Sprintf("fileSizeLimit (%d) exceeds maximum (%d)", fileSizeLimit, shared.ConfigFileSizeLimitMax),
		)
	}

	return validationErrors
}

// validateIgnoreDirectories validates the ignore directories setting.
func validateIgnoreDirectories() []string {
	var validationErrors []string

	ignoreDirectories := viper.GetStringSlice(shared.ConfigKeyIgnoreDirectories)
	for i, dir := range ignoreDirectories {
		if errMsg := validateEmptyElement(shared.ConfigKeyIgnoreDirectories, dir, i); errMsg != "" {
			validationErrors = append(validationErrors, errMsg)

			continue
		}
		dir = strings.TrimSpace(dir)
		if strings.Contains(dir, "/") {
			validationErrors = append(
				validationErrors,
				fmt.Sprintf(
					"ignoreDirectories[%d] (%s) contains path separator - only directory names are allowed", i, dir,
				),
			)
		}
		if strings.HasPrefix(dir, ".") && dir != ".git" && dir != ".vscode" && dir != ".idea" {
			validationErrors = append(
				validationErrors,
				fmt.Sprintf("ignoreDirectories[%d] (%s) starts with dot - this may cause unexpected behavior", i, dir),
			)
		}
	}

	return validationErrors
}

// validateSupportedFormats validates the supported formats setting.
func validateSupportedFormats() []string {
	var validationErrors []string

	if !viper.IsSet(shared.ConfigKeySupportedFormats) {
		return validationErrors
	}

	supportedFormats := viper.GetStringSlice(shared.ConfigKeySupportedFormats)
	validFormats := map[string]bool{shared.FormatJSON: true, shared.FormatYAML: true, shared.FormatMarkdown: true}
	for i, format := range supportedFormats {
		format = strings.ToLower(strings.TrimSpace(format))
		if !validFormats[format] {
			validationErrors = append(
				validationErrors,
				fmt.Sprintf("supportedFormats[%d] (%s) is not a valid format (json, yaml, markdown)", i, format),
			)
		}
	}

	return validationErrors
}

// validateConcurrencySettings validates the concurrency settings.
func validateConcurrencySettings() []string {
	var validationErrors []string

	if !viper.IsSet(shared.ConfigKeyMaxConcurrency) {
		return validationErrors
	}

	maxConcurrency := viper.GetInt(shared.ConfigKeyMaxConcurrency)
	if maxConcurrency < 1 {
		validationErrors = append(
			validationErrors, fmt.Sprintf("maxConcurrency (%d) must be at least 1", maxConcurrency),
		)
	}
	if maxConcurrency > 100 {
		validationErrors = append(
			validationErrors,
			fmt.Sprintf("maxConcurrency (%d) is unreasonably high (max 100)", maxConcurrency),
		)
	}

	return validationErrors
}

// validateFilePatterns validates the file patterns setting.
func validateFilePatterns() []string {
	var validationErrors []string

	if !viper.IsSet(shared.ConfigKeyFilePatterns) {
		return validationErrors
	}

	filePatterns := viper.GetStringSlice(shared.ConfigKeyFilePatterns)
	for i, pattern := range filePatterns {
		if errMsg := validateEmptyElement(shared.ConfigKeyFilePatterns, pattern, i); errMsg != "" {
			validationErrors = append(validationErrors, errMsg)

			continue
		}
		pattern = strings.TrimSpace(pattern)
		// Basic validation - patterns should contain at least one alphanumeric character
		if !strings.ContainsAny(pattern, "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789") {
			validationErrors = append(
				validationErrors,
				fmt.Sprintf("filePatterns[%d] (%s) appears to be invalid", i, pattern),
			)
		}
	}

	return validationErrors
}

// validateFileTypeSettings validates file type configuration settings.
func validateFileTypeSettings() []string {
	var validationErrors []string

	validationErrors = append(validationErrors, validateCustomImageExtensions()...)
	validationErrors = append(validationErrors, validateCustomBinaryExtensions()...)
	validationErrors = append(validationErrors, validateCustomLanguages()...)

	return validationErrors
}

// validateCustomImageExtensions validates custom image extensions.
func validateCustomImageExtensions() []string {
	var validationErrors []string

	if !viper.IsSet(shared.ConfigKeyFileTypesCustomImageExtensions) {
		return validationErrors
	}

	customImages := viper.GetStringSlice(shared.ConfigKeyFileTypesCustomImageExtensions)
	for i, ext := range customImages {
		if errMsg := validateEmptyElement(shared.ConfigKeyFileTypesCustomImageExtensions, ext, i); errMsg != "" {
			validationErrors = append(validationErrors, errMsg)

			continue
		}
		ext = strings.TrimSpace(ext)
		if errMsg := validateDotPrefix(shared.ConfigKeyFileTypesCustomImageExtensions, ext, i); errMsg != "" {
			validationErrors = append(validationErrors, errMsg)
		}
	}

	return validationErrors
}

// validateCustomBinaryExtensions validates custom binary extensions.
func validateCustomBinaryExtensions() []string {
	var validationErrors []string

	if !viper.IsSet(shared.ConfigKeyFileTypesCustomBinaryExtensions) {
		return validationErrors
	}

	customBinary := viper.GetStringSlice(shared.ConfigKeyFileTypesCustomBinaryExtensions)
	for i, ext := range customBinary {
		if errMsg := validateEmptyElement(shared.ConfigKeyFileTypesCustomBinaryExtensions, ext, i); errMsg != "" {
			validationErrors = append(validationErrors, errMsg)

			continue
		}
		ext = strings.TrimSpace(ext)
		if errMsg := validateDotPrefix(shared.ConfigKeyFileTypesCustomBinaryExtensions, ext, i); errMsg != "" {
			validationErrors = append(validationErrors, errMsg)
		}
	}

	return validationErrors
}

// validateCustomLanguages validates custom language mappings.
func validateCustomLanguages() []string {
	var validationErrors []string

	if !viper.IsSet(shared.ConfigKeyFileTypesCustomLanguages) {
		return validationErrors
	}

	customLangs := viper.GetStringMapString(shared.ConfigKeyFileTypesCustomLanguages)
	for ext, lang := range customLangs {
		ext = strings.TrimSpace(ext)
		if ext == "" {
			validationErrors = append(
				validationErrors,
				shared.ConfigKeyFileTypesCustomLanguages+" contains empty extension key",
			)

			continue
		}
		if errMsg := validateDotPrefixMap(shared.ConfigKeyFileTypesCustomLanguages, ext); errMsg != "" {
			validationErrors = append(validationErrors, errMsg)
		}
		if errMsg := validateEmptyMapValue(shared.ConfigKeyFileTypesCustomLanguages, ext, lang); errMsg != "" {
			validationErrors = append(validationErrors, errMsg)
		}
	}

	return validationErrors
}

// validateBackpressureSettings validates back-pressure configuration settings.
func validateBackpressureSettings() []string {
	var validationErrors []string

	validationErrors = append(validationErrors, validateMaxPendingFiles()...)
	validationErrors = append(validationErrors, validateMaxPendingWrites()...)
	validationErrors = append(validationErrors, validateMaxMemoryUsage()...)
	validationErrors = append(validationErrors, validateMemoryCheckInterval()...)

	return validationErrors
}

// validateMaxPendingFiles validates backpressure.maxPendingFiles setting.
func validateMaxPendingFiles() []string {
	var validationErrors []string

	if !viper.IsSet(shared.ConfigKeyBackpressureMaxPendingFiles) {
		return validationErrors
	}

	maxPendingFiles := viper.GetInt(shared.ConfigKeyBackpressureMaxPendingFiles)
	if maxPendingFiles < 1 {
		validationErrors = append(
			validationErrors,
			fmt.Sprintf("backpressure.maxPendingFiles (%d) must be at least 1", maxPendingFiles),
		)
	}
	if maxPendingFiles > 100000 {
		validationErrors = append(
			validationErrors,
			fmt.Sprintf("backpressure.maxPendingFiles (%d) is unreasonably high (max 100000)", maxPendingFiles),
		)
	}

	return validationErrors
}

// validateMaxPendingWrites validates backpressure.maxPendingWrites setting.
func validateMaxPendingWrites() []string {
	var validationErrors []string

	if !viper.IsSet(shared.ConfigKeyBackpressureMaxPendingWrites) {
		return validationErrors
	}

	maxPendingWrites := viper.GetInt(shared.ConfigKeyBackpressureMaxPendingWrites)
	if maxPendingWrites < 1 {
		validationErrors = append(
			validationErrors,
			fmt.Sprintf("backpressure.maxPendingWrites (%d) must be at least 1", maxPendingWrites),
		)
	}
	if maxPendingWrites > 10000 {
		validationErrors = append(
			validationErrors,
			fmt.Sprintf("backpressure.maxPendingWrites (%d) is unreasonably high (max 10000)", maxPendingWrites),
		)
	}

	return validationErrors
}

// validateMaxMemoryUsage validates backpressure.maxMemoryUsage setting.
func validateMaxMemoryUsage() []string {
	var validationErrors []string

	if !viper.IsSet(shared.ConfigKeyBackpressureMaxMemoryUsage) {
		return validationErrors
	}

	maxMemoryUsage := viper.GetInt64(shared.ConfigKeyBackpressureMaxMemoryUsage)
	minMemory := int64(shared.BytesPerMB)      // 1MB minimum
	maxMemory := int64(10 * shared.BytesPerGB) // 10GB maximum
	if maxMemoryUsage < minMemory {
		validationErrors = append(
			validationErrors,
			fmt.Sprintf("backpressure.maxMemoryUsage (%d) must be at least 1MB (%d bytes)", maxMemoryUsage, minMemory),
		)
	}
	if maxMemoryUsage > maxMemory { // 10GB maximum
		validationErrors = append(
			validationErrors,
			fmt.Sprintf("backpressure.maxMemoryUsage (%d) is unreasonably high (max 10GB)", maxMemoryUsage),
		)
	}

	return validationErrors
}

// validateMemoryCheckInterval validates backpressure.memoryCheckInterval setting.
func validateMemoryCheckInterval() []string {
	var validationErrors []string

	if !viper.IsSet(shared.ConfigKeyBackpressureMemoryCheckInt) {
		return validationErrors
	}

	interval := viper.GetInt(shared.ConfigKeyBackpressureMemoryCheckInt)
	if interval < 1 {
		validationErrors = append(
			validationErrors,
			fmt.Sprintf("backpressure.memoryCheckInterval (%d) must be at least 1", interval),
		)
	}
	if interval > 100000 {
		validationErrors = append(
			validationErrors,
			fmt.Sprintf("backpressure.memoryCheckInterval (%d) is unreasonably high (max 100000)", interval),
		)
	}

	return validationErrors
}

// validateResourceLimitSettings validates resource limit configuration settings.
func validateResourceLimitSettings() []string {
	var validationErrors []string

	validationErrors = append(validationErrors, validateMaxFilesLimit()...)
	validationErrors = append(validationErrors, validateMaxTotalSizeLimit()...)
	validationErrors = append(validationErrors, validateTimeoutLimits()...)
	validationErrors = append(validationErrors, validateConcurrencyLimits()...)
	validationErrors = append(validationErrors, validateMemoryLimits()...)

	return validationErrors
}

// validateMaxFilesLimit validates resourceLimits.maxFiles setting.
func validateMaxFilesLimit() []string {
	var validationErrors []string

	if !viper.IsSet(shared.ConfigKeyResourceLimitsMaxFiles) {
		return validationErrors
	}

	maxFiles := viper.GetInt(shared.ConfigKeyResourceLimitsMaxFiles)
	if maxFiles < shared.ConfigMaxFilesMin {
		validationErrors = append(
			validationErrors,
			fmt.Sprintf("resourceLimits.maxFiles (%d) must be at least %d", maxFiles, shared.ConfigMaxFilesMin),
		)
	}
	if maxFiles > shared.ConfigMaxFilesMax {
		validationErrors = append(
			validationErrors,
			fmt.Sprintf("resourceLimits.maxFiles (%d) exceeds maximum (%d)", maxFiles, shared.ConfigMaxFilesMax),
		)
	}

	return validationErrors
}

// validateMaxTotalSizeLimit validates resourceLimits.maxTotalSize setting.
func validateMaxTotalSizeLimit() []string {
	var validationErrors []string

	if !viper.IsSet(shared.ConfigKeyResourceLimitsMaxTotalSize) {
		return validationErrors
	}

	maxTotalSize := viper.GetInt64(shared.ConfigKeyResourceLimitsMaxTotalSize)
	minTotalSize := int64(shared.ConfigMaxTotalSizeMin)
	maxTotalSizeLimit := int64(shared.ConfigMaxTotalSizeMax)
	if maxTotalSize < minTotalSize {
		validationErrors = append(
			validationErrors,
			fmt.Sprintf("resourceLimits.maxTotalSize (%d) must be at least %d", maxTotalSize, minTotalSize),
		)
	}
	if maxTotalSize > maxTotalSizeLimit {
		validationErrors = append(
			validationErrors,
			fmt.Sprintf("resourceLimits.maxTotalSize (%d) exceeds maximum (%d)", maxTotalSize, maxTotalSizeLimit),
		)
	}

	return validationErrors
}

// validateTimeoutLimits validates timeout-related resource limit settings.
func validateTimeoutLimits() []string {
	var validationErrors []string

	if viper.IsSet(shared.ConfigKeyResourceLimitsFileProcessingTO) {
		timeout := viper.GetInt(shared.ConfigKeyResourceLimitsFileProcessingTO)
		if timeout < shared.ConfigFileProcessingTimeoutSecMin {
			validationErrors = append(
				validationErrors,
				fmt.Sprintf(
					"resourceLimits.fileProcessingTimeoutSec (%d) must be at least %d",
					timeout,
					shared.ConfigFileProcessingTimeoutSecMin,
				),
			)
		}
		if timeout > shared.ConfigFileProcessingTimeoutSecMax {
			validationErrors = append(
				validationErrors,
				fmt.Sprintf(
					"resourceLimits.fileProcessingTimeoutSec (%d) exceeds maximum (%d)",
					timeout,
					shared.ConfigFileProcessingTimeoutSecMax,
				),
			)
		}
	}

	if viper.IsSet(shared.ConfigKeyResourceLimitsOverallTO) {
		timeout := viper.GetInt(shared.ConfigKeyResourceLimitsOverallTO)
		minTimeout := shared.ConfigOverallTimeoutSecMin
		maxTimeout := shared.ConfigOverallTimeoutSecMax
		if timeout < minTimeout {
			validationErrors = append(
				validationErrors,
				fmt.Sprintf("resourceLimits.overallTimeoutSec (%d) must be at least %d", timeout, minTimeout),
			)
		}
		if timeout > maxTimeout {
			validationErrors = append(
				validationErrors,
				fmt.Sprintf("resourceLimits.overallTimeoutSec (%d) exceeds maximum (%d)", timeout, maxTimeout),
			)
		}
	}

	return validationErrors
}

// validateConcurrencyLimits validates concurrency-related resource limit settings.
func validateConcurrencyLimits() []string {
	var validationErrors []string

	if viper.IsSet(shared.ConfigKeyResourceLimitsMaxConcurrentReads) {
		maxReads := viper.GetInt(shared.ConfigKeyResourceLimitsMaxConcurrentReads)
		minReads := shared.ConfigMaxConcurrentReadsMin
		maxReadsLimit := shared.ConfigMaxConcurrentReadsMax
		if maxReads < minReads {
			validationErrors = append(
				validationErrors,
				fmt.Sprintf("resourceLimits.maxConcurrentReads (%d) must be at least %d", maxReads, minReads),
			)
		}
		if maxReads > maxReadsLimit {
			validationErrors = append(
				validationErrors,
				fmt.Sprintf("resourceLimits.maxConcurrentReads (%d) exceeds maximum (%d)", maxReads, maxReadsLimit),
			)
		}
	}

	if viper.IsSet(shared.ConfigKeyResourceLimitsRateLimitFilesPerSec) {
		rateLimit := viper.GetInt(shared.ConfigKeyResourceLimitsRateLimitFilesPerSec)
		minRate := shared.ConfigRateLimitFilesPerSecMin
		maxRate := shared.ConfigRateLimitFilesPerSecMax
		if rateLimit < minRate {
			validationErrors = append(
				validationErrors,
				fmt.Sprintf("resourceLimits.rateLimitFilesPerSec (%d) must be at least %d", rateLimit, minRate),
			)
		}
		if rateLimit > maxRate {
			validationErrors = append(
				validationErrors,
				fmt.Sprintf("resourceLimits.rateLimitFilesPerSec (%d) exceeds maximum (%d)", rateLimit, maxRate),
			)
		}
	}

	return validationErrors
}

// validateMemoryLimits validates memory-related resource limit settings.
func validateMemoryLimits() []string {
	var validationErrors []string

	if !viper.IsSet(shared.ConfigKeyResourceLimitsHardMemoryLimitMB) {
		return validationErrors
	}

	memLimit := viper.GetInt(shared.ConfigKeyResourceLimitsHardMemoryLimitMB)
	minMemLimit := shared.ConfigHardMemoryLimitMBMin
	maxMemLimit := shared.ConfigHardMemoryLimitMBMax
	if memLimit < minMemLimit {
		validationErrors = append(
			validationErrors,
			fmt.Sprintf("resourceLimits.hardMemoryLimitMB (%d) must be at least %d", memLimit, minMemLimit),
		)
	}
	if memLimit > maxMemLimit {
		validationErrors = append(
			validationErrors,
			fmt.Sprintf("resourceLimits.hardMemoryLimitMB (%d) exceeds maximum (%d)", memLimit, maxMemLimit),
		)
	}

	return validationErrors
}

// ValidateFileSize checks if a file size is within the configured limit.
func ValidateFileSize(size int64) error {
	limit := FileSizeLimit()
	if size > limit {
		return shared.NewStructuredError(
			shared.ErrorTypeValidation,
			shared.CodeValidationSize,
			fmt.Sprintf(shared.FileProcessingMsgSizeExceeds, size, limit),
			"",
			map[string]any{"file_size": size, "size_limit": limit},
		)
	}

	return nil
}

// ValidateOutputFormat checks if an output format is valid.
func ValidateOutputFormat(format string) error {
	if !IsValidFormat(format) {
		return shared.NewStructuredError(
			shared.ErrorTypeValidation,
			shared.CodeValidationFormat,
			fmt.Sprintf("unsupported output format: %s (supported: json, yaml, markdown)", format),
			"",
			map[string]any{"format": format},
		)
	}

	return nil
}

// ValidateConcurrency checks if a concurrency level is valid.
func ValidateConcurrency(concurrency int) error {
	if concurrency < 1 {
		return shared.NewStructuredError(
			shared.ErrorTypeValidation,
			shared.CodeValidationFormat,
			fmt.Sprintf("concurrency (%d) must be at least 1", concurrency),
			"",
			map[string]any{"concurrency": concurrency},
		)
	}

	if viper.IsSet(shared.ConfigKeyMaxConcurrency) {
		maxConcurrency := MaxConcurrency()
		if concurrency > maxConcurrency {
			return shared.NewStructuredError(
				shared.ErrorTypeValidation,
				shared.CodeValidationFormat,
				fmt.Sprintf("concurrency (%d) exceeds maximum (%d)", concurrency, maxConcurrency),
				"",
				map[string]any{"concurrency": concurrency, "max_concurrency": maxConcurrency},
			)
		}
	}

	return nil
}
