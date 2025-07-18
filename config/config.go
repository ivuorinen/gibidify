// Package config handles application configuration using Viper.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/ivuorinen/gibidify/utils"
)

const (
	// DefaultFileSizeLimit is the default maximum file size (5MB).
	DefaultFileSizeLimit = 5242880
	// MinFileSizeLimit is the minimum allowed file size limit (1KB).
	MinFileSizeLimit = 1024
	// MaxFileSizeLimit is the maximum allowed file size limit (100MB).
	MaxFileSizeLimit = 104857600

	// Resource Limit Constants

	// DefaultMaxFiles is the default maximum number of files to process.
	DefaultMaxFiles = 10000
	// MinMaxFiles is the minimum allowed file count limit.
	MinMaxFiles = 1
	// MaxMaxFiles is the maximum allowed file count limit.
	MaxMaxFiles = 1000000

	// DefaultMaxTotalSize is the default maximum total size of files (1GB).
	DefaultMaxTotalSize = 1073741824
	// MinMaxTotalSize is the minimum allowed total size limit (1MB).
	MinMaxTotalSize = 1048576
	// MaxMaxTotalSize is the maximum allowed total size limit (100GB).
	MaxMaxTotalSize = 107374182400

	// DefaultFileProcessingTimeoutSec is the default timeout for individual file processing (30 seconds).
	DefaultFileProcessingTimeoutSec = 30
	// MinFileProcessingTimeoutSec is the minimum allowed file processing timeout (1 second).
	MinFileProcessingTimeoutSec = 1
	// MaxFileProcessingTimeoutSec is the maximum allowed file processing timeout (300 seconds).
	MaxFileProcessingTimeoutSec = 300

	// DefaultOverallTimeoutSec is the default timeout for overall processing (3600 seconds = 1 hour).
	DefaultOverallTimeoutSec = 3600
	// MinOverallTimeoutSec is the minimum allowed overall timeout (10 seconds).
	MinOverallTimeoutSec = 10
	// MaxOverallTimeoutSec is the maximum allowed overall timeout (86400 seconds = 24 hours).
	MaxOverallTimeoutSec = 86400

	// DefaultMaxConcurrentReads is the default maximum concurrent file reading operations.
	DefaultMaxConcurrentReads = 10
	// MinMaxConcurrentReads is the minimum allowed concurrent reads.
	MinMaxConcurrentReads = 1
	// MaxMaxConcurrentReads is the maximum allowed concurrent reads.
	MaxMaxConcurrentReads = 100

	// DefaultRateLimitFilesPerSec is the default rate limit for file processing (0 = disabled).
	DefaultRateLimitFilesPerSec = 0
	// MinRateLimitFilesPerSec is the minimum rate limit.
	MinRateLimitFilesPerSec = 0
	// MaxRateLimitFilesPerSec is the maximum rate limit.
	MaxRateLimitFilesPerSec = 10000

	// DefaultHardMemoryLimitMB is the default hard memory limit (512MB).
	DefaultHardMemoryLimitMB = 512
	// MinHardMemoryLimitMB is the minimum hard memory limit (64MB).
	MinHardMemoryLimitMB = 64
	// MaxHardMemoryLimitMB is the maximum hard memory limit (8192MB = 8GB).
	MaxHardMemoryLimitMB = 8192
)

// LoadConfig reads configuration from a YAML file.
// It looks for config in the following order:
// 1. $XDG_CONFIG_HOME/gibidify/config.yaml
// 2. $HOME/.config/gibidify/config.yaml
// 3. The current directory as fallback.
func LoadConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	if xdgConfig := os.Getenv("XDG_CONFIG_HOME"); xdgConfig != "" {
		// Validate XDG_CONFIG_HOME for path traversal attempts
		if err := utils.ValidateConfigPath(xdgConfig); err != nil {
			logrus.Warnf("Invalid XDG_CONFIG_HOME path, using default config: %v", err)
		} else {
			configPath := filepath.Join(xdgConfig, "gibidify")
			viper.AddConfigPath(configPath)
		}
	} else if home, err := os.UserHomeDir(); err == nil {
		viper.AddConfigPath(filepath.Join(home, ".config", "gibidify"))
	}
	// Only add current directory if no config file named gibidify.yaml exists
	// to avoid conflicts with the project's output file
	if _, err := os.Stat("gibidify.yaml"); os.IsNotExist(err) {
		viper.AddConfigPath(".")
	}

	if err := viper.ReadInConfig(); err != nil {
		logrus.Infof("Config file not found, using default values: %v", err)
		setDefaultConfig()
	} else {
		logrus.Infof("Using config file: %s", viper.ConfigFileUsed())
		// Validate configuration after loading
		if err := ValidateConfig(); err != nil {
			logrus.Warnf("Configuration validation failed: %v", err)
			logrus.Info("Falling back to default configuration")
			// Reset viper and set defaults when validation fails
			viper.Reset()
			setDefaultConfig()
		}
	}
}

