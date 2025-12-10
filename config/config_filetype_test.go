package config

import (
	"testing"

	"github.com/spf13/viper"

	"github.com/ivuorinen/gibidify/shared"
)

// TestFileTypeRegistryDefaultValues tests default configuration values.
func TestFileTypeRegistryDefaultValues(t *testing.T) {
	viper.Reset()
	SetDefaultConfig()

	verifyDefaultValues(t)
}

// TestFileTypeRegistrySetGet tests configuration setting and getting.
func TestFileTypeRegistrySetGet(t *testing.T) {
	viper.Reset()

	// Set test values
	setTestConfiguration()

	// Test getter functions
	verifyTestConfiguration(t)
}

// TestFileTypeRegistryValidationSuccess tests successful validation.
func TestFileTypeRegistryValidationSuccess(t *testing.T) {
	viper.Reset()
	SetDefaultConfig()

	// Set valid configuration
	setValidConfiguration()

	err := ValidateConfig()
	if err != nil {
		t.Errorf("Expected validation to pass with valid config, got error: %v", err)
	}
}

// TestFileTypeRegistryValidationFailure tests validation failures.
func TestFileTypeRegistryValidationFailure(t *testing.T) {
	// Test invalid custom image extensions
	testInvalidImageExtensions(t)

	// Test invalid custom binary extensions
	testInvalidBinaryExtensions(t)

	// Test invalid custom languages
	testInvalidCustomLanguages(t)
}

// verifyDefaultValues verifies that default values are correct.
func verifyDefaultValues(t *testing.T) {
	t.Helper()

	if !FileTypesEnabled() {
		t.Error("Expected file types to be enabled by default")
	}

	verifyEmptySlice(t, CustomImageExtensions(), "custom image extensions")
	verifyEmptySlice(t, CustomBinaryExtensions(), "custom binary extensions")
	verifyEmptyMap(t, CustomLanguages(), "custom languages")
	verifyEmptySlice(t, DisabledImageExtensions(), "disabled image extensions")
	verifyEmptySlice(t, DisabledBinaryExtensions(), "disabled binary extensions")
	verifyEmptySlice(t, DisabledLanguageExtensions(), "disabled language extensions")
}

// setTestConfiguration sets test configuration values.
func setTestConfiguration() {
	viper.Set("fileTypes.enabled", false)
	viper.Set(shared.ConfigKeyFileTypesCustomImageExtensions, []string{".webp", ".avif"})
	viper.Set(shared.ConfigKeyFileTypesCustomBinaryExtensions, []string{shared.TestExtensionCustom, ".mybin"})
	viper.Set(
		shared.ConfigKeyFileTypesCustomLanguages, map[string]string{
			".zig": "zig",
			".v":   "vlang",
		},
	)
	viper.Set("fileTypes.disabledImageExtensions", []string{".gif", ".bmp"})
	viper.Set("fileTypes.disabledBinaryExtensions", []string{".exe", ".dll"})
	viper.Set("fileTypes.disabledLanguageExtensions", []string{".rb", ".pl"})
}

// verifyTestConfiguration verifies that test configuration is retrieved correctly.
func verifyTestConfiguration(t *testing.T) {
	t.Helper()

	if FileTypesEnabled() {
		t.Error("Expected file types to be disabled")
	}

	verifyStringSlice(t, CustomImageExtensions(), []string{".webp", ".avif"}, "custom image extensions")
	verifyStringSlice(t, CustomBinaryExtensions(), []string{".custom", ".mybin"}, "custom binary extensions")

	expectedLangs := map[string]string{
		".zig": "zig",
		".v":   "vlang",
	}
	verifyStringMap(t, CustomLanguages(), expectedLangs, "custom languages")

	verifyStringSliceLength(t, DisabledImageExtensions(), []string{".gif", ".bmp"}, "disabled image extensions")
	verifyStringSliceLength(t, DisabledBinaryExtensions(), []string{".exe", ".dll"}, "disabled binary extensions")
	verifyStringSliceLength(t, DisabledLanguageExtensions(), []string{".rb", ".pl"}, "disabled language extensions")
}

