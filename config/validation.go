package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"

	"github.com/ivuorinen/gibidify/gibidiutils"
)

// validateFileSizeLimit validates the file size limit configuration.
func validateFileSizeLimit() []string {
	var errors []string
	fileSizeLimit := viper.GetInt64("fileSizeLimit")
	if fileSizeLimit < MinFileSizeLimit {
		errors = append(
			errors,
			fmt.Sprintf("fileSizeLimit (%d) is below minimum (%d)", fileSizeLimit, MinFileSizeLimit),
		)
	}
	if fileSizeLimit > MaxFileSizeLimit {
		errors = append(
			errors,
			fmt.Sprintf("fileSizeLimit (%d) exceeds maximum (%d)", fileSizeLimit, MaxFileSizeLimit),
		)
	}
	return errors
}

// validateIgnoreDirectories validates the ignore directories configuration.
func validateIgnoreDirectories() []string {
	var errors []string
	ignoreDirectories := viper.GetStringSlice("ignoreDirectories")
	for i, dir := range ignoreDirectories {
		dir = strings.TrimSpace(dir)
		if dir == "" {
			errors = append(errors, fmt.Sprintf("ignoreDirectories[%d] is empty", i))
			continue
		}
		if strings.Contains(dir, "/") {
			errors = append(
				errors,
				fmt.Sprintf(
					"ignoreDirectories[%d] (%s) contains path separator - only directory names are allowed",
					i,
					dir,
				),
			)
		}
		if strings.HasPrefix(dir, ".") && dir != ".git" && dir != ".vscode" && dir != ".idea" {
			errors = append(
				errors,
				fmt.Sprintf("ignoreDirectories[%d] (%s) starts with dot - this may cause unexpected behavior", i, dir),
			)
		}
	}
	return errors
}

// validateSupportedFormats validates the supported output formats configuration.
func validateSupportedFormats() []string {
	var errors []string
	if viper.IsSet("supportedFormats") {
		supportedFormats := viper.GetStringSlice("supportedFormats")
		validFormats := map[string]bool{"json": true, "yaml": true, "markdown": true}
		for i, format := range supportedFormats {
			format = strings.ToLower(strings.TrimSpace(format))
			if !validFormats[format] {
				errors = append(
					errors,
					fmt.Sprintf("supportedFormats[%d] (%s) is not a valid format (json, yaml, markdown)", i, format),
				)
			}
		}
	}
	return errors
}

// validateConcurrencySettings validates the concurrency settings configuration.
func validateConcurrencySettings() []string {
	var errors []string
	if viper.IsSet("maxConcurrency") {
		maxConcurrency := viper.GetInt("maxConcurrency")
		if maxConcurrency < 1 {
			errors = append(
				errors,
				fmt.Sprintf("maxConcurrency (%d) must be at least 1", maxConcurrency),
			)
		}
		if maxConcurrency > 100 {
			errors = append(
				errors,
				fmt.Sprintf("maxConcurrency (%d) is unreasonably high (max 100)", maxConcurrency),
			)
		}
	}
	return errors
}

// validateFilePatterns validates the file patterns configuration.
func validateFilePatterns() []string {
	var errors []string
	if viper.IsSet("filePatterns") {
		filePatterns := viper.GetStringSlice("filePatterns")
		for i, pattern := range filePatterns {
			pattern = strings.TrimSpace(pattern)
			if pattern == "" {
				errors = append(errors, fmt.Sprintf("filePatterns[%d] is empty", i))
				continue
			}
			// Basic validation - patterns should contain at least one alphanumeric character
			if !strings.ContainsAny(pattern, "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789") {
				errors = append(
					errors,
					fmt.Sprintf("filePatterns[%d] (%s) appears to be invalid", i, pattern),
				)
			}
		}
	}
	return errors
}

// validateFileTypes validates the FileTypeRegistry configuration.
// validateCustomImageExtensions validates custom image extensions configuration.
func validateCustomImageExtensions() []string {
	var errors []string
	if !viper.IsSet("fileTypes.customImageExtensions") {
		return errors
	}

	customImages := viper.GetStringSlice("fileTypes.customImageExtensions")
	for i, ext := range customImages {
		ext = strings.TrimSpace(ext)
		if ext == "" {
			errors = append(
				errors,
				fmt.Sprintf("fileTypes.customImageExtensions[%d] is empty", i),
			)
			continue
		}
		if !strings.HasPrefix(ext, ".") {
			errors = append(
				errors,
				fmt.Sprintf("fileTypes.customImageExtensions[%d] (%s) must start with a dot", i, ext),
			)
		}
	}
	return errors
}

