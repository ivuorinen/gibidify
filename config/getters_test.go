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
		{
			name:           "GetSupportedFormats",
			configKey:      "supportedFormats",
			configValue:    []string{"json", "yaml", "markdown"},
			getterFunc:     func() any { return config.SupportedFormats() },
			expectedResult: []string{"json", "yaml", "markdown"},
		},
		{
			name:           "GetFilePatterns",
			configKey:      "filePatterns",
			configValue:    []string{"*.go", "*.js", "*.py"},
			getterFunc:     func() any { return config.FilePatterns() },
			expectedResult: []string{"*.go", "*.js", "*.py"},
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

		// Backpressure configuration getters
		{
			name:           "GetBackpressureEnabled",
			configKey:      "backpressure.enabled",
			configValue:    true,
			getterFunc:     func() any { return config.BackpressureEnabled() },
			expectedResult: true,
		},
		{
			name:           "GetMaxPendingFiles",
			configKey:      "backpressure.maxPendingFiles",
			configValue:    1000,
			getterFunc:     func() any { return config.MaxPendingFiles() },
			expectedResult: 1000,
		},
		{
			name:           "GetMaxPendingWrites",
			configKey:      "backpressure.maxPendingWrites",
			configValue:    100,
			getterFunc:     func() any { return config.MaxPendingWrites() },
			expectedResult: 100,
		},
		{
			name:           "GetMaxMemoryUsage",
			configKey:      "backpressure.maxMemoryUsage",
			configValue:    int64(104857600),
			getterFunc:     func() any { return config.MaxMemoryUsage() },
			expectedResult: int64(104857600),
		},
		{
			name:           "GetMemoryCheckInterval",
			configKey:      "backpressure.memoryCheckInterval",
			configValue:    500,
			getterFunc:     func() any { return config.MemoryCheckInterval() },
			expectedResult: 500,
		},

		// Resource limits configuration getters
		{
			name:           "GetResourceLimitsEnabled",
			configKey:      "resourceLimits.enabled",
			configValue:    true,
			getterFunc:     func() any { return config.ResourceLimitsEnabled() },
			expectedResult: true,
		},
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
		{
			name:           "GetFileProcessingTimeoutSec",
			configKey:      "resourceLimits.fileProcessingTimeoutSec",
			configValue:    30,
			getterFunc:     func() any { return config.FileProcessingTimeoutSec() },
			expectedResult: 30,
		},
		{
			name:           "GetOverallTimeoutSec",
			configKey:      "resourceLimits.overallTimeoutSec",
			configValue:    1800,
			getterFunc:     func() any { return config.OverallTimeoutSec() },
			expectedResult: 1800,
		},
		{
			name:           "GetMaxConcurrentReads",
			configKey:      "resourceLimits.maxConcurrentReads",
			configValue:    10,
			getterFunc:     func() any { return config.MaxConcurrentReads() },
			expectedResult: 10,
		},
		{
			name:           "GetRateLimitFilesPerSec",
			configKey:      "resourceLimits.rateLimitFilesPerSec",
			configValue:    100,
			getterFunc:     func() any { return config.RateLimitFilesPerSec() },
			expectedResult: 100,
		},
		{
			name:           "GetHardMemoryLimitMB",
			configKey:      "resourceLimits.hardMemoryLimitMB",
			configValue:    512,
			getterFunc:     func() any { return config.HardMemoryLimitMB() },
			expectedResult: 512,
		},
		{
			name:           "GetEnableGracefulDegradation",
			configKey:      "resourceLimits.enableGracefulDegradation",
			configValue:    true,
			getterFunc:     func() any { return config.EnableGracefulDegradation() },
			expectedResult: true,
		},
		{
			name:           "GetEnableResourceMonitoring",
			configKey:      "resourceLimits.enableResourceMonitoring",
			configValue:    true,
			getterFunc:     func() any { return config.EnableResourceMonitoring() },
			expectedResult: true,
		},

		// Template system configuration getters
		{
			name:           "GetOutputTemplate",
			configKey:      "output.template",
			configValue:    "detailed",
			getterFunc:     func() any { return config.OutputTemplate() },
			expectedResult: "detailed",
		},
		{
			name:           "GetTemplateMetadataIncludeStats",
			configKey:      "output.metadata.includeStats",
			configValue:    true,
			getterFunc:     func() any { return config.TemplateMetadataIncludeStats() },
			expectedResult: true,
		},
		{
			name:           "GetTemplateMetadataIncludeTimestamp",
			configKey:      "output.metadata.includeTimestamp",
			configValue:    false,
			getterFunc:     func() any { return config.TemplateMetadataIncludeTimestamp() },
			expectedResult: false,
		},
		{
			name:           "GetTemplateMetadataIncludeFileCount",
			configKey:      "output.metadata.includeFileCount",
			configValue:    true,
			getterFunc:     func() any { return config.TemplateMetadataIncludeFileCount() },
			expectedResult: true,
		},
		{
			name:           "GetTemplateMetadataIncludeSourcePath",
			configKey:      "output.metadata.includeSourcePath",
			configValue:    false,
			getterFunc:     func() any { return config.TemplateMetadataIncludeSourcePath() },
			expectedResult: false,
		},
		{
			name:           "GetTemplateMetadataIncludeFileTypes",
			configKey:      "output.metadata.includeFileTypes",
			configValue:    true,
			getterFunc:     func() any { return config.TemplateMetadataIncludeFileTypes() },
			expectedResult: true,
		},
		{
			name:           "GetTemplateMetadataIncludeProcessingTime",
			configKey:      "output.metadata.includeProcessingTime",
			configValue:    false,
			getterFunc:     func() any { return config.TemplateMetadataIncludeProcessingTime() },
			expectedResult: false,
		},
		{
			name:           "GetTemplateMetadataIncludeTotalSize",
			configKey:      "output.metadata.includeTotalSize",
			configValue:    true,
			getterFunc:     func() any { return config.TemplateMetadataIncludeTotalSize() },
			expectedResult: true,
		},
		{
			name:           "GetTemplateMetadataIncludeMetrics",
			configKey:      "output.metadata.includeMetrics",
			configValue:    false,
			getterFunc:     func() any { return config.TemplateMetadataIncludeMetrics() },
			expectedResult: false,
		},

		// Markdown template configuration getters
		{
			name:           "GetTemplateMarkdownUseCodeBlocks",
			configKey:      "output.markdown.useCodeBlocks",
			configValue:    true,
			getterFunc:     func() any { return config.TemplateMarkdownUseCodeBlocks() },
			expectedResult: true,
		},
		{
			name:           "GetTemplateMarkdownIncludeLanguage",
			configKey:      "output.markdown.includeLanguage",
			configValue:    false,
			getterFunc:     func() any { return config.TemplateMarkdownIncludeLanguage() },
			expectedResult: false,
		},
		{
			name:           "GetTemplateMarkdownHeaderLevel",
			configKey:      "output.markdown.headerLevel",
			configValue:    3,
			getterFunc:     func() any { return config.TemplateMarkdownHeaderLevel() },
			expectedResult: 3,
		},
		{
			name:           "GetTemplateMarkdownTableOfContents",
			configKey:      "output.markdown.tableOfContents",
			configValue:    true,
			getterFunc:     func() any { return config.TemplateMarkdownTableOfContents() },
			expectedResult: true,
		},
		{
			name:           "GetTemplateMarkdownUseCollapsible",
			configKey:      "output.markdown.useCollapsible",
			configValue:    false,
			getterFunc:     func() any { return config.TemplateMarkdownUseCollapsible() },
			expectedResult: false,
		},
		{
			name:           "GetTemplateMarkdownSyntaxHighlighting",
			configKey:      "output.markdown.syntaxHighlighting",
			configValue:    true,
			getterFunc:     func() any { return config.TemplateMarkdownSyntaxHighlighting() },
			expectedResult: true,
		},
		{
			name:           "GetTemplateMarkdownLineNumbers",
			configKey:      "output.markdown.lineNumbers",
			configValue:    false,
			getterFunc:     func() any { return config.TemplateMarkdownLineNumbers() },
			expectedResult: false,
		},
		{
			name:           "GetTemplateMarkdownFoldLongFiles",
			configKey:      "output.markdown.foldLongFiles",
			configValue:    true,
			getterFunc:     func() any { return config.TemplateMarkdownFoldLongFiles() },
			expectedResult: true,
		},
		{
			name:           "GetTemplateMarkdownMaxLineLength",
			configKey:      "output.markdown.maxLineLength",
			configValue:    120,
			getterFunc:     func() any { return config.TemplateMarkdownMaxLineLength() },
			expectedResult: 120,
		},
		{
			name:           "GetTemplateCustomCSS",
			configKey:      "output.markdown.customCSS",
			configValue:    "body { color: blue; }",
			getterFunc:     func() any { return config.TemplateCustomCSS() },
			expectedResult: "body { color: blue; }",
		},

		// Custom template configuration getters
		{
			name:           "GetTemplateCustomHeader",
			configKey:      "output.custom.header",
			configValue:    "# Custom Header\n",
			getterFunc:     func() any { return config.TemplateCustomHeader() },
			expectedResult: "# Custom Header\n",
		},
		{
			name:           "GetTemplateCustomFooter",
			configKey:      "output.custom.footer",
			configValue:    "---\nFooter content",
			getterFunc:     func() any { return config.TemplateCustomFooter() },
			expectedResult: "---\nFooter content",
		},
		{
			name:           "GetTemplateCustomFileHeader",
			configKey:      "output.custom.fileHeader",
			configValue:    "## File: {{ .Path }}",
			getterFunc:     func() any { return config.TemplateCustomFileHeader() },
			expectedResult: "## File: {{ .Path }}",
		},
		{
			name:           "GetTemplateCustomFileFooter",
			configKey:      "output.custom.fileFooter",
			configValue:    "---",
			getterFunc:     func() any { return config.TemplateCustomFileFooter() },
			expectedResult: "---",
		},

		// Custom languages map getter
		{
			name:           "GetCustomLanguages",
			configKey:      "fileTypes.customLanguages",
			configValue:    map[string]string{".vue": "vue", ".svelte": "svelte"},
			getterFunc:     func() any { return config.CustomLanguages() },
			expectedResult: map[string]string{".vue": "vue", ".svelte": "svelte"},
		},

		// Template variables map getter
		{
			name:           "GetTemplateVariables",
			configKey:      "output.variables",
			configValue:    map[string]string{"project": "gibidify", "version": "1.0"},
			getterFunc:     func() any { return config.TemplateVariables() },
			expectedResult: map[string]string{"project": "gibidify", "version": "1.0"},
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
		assertIntGetter(t, "TemplateMarkdownHeaderLevel", config.TemplateMarkdownHeaderLevel,
			shared.ConfigMarkdownHeaderLevelDefault)
		assertIntGetter(t, "MaxFiles", config.MaxFiles, shared.ConfigMaxFilesDefault)
		assertInt64Getter(t, "MaxTotalSize", config.MaxTotalSize, shared.ConfigMaxTotalSizeDefault)
		assertIntGetter(t, "FileProcessingTimeoutSec", config.FileProcessingTimeoutSec,
			shared.ConfigFileProcessingTimeoutSecDefault)
		assertIntGetter(t, "OverallTimeoutSec", config.OverallTimeoutSec, shared.ConfigOverallTimeoutSecDefault)
		assertIntGetter(t, "MaxConcurrentReads", config.MaxConcurrentReads, shared.ConfigMaxConcurrentReadsDefault)
		assertIntGetter(t, "HardMemoryLimitMB", config.HardMemoryLimitMB, shared.ConfigHardMemoryLimitMBDefault)
	})

	// Test boolean getters with concrete default assertions
	t.Run("boolean_getters", func(t *testing.T) {
		assertBoolGetter(t, "FileTypesEnabled", config.FileTypesEnabled, shared.ConfigFileTypesEnabledDefault)
		assertBoolGetter(t, "BackpressureEnabled", config.BackpressureEnabled, shared.ConfigBackpressureEnabledDefault)
		assertBoolGetter(t, "ResourceLimitsEnabled", config.ResourceLimitsEnabled,
			shared.ConfigResourceLimitsEnabledDefault)
		assertBoolGetter(t, "EnableGracefulDegradation", config.EnableGracefulDegradation,
			shared.ConfigEnableGracefulDegradationDefault)
		assertBoolGetter(t, "TemplateMarkdownUseCodeBlocks", config.TemplateMarkdownUseCodeBlocks,
			shared.ConfigMarkdownUseCodeBlocksDefault)
		assertBoolGetter(t, "TemplateMarkdownTableOfContents", config.TemplateMarkdownTableOfContents,
			shared.ConfigMarkdownTableOfContentsDefault)
	})

	// Test string getters with concrete default assertions
	t.Run("string_getters", func(t *testing.T) {
		assertStringGetter(t, "OutputTemplate", config.OutputTemplate, shared.ConfigOutputTemplateDefault)
		assertStringGetter(t, "TemplateCustomCSS", config.TemplateCustomCSS, shared.ConfigMarkdownCustomCSSDefault)
		assertStringGetter(t, "TemplateCustomHeader", config.TemplateCustomHeader, shared.ConfigCustomHeaderDefault)
		assertStringGetter(t, "TemplateCustomFooter", config.TemplateCustomFooter, shared.ConfigCustomFooterDefault)
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

// assertStringGetter tests a string getter returns the expected default value.
func assertStringGetter(t *testing.T, name string, getter func() string, expected string) {
	t.Helper()
	result := getter()
	if result != expected {
		t.Errorf("%s: expected %q, got %q", name, expected, result)
	}
}
