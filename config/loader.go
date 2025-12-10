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
	viper.SetConfigType(shared.FormatYAML)

	logger := shared.GetLogger()

	if xdgConfig := os.Getenv("XDG_CONFIG_HOME"); xdgConfig != "" {
		// Validate XDG_CONFIG_HOME for path traversal attempts
		if err := shared.ValidateConfigPath(xdgConfig); err != nil {
			logger.Warnf("Invalid XDG_CONFIG_HOME path, using default config: %v", err)
		} else {
			configPath := filepath.Join(xdgConfig, shared.AppName)
			viper.AddConfigPath(configPath)
		}
	} else if home, err := os.UserHomeDir(); err == nil {
		viper.AddConfigPath(filepath.Join(home, ".config", shared.AppName))
	}
	// Only add current directory if no config file named gibidify.yaml exists
	// to avoid conflicts with the project's output file
	if _, err := os.Stat(shared.AppName + ".yaml"); os.IsNotExist(err) {
		viper.AddConfigPath(".")
	}

	if err := viper.ReadInConfig(); err != nil {
		logger.Infof("Config file not found, using default values: %v", err)
		SetDefaultConfig()
	} else {
		logger.Infof("Using config file: %s", viper.ConfigFileUsed())
		// Validate configuration after loading
		if err := ValidateConfig(); err != nil {
			logger.Warnf("Configuration validation failed: %v", err)
			logger.Info("Falling back to default configuration")
			// Reset viper and set defaults when validation fails
			viper.Reset()
			SetDefaultConfig()
		}
	}
}

