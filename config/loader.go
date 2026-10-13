// Package config handles application configuration management.
package config

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/ivuorinen/gibidify/shared"
)

// LoadConfig reads configuration from a YAML file.
// It looks for config.yaml in, in order:
// 1. $XDG_CONFIG_HOME/gibidify/ (only when XDG_CONFIG_HOME is absolute, per the XDG spec)
// 2. $HOME/.config/gibidify/
//
// The current directory is not searched: an unrelated project's config.yaml
// there was silently parsed as gibidify configuration.
//
// A missing file is logged at Info and the defaults apply. A file that exists
// but cannot be read or parsed, or that fails validation, is returned as an
// error: falling back to defaults silently dropped the user's ignore list, so
// files they excluded (credentials included) reached the output. Unknown keys
// are only warned about, since ignoring them changes nothing the user set.
func LoadConfig() error {
	logger := shared.GetLogger()

	// Populate defaults first so keys absent from a loaded config file still fall
	// back to their defaults (the store's lookup checks values, then defaults).
	SetDefaultConfig()

	xdgConfig := os.Getenv("XDG_CONFIG_HOME")
	switch {
	case filepath.IsAbs(xdgConfig):
		AddConfigPath(filepath.Join(xdgConfig, shared.AppName))
	case xdgConfig != "":
		logger.Warnf("Ignoring relative XDG_CONFIG_HOME %q; the XDG spec requires an absolute path", xdgConfig)

		fallthrough
	default:
		if home, err := os.UserHomeDir(); err == nil {
			AddConfigPath(filepath.Join(home, ".config", shared.AppName))
		}
	}

	if err := ReadInConfig(); err != nil {
		if errors.Is(err, errNoConfigFile) {
			logger.Infof("No config file, using default values: %v", err)

			return nil
		}

		return shared.WrapError(
			err, shared.ErrorTypeConfiguration, shared.CodeConfigValidation, "cannot load config file",
		)
	}

	logger.Infof("Using config file: %s", FileUsed())
	for _, key := range unknownKeys() {
		logger.Warnf("Unknown configuration key %s in %s is ignored", key, FileUsed())
	}

	if err := ValidateConfig(); err != nil {
		var structErr *shared.StructuredError
		if errors.As(err, &structErr) {
			return structErr.WithFilePath(FileUsed())
		}

		return err
	}

	return nil
}

// SetDefaultConfig sets default configuration values.
// maxConcurrency has no default: it is an optional ceiling on -concurrency and
// applies only when the user sets it.
func SetDefaultConfig() {
	// File size limits
	SetDefault(shared.ConfigKeyFileSizeLimit, shared.ConfigFileSizeLimitDefault)
	SetDefault(shared.ConfigKeyIgnoreDirectories, shared.ConfigIgnoredDirectoriesDefault)

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
