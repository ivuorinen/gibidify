// Package config handles application configuration using Viper.
package config

import (
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
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
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		logrus.Infof("Config file not found, using default values: %v", err)
		setDefaultConfig()
	} else {
		logrus.Infof("Using config file: %s", viper.ConfigFileUsed())
	}
}

// setDefaultConfig sets default configuration values.
func setDefaultConfig() {
	viper.SetDefault("fileSizeLimit", 5242880) // 5 MB
	// Default ignored directories.
	viper.SetDefault("ignoreDirectories", []string{
		"vendor", "node_modules", ".git", "dist", "build", "target", "bower_components", "cache", "tmp",
	})
}

// GetFileSizeLimit returns the file size limit from configuration.
func GetFileSizeLimit() int64 {
	return viper.GetInt64("fileSizeLimit")
}

// GetIgnoredDirectories returns the list of directories to ignore.
func GetIgnoredDirectories() []string {
	return viper.GetStringSlice("ignoreDirectories")
}