// validateCustomBinaryExtensions validates custom binary extensions configuration.
func validateCustomBinaryExtensions() []string {
	var errors []string
	if !viper.IsSet("fileTypes.customBinaryExtensions") {
		return errors
	}

	customBinary := viper.GetStringSlice("fileTypes.customBinaryExtensions")
	for i, ext := range customBinary {
		ext = strings.TrimSpace(ext)
		if ext == "" {
			errors = append(
				errors,
				fmt.Sprintf("fileTypes.customBinaryExtensions[%d] is empty", i),
			)
			continue
		}
		if !strings.HasPrefix(ext, ".") {
			errors = append(
				errors,
				fmt.Sprintf("fileTypes.customBinaryExtensions[%d] (%s) must start with a dot", i, ext),
			)
		}
	}
	return errors
}

// validateCustomLanguages validates custom languages configuration.
func validateCustomLanguages() []string {
	var errors []string
	if !viper.IsSet("fileTypes.customLanguages") {
		return errors
	}

	customLangs := viper.GetStringMapString("fileTypes.customLanguages")
	for ext, lang := range customLangs {
		ext = strings.TrimSpace(ext)
		lang = strings.TrimSpace(lang)
		if ext == "" {
			errors = append(errors, "fileTypes.customLanguages contains empty extension key")
			continue
		}
		if !strings.HasPrefix(ext, ".") {
			errors = append(
				errors,
				fmt.Sprintf("fileTypes.customLanguages extension (%s) must start with a dot", ext),
			)
		}
		if lang == "" {
			errors = append(
				errors,
				fmt.Sprintf("fileTypes.customLanguages[%s] has empty language value", ext),
			)
		}
	}
	return errors
}

// validateFileTypes validates the FileTypeRegistry configuration.
func validateFileTypes() []string {
	var errors []string
	errors = append(errors, validateCustomImageExtensions()...)
	errors = append(errors, validateCustomBinaryExtensions()...)
	errors = append(errors, validateCustomLanguages()...)
	return errors
}

// validateBackpressureConfig validates the back-pressure configuration.
// validateBackpressureMaxPendingFiles validates max pending files configuration.
func validateBackpressureMaxPendingFiles() []string {
	var errors []string
	if !viper.IsSet("backpressure.maxPendingFiles") {
		return errors
	}

	maxPendingFiles := viper.GetInt("backpressure.maxPendingFiles")
	if maxPendingFiles < 1 {
		errors = append(
			errors,
			fmt.Sprintf("backpressure.maxPendingFiles (%d) must be at least 1", maxPendingFiles),
		)
	}
	if maxPendingFiles > 100000 {
		errors = append(
			errors,
			fmt.Sprintf("backpressure.maxPendingFiles (%d) is unreasonably high (max 100000)", maxPendingFiles),
		)
	}
	return errors
}

// validateBackpressureMaxPendingWrites validates max pending writes configuration.
func validateBackpressureMaxPendingWrites() []string {
	var errors []string
	if !viper.IsSet("backpressure.maxPendingWrites") {
		return errors
	}

	maxPendingWrites := viper.GetInt("backpressure.maxPendingWrites")
	if maxPendingWrites < 1 {
		errors = append(
			errors,
			fmt.Sprintf("backpressure.maxPendingWrites (%d) must be at least 1", maxPendingWrites),
		)
	}
	if maxPendingWrites > 10000 {
		errors = append(
			errors,
			fmt.Sprintf("backpressure.maxPendingWrites (%d) is unreasonably high (max 10000)", maxPendingWrites),
		)
	}
	return errors
}

// validateBackpressureMaxMemoryUsage validates max memory usage configuration.
func validateBackpressureMaxMemoryUsage() []string {
	var errors []string
	if !viper.IsSet("backpressure.maxMemoryUsage") {
		return errors
	}

	maxMemoryUsage := viper.GetInt64("backpressure.maxMemoryUsage")
	if maxMemoryUsage < 1048576 { // 1MB minimum
		errors = append(
			errors,
			fmt.Sprintf("backpressure.maxMemoryUsage (%d) must be at least 1MB (1048576 bytes)", maxMemoryUsage),
		)
	}
	if maxMemoryUsage > 104857600 { // 100MB maximum
		errors = append(
			errors,
			fmt.Sprintf("backpressure.maxMemoryUsage (%d) is unreasonably high (max 100MB)", maxMemoryUsage),
		)
	}
	return errors
}

// validateBackpressureMemoryCheckInterval validates memory check interval configuration.
func validateBackpressureMemoryCheckInterval() []string {
	var errors []string
	if !viper.IsSet("backpressure.memoryCheckInterval") {
		return errors
	}

	interval := viper.GetInt("backpressure.memoryCheckInterval")
	if interval < 1 {
		errors = append(
			errors,
			fmt.Sprintf("backpressure.memoryCheckInterval (%d) must be at least 1", interval),
		)
	}
	if interval > 100000 {
		errors = append(
			errors,
			fmt.Sprintf("backpressure.memoryCheckInterval (%d) is unreasonably high (max 100000)", interval),
		)
	}
	return errors
}

