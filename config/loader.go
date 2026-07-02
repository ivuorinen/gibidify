// Package config handles application configuration management.
package config

import (
	"os"
	"path/filepath"

	"github.com/ivuorinen/gibidify/shared"
)

// LoadConfig reads configuration from a YAML file.
// It looks for config in the following order:
// 1. $XDG_CONFIG_HOME/gibidify/config.yaml
// 2. $HOME/.config/gibidify/config.yaml
// 3. The current directory as fallback.
func LoadConfig() {
	logger := shared.GetLogger()

	if xdgConfig := os.Getenv("XDG_CONFIG_HOME"); xdgConfig != "" {
		// Validate XDG_CONFIG_HOME for path traversal attempts
		if err := shared.ValidateConfigPath(xdgConfig); err != nil {
			logger.Warnf("Invalid XDG_CONFIG_HOME path, using default config: %v", err)
		} else {
			AddConfigPath(filepath.Join(xdgConfig, shared.AppName))
		}
	} else if home, err := os.UserHomeDir(); err == nil {
		AddConfigPath(filepath.Join(home, ".config", shared.AppName))
	}
	// Only add current directory if no config file named gibidify.yaml exists
	// to avoid conflicts with the project's output file
	if _, err := os.Stat(shared.AppName + ".yaml"); os.IsNotExist(err) {
		AddConfigPath(".")
	}

	if err := ReadInConfig(); err != nil {
		logger.Infof("Config file not found, using default values: %v", err)
		SetDefaultConfig()

		return
	}

	logger.Infof("Using config file: %s", FileUsed())
	// Validate configuration after loading
	if err := ValidateConfig(); err != nil {
		logger.Warnf("Configuration validation failed: %v", err)
		logger.Info("Falling back to default configuration")
		// Reset and set defaults when validation fails
		Reset()
		SetDefaultConfig()
	}
}

// SetDefaultConfig sets default configuration values.
func SetDefaultConfig() {
	// File size limits
	SetDefault(shared.ConfigKeyFileSizeLimit, shared.ConfigFileSizeLimitDefault)
	SetDefault(shared.ConfigKeyIgnoreDirectories, shared.ConfigIgnoredDirectoriesDefault)
	SetDefault(shared.ConfigKeyMaxConcurrency, shared.ConfigMaxConcurrencyDefault)

	// FileTypeRegistry defaults
	SetDefault(shared.ConfigKeyFileTypesEnabled, shared.ConfigFileTypesEnabledDefault)
	SetDefault(shared.ConfigKeyFileTypesCustomImageExtensions, shared.ConfigCustomImageExtensionsDefault)
	SetDefault(shared.ConfigKeyFileTypesCustomBinaryExtensions, shared.ConfigCustomBinaryExtensionsDefault)
	SetDefault(shared.ConfigKeyFileTypesCustomLanguages, shared.ConfigCustomLanguagesDefault)
	SetDefault(shared.ConfigKeyFileTypesDisabledImageExtensions, shared.ConfigDisabledImageExtensionsDefault)
	SetDefault(shared.ConfigKeyFileTypesDisabledBinaryExtensions, shared.ConfigDisabledBinaryExtensionsDefault)
	SetDefault(shared.ConfigKeyFileTypesDisabledLanguageExts, shared.ConfigDisabledLanguageExtensionsDefault)

	// Resource limit defaults (size caps)
	SetDefault(shared.ConfigKeyResourceLimitsMaxFiles, shared.ConfigMaxFilesDefault)
	SetDefault(shared.ConfigKeyResourceLimitsMaxTotalSize, shared.ConfigMaxTotalSizeDefault)
}
