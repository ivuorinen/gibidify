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
		viper.AddConfigPath(filepath.Join(xdgConfig, "gibidify"))
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

	if len(validationErrors) > 0 {
		return utils.NewStructuredError(
			utils.ErrorTypeConfiguration,
			utils.CodeConfigValidation,
			"configuration validation failed: "+strings.Join(validationErrors, "; "),
		).WithContext("validation_errors", validationErrors)
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
		).WithContext("file_size", size).WithContext("size_limit", limit)
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
		).WithContext("format", format)
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
		).WithContext("concurrency", concurrency)
	}

	if viper.IsSet("maxConcurrency") {
		maxConcurrency := GetMaxConcurrency()
		if concurrency > maxConcurrency {
			return utils.NewStructuredError(
				utils.ErrorTypeValidation,
				utils.CodeValidationFormat,
				fmt.Sprintf("concurrency (%d) exceeds maximum (%d)", concurrency, maxConcurrency),
			).WithContext("concurrency", concurrency).WithContext("max_concurrency", maxConcurrency)
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
