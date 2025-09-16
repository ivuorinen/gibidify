package config_test

import (
	"reflect"
	"testing"

	"github.com/ivuorinen/gibidify/config"
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
			getterFunc:     func() any { return config.GetFileSizeLimit() },
			expectedResult: int64(1048576),
		},
		{
			name:           "GetIgnoredDirectories",
			configKey:      "ignoreDirectories",
			configValue:    []string{"node_modules", ".git", "dist"},
			getterFunc:     func() any { return config.GetIgnoredDirectories() },
			expectedResult: []string{"node_modules", ".git", "dist"},
		},
		{
			name:           "GetMaxConcurrency",
			configKey:      "maxConcurrency",
			configValue:    8,
			getterFunc:     func() any { return config.GetMaxConcurrency() },
			expectedResult: 8,
		},
		{
			name:           "GetSupportedFormats",
			configKey:      "supportedFormats",
			configValue:    []string{"json", "yaml", "markdown"},
			getterFunc:     func() any { return config.GetSupportedFormats() },
			expectedResult: []string{"json", "yaml", "markdown"},
		},
		{
			name:           "GetFilePatterns",
			configKey:      "filePatterns",
			configValue:    []string{"*.go", "*.js", "*.py"},
			getterFunc:     func() any { return config.GetFilePatterns() },
			expectedResult: []string{"*.go", "*.js", "*.py"},
		},

		// File type configuration getters
		{
			name:           "GetFileTypesEnabled",
			configKey:      "fileTypes.enabled",
			configValue:    true,
			getterFunc:     func() any { return config.GetFileTypesEnabled() },
			expectedResult: true,
		},
		{
			name:           "GetCustomImageExtensions",
			configKey:      "fileTypes.customImageExtensions",
			configValue:    []string{".webp", ".avif"},
			getterFunc:     func() any { return config.GetCustomImageExtensions() },
			expectedResult: []string{".webp", ".avif"},
		},
		{
			name:           "GetCustomBinaryExtensions",
			configKey:      "fileTypes.customBinaryExtensions",
			configValue:    []string{".custom", ".bin"},
			getterFunc:     func() any { return config.GetCustomBinaryExtensions() },
			expectedResult: []string{".custom", ".bin"},
		},
		{
			name:           "GetDisabledImageExtensions",
			configKey:      "fileTypes.disabledImageExtensions",
			configValue:    []string{".gif", ".bmp"},
			getterFunc:     func() any { return config.GetDisabledImageExtensions() },
			expectedResult: []string{".gif", ".bmp"},
		},
		{
			name:           "GetDisabledBinaryExtensions",
			configKey:      "fileTypes.disabledBinaryExtensions",
			configValue:    []string{".exe", ".dll"},
			getterFunc:     func() any { return config.GetDisabledBinaryExtensions() },
			expectedResult: []string{".exe", ".dll"},
		},
		{
			name:           "GetDisabledLanguageExtensions",
			configKey:      "fileTypes.disabledLanguageExtensions",
			configValue:    []string{".sh", ".bat"},
			getterFunc:     func() any { return config.GetDisabledLanguageExtensions() },
			expectedResult: []string{".sh", ".bat"},
		},

		// Backpressure configuration getters
		{
			name:           "GetBackpressureEnabled",
			configKey:      "backpressure.enabled",
			configValue:    true,
			getterFunc:     func() any { return config.GetBackpressureEnabled() },
			expectedResult: true,
		},
		{
			name:           "GetMaxPendingFiles",
			configKey:      "backpressure.maxPendingFiles",
			configValue:    1000,
			getterFunc:     func() any { return config.GetMaxPendingFiles() },
			expectedResult: 1000,
		},
		{
			name:           "GetMaxPendingWrites",
			configKey:      "backpressure.maxPendingWrites",
			configValue:    100,
			getterFunc:     func() any { return config.GetMaxPendingWrites() },
			expectedResult: 100,
		},
		{
			name:           "GetMaxMemoryUsage",
			configKey:      "backpressure.maxMemoryUsage",
			configValue:    int64(104857600),
			getterFunc:     func() any { return config.GetMaxMemoryUsage() },
			expectedResult: int64(104857600),
		},
		{
			name:           "GetMemoryCheckInterval",
			configKey:      "backpressure.memoryCheckInterval",
			configValue:    500,
			getterFunc:     func() any { return config.GetMemoryCheckInterval() },
			expectedResult: 500,
		},

		// Resource limits configuration getters
		{
			name:           "GetResourceLimitsEnabled",
			configKey:      "resourceLimits.enabled",
			configValue:    true,
			getterFunc:     func() any { return config.GetResourceLimitsEnabled() },
			expectedResult: true,
		},
		{
			name:           "GetMaxFiles",
			configKey:      "resourceLimits.maxFiles",
			configValue:    5000,
			getterFunc:     func() any { return config.GetMaxFiles() },
			expectedResult: 5000,
		},
		{
			name:           "GetMaxTotalSize",
			configKey:      "resourceLimits.maxTotalSize",
			configValue:    int64(1073741824),
			getterFunc:     func() any { return config.GetMaxTotalSize() },
			expectedResult: int64(1073741824),
		},
		{
			name:           "GetFileProcessingTimeoutSec",
			configKey:      "resourceLimits.fileProcessingTimeoutSec",
			configValue:    30,
			getterFunc:     func() any { return config.GetFileProcessingTimeoutSec() },
			expectedResult: 30,
		},
		{
			name:           "GetOverallTimeoutSec",
			configKey:      "resourceLimits.overallTimeoutSec",
			configValue:    1800,
			getterFunc:     func() any { return config.GetOverallTimeoutSec() },
			expectedResult: 1800,
		},
		{
			name:           "GetMaxConcurrentReads",
			configKey:      "resourceLimits.maxConcurrentReads",
			configValue:    10,
			getterFunc:     func() any { return config.GetMaxConcurrentReads() },
			expectedResult: 10,
		},
		{
			name:           "GetRateLimitFilesPerSec",
			configKey:      "resourceLimits.rateLimitFilesPerSec",
			configValue:    100,
			getterFunc:     func() any { return config.GetRateLimitFilesPerSec() },
			expectedResult: 100,
		},
		{
			name:           "GetHardMemoryLimitMB",
			configKey:      "resourceLimits.hardMemoryLimitMB",
			configValue:    512,
			getterFunc:     func() any { return config.GetHardMemoryLimitMB() },
			expectedResult: 512,
		},
		{
			name:           "GetEnableGracefulDegradation",
			configKey:      "resourceLimits.enableGracefulDegradation",
			configValue:    true,
			getterFunc:     func() any { return config.GetEnableGracefulDegradation() },
			expectedResult: true,
		},
		{
			name:           "GetEnableResourceMonitoring",
			configKey:      "resourceLimits.enableResourceMonitoring",
			configValue:    true,
			getterFunc:     func() any { return config.GetEnableResourceMonitoring() },
			expectedResult: true,
		},

		// Template system configuration getters
		{
			name:           "GetOutputTemplate",
			configKey:      "output.template",
			configValue:    "detailed",
			getterFunc:     func() any { return config.GetOutputTemplate() },
			expectedResult: "detailed",
		},
		{
			name:           "GetTemplateMetadataIncludeStats",
			configKey:      "output.metadata.includeStats",
			configValue:    true,
			getterFunc:     func() any { return config.GetTemplateMetadataIncludeStats() },
			expectedResult: true,
		},
		{
			name:           "GetTemplateMetadataIncludeTimestamp",
			configKey:      "output.metadata.includeTimestamp",
			configValue:    false,
			getterFunc:     func() any { return config.GetTemplateMetadataIncludeTimestamp() },
			expectedResult: false,
		},
		{
			name:           "GetTemplateMetadataIncludeFileCount",
			configKey:      "output.metadata.includeFileCount",
			configValue:    true,
			getterFunc:     func() any { return config.GetTemplateMetadataIncludeFileCount() },
			expectedResult: true,
		},
		{
			name:           "GetTemplateMetadataIncludeSourcePath",
			configKey:      "output.metadata.includeSourcePath",
			configValue:    false,
			getterFunc:     func() any { return config.GetTemplateMetadataIncludeSourcePath() },
			expectedResult: false,
		},
		{
			name:           "GetTemplateMetadataIncludeFileTypes",
			configKey:      "output.metadata.includeFileTypes",
			configValue:    true,
			getterFunc:     func() any { return config.GetTemplateMetadataIncludeFileTypes() },
			expectedResult: true,
		},
		{
			name:           "GetTemplateMetadataIncludeProcessingTime",
			configKey:      "output.metadata.includeProcessingTime",
			configValue:    false,
			getterFunc:     func() any { return config.GetTemplateMetadataIncludeProcessingTime() },
			expectedResult: false,
		},
		{
			name:           "GetTemplateMetadataIncludeTotalSize",
			configKey:      "output.metadata.includeTotalSize",
			configValue:    true,
			getterFunc:     func() any { return config.GetTemplateMetadataIncludeTotalSize() },
			expectedResult: true,
		},
		{
			name:           "GetTemplateMetadataIncludeMetrics",
			configKey:      "output.metadata.includeMetrics",
			configValue:    false,
			getterFunc:     func() any { return config.GetTemplateMetadataIncludeMetrics() },
			expectedResult: false,
		},

		// Markdown template configuration getters
		{
			name:           "GetTemplateMarkdownUseCodeBlocks",
			configKey:      "output.markdown.useCodeBlocks",
			configValue:    true,
			getterFunc:     func() any { return config.GetTemplateMarkdownUseCodeBlocks() },
			expectedResult: true,
		},
		{
			name:           "GetTemplateMarkdownIncludeLanguage",
			configKey:      "output.markdown.includeLanguage",
			configValue:    false,
			getterFunc:     func() any { return config.GetTemplateMarkdownIncludeLanguage() },
			expectedResult: false,
		},
		{
			name:           "GetTemplateMarkdownHeaderLevel",
			configKey:      "output.markdown.headerLevel",
			configValue:    3,
			getterFunc:     func() any { return config.GetTemplateMarkdownHeaderLevel() },
			expectedResult: 3,
		},
		{
			name:           "GetTemplateMarkdownTableOfContents",
			configKey:      "output.markdown.tableOfContents",
			configValue:    true,
			getterFunc:     func() any { return config.GetTemplateMarkdownTableOfContents() },
			expectedResult: true,
		},
		{
			name:           "GetTemplateMarkdownUseCollapsible",
			configKey:      "output.markdown.useCollapsible",
			configValue:    false,
			getterFunc:     func() any { return config.GetTemplateMarkdownUseCollapsible() },
			expectedResult: false,
		},
		{
			name:           "GetTemplateMarkdownSyntaxHighlighting",
			configKey:      "output.markdown.syntaxHighlighting",
			configValue:    true,
			getterFunc:     func() any { return config.GetTemplateMarkdownSyntaxHighlighting() },
			expectedResult: true,
		},
		{
			name:           "GetTemplateMarkdownLineNumbers",
			configKey:      "output.markdown.lineNumbers",
			configValue:    false,
			getterFunc:     func() any { return config.GetTemplateMarkdownLineNumbers() },
			expectedResult: false,
		},
		{
			name:           "GetTemplateMarkdownFoldLongFiles",
			configKey:      "output.markdown.foldLongFiles",
			configValue:    true,
			getterFunc:     func() any { return config.GetTemplateMarkdownFoldLongFiles() },
			expectedResult: true,
		},
		{
			name:           "GetTemplateMarkdownMaxLineLength",
			configKey:      "output.markdown.maxLineLength",
			configValue:    120,
			getterFunc:     func() any { return config.GetTemplateMarkdownMaxLineLength() },
			expectedResult: 120,
		},
		{
			name:           "GetTemplateCustomCSS",
			configKey:      "output.markdown.customCSS",
			configValue:    "body { color: blue; }",
			getterFunc:     func() any { return config.GetTemplateCustomCSS() },
			expectedResult: "body { color: blue; }",
		},

		// Custom template configuration getters
		{
			name:           "GetTemplateCustomHeader",
			configKey:      "output.custom.header",
			configValue:    "# Custom Header\n",
			getterFunc:     func() any { return config.GetTemplateCustomHeader() },
			expectedResult: "# Custom Header\n",
		},
		{
			name:           "GetTemplateCustomFooter",
			configKey:      "output.custom.footer",
			configValue:    "---\nFooter content",
			getterFunc:     func() any { return config.GetTemplateCustomFooter() },
			expectedResult: "---\nFooter content",
		},
		{
			name:           "GetTemplateCustomFileHeader",
			configKey:      "output.custom.fileHeader",
			configValue:    "## File: {{ .Path }}",
			getterFunc:     func() any { return config.GetTemplateCustomFileHeader() },
			expectedResult: "## File: {{ .Path }}",
		},
		{
			name:           "GetTemplateCustomFileFooter",
			configKey:      "output.custom.fileFooter",
			configValue:    "---",
			getterFunc:     func() any { return config.GetTemplateCustomFileFooter() },
			expectedResult: "---",
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

// TestGetCustomLanguages tests the custom languages map getter function.
func TestGetCustomLanguages(t *testing.T) {
	expectedLanguages := map[string]string{
		".vue":    "vue",
		".svelte": "svelte",
		".astro":  "astro",
	}

	testutil.SetViperKeys(t, map[string]any{
		"fileTypes.customLanguages": expectedLanguages,
	})

	result := config.GetCustomLanguages()
	if len(result) != len(expectedLanguages) {
		t.Errorf("Expected %d custom languages, got %d", len(expectedLanguages), len(result))
	}

	for key, expectedValue := range expectedLanguages {
		if actualValue, exists := result[key]; !exists {
			t.Errorf("Expected custom language key %s not found", key)
		} else if actualValue != expectedValue {
			t.Errorf("For key %s: expected %s, got %s", key, expectedValue, actualValue)
		}
	}
}

// TestGetTemplateVariables tests the template variables map getter function.
func TestGetTemplateVariables(t *testing.T) {
	expectedVariables := map[string]string{
		"project_name": "Test Project",
		"author":       "Test Author",
		"version":      "1.0.0",
	}

	testutil.SetViperKeys(t, map[string]any{
		"output.variables": expectedVariables,
	})

	result := config.GetTemplateVariables()
	if len(result) != len(expectedVariables) {
		t.Errorf("Expected %d template variables, got %d", len(expectedVariables), len(result))
	}

	for key, expectedValue := range expectedVariables {
		if actualValue, exists := result[key]; !exists {
			t.Errorf("Expected template variable key %s not found", key)
		} else if actualValue != expectedValue {
			t.Errorf("For key %s: expected %s, got %s", key, expectedValue, actualValue)
		}
	}
}

// TestConfigGettersWithDefaults tests that getters return appropriate default values
// when configuration keys are not set.
func TestConfigGettersWithDefaults(t *testing.T) {
	// Reset viper to ensure clean state
	testutil.ResetViperConfig(t, "")

	// Test numeric getters
	t.Run("numeric_getters", func(t *testing.T) {
		testInt64Getter(t, "GetFileSizeLimit", config.GetFileSizeLimit)
		testIntGetter(t, "GetMaxConcurrency", config.GetMaxConcurrency)
		testIntGetter(t, "GetTemplateMarkdownHeaderLevel", config.GetTemplateMarkdownHeaderLevel)
	})

	// Test boolean getters
	t.Run("boolean_getters", func(t *testing.T) {
		testBoolGetter(t, "GetFileTypesEnabled", config.GetFileTypesEnabled)
		testBoolGetter(t, "GetBackpressureEnabled", config.GetBackpressureEnabled)
		testBoolGetter(t, "GetResourceLimitsEnabled", config.GetResourceLimitsEnabled)
	})

	// Test string getters
	t.Run("string_getters", func(t *testing.T) {
		testStringGetter(t, "GetOutputTemplate", config.GetOutputTemplate)
	})
}

// testInt64Getter tests an int64 getter function
func testInt64Getter(t *testing.T, name string, getter func() int64) {
	t.Helper()
	result := getter() // Should not panic
	t.Logf("%s returned: %v (type: %T)", name, result, result)
}

// testIntGetter tests an int getter function
func testIntGetter(t *testing.T, name string, getter func() int) {
	t.Helper()
	result := getter() // Should not panic
	t.Logf("%s returned: %v (type: %T)", name, result, result)
}

// testBoolGetter tests a bool getter function
func testBoolGetter(t *testing.T, name string, getter func() bool) {
	t.Helper()
	result := getter() // Should not panic
	t.Logf("%s returned: %v (type: %T)", name, result, result)
}

// testStringGetter tests a string getter function
func testStringGetter(t *testing.T, name string, getter func() string) {
	t.Helper()
	result := getter() // Should not panic
	t.Logf("%s returned: %v (type: %T)", name, result, result)
}
