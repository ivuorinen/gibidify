package config_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/spf13/viper"

	"github.com/ivuorinen/gibidify/config"
	"github.com/ivuorinen/gibidify/shared"
)

// TestValidateConfig tests the configuration validation functionality.
func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name        string
		config      map[string]any
		wantErr     bool
		errContains string
	}{
		{
			name: "valid default config",
			config: map[string]any{
				"fileSizeLimit":     shared.ConfigFileSizeLimitDefault,
				"ignoreDirectories": []string{"node_modules", ".git"},
			},
			wantErr: false,
		},
		{
			name: "file size limit too small",
			config: map[string]any{
				"fileSizeLimit": shared.ConfigFileSizeLimitMin - 1,
			},
			wantErr:     true,
			errContains: "fileSizeLimit",
		},
		{
			name: "file size limit too large",
			config: map[string]any{
				"fileSizeLimit": shared.ConfigFileSizeLimitMax + 1,
			},
			wantErr:     true,
			errContains: "fileSizeLimit",
		},
		{
			name: "empty ignore directory",
			config: map[string]any{
				"ignoreDirectories": []string{"node_modules", "", ".git"},
			},
			wantErr:     true,
			errContains: "ignoreDirectories",
		},
		{
			name: "ignore directory with path separator",
			config: map[string]any{
				"ignoreDirectories": []string{"node_modules", "src/build", ".git"},
			},
			wantErr:     true,
			errContains: "path separator",
		},
		{
			name: "invalid supported format",
			config: map[string]any{
				"supportedFormats": []string{"json", "xml", "yaml"},
			},
			wantErr:     true,
			errContains: "not a valid format",
		},
		{
			name: "invalid max concurrency",
			config: map[string]any{
				"maxConcurrency": 0,
			},
			wantErr:     true,
			errContains: "maxConcurrency",
		},
		{
			name: "valid comprehensive config",
			config: map[string]any{
				"fileSizeLimit":     shared.ConfigFileSizeLimitDefault,
				"ignoreDirectories": []string{"node_modules", ".git", ".vscode"},
				"supportedFormats":  []string{"json", "yaml", "markdown"},
				"maxConcurrency":    8,
				"filePatterns":      []string{"*.go", "*.js", "*.py"},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				// Reset viper for each test
				viper.Reset()

				// Set test configuration
				for key, value := range tt.config {
					viper.Set(key, value)
				}

				// Set defaults for missing values without touching disk
				config.SetDefaultConfig()

				err := config.ValidateConfig()

				if tt.wantErr {
					validateExpectedError(t, err, tt.errContains)
				} else if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			},
		)
	}
}

// TestIsValidFormat tests the IsValidFormat function.
func TestIsValidFormat(t *testing.T) {
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
}

// TestValidateFileSize tests the ValidateFileSize function.
func TestValidateFileSize(t *testing.T) {
	viper.Reset()
	viper.Set("fileSizeLimit", shared.ConfigFileSizeLimitDefault)

	tests := []struct {
		name    string
		size    int64
		wantErr bool
	}{
		{"size within limit", shared.ConfigFileSizeLimitDefault - 1, false},
		{"size at limit", shared.ConfigFileSizeLimitDefault, false},
		{"size exceeds limit", shared.ConfigFileSizeLimitDefault + 1, true},
		{"zero size", 0, false},
	}

	for _, tt := range tests {
		err := config.ValidateFileSize(tt.size)
		if (err != nil) != tt.wantErr {
			t.Errorf("%s: ValidateFileSize(%d) error = %v, wantErr %v", tt.name, tt.size, err, tt.wantErr)
		}
	}
}

// TestValidateOutputFormat tests the ValidateOutputFormat function.
func TestValidateOutputFormat(t *testing.T) {
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
}

// TestValidateConcurrency tests the ValidateConcurrency function.
func TestValidateConcurrency(t *testing.T) {
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
}

// validateExpectedError validates that an error occurred and matches expectations.
func validateExpectedError(t *testing.T, err error, errContains string) {
	t.Helper()
	if err == nil {
		t.Error(shared.TestMsgExpectedError)

		return
	}
	if errContains != "" && !strings.Contains(err.Error(), errContains) {
		t.Errorf("Expected error to contain %q, got %q", errContains, err.Error())
	}

	// Check that it's a structured error
	var structErr *shared.StructuredError
	if !errorAs(err, &structErr) {
		t.Errorf("Expected structured error, got %T", err)

		return
	}
	if structErr.Type != shared.ErrorTypeConfiguration {
		t.Errorf("Expected error type %v, got %v", shared.ErrorTypeConfiguration, structErr.Type)
	}
	if structErr.Code != shared.CodeConfigValidation {
		t.Errorf("Expected error code %v, got %v", shared.CodeConfigValidation, structErr.Code)
	}
}

func errorAs(err error, target any) bool {
	if err == nil {
		return false
	}
	structErr := &shared.StructuredError{}
	if errors.As(err, &structErr) {
		if ptr, ok := target.(**shared.StructuredError); ok {
			*ptr = structErr

			return true
		}
	}

	return false
}
