package config_test

import (
	"strings"
	"testing"

	"github.com/spf13/viper"

	"github.com/ivuorinen/gibidify/config"
	"github.com/ivuorinen/gibidify/utils"
)

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