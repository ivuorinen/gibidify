// Package config handles application configuration management.
package config

import (
	"fmt"
	"strings"

	"github.com/ivuorinen/gibidify/shared"
)

// ValidateConfig validates the loaded configuration.
func ValidateConfig() error {
	var validationErrors []string

	// Validate basic settings
	validationErrors = append(validationErrors, validateBasicSettings()...)
	validationErrors = append(validationErrors, validateFileTypeSettings()...)
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

	fileSizeLimit := GetInt64(shared.ConfigKeyFileSizeLimit)
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

	ignoreDirectories := GetStringSlice(shared.ConfigKeyIgnoreDirectories)
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

	if !IsSet(shared.ConfigKeySupportedFormats) {
		return validationErrors
	}

	supportedFormats := GetStringSlice(shared.ConfigKeySupportedFormats)
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

	if !IsSet(shared.ConfigKeyMaxConcurrency) {
		return validationErrors
	}

	maxConcurrency := GetInt(shared.ConfigKeyMaxConcurrency)
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

	if !IsSet(shared.ConfigKeyFilePatterns) {
		return validationErrors
	}

	filePatterns := GetStringSlice(shared.ConfigKeyFilePatterns)
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

	validationErrors = append(validationErrors,
		validateCustomExtensions(shared.ConfigKeyFileTypesCustomImageExtensions)...)
	validationErrors = append(validationErrors,
		validateCustomExtensions(shared.ConfigKeyFileTypesCustomBinaryExtensions)...)
	validationErrors = append(validationErrors, validateCustomLanguages()...)

	return validationErrors
}

// validateCustomExtensions validates a custom extension slice under the given config key
// (image or binary): each element must be non-empty and dot-prefixed.
func validateCustomExtensions(key string) []string {
	var validationErrors []string

	if !IsSet(key) {
		return validationErrors
	}

	for i, ext := range GetStringSlice(key) {
		if errMsg := validateEmptyElement(key, ext, i); errMsg != "" {
			validationErrors = append(validationErrors, errMsg)

			continue
		}
		ext = strings.TrimSpace(ext)
		if errMsg := validateDotPrefix(key, ext, i); errMsg != "" {
			validationErrors = append(validationErrors, errMsg)
		}
	}

	return validationErrors
}

// validateCustomLanguages validates custom language mappings.
func validateCustomLanguages() []string {
	var validationErrors []string

	if !IsSet(shared.ConfigKeyFileTypesCustomLanguages) {
		return validationErrors
	}

	customLangs := GetStringMapString(shared.ConfigKeyFileTypesCustomLanguages)
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

// validateResourceLimitSettings validates the file-count and total-size caps.
func validateResourceLimitSettings() []string {
	var validationErrors []string

	validationErrors = append(validationErrors, validateMaxFilesLimit()...)
	validationErrors = append(validationErrors, validateMaxTotalSizeLimit()...)

	return validationErrors
}

// validateMaxFilesLimit validates resourceLimits.maxFiles setting.
func validateMaxFilesLimit() []string {
	var validationErrors []string

	if !IsSet(shared.ConfigKeyResourceLimitsMaxFiles) {
		return validationErrors
	}

	maxFiles := GetInt(shared.ConfigKeyResourceLimitsMaxFiles)
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

	if !IsSet(shared.ConfigKeyResourceLimitsMaxTotalSize) {
		return validationErrors
	}

	maxTotalSize := GetInt64(shared.ConfigKeyResourceLimitsMaxTotalSize)
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

	if IsSet(shared.ConfigKeyMaxConcurrency) {
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