// validateBackpressureConfig validates the back-pressure configuration.
func validateBackpressureConfig() []string {
	var errors []string
	errors = append(errors, validateBackpressureMaxPendingFiles()...)
	errors = append(errors, validateBackpressureMaxPendingWrites()...)
	errors = append(errors, validateBackpressureMaxMemoryUsage()...)
	errors = append(errors, validateBackpressureMemoryCheckInterval()...)
	return errors
}

// validateResourceLimits validates the resource limits configuration.
// validateResourceLimitsMaxFiles validates max files configuration.
func validateResourceLimitsMaxFiles() []string {
	var errors []string
	if !viper.IsSet("resourceLimits.maxFiles") {
		return errors
	}

	maxFiles := viper.GetInt("resourceLimits.maxFiles")
	if maxFiles < MinMaxFiles {
		errors = append(
			errors,
			fmt.Sprintf("resourceLimits.maxFiles (%d) must be at least %d", maxFiles, MinMaxFiles),
		)
	}
	if maxFiles > MaxMaxFiles {
		errors = append(
			errors,
			fmt.Sprintf("resourceLimits.maxFiles (%d) exceeds maximum (%d)", maxFiles, MaxMaxFiles),
		)
	}
	return errors
}

// validateResourceLimitsMaxTotalSize validates max total size configuration.
func validateResourceLimitsMaxTotalSize() []string {
	var errors []string
	if !viper.IsSet("resourceLimits.maxTotalSize") {
		return errors
	}

	maxTotalSize := viper.GetInt64("resourceLimits.maxTotalSize")
	if maxTotalSize < MinMaxTotalSize {
		errors = append(
			errors,
			fmt.Sprintf("resourceLimits.maxTotalSize (%d) must be at least %d", maxTotalSize, MinMaxTotalSize),
		)
	}
	if maxTotalSize > MaxMaxTotalSize {
		errors = append(
			errors,
			fmt.Sprintf("resourceLimits.maxTotalSize (%d) exceeds maximum (%d)", maxTotalSize, MaxMaxTotalSize),
		)
	}
	return errors
}