// SetDefaultConfig sets default configuration values.
func SetDefaultConfig() {
	// File size limits
	viper.SetDefault(shared.ConfigKeyFileSizeLimit, shared.ConfigFileSizeLimitDefault)
	viper.SetDefault(shared.ConfigKeyIgnoreDirectories, shared.ConfigIgnoredDirectoriesDefault)
	viper.SetDefault(shared.ConfigKeyMaxConcurrency, shared.ConfigMaxConcurrencyDefault)
	viper.SetDefault(shared.ConfigKeySupportedFormats, shared.ConfigSupportedFormatsDefault)
	viper.SetDefault(shared.ConfigKeyFilePatterns, shared.ConfigFilePatternsDefault)

	// FileTypeRegistry defaults
	viper.SetDefault(shared.ConfigKeyFileTypesEnabled, shared.ConfigFileTypesEnabledDefault)
	viper.SetDefault(shared.ConfigKeyFileTypesCustomImageExtensions, shared.ConfigCustomImageExtensionsDefault)
	viper.SetDefault(shared.ConfigKeyFileTypesCustomBinaryExtensions, shared.ConfigCustomBinaryExtensionsDefault)
	viper.SetDefault(shared.ConfigKeyFileTypesCustomLanguages, shared.ConfigCustomLanguagesDefault)
	viper.SetDefault(shared.ConfigKeyFileTypesDisabledImageExtensions, shared.ConfigDisabledImageExtensionsDefault)
	viper.SetDefault(shared.ConfigKeyFileTypesDisabledBinaryExtensions, shared.ConfigDisabledBinaryExtensionsDefault)
	viper.SetDefault(shared.ConfigKeyFileTypesDisabledLanguageExts, shared.ConfigDisabledLanguageExtensionsDefault)

	// Backpressure and memory management defaults
	viper.SetDefault(shared.ConfigKeyBackpressureEnabled, shared.ConfigBackpressureEnabledDefault)
	viper.SetDefault(shared.ConfigKeyBackpressureMaxPendingFiles, shared.ConfigMaxPendingFilesDefault)
	viper.SetDefault(shared.ConfigKeyBackpressureMaxPendingWrites, shared.ConfigMaxPendingWritesDefault)
	viper.SetDefault(shared.ConfigKeyBackpressureMaxMemoryUsage, shared.ConfigMaxMemoryUsageDefault)
	viper.SetDefault(shared.ConfigKeyBackpressureMemoryCheckInt, shared.ConfigMemoryCheckIntervalDefault)

	// Resource limit defaults
	viper.SetDefault(shared.ConfigKeyResourceLimitsEnabled, shared.ConfigResourceLimitsEnabledDefault)
	viper.SetDefault(shared.ConfigKeyResourceLimitsMaxFiles, shared.ConfigMaxFilesDefault)
	viper.SetDefault(shared.ConfigKeyResourceLimitsMaxTotalSize, shared.ConfigMaxTotalSizeDefault)
	viper.SetDefault(shared.ConfigKeyResourceLimitsFileProcessingTO, shared.ConfigFileProcessingTimeoutSecDefault)
	viper.SetDefault(shared.ConfigKeyResourceLimitsOverallTO, shared.ConfigOverallTimeoutSecDefault)
	viper.SetDefault(shared.ConfigKeyResourceLimitsMaxConcurrentReads, shared.ConfigMaxConcurrentReadsDefault)
	viper.SetDefault(shared.ConfigKeyResourceLimitsRateLimitFilesPerSec, shared.ConfigRateLimitFilesPerSecDefault)
	viper.SetDefault(shared.ConfigKeyResourceLimitsHardMemoryLimitMB, shared.ConfigHardMemoryLimitMBDefault)
	viper.SetDefault(shared.ConfigKeyResourceLimitsEnableGracefulDeg, shared.ConfigEnableGracefulDegradationDefault)
	viper.SetDefault(shared.ConfigKeyResourceLimitsEnableMonitoring, shared.ConfigEnableResourceMonitoringDefault)

	// Output configuration defaults
	viper.SetDefault(shared.ConfigKeyOutputTemplate, shared.ConfigOutputTemplateDefault)
	viper.SetDefault("output.metadata.includeStats", shared.ConfigMetadataIncludeStatsDefault)
	viper.SetDefault("output.metadata.includeTimestamp", shared.ConfigMetadataIncludeTimestampDefault)
	viper.SetDefault("output.metadata.includeFileCount", shared.ConfigMetadataIncludeFileCountDefault)
	viper.SetDefault("output.metadata.includeSourcePath", shared.ConfigMetadataIncludeSourcePathDefault)
	viper.SetDefault("output.metadata.includeFileTypes", shared.ConfigMetadataIncludeFileTypesDefault)
	viper.SetDefault("output.metadata.includeProcessingTime", shared.ConfigMetadataIncludeProcessingTimeDefault)
	viper.SetDefault("output.metadata.includeTotalSize", shared.ConfigMetadataIncludeTotalSizeDefault)
	viper.SetDefault("output.metadata.includeMetrics", shared.ConfigMetadataIncludeMetricsDefault)
	viper.SetDefault("output.markdown.useCodeBlocks", shared.ConfigMarkdownUseCodeBlocksDefault)
	viper.SetDefault("output.markdown.includeLanguage", shared.ConfigMarkdownIncludeLanguageDefault)
	viper.SetDefault(shared.ConfigKeyOutputMarkdownHeaderLevel, shared.ConfigMarkdownHeaderLevelDefault)
	viper.SetDefault("output.markdown.tableOfContents", shared.ConfigMarkdownTableOfContentsDefault)
	viper.SetDefault("output.markdown.useCollapsible", shared.ConfigMarkdownUseCollapsibleDefault)
	viper.SetDefault("output.markdown.syntaxHighlighting", shared.ConfigMarkdownSyntaxHighlightingDefault)
	viper.SetDefault("output.markdown.lineNumbers", shared.ConfigMarkdownLineNumbersDefault)
	viper.SetDefault("output.markdown.foldLongFiles", shared.ConfigMarkdownFoldLongFilesDefault)
	viper.SetDefault(shared.ConfigKeyOutputMarkdownMaxLineLen, shared.ConfigMarkdownMaxLineLengthDefault)
	viper.SetDefault(shared.ConfigKeyOutputMarkdownCustomCSS, shared.ConfigMarkdownCustomCSSDefault)
	viper.SetDefault(shared.ConfigKeyOutputCustomHeader, shared.ConfigCustomHeaderDefault)
	viper.SetDefault(shared.ConfigKeyOutputCustomFooter, shared.ConfigCustomFooterDefault)
	viper.SetDefault(shared.ConfigKeyOutputCustomFileHeader, shared.ConfigCustomFileHeaderDefault)
	viper.SetDefault(shared.ConfigKeyOutputCustomFileFooter, shared.ConfigCustomFileFooterDefault)
	viper.SetDefault(shared.ConfigKeyOutputVariables, shared.ConfigTemplateVariablesDefault)
}
