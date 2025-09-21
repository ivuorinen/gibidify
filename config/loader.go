// Package config handles application configuration management.
package config

import (
	"os"
	"path/filepath"

	"github.com/spf13/viper"

	"github.com/ivuorinen/gibidify/shared"
)

// LoadConfig reads configuration from a YAML file.
// It looks for config in the following order:
// 1. $XDG_CONFIG_HOME/gibidify/config.yaml
// 2. $HOME/.config/gibidify/config.yaml
// 3. The current directory as fallback.
func LoadConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	logger := shared.GetLogger()

	if xdgConfig := os.Getenv("XDG_CONFIG_HOME"); xdgConfig != "" {
		// Validate XDG_CONFIG_HOME for path traversal attempts
		if err := shared.ValidateConfigPath(xdgConfig); err != nil {
			logger.Warnf("Invalid XDG_CONFIG_HOME path, using default config: %v", err)
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
		logger.Infof("Config file not found, using default values: %v", err)
		setDefaultConfig()
	} else {
		logger.Infof("Using config file: %s", viper.ConfigFileUsed())
		// Validate configuration after loading
		if err := ValidateConfig(); err != nil {
			logger.Warnf("Configuration validation failed: %v", err)
			logger.Info("Falling back to default configuration")
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
	viper.SetDefault(
		"ignoreDirectories", []string{
			"vendor", "node_modules", ".git", "dist", "build", "target", "bower_components", "cache", "tmp",
		},
	)

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

	// Output configuration defaults
	viper.SetDefault("output.template", "")
	viper.SetDefault("output.metadata.includeStats", false)
	viper.SetDefault("output.metadata.includeTimestamp", false)
	viper.SetDefault("output.metadata.includeFileCount", false)
	viper.SetDefault("output.metadata.includeSourcePath", false)
	viper.SetDefault("output.metadata.includeFileTypes", false)
	viper.SetDefault("output.metadata.includeProcessingTime", false)
	viper.SetDefault("output.metadata.includeTotalSize", false)
	viper.SetDefault("output.metadata.includeMetrics", false)
	viper.SetDefault("output.markdown.useCodeBlocks", false)
	viper.SetDefault("output.markdown.includeLanguage", false)
	viper.SetDefault("output.markdown.headerLevel", 0)
	viper.SetDefault("output.markdown.tableOfContents", false)
	viper.SetDefault("output.markdown.useCollapsible", false)
	viper.SetDefault("output.markdown.syntaxHighlighting", false)
	viper.SetDefault("output.markdown.lineNumbers", false)
	viper.SetDefault("output.markdown.foldLongFiles", false)
	viper.SetDefault("output.markdown.maxLineLength", 0)
	viper.SetDefault("output.markdown.customCSS", "")
	viper.SetDefault("output.custom.header", "")
	viper.SetDefault("output.custom.footer", "")
	viper.SetDefault("output.custom.fileHeader", "")
	viper.SetDefault("output.custom.fileFooter", "")
	viper.SetDefault("output.variables", map[string]string{})
}