// validateResourceLimitsTimeouts validates timeout configurations.
func validateResourceLimitsTimeouts() []string {
	var errors []string

	if viper.IsSet("resourceLimits.fileProcessingTimeoutSec") {
		timeout := viper.GetInt("resourceLimits.fileProcessingTimeoutSec")
		if timeout < MinFileProcessingTimeoutSec {
			errors = append(
				errors,
				fmt.Sprintf(
					"resourceLimits.fileProcessingTimeoutSec (%d) must be at least %d",
					timeout,
					MinFileProcessingTimeoutSec,
				),
			)
		}
		if timeout > MaxFileProcessingTimeoutSec {
			errors = append(
				errors,
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
			errors = append(
				errors,
				fmt.Sprintf("resourceLimits.overallTimeoutSec (%d) must be at least %d", timeout, MinOverallTimeoutSec),
			)
		}
		if timeout > MaxOverallTimeoutSec {
			errors = append(
				errors,
				fmt.Sprintf(
					"resourceLimits.overallTimeoutSec (%d) exceeds maximum (%d)",
					timeout,
					MaxOverallTimeoutSec,
				),
			)
		}
	}

	return errors
}

// validateResourceLimitsConcurrency validates concurrency configurations.
func validateResourceLimitsConcurrency() []string {
	var errors []string

	if viper.IsSet("resourceLimits.maxConcurrentReads") {
		maxReads := viper.GetInt("resourceLimits.maxConcurrentReads")
		if maxReads < MinMaxConcurrentReads {
			errors = append(
				errors,
				fmt.Sprintf(
					"resourceLimits.maxConcurrentReads (%d) must be at least %d",
					maxReads,
					MinMaxConcurrentReads,
				),
			)
		}
		if maxReads > MaxMaxConcurrentReads {
			errors = append(
				errors,
				fmt.Sprintf(
					"resourceLimits.maxConcurrentReads (%d) exceeds maximum (%d)",
					maxReads,
					MaxMaxConcurrentReads,
				),
			)
		}
	}

	if viper.IsSet("resourceLimits.rateLimitFilesPerSec") {
		rateLimit := viper.GetInt("resourceLimits.rateLimitFilesPerSec")
		if rateLimit < MinRateLimitFilesPerSec {
			errors = append(
				errors,
				fmt.Sprintf(
					"resourceLimits.rateLimitFilesPerSec (%d) must be at least %d",
					rateLimit,
					MinRateLimitFilesPerSec,
				),
			)
		}
		if rateLimit > MaxRateLimitFilesPerSec {
			errors = append(
				errors,
				fmt.Sprintf(
					"resourceLimits.rateLimitFilesPerSec (%d) exceeds maximum (%d)",
					rateLimit,
					MaxRateLimitFilesPerSec,
				),
			)
		}
	}

	return errors
}

// validateResourceLimitsMemory validates memory limit configuration.
func validateResourceLimitsMemory() []string {
	var errors []string
	if !viper.IsSet("resourceLimits.hardMemoryLimitMB") {
		return errors
	}

	memLimit := viper.GetInt("resourceLimits.hardMemoryLimitMB")
	if memLimit < MinHardMemoryLimitMB {
		errors = append(
			errors,
			fmt.Sprintf(
				"resourceLimits.hardMemoryLimitMB (%d) must be at least %d",
				memLimit,
				MinHardMemoryLimitMB,
			),
		)
	}
	if memLimit > MaxHardMemoryLimitMB {
		errors = append(
			errors,
			fmt.Sprintf(
				"resourceLimits.hardMemoryLimitMB (%d) exceeds maximum (%d)",
				memLimit,
				MaxHardMemoryLimitMB,
			),
		)
	}
	return errors
}

// validateResourceLimits validates the resource limits configuration.
func validateResourceLimits() []string {
	var errors []string
	errors = append(errors, validateResourceLimitsMaxFiles()...)
	errors = append(errors, validateResourceLimitsMaxTotalSize()...)
	errors = append(errors, validateResourceLimitsTimeouts()...)
	errors = append(errors, validateResourceLimitsConcurrency()...)
	errors = append(errors, validateResourceLimitsMemory()...)
	return errors
}

// ValidateConfig validates the loaded configuration.
func ValidateConfig() error {
	var validationErrors []string

	// Collect validation errors from all validation helpers
	validationErrors = append(validationErrors, validateFileSizeLimit()...)
	validationErrors = append(validationErrors, validateIgnoreDirectories()...)
	validationErrors = append(validationErrors, validateSupportedFormats()...)
	validationErrors = append(validationErrors, validateConcurrencySettings()...)
	validationErrors = append(validationErrors, validateFilePatterns()...)
	validationErrors = append(validationErrors, validateFileTypes()...)
	validationErrors = append(validationErrors, validateBackpressureConfig()...)
	validationErrors = append(validationErrors, validateResourceLimits()...)

	if len(validationErrors) > 0 {
		return gibidiutils.NewStructuredError(
			gibidiutils.ErrorTypeConfiguration,
			gibidiutils.CodeConfigValidation,
			"configuration validation failed: "+strings.Join(validationErrors, "; "),
			"",
			map[string]interface{}{"validation_errors": validationErrors},
		)
	}

	return nil
}

// ValidateFileSize checks if a file size is within the configured limit.
func ValidateFileSize(size int64) error {
	limit := GetFileSizeLimit()
	if size > limit {
		return gibidiutils.NewStructuredError(
			gibidiutils.ErrorTypeValidation,
			gibidiutils.CodeValidationSize,
			fmt.Sprintf("file size (%d bytes) exceeds limit (%d bytes)", size, limit),
			"",
			map[string]interface{}{"file_size": size, "size_limit": limit},
		)
	}
	return nil
}

// ValidateOutputFormat checks if an output format is valid.
func ValidateOutputFormat(format string) error {
	if !IsValidFormat(format) {
		return gibidiutils.NewStructuredError(
			gibidiutils.ErrorTypeValidation,
			gibidiutils.CodeValidationFormat,
			fmt.Sprintf("unsupported output format: %s (supported: json, yaml, markdown)", format),
			"",
			map[string]interface{}{"format": format},
		)
	}
	return nil
}

// ValidateConcurrency checks if a concurrency level is valid.
func ValidateConcurrency(concurrency int) error {
	if concurrency < 1 {
		return gibidiutils.NewStructuredError(
			gibidiutils.ErrorTypeValidation,
			gibidiutils.CodeValidationFormat,
			fmt.Sprintf("concurrency (%d) must be at least 1", concurrency),
			"",
			map[string]interface{}{"concurrency": concurrency},
		)
	}

	if viper.IsSet("maxConcurrency") {
		maxConcurrency := GetMaxConcurrency()
		if concurrency > maxConcurrency {
			return gibidiutils.NewStructuredError(
				gibidiutils.ErrorTypeValidation,
				gibidiutils.CodeValidationFormat,
				fmt.Sprintf("concurrency (%d) exceeds maximum (%d)", concurrency, maxConcurrency),
				"",
				map[string]interface{}{"concurrency": concurrency, "max_concurrency": maxConcurrency},
			)
		}
	}

	return nil
}