// setDefaultConfig sets default configuration values.
func setDefaultConfig() {
	viper.SetDefault("fileSizeLimit", DefaultFileSizeLimit)
	// Default ignored directories.
	viper.SetDefault("ignoreDirectories", []string{
		"vendor", "node_modules", ".git", "dist", "build", "target", "bower_components", "cache", "tmp",
	})

	// FileTypeRegistry defaults
	viper.SetDefault("fileTypes.enabled", true)
	viper.SetDefault("fileTypes.customImageExtensions", []string{})
	viper.SetDefault("fileTypes.customBinaryExtensions", []string{})
	viper.SetDefault("fileTypes.customLanguages", map[string]string{})
	viper.SetDefault("fileTypes.disabledImageExtensions", []string{})
	viper.SetDefault("fileTypes.disabledBinaryExtensions", []string{})
	viper.SetDefault("fileTypes.disabledLanguageExtensions", []string{})

	// Back-pressure and memory management defaults
	viper.SetDefault("backpressure.enabled", true)
	viper.SetDefault("backpressure.maxPendingFiles", 1000)     // Max files in file channel buffer
	viper.SetDefault("backpressure.maxPendingWrites", 100)     // Max writes in write channel buffer
	viper.SetDefault("backpressure.maxMemoryUsage", 104857600) // 100MB max memory usage
	viper.SetDefault("backpressure.memoryCheckInterval", 1000) // Check memory every 1000 files

	// Resource limit defaults
	viper.SetDefault("resourceLimits.enabled", true)
	viper.SetDefault("resourceLimits.maxFiles", DefaultMaxFiles)
	viper.SetDefault("resourceLimits.maxTotalSize", DefaultMaxTotalSize)
	viper.SetDefault("resourceLimits.fileProcessingTimeoutSec", DefaultFileProcessingTimeoutSec)
	viper.SetDefault("resourceLimits.overallTimeoutSec", DefaultOverallTimeoutSec)
	viper.SetDefault("resourceLimits.maxConcurrentReads", DefaultMaxConcurrentReads)
	viper.SetDefault("resourceLimits.rateLimitFilesPerSec", DefaultRateLimitFilesPerSec)
	viper.SetDefault("resourceLimits.hardMemoryLimitMB", DefaultHardMemoryLimitMB)
	viper.SetDefault("resourceLimits.enableGracefulDegradation", true)
	viper.SetDefault("resourceLimits.enableResourceMonitoring", true)
}

// GetFileSizeLimit returns the file size limit from configuration.
func GetFileSizeLimit() int64 {
	return viper.GetInt64("fileSizeLimit")
}

// GetIgnoredDirectories returns the list of directories to ignore.
func GetIgnoredDirectories() []string {
	return viper.GetStringSlice("ignoreDirectories")
}

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
		return utils.NewStructuredError(
			utils.ErrorTypeConfiguration,
			utils.CodeConfigValidation,
			"configuration validation failed: "+strings.Join(validationErrors, "; "),
			"",
			map[string]interface{}{"validation_errors": validationErrors},
		)
	}

	return nil
}

// GetMaxConcurrency returns the maximum concurrency limit from configuration.
func GetMaxConcurrency() int {
	return viper.GetInt("maxConcurrency")
}

// GetSupportedFormats returns the supported output formats from configuration.
func GetSupportedFormats() []string {
	return viper.GetStringSlice("supportedFormats")
}

// GetFilePatterns returns the file patterns from configuration.
func GetFilePatterns() []string {
	return viper.GetStringSlice("filePatterns")
}

