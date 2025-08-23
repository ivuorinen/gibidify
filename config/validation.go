package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"

	"github.com/ivuorinen/gibidify/utils"
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
		return utils.NewStructuredError(
			utils.ErrorTypeConfiguration,
			utils.CodeConfigValidation,
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

	fileSizeLimit := viper.GetInt64("fileSizeLimit")
	if fileSizeLimit < MinFileSizeLimit {
		validationErrors = append(
			validationErrors,
			fmt.Sprintf("fileSizeLimit (%d) is below minimum (%d)", fileSizeLimit, MinFileSizeLimit),
		)
	}
	if fileSizeLimit > MaxFileSizeLimit {
		validationErrors = append(
			validationErrors,
			fmt.Sprintf("fileSizeLimit (%d) exceeds maximum (%d)", fileSizeLimit, MaxFileSizeLimit),
		)
	}

	return validationErrors
}

// validateIgnoreDirectories validates the ignore directories setting.
func validateIgnoreDirectories() []string {
	var validationErrors []string

	ignoreDirectories := viper.GetStringSlice("ignoreDirectories")
	for i, dir := range ignoreDirectories {
		dir = strings.TrimSpace(dir)
		if dir == "" {
			validationErrors = append(validationErrors, fmt.Sprintf("ignoreDirectories[%d] is empty", i))

			continue
		}
		if strings.Contains(dir, "/") {
			validationErrors = append(
				validationErrors,
				fmt.Sprintf("ignoreDirectories[%d] (%s) contains path separator - only directory names are allowed", i, dir),
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

	if !viper.IsSet("supportedFormats") {
		return validationErrors
	}

	supportedFormats := viper.GetStringSlice("supportedFormats")
	validFormats := map[string]bool{"json": true, "yaml": true, "markdown": true}
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

	if !viper.IsSet("maxConcurrency") {
		return validationErrors
	}

	maxConcurrency := viper.GetInt("maxConcurrency")
	if maxConcurrency < 1 {
		validationErrors = append(validationErrors, fmt.Sprintf("maxConcurrency (%d) must be at least 1", maxConcurrency))
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

	if !viper.IsSet("filePatterns") {
		return validationErrors
	}

	filePatterns := viper.GetStringSlice("filePatterns")
	for i, pattern := range filePatterns {
		pattern = strings.TrimSpace(pattern)
		if pattern == "" {
			validationErrors = append(validationErrors, fmt.Sprintf("filePatterns[%d] is empty", i))

			continue
		}
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

	if !viper.IsSet("fileTypes.customImageExtensions") {
		return validationErrors
	}

	customImages := viper.GetStringSlice("fileTypes.customImageExtensions")
	for i, ext := range customImages {
		ext = strings.TrimSpace(ext)
		if ext == "" {
			validationErrors = append(validationErrors, fmt.Sprintf("fileTypes.customImageExtensions[%d] is empty", i))

			continue
		}
		if !strings.HasPrefix(ext, ".") {
			validationErrors = append(
				validationErrors,
				fmt.Sprintf("fileTypes.customImageExtensions[%d] (%s) must start with a dot", i, ext),
			)
		}
	}

	return validationErrors
}

// validateCustomBinaryExtensions validates custom binary extensions.
func validateCustomBinaryExtensions() []string {
	var validationErrors []string

	if !viper.IsSet("fileTypes.customBinaryExtensions") {
		return validationErrors
	}

	customBinary := viper.GetStringSlice("fileTypes.customBinaryExtensions")
	for i, ext := range customBinary {
		ext = strings.TrimSpace(ext)
		if ext == "" {
			validationErrors = append(validationErrors, fmt.Sprintf("fileTypes.customBinaryExtensions[%d] is empty", i))

			continue
		}
		if !strings.HasPrefix(ext, ".") {
			validationErrors = append(
				validationErrors,
				fmt.Sprintf("fileTypes.customBinaryExtensions[%d] (%s) must start with a dot", i, ext),
			)
		}
	}

	return validationErrors
}

// validateCustomLanguages validates custom language mappings.
func validateCustomLanguages() []string {
	var validationErrors []string

	if !viper.IsSet("fileTypes.customLanguages") {
		return validationErrors
	}

	customLangs := viper.GetStringMapString("fileTypes.customLanguages")
	for ext, lang := range customLangs {
		ext = strings.TrimSpace(ext)
		lang = strings.TrimSpace(lang)
		if ext == "" {
			validationErrors = append(validationErrors, "fileTypes.customLanguages contains empty extension key")

			continue
		}
		if !strings.HasPrefix(ext, ".") {
			validationErrors = append(
				validationErrors,
				fmt.Sprintf("fileTypes.customLanguages extension (%s) must start with a dot", ext),
			)
		}
		if lang == "" {
			validationErrors = append(
				validationErrors,
				fmt.Sprintf("fileTypes.customLanguages[%s] has empty language value", ext),
			)
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

	if !viper.IsSet("backpressure.maxPendingFiles") {
		return validationErrors
	}

	maxPendingFiles := viper.GetInt("backpressure.maxPendingFiles")
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

	if !viper.IsSet("backpressure.maxPendingWrites") {
		return validationErrors
	}

	maxPendingWrites := viper.GetInt("backpressure.maxPendingWrites")
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

	if !viper.IsSet("backpressure.maxMemoryUsage") {
		return validationErrors
	}

	maxMemoryUsage := viper.GetInt64("backpressure.maxMemoryUsage")
	if maxMemoryUsage < 1048576 { // 1MB minimum
		validationErrors = append(
			validationErrors,
			fmt.Sprintf("backpressure.maxMemoryUsage (%d) must be at least 1MB (1048576 bytes)", maxMemoryUsage),
		)
	}
	if maxMemoryUsage > 10737418240 { // 10GB maximum
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

	if !viper.IsSet("backpressure.memoryCheckInterval") {
		return validationErrors
	}

	interval := viper.GetInt("backpressure.memoryCheckInterval")
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

	if !viper.IsSet("resourceLimits.maxFiles") {
		return validationErrors
	}

	maxFiles := viper.GetInt("resourceLimits.maxFiles")
	if maxFiles < MinMaxFiles {
		validationErrors = append(
			validationErrors,
			fmt.Sprintf("resourceLimits.maxFiles (%d) must be at least %d", maxFiles, MinMaxFiles),
		)
	}
	if maxFiles > MaxMaxFiles {
		validationErrors = append(
			validationErrors,
			fmt.Sprintf("resourceLimits.maxFiles (%d) exceeds maximum (%d)", maxFiles, MaxMaxFiles),
		)
	}

	return validationErrors
}

// validateMaxTotalSizeLimit validates resourceLimits.maxTotalSize setting.
func validateMaxTotalSizeLimit() []string {
	var validationErrors []string

	if !viper.IsSet("resourceLimits.maxTotalSize") {
		return validationErrors
	}

	maxTotalSize := viper.GetInt64("resourceLimits.maxTotalSize")
	if maxTotalSize < MinMaxTotalSize {
		validationErrors = append(
			validationErrors,
			fmt.Sprintf("resourceLimits.maxTotalSize (%d) must be at least %d", maxTotalSize, MinMaxTotalSize),
		)
	}
	if maxTotalSize > MaxMaxTotalSize {
		validationErrors = append(
			validationErrors,
			fmt.Sprintf("resourceLimits.maxTotalSize (%d) exceeds maximum (%d)", maxTotalSize, MaxMaxTotalSize),
		)
	}

	return validationErrors
}

// validateTimeoutLimits validates timeout-related resource limit settings.
func validateTimeoutLimits() []string {
	var validationErrors []string

	if viper.IsSet("resourceLimits.fileProcessingTimeoutSec") {
		timeout := viper.GetInt("resourceLimits.fileProcessingTimeoutSec")
		if timeout < MinFileProcessingTimeoutSec {
			validationErrors = append(
				validationErrors,
				fmt.Sprintf(
					"resourceLimits.fileProcessingTimeoutSec (%d) must be at least %d",
					timeout,
					MinFileProcessingTimeoutSec,
				),
			)
		}
		if timeout > MaxFileProcessingTimeoutSec {
			validationErrors = append(
				validationErrors,
				fmt.Sprintf(
					"resourceLimits.fileProcessingTimeoutSec (%d) exceeds maximum (%d)",
					timeout,
					MaxFileProcessingTimeoutSec,
				),
			)
		}
	}

	if viper.IsSet("resourceLimits.overallTimeoutSec") {
		timeout := viper.GetInt("resourceLimits.overallTimeoutSec")
		if timeout < MinOverallTimeoutSec {
			validationErrors = append(
				validationErrors,
				fmt.Sprintf("resourceLimits.overallTimeoutSec (%d) must be at least %d", timeout, MinOverallTimeoutSec),
			)
		}
		if timeout > MaxOverallTimeoutSec {
			validationErrors = append(
				validationErrors,
				fmt.Sprintf("resourceLimits.overallTimeoutSec (%d) exceeds maximum (%d)", timeout, MaxOverallTimeoutSec),
			)
		}
	}

	return validationErrors
}

// validateConcurrencyLimits validates concurrency-related resource limit settings.
func validateConcurrencyLimits() []string {
	var validationErrors []string

	if viper.IsSet("resourceLimits.maxConcurrentReads") {
		maxReads := viper.GetInt("resourceLimits.maxConcurrentReads")
		if maxReads < MinMaxConcurrentReads {
			validationErrors = append(
				validationErrors,
				fmt.Sprintf("resourceLimits.maxConcurrentReads (%d) must be at least %d", maxReads, MinMaxConcurrentReads),
			)
		}
		if maxReads > MaxMaxConcurrentReads {
			validationErrors = append(
				validationErrors,
				fmt.Sprintf("resourceLimits.maxConcurrentReads (%d) exceeds maximum (%d)", maxReads, MaxMaxConcurrentReads),
			)
		}
	}

	if viper.IsSet("resourceLimits.rateLimitFilesPerSec") {
		rateLimit := viper.GetInt("resourceLimits.rateLimitFilesPerSec")
		if rateLimit < MinRateLimitFilesPerSec {
			validationErrors = append(
				validationErrors,
				fmt.Sprintf("resourceLimits.rateLimitFilesPerSec (%d) must be at least %d", rateLimit, MinRateLimitFilesPerSec),
			)
		}
		if rateLimit > MaxRateLimitFilesPerSec {
			validationErrors = append(
				validationErrors,
				fmt.Sprintf(
					"resourceLimits.rateLimitFilesPerSec (%d) exceeds maximum (%d)",
					rateLimit,
					MaxRateLimitFilesPerSec,
				),
			)
		}
	}

	return validationErrors
}

// validateMemoryLimits validates memory-related resource limit settings.
func validateMemoryLimits() []string {
	var validationErrors []string

	if !viper.IsSet("resourceLimits.hardMemoryLimitMB") {
		return validationErrors
	}

	memLimit := viper.GetInt("resourceLimits.hardMemoryLimitMB")
	if memLimit < MinHardMemoryLimitMB {
		validationErrors = append(
			validationErrors,
			fmt.Sprintf("resourceLimits.hardMemoryLimitMB (%d) must be at least %d", memLimit, MinHardMemoryLimitMB),
		)
	}
	if memLimit > MaxHardMemoryLimitMB {
		validationErrors = append(
			validationErrors,
			fmt.Sprintf("resourceLimits.hardMemoryLimitMB (%d) exceeds maximum (%d)", memLimit, MaxHardMemoryLimitMB),
		)
	}

	return validationErrors
}

// ValidateFileSize checks if a file size is within the configured limit.
func ValidateFileSize(size int64) error {
	limit := GetFileSizeLimit()
	if size > limit {
		return utils.NewStructuredError(
			utils.ErrorTypeValidation,
			utils.CodeValidationSize,
			fmt.Sprintf("file size (%d bytes) exceeds limit (%d bytes)", size, limit),
			"",
			map[string]any{"file_size": size, "size_limit": limit},
		)
	}

	return nil
}

// ValidateOutputFormat checks if an output format is valid.
func ValidateOutputFormat(format string) error {
	if !IsValidFormat(format) {
		return utils.NewStructuredError(
			utils.ErrorTypeValidation,
			utils.CodeValidationFormat,
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
		return utils.NewStructuredError(
			utils.ErrorTypeValidation,
			utils.CodeValidationFormat,
			fmt.Sprintf("concurrency (%d) must be at least 1", concurrency),
			"",
			map[string]any{"concurrency": concurrency},
		)
	}

	if viper.IsSet("maxConcurrency") {
		maxConcurrency := GetMaxConcurrency()
		if concurrency > maxConcurrency {
			return utils.NewStructuredError(
				utils.ErrorTypeValidation,
				utils.CodeValidationFormat,
				fmt.Sprintf("concurrency (%d) exceeds maximum (%d)", concurrency, maxConcurrency),
				"",
				map[string]any{"concurrency": concurrency, "max_concurrency": maxConcurrency},
			)
		}
	}

	return nil
}
