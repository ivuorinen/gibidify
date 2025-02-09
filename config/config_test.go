package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
)

// TestDefaultConfig verifies that if no config file is found,
// the default configuration values are correctly set.
func TestDefaultConfig(t *testing.T) {
	// Create a temporary directory to ensure no config file is present.
	tmpDir, err := os.MkdirTemp("", "gibidify_config_test_default")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Point Viper to the temp directory with no config file.
	originalConfigPaths := viper.ConfigFileUsed()
	viper.Reset()
	viper.AddConfigPath(tmpDir)
	LoadConfig()

	// Check defaults
	defaultSizeLimit := GetFileSizeLimit()
	if defaultSizeLimit != 5242880 {
		t.Errorf("Expected default file size limit of 5242880, got %d", defaultSizeLimit)
	}

	ignoredDirs := GetIgnoredDirectories()
	if len(ignoredDirs) == 0 {
		t.Errorf("Expected some default ignored directories, got none")
	}

	// Restore Viper state
	viper.SetConfigFile(originalConfigPaths)
}

// TestLoadConfigFile verifies that when a valid config file is present,
// viper loads the specified values correctly.
func TestLoadConfigFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "gibidify_config_test_file")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Prepare a minimal config file
	configContent := []byte(`
fileSizeLimit: 123456
ignoreDirectories:
  - "testdir1"
  - "testdir2"
`)

	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, configContent, 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Reset viper and point to the new config path
	viper.Reset()
	viper.AddConfigPath(tmpDir)

	// Force Viper to read our config file
	if err := viper.ReadInConfig(); err != nil {
		t.Fatalf("Could not read config file: %v", err)
	}

	// Validate loaded data
	if got := viper.GetInt64("fileSizeLimit"); got != 123456 {
		t.Errorf("Expected fileSizeLimit=123456, got %d", got)
	}

	ignored := viper.GetStringSlice("ignoreDirectories")
	if len(ignored) != 2 || ignored[0] != "testdir1" || ignored[1] != "testdir2" {
		t.Errorf("Expected [\"testdir1\", \"testdir2\"], got %v", ignored)
	}
}
