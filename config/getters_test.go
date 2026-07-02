package config_test

import (
	"reflect"
	"testing"

	"github.com/ivuorinen/gibidify/config"
	"github.com/ivuorinen/gibidify/shared"
	"github.com/ivuorinen/gibidify/testutil"
)

// TestConfigGetters tests all configuration getter functions with comprehensive test coverage.
func TestConfigGetters(t *testing.T) {
	tests := []struct {
		name           string
		configKey      string
		configValue    any
		getterFunc     func() any
		expectedResult any
	}{
		// Basic configuration getters
		{
			name:           "GetFileSizeLimit",
			configKey:      "fileSizeLimit",
			configValue:    int64(1048576),
			getterFunc:     func() any { return config.FileSizeLimit() },
			expectedResult: int64(1048576),
		},
		{
			name:           "GetIgnoredDirectories",
			configKey:      "ignoreDirectories",
			configValue:    []string{"node_modules", ".git", "dist"},
			getterFunc:     func() any { return config.IgnoredDirectories() },
			expectedResult: []string{"node_modules", ".git", "dist"},
		},
		{
			name:           "GetMaxConcurrency",
			configKey:      "maxConcurrency",
			configValue:    8,
			getterFunc:     func() any { return config.MaxConcurrency() },
			expectedResult: 8,
		},
		// File type configuration getters
		{
			name:           "GetFileTypesEnabled",
			configKey:      "fileTypes.enabled",
			configValue:    true,
			getterFunc:     func() any { return config.FileTypesEnabled() },
			expectedResult: true,
		},
		{
			name:           "GetCustomImageExtensions",
			configKey:      "fileTypes.customImageExtensions",
			configValue:    []string{".webp", ".avif"},
			getterFunc:     func() any { return config.CustomImageExtensions() },
			expectedResult: []string{".webp", ".avif"},
		},
		{
			name:           "GetCustomBinaryExtensions",
			configKey:      "fileTypes.customBinaryExtensions",
			configValue:    []string{".custom", ".bin"},
			getterFunc:     func() any { return config.CustomBinaryExtensions() },
			expectedResult: []string{".custom", ".bin"},
		},
		{
			name:           "GetDisabledImageExtensions",
			configKey:      "fileTypes.disabledImageExtensions",
			configValue:    []string{".gif", ".bmp"},
			getterFunc:     func() any { return config.DisabledImageExtensions() },
			expectedResult: []string{".gif", ".bmp"},
		},
		{
			name:           "GetDisabledBinaryExtensions",
			configKey:      "fileTypes.disabledBinaryExtensions",
			configValue:    []string{".exe", ".dll"},
			getterFunc:     func() any { return config.DisabledBinaryExtensions() },
			expectedResult: []string{".exe", ".dll"},
		},
		{
			name:           "GetDisabledLanguageExtensions",
			configKey:      "fileTypes.disabledLanguageExtensions",
			configValue:    []string{".sh", ".bat"},
			getterFunc:     func() any { return config.DisabledLanguageExtensions() },
			expectedResult: []string{".sh", ".bat"},
		},

		// Resource limits configuration getters (size caps)
		{
			name:           "GetMaxFiles",
			configKey:      "resourceLimits.maxFiles",
			configValue:    5000,
			getterFunc:     func() any { return config.MaxFiles() },
			expectedResult: 5000,
		},
		{
			name:           "GetMaxTotalSize",
			configKey:      "resourceLimits.maxTotalSize",
			configValue:    int64(1073741824),
			getterFunc:     func() any { return config.MaxTotalSize() },
			expectedResult: int64(1073741824),
		},

		// Custom languages map getter
		{
			name:           "GetCustomLanguages",
			configKey:      "fileTypes.customLanguages",
			configValue:    map[string]string{".vue": "vue", ".svelte": "svelte"},
			getterFunc:     func() any { return config.CustomLanguages() },
			expectedResult: map[string]string{".vue": "vue", ".svelte": "svelte"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset viper and set the specific configuration
			testutil.SetViperKeys(t, map[string]any{
				tt.configKey: tt.configValue,
			})

			// Call the getter function and compare results
			result := tt.getterFunc()
			if !reflect.DeepEqual(result, tt.expectedResult) {
				t.Errorf("Test %s: expected %v (type %T), got %v (type %T)",
					tt.name, tt.expectedResult, tt.expectedResult, result, result)
			}
		})
	}
}

// TestConfigGettersWithDefaults tests that getters return appropriate default values
// when configuration keys are not set.
func TestConfigGettersWithDefaults(t *testing.T) {
	// Reset viper to ensure clean state
	testutil.ResetViperConfig(t, "")

	// Test numeric getters with concrete default assertions
	t.Run("numeric_getters", func(t *testing.T) {
		assertInt64Getter(t, "FileSizeLimit", config.FileSizeLimit, shared.ConfigFileSizeLimitDefault)
		assertIntGetter(t, "MaxConcurrency", config.MaxConcurrency, shared.ConfigMaxConcurrencyDefault)
		assertIntGetter(t, "MaxFiles", config.MaxFiles, shared.ConfigMaxFilesDefault)
		assertInt64Getter(t, "MaxTotalSize", config.MaxTotalSize, shared.ConfigMaxTotalSizeDefault)
	})

	// Test boolean getters with concrete default assertions
	t.Run("boolean_getters", func(t *testing.T) {
		assertBoolGetter(t, "FileTypesEnabled", config.FileTypesEnabled, shared.ConfigFileTypesEnabledDefault)
	})
}

// assertInt64Getter tests an int64 getter returns the expected default value.
func assertInt64Getter(t *testing.T, name string, getter func() int64, expected int64) {
	t.Helper()
	result := getter()
	if result != expected {
		t.Errorf("%s: expected %d, got %d", name, expected, result)
	}
}

// assertIntGetter tests an int getter returns the expected default value.
func assertIntGetter(t *testing.T, name string, getter func() int, expected int) {
	t.Helper()
	result := getter()
	if result != expected {
		t.Errorf("%s: expected %d, got %d", name, expected, result)
	}
}

// assertBoolGetter tests a bool getter returns the expected default value.
func assertBoolGetter(t *testing.T, name string, getter func() bool, expected bool) {
	t.Helper()
	result := getter()
	if result != expected {
		t.Errorf("%s: expected %v, got %v", name, expected, result)
	}
}