// setValidConfiguration sets valid configuration for validation tests.
func setValidConfiguration() {
	viper.Set(shared.ConfigKeyFileTypesCustomImageExtensions, []string{".webp", ".avif"})
	viper.Set(shared.ConfigKeyFileTypesCustomBinaryExtensions, []string{shared.TestExtensionCustom})
	viper.Set(
		shared.ConfigKeyFileTypesCustomLanguages, map[string]string{
			".zig": "zig",
			".v":   "vlang",
		},
	)
}

// testInvalidImageExtensions tests validation failure with invalid image extensions.
func testInvalidImageExtensions(t *testing.T) {
	t.Helper()

	viper.Reset()
	SetDefaultConfig()
	viper.Set(shared.ConfigKeyFileTypesCustomImageExtensions, []string{"", "webp"}) // Empty and missing dot

	err := ValidateConfig()
	if err == nil {
		t.Error("Expected validation to fail with invalid custom image extensions")
	}
}

// testInvalidBinaryExtensions tests validation failure with invalid binary extensions.
func testInvalidBinaryExtensions(t *testing.T) {
	t.Helper()

	viper.Reset()
	SetDefaultConfig()
	viper.Set(shared.ConfigKeyFileTypesCustomBinaryExtensions, []string{"custom"}) // Missing dot

	err := ValidateConfig()
	if err == nil {
		t.Error("Expected validation to fail with invalid custom binary extensions")
	}
}

// testInvalidCustomLanguages tests validation failure with invalid custom languages.
func testInvalidCustomLanguages(t *testing.T) {
	t.Helper()

	viper.Reset()
	SetDefaultConfig()
	viper.Set(
		shared.ConfigKeyFileTypesCustomLanguages, map[string]string{
			"zig": "zig", // Missing dot in extension
			".v":  "",    // Empty language
		},
	)

	err := ValidateConfig()
	if err == nil {
		t.Error("Expected validation to fail with invalid custom languages")
	}
}

// verifyEmptySlice verifies that a slice is empty.
func verifyEmptySlice(t *testing.T, slice []string, name string) {
	t.Helper()

	if len(slice) != 0 {
		t.Errorf("Expected %s to be empty by default", name)
	}
}

// verifyEmptyMap verifies that a map is empty.
func verifyEmptyMap(t *testing.T, m map[string]string, name string) {
	t.Helper()

	if len(m) != 0 {
		t.Errorf("Expected %s to be empty by default", name)
	}
}

// verifyStringSlice verifies that a string slice matches expected values.
func verifyStringSlice(t *testing.T, actual, expected []string, name string) {
	t.Helper()

	if len(actual) != len(expected) {
		t.Errorf(shared.TestFmtExpectedCount, len(expected), name, len(actual))

		return
	}
	for i, ext := range expected {
		if actual[i] != ext {
			t.Errorf("Expected %s %s, got %s", name, ext, actual[i])
		}
	}
}

// verifyStringMap verifies that a string map matches expected values.
func verifyStringMap(t *testing.T, actual, expected map[string]string, name string) {
	t.Helper()

	if len(actual) != len(expected) {
		t.Errorf(shared.TestFmtExpectedCount, len(expected), name, len(actual))

		return
	}
	for ext, lang := range expected {
		if actual[ext] != lang {
			t.Errorf("Expected %s %s -> %s, got %s", name, ext, lang, actual[ext])
		}
	}
}

// verifyStringSliceLength verifies that a string slice has the expected length.
func verifyStringSliceLength(t *testing.T, actual, expected []string, name string) {
	t.Helper()

	if len(actual) != len(expected) {
		t.Errorf(shared.TestFmtExpectedCount, len(expected), name, len(actual))
	}
}
