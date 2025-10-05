package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"

	"github.com/ivuorinen/gibidify/gibidiutils"
)

// ValidateConfig validates the loaded configuration.
func ValidateConfig() error {
	var validationErrors []string

	// Validate file size limit
	fileSizeLimit := viper.GetInt64("fileSizeLimit")
	if fileSizeLimit < MinFileSizeLimit {
		validationErrors = append(validationErrors, fmt.Sprintf("fileSizeLimit (%d) is below minimum (%d)", fileSizeLimit, MinFileSizeLimit))
	}
	if fileSizeLimit > MaxFileSizeLimit {
		validationErrors = append(validationErrors, fmt.Sprintf("fileSizeLimit (%d) exceeds maximum (%d)", fileSizeLimit, MaxFileSizeLimit))
	}

	// Validate ignore directories
	ignoreDirectories := viper.GetStringSlice("ignoreDirectories")
	for i, dir := range ignoreDirectories {
		dir = strings.TrimSpace(dir)
		if dir == "" {
			validationErrors = append(validationErrors, fmt.Sprintf("ignoreDirectories[%d] is empty", i))
			continue
		}
		if strings.Contains(dir, "/") {
			validationErrors = append(validationErrors, fmt.Sprintf("ignoreDirectories[%d] (%s) contains path separator - only directory names are allowed", i, dir))
		}
		if strings.HasPrefix(dir, ".") && dir != ".git" && dir != ".vscode" && dir != ".idea" {
			validationErrors = append(validationErrors, fmt.Sprintf("ignoreDirectories[%d] (%s) starts with dot - this may cause unexpected behavior", i, dir))
		}
	}

	// Validate supported output formats if configured
	if viper.IsSet("supportedFormats") {
		supportedFormats := viper.GetStringSlice("supportedFormats")
		validFormats := map[string]bool{"json": true, "yaml": true, "markdown": true}
		for i, format := range supportedFormats {
			format = strings.ToLower(strings.TrimSpace(format))
			if !validFormats[format] {
				validationErrors = append(validationErrors, fmt.Sprintf("supportedFormats[%d] (%s) is not a valid format (json, yaml, markdown)", i, format))
			}
		}
	}

	// Validate concurrency settings if configured
	if viper.IsSet("maxConcurrency") {
		maxConcurrency := viper.GetInt("maxConcurrency")
		if maxConcurrency < 1 {
			validationErrors = append(validationErrors, fmt.Sprintf("maxConcurrency (%d) must be at least 1", maxConcurrency))
		}
		if maxConcurrency > 100 {
			validationErrors = append(validationErrors, fmt.Sprintf("maxConcurrency (%d) is unreasonably high (max 100)", maxConcurrency))
		}
	}

	// Validate file patterns if configured
	if viper.IsSet("filePatterns") {
		filePatterns := viper.GetStringSlice("filePatterns")
		for i, pattern := range filePatterns {
			pattern = strings.TrimSpace(pattern)
			if pattern == "" {
				validationErrors = append(validationErrors, fmt.Sprintf("filePatterns[%d] is empty", i))
				continue
			}
			// Basic validation - patterns should contain at least one alphanumeric character
			if !strings.ContainsAny(pattern, "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789") {
				validationErrors = append(validationErrors, fmt.Sprintf("filePatterns[%d] (%s) appears to be invalid", i, pattern))
			}
		}
	}

	// Validate FileTypeRegistry configuration
	if viper.IsSet("fileTypes.customImageExtensions") {
		customImages := viper.GetStringSlice("fileTypes.customImageExtensions")
		for i, ext := range customImages {
			ext = strings.TrimSpace(ext)
			if ext == "" {
				validationErrors = append(validationErrors, fmt.Sprintf("fileTypes.customImageExtensions[%d] is empty", i))
				continue
			}
			if !strings.HasPrefix(ext, ".") {
				validationErrors = append(validationErrors, fmt.Sprintf("fileTypes.customImageExtensions[%d] (%s) must start with a dot", i, ext))
			}
		}
	}

	if viper.IsSet("fileTypes.customBinaryExtensions") {
		customBinary := viper.GetStringSlice("fileTypes.customBinaryExtensions")
		for i, ext := range customBinary {
			ext = strings.TrimSpace(ext)
			if ext == "" {
				validationErrors = append(validationErrors, fmt.Sprintf("fileTypes.customBinaryExtensions[%d] is empty", i))
				continue
			}
			if !strings.HasPrefix(ext, ".") {
				validationErrors = append(validationErrors, fmt.Sprintf("fileTypes.customBinaryExtensions[%d] (%s) must start with a dot", i, ext))
			}
		}
	}

	if viper.IsSet("fileTypes.customLanguages") {
		customLangs := viper.GetStringMapString("fileTypes.customLanguages")
		for ext, lang := range customLangs {
			ext = strings.TrimSpace(ext)
			lang = strings.TrimSpace(lang)
			if ext == "" {
				validationErrors = append(validationErrors, "fileTypes.customLanguages contains empty extension key")
				continue
			}
			if !strings.HasPrefix(ext, ".") {
				validationErrors = append(validationErrors, fmt.Sprintf("fileTypes.customLanguages extension (%s) must start with a dot", ext))
			}
			if lang == "" {
				validationErrors = append(validationErrors, fmt.Sprintf("fileTypes.customLanguages[%s] has empty language value", ext))
			}
		}
	}

	// Validate back-pressure configuration
	if viper.IsSet("backpressure.maxPendingFiles") {
		maxPendingFiles := viper.GetInt("backpressure.maxPendingFiles")
		if maxPendingFiles < 1 {
			validationErrors = append(validationErrors, fmt.Sprintf("backpressure.maxPendingFiles (%d) must be at least 1", maxPendingFiles))
		}
		if maxPendingFiles > 100000 {
			validationErrors = append(validationErrors, fmt.Sprintf("backpressure.maxPendingFiles (%d) is unreasonably high (max 100000)", maxPendingFiles))
		}
	}

	if viper.IsSet("backpressure.maxPendingWrites") {
		maxPendingWrites := viper.GetInt("backpressure.maxPendingWrites")
		if maxPendingWrites < 1 {
			validationErrors = append(validationErrors, fmt.Sprintf("backpressure.maxPendingWrites (%d) must be at least 1", maxPendingWrites))
		}
		if maxPendingWrites > 10000 {
			validationErrors = append(validationErrors, fmt.Sprintf("backpressure.maxPendingWrites (%d) is unreasonably high (max 10000)", maxPendingWrites))
		}
	}

	if viper.IsSet("backpressure.maxMemoryUsage") {
		maxMemoryUsage := viper.GetInt64("backpressure.maxMemoryUsage")
		if maxMemoryUsage < 1048576 { // 1MB minimum
			validationErrors = append(validationErrors, fmt.Sprintf("backpressure.maxMemoryUsage (%d) must be at least 1MB (1048576 bytes)", maxMemoryUsage))
		}
		if maxMemoryUsage > 10737418240 { // 10GB maximum
			validationErrors = append(validationErrors, fmt.Sprintf("backpressure.maxMemoryUsage (%d) is unreasonably high (max 10GB)", maxMemoryUsage))
		}
	}

	if viper.IsSet("backpressure.memoryCheckInterval") {
		interval := viper.GetInt("backpressure.memoryCheckInterval")
		if interval < 1 {
			validationErrors = append(validationErrors, fmt.Sprintf("backpressure.memoryCheckInterval (%d) must be at least 1", interval))
		}
		if interval > 100000 {
			validationErrors = append(validationErrors, fmt.Sprintf("backpressure.memoryCheckInterval (%d) is unreasonably high (max 100000)", interval))
		}
	}

	// Validate resource limits configuration
	if viper.IsSet("resourceLimits.maxFiles") {
		maxFiles := viper.GetInt("resourceLimits.maxFiles")
		if maxFiles < MinMaxFiles {
			validationErrors = append(validationErrors, fmt.Sprintf("resourceLimits.maxFiles (%d) must be at least %d", maxFiles, MinMaxFiles))
		}
		if maxFiles > MaxMaxFiles {
			validationErrors = append(validationErrors, fmt.Sprintf("resourceLimits.maxFiles (%d) exceeds maximum (%d)", maxFiles, MaxMaxFiles))
		}
	}

	if viper.IsSet("resourceLimits.maxTotalSize") {
		maxTotalSize := viper.GetInt64("resourceLimits.maxTotalSize")
		if maxTotalSize < MinMaxTotalSize {
			validationErrors = append(validationErrors, fmt.Sprintf("resourceLimits.maxTotalSize (%d) must be at least %d", maxTotalSize, MinMaxTotalSize))
		}
		if maxTotalSize > MaxMaxTotalSize {
			validationErrors = append(validationErrors, fmt.Sprintf("resourceLimits.maxTotalSize (%d) exceeds maximum (%d)", maxTotalSize, MaxMaxTotalSize))
		}
	}

	if viper.IsSet("resourceLimits.fileProcessingTimeoutSec") {
		timeout := viper.GetInt("resourceLimits.fileProcessingTimeoutSec")
		if timeout < MinFileProcessingTimeoutSec {
			validationErrors = append(validationErrors, fmt.Sprintf("resourceLimits.fileProcessingTimeoutSec (%d) must be at least %d", timeout, MinFileProcessingTimeoutSec))
		}
		if timeout > MaxFileProcessingTimeoutSec {
			validationErrors = append(validationErrors, fmt.Sprintf("resourceLimits.fileProcessingTimeoutSec (%d) exceeds maximum (%d)", timeout, MaxFileProcessingTimeoutSec))
		}
	}

	if viper.IsSet("resourceLimits.overallTimeoutSec") {
		timeout := viper.GetInt("resourceLimits.overallTimeoutSec")
		if timeout < MinOverallTimeoutSec {
			validationErrors = append(validationErrors, fmt.Sprintf("resourceLimits.overallTimeoutSec (%d) must be at least %d", timeout, MinOverallTimeoutSec))
		}
		if timeout > MaxOverallTimeoutSec {
			validationErrors = append(validationErrors, fmt.Sprintf("resourceLimits.overallTimeoutSec (%d) exceeds maximum (%d)", timeout, MaxOverallTimeoutSec))
		}
	}

	if viper.IsSet("resourceLimits.maxConcurrentReads") {
		maxReads := viper.GetInt("resourceLimits.maxConcurrentReads")
		if maxReads < MinMaxConcurrentReads {
			validationErrors = append(validationErrors, fmt.Sprintf("resourceLimits.maxConcurrentReads (%d) must be at least %d", maxReads, MinMaxConcurrentReads))
		}
		if maxReads > MaxMaxConcurrentReads {
			validationErrors = append(validationErrors, fmt.Sprintf("resourceLimits.maxConcurrentReads (%d) exceeds maximum (%d)", maxReads, MaxMaxConcurrentReads))
		}
	}

	if viper.IsSet("resourceLimits.rateLimitFilesPerSec") {
		rateLimit := viper.GetInt("resourceLimits.rateLimitFilesPerSec")
		if rateLimit < MinRateLimitFilesPerSec {
			validationErrors = append(validationErrors, fmt.Sprintf("resourceLimits.rateLimitFilesPerSec (%d) must be at least %d", rateLimit, MinRateLimitFilesPerSec))
		}
		if rateLimit > MaxRateLimitFilesPerSec {
			validationErrors = append(validationErrors, fmt.Sprintf("resourceLimits.rateLimitFilesPerSec (%d) exceeds maximum (%d)", rateLimit, MaxRateLimitFilesPerSec))
		}
	}

	if viper.IsSet("resourceLimits.hardMemoryLimitMB") {
		memLimit := viper.GetInt("resourceLimits.hardMemoryLimitMB")
		if memLimit < MinHardMemoryLimitMB {
			validationErrors = append(validationErrors, fmt.Sprintf("resourceLimits.hardMemoryLimitMB (%d) must be at least %d", memLimit, MinHardMemoryLimitMB))
		}
		if memLimit > MaxHardMemoryLimitMB {
			validationErrors = append(validationErrors, fmt.Sprintf("resourceLimits.hardMemoryLimitMB (%d) exceeds maximum (%d)", memLimit, MaxHardMemoryLimitMB))
		}
	}

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
