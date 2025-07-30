package config_test

import (
	"os"
	"testing"

	"github.com/spf13/viper"

	"github.com/ivuorinen/gibidify/config"
	"github.com/ivuorinen/gibidify/testutil"
)

const (
	defaultFileSizeLimit = 5242880
	testFileSizeLimit    = 123456
)

// TestDefaultConfig verifies that if no config file is found,
// the default configuration values are correctly set.
func TestDefaultConfig(t *testing.T) {
	// Create a temporary directory to ensure no config file is present.
	tmpDir := t.TempDir()

	// Point Viper to the temp directory with no config file.
	originalConfigPaths := viper.ConfigFileUsed()
	testutil.ResetViperConfig(t, tmpDir)

	// Check defaults
	defaultSizeLimit := config.GetFileSizeLimit()
	if defaultSizeLimit != defaultFileSizeLimit {
		t.Errorf("Expected default file size limit of 5242880, got %d", defaultSizeLimit)
	}

	ignoredDirs := config.GetIgnoredDirectories()
	if len(ignoredDirs) == 0 {
		t.Errorf("Expected some default ignored directories, got none")
	}

	// Restore Viper state
	viper.SetConfigFile(originalConfigPaths)
}

// TestLoadConfigFile verifies that when a valid config file is present,
// viper loads the specified values correctly.
func TestLoadConfigFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Prepare a minimal config file
	configContent := []byte(`---
fileSizeLimit: 123456
ignoreDirectories:
- "testdir1"
- "testdir2"
`)

	testutil.CreateTestFile(t, tmpDir, "config.yaml", configContent)

	// Reset viper and point to the new config path
	viper.Reset()
	viper.AddConfigPath(tmpDir)

	// Force Viper to read our config file
	testutil.MustSucceed(t, viper.ReadInConfig(), "reading config file")

	// Validate loaded data
	if got := viper.GetInt64("fileSizeLimit"); got != testFileSizeLimit {
		t.Errorf("Expected fileSizeLimit=123456, got %d", got)
	}

	ignored := viper.GetStringSlice("ignoreDirectories")
	if len(ignored) != 2 || ignored[0] != "testdir1" || ignored[1] != "testdir2" {
		t.Errorf("Expected [\"testdir1\", \"testdir2\"], got %v", ignored)
	}
}

// TestLoadConfigWithValidation tests that invalid config files fall back to defaults.
func TestLoadConfigWithValidation(t *testing.T) {
	// Create a temporary config file with invalid content
	configContent := `
fileSizeLimit: 100
ignoreDirectories:
  - node_modules
  - ""
  - .git
`

	tempDir := t.TempDir()
	configFile := tempDir + "/config.yaml"

	err := os.WriteFile(configFile, []byte(configContent), 0o644)
	if err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Reset viper and set config path
	viper.Reset()
	viper.AddConfigPath(tempDir)

	// This should load the config but validation should fail and fall back to defaults
	config.LoadConfig()

	// Should have fallen back to defaults due to validation failure
	if config.GetFileSizeLimit() != int64(config.DefaultFileSizeLimit) {
		t.Errorf("Expected default file size limit after validation failure, got %d", config.GetFileSizeLimit())
	}
	if containsString(config.GetIgnoredDirectories(), "") {
		t.Errorf("Expected ignored directories not to contain empty string after validation failure, got %v", config.GetIgnoredDirectories())
	}
}

// Helper functions

func containsString(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}