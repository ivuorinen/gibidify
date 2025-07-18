package config_test

import (
	"os"
	"strings"
	"testing"

	"github.com/spf13/viper"

	"github.com/ivuorinen/gibidify/config"
	"github.com/ivuorinen/gibidify/testutil"
	"github.com/ivuorinen/gibidify/utils"
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

// TestValidateConfig tests the configuration validation functionality.
func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name        string
		config      map[string]interface{}
		wantErr     bool
		errContains string
	}{
		{
			name: "valid default config",
			config: map[string]interface{}{
				"fileSizeLimit":     config.DefaultFileSizeLimit,
				"ignoreDirectories": []string{"node_modules", ".git"},
			},
			wantErr: false,
		},
		{
			name: "file size limit too small",
			config: map[string]interface{}{
				"fileSizeLimit": config.MinFileSizeLimit - 1,
			},
			wantErr:     true,
			errContains: "fileSizeLimit",
		},
		{
			name: "file size limit too large",
			config: map[string]interface{}{
				"fileSizeLimit": config.MaxFileSizeLimit + 1,
			},
			wantErr:     true,
			errContains: "fileSizeLimit",
		},
		{
			name: "empty ignore directory",
			config: map[string]interface{}{
				"ignoreDirectories": []string{"node_modules", "", ".git"},
			},
			wantErr:     true,
			errContains: "ignoreDirectories",
		},
		{
			name: "ignore directory with path separator",
			config: map[string]interface{}{
				"ignoreDirectories": []string{"node_modules", "src/build", ".git"},
			},
			wantErr:     true,
			errContains: "path separator",
		},
		{
			name: "invalid supported format",
			config: map[string]interface{}{
				"supportedFormats": []string{"json", "xml", "yaml"},
			},
			wantErr:     true,
			errContains: "not a valid format",
		},
		{
			name: "invalid max concurrency",
			config: map[string]interface{}{
				"maxConcurrency": 0,
			},
			wantErr:     true,
			errContains: "maxConcurrency",
		},
		{
			name: "valid comprehensive config",
			config: map[string]interface{}{
				"fileSizeLimit":     config.DefaultFileSizeLimit,
				"ignoreDirectories": []string{"node_modules", ".git", ".vscode"},
				"supportedFormats":  []string{"json", "yaml", "markdown"},
				"maxConcurrency":    8,
				"filePatterns":      []string{"*.go", "*.js", "*.py"},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset viper for each test
			viper.Reset()

			// Set test configuration
			for key, value := range tt.config {
				viper.Set(key, value)
			}

			// Load defaults for missing values
			config.LoadConfig()

			err := config.ValidateConfig()

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("Expected error to contain %q, got %q", tt.errContains, err.Error())
				}

				// Check that it's a structured error
				var structErr *utils.StructuredError
				if !errorAs(err, &structErr) {
					t.Errorf("Expected structured error, got %T", err)
					return
				}
				if structErr.Type != utils.ErrorTypeConfiguration {
					t.Errorf("Expected error type %v, got %v", utils.ErrorTypeConfiguration, structErr.Type)
				}
				if structErr.Code != utils.CodeConfigValidation {
					t.Errorf("Expected error code %v, got %v", utils.CodeConfigValidation, structErr.Code)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

// TestValidationFunctions tests individual validation functions.
func TestValidationFunctions(t *testing.T) {
	t.Run("IsValidFormat", func(t *testing.T) {
		tests := []struct {
			format string
			valid  bool
		}{
			{"json", true},
			{"yaml", true},
			{"markdown", true},
			{"JSON", true},
			{"xml", false},
			{"txt", false},
			{"", false},
			{"  json  ", true},
		}

		for _, tt := range tests {
			result := config.IsValidFormat(tt.format)
			if result != tt.valid {
				t.Errorf("IsValidFormat(%q) = %v, want %v", tt.format, result, tt.valid)
			}
		}
	})

	t.Run("ValidateFileSize", func(t *testing.T) {
		viper.Reset()
		viper.Set("fileSizeLimit", config.DefaultFileSizeLimit)

		tests := []struct {
			name    string
			size    int64
			wantErr bool
		}{
			{"size within limit", config.DefaultFileSizeLimit - 1, false},
			{"size at limit", config.DefaultFileSizeLimit, false},
			{"size exceeds limit", config.DefaultFileSizeLimit + 1, true},
			{"zero size", 0, false},
		}

		for _, tt := range tests {
			err := config.ValidateFileSize(tt.size)
			if (err != nil) != tt.wantErr {
				t.Errorf("%s: ValidateFileSize(%d) error = %v, wantErr %v", tt.name, tt.size, err, tt.wantErr)
			}
		}
	})

	t.Run("ValidateOutputFormat", func(t *testing.T) {
		tests := []struct {
			format  string
			wantErr bool
		}{
			{"json", false},
			{"yaml", false},
			{"markdown", false},
			{"xml", true},
			{"txt", true},
			{"", true},
		}

		for _, tt := range tests {
			err := config.ValidateOutputFormat(tt.format)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateOutputFormat(%q) error = %v, wantErr %v", tt.format, err, tt.wantErr)
			}
		}
	})

	t.Run("ValidateConcurrency", func(t *testing.T) {
		tests := []struct {
			name           string
			concurrency    int
			maxConcurrency int
			setMax         bool
			wantErr        bool
		}{
			{"valid concurrency", 4, 0, false, false},
			{"minimum concurrency", 1, 0, false, false},
			{"zero concurrency", 0, 0, false, true},
			{"negative concurrency", -1, 0, false, true},
			{"concurrency within max", 4, 8, true, false},
			{"concurrency exceeds max", 16, 8, true, true},
		}

		for _, tt := range tests {
			viper.Reset()
			if tt.setMax {
				viper.Set("maxConcurrency", tt.maxConcurrency)
			}

			err := config.ValidateConcurrency(tt.concurrency)
			if (err != nil) != tt.wantErr {
				t.Errorf("%s: ValidateConcurrency(%d) error = %v, wantErr %v", tt.name, tt.concurrency, err, tt.wantErr)
			}
		}
	})
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

func errorAs(err error, target interface{}) bool {
	if err == nil {
		return false
	}
	if structErr, ok := err.(*utils.StructuredError); ok {
		if ptr, ok := target.(**utils.StructuredError); ok {
			*ptr = structErr
			return true
		}
	}
	return false
}