// IsValidFormat checks if a format is supported.
func IsValidFormat(format string) bool {
	format = strings.ToLower(strings.TrimSpace(format))
	validFormats := map[string]bool{"json": true, "yaml": true, "markdown": true}
	return validFormats[format]
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
			map[string]interface{}{"file_size": size, "size_limit": limit},
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
			map[string]interface{}{"format": format},
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
			map[string]interface{}{"concurrency": concurrency},
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
				map[string]interface{}{"concurrency": concurrency, "max_concurrency": maxConcurrency},
			)
		}
	}

	return nil
}

// GetFileTypesEnabled returns whether file type detection is enabled.
func GetFileTypesEnabled() bool {
	return viper.GetBool("fileTypes.enabled")
}

// GetCustomImageExtensions returns custom image extensions from configuration.
func GetCustomImageExtensions() []string {
	return viper.GetStringSlice("fileTypes.customImageExtensions")
}

// GetCustomBinaryExtensions returns custom binary extensions from configuration.
func GetCustomBinaryExtensions() []string {
	return viper.GetStringSlice("fileTypes.customBinaryExtensions")
}

// GetCustomLanguages returns custom language mappings from configuration.
func GetCustomLanguages() map[string]string {
	return viper.GetStringMapString("fileTypes.customLanguages")
}

// GetDisabledImageExtensions returns disabled image extensions from configuration.
func GetDisabledImageExtensions() []string {
	return viper.GetStringSlice("fileTypes.disabledImageExtensions")
}

// GetDisabledBinaryExtensions returns disabled binary extensions from configuration.
func GetDisabledBinaryExtensions() []string {
	return viper.GetStringSlice("fileTypes.disabledBinaryExtensions")
}

// GetDisabledLanguageExtensions returns disabled language extensions from configuration.
func GetDisabledLanguageExtensions() []string {
	return viper.GetStringSlice("fileTypes.disabledLanguageExtensions")
}

// Back-pressure configuration getters

// GetBackpressureEnabled returns whether back-pressure management is enabled.
func GetBackpressureEnabled() bool {
	return viper.GetBool("backpressure.enabled")
}

// GetMaxPendingFiles returns the maximum number of files that can be pending in the file channel.
func GetMaxPendingFiles() int {
	return viper.GetInt("backpressure.maxPendingFiles")
}

// GetMaxPendingWrites returns the maximum number of writes that can be pending in the write channel.
func GetMaxPendingWrites() int {
	return viper.GetInt("backpressure.maxPendingWrites")
}

// GetMaxMemoryUsage returns the maximum memory usage in bytes before back-pressure kicks in.
func GetMaxMemoryUsage() int64 {
	return viper.GetInt64("backpressure.maxMemoryUsage")
}

// GetMemoryCheckInterval returns how often to check memory usage (in number of files processed).
func GetMemoryCheckInterval() int {
	return viper.GetInt("backpressure.memoryCheckInterval")
}

// Resource Limit Configuration Getters

// GetResourceLimitsEnabled returns whether resource limits are enabled.
func GetResourceLimitsEnabled() bool {
	return viper.GetBool("resourceLimits.enabled")
}

// GetMaxFiles returns the maximum number of files that can be processed.
func GetMaxFiles() int {
	return viper.GetInt("resourceLimits.maxFiles")
}

// GetMaxTotalSize returns the maximum total size of files that can be processed.
func GetMaxTotalSize() int64 {
	return viper.GetInt64("resourceLimits.maxTotalSize")
}

// GetFileProcessingTimeoutSec returns the timeout for individual file processing in seconds.
func GetFileProcessingTimeoutSec() int {
	return viper.GetInt("resourceLimits.fileProcessingTimeoutSec")
}

// GetOverallTimeoutSec returns the timeout for overall processing in seconds.
func GetOverallTimeoutSec() int {
	return viper.GetInt("resourceLimits.overallTimeoutSec")
}

// GetMaxConcurrentReads returns the maximum number of concurrent file reading operations.
func GetMaxConcurrentReads() int {
	return viper.GetInt("resourceLimits.maxConcurrentReads")
}

// GetRateLimitFilesPerSec returns the rate limit for file processing (files per second).
func GetRateLimitFilesPerSec() int {
	return viper.GetInt("resourceLimits.rateLimitFilesPerSec")
}

// GetHardMemoryLimitMB returns the hard memory limit in megabytes.
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
