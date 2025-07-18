package config

import (
	"testing"

	"github.com/spf13/viper"
)

// TestFileTypeRegistryConfig tests the FileTypeRegistry configuration functionality.
func TestFileTypeRegistryConfig(t *testing.T) {
	// Test default values
	t.Run("DefaultValues", func(t *testing.T) {
		viper.Reset()
		setDefaultConfig()

		if !GetFileTypesEnabled() {
			t.Error("Expected file types to be enabled by default")
		}

		if len(GetCustomImageExtensions()) != 0 {
			t.Error("Expected custom image extensions to be empty by default")
		}

		if len(GetCustomBinaryExtensions()) != 0 {
			t.Error("Expected custom binary extensions to be empty by default")
		}

		if len(GetCustomLanguages()) != 0 {
			t.Error("Expected custom languages to be empty by default")
		}

		if len(GetDisabledImageExtensions()) != 0 {
			t.Error("Expected disabled image extensions to be empty by default")
		}

		if len(GetDisabledBinaryExtensions()) != 0 {
			t.Error("Expected disabled binary extensions to be empty by default")
		}

		if len(GetDisabledLanguageExtensions()) != 0 {
			t.Error("Expected disabled language extensions to be empty by default")
		}
	})

	// Test configuration setting and getting
	t.Run("ConfigurationSetGet", func(t *testing.T) {
		viper.Reset()

		// Set test values
		viper.Set("fileTypes.enabled", false)
		viper.Set("fileTypes.customImageExtensions", []string{".webp", ".avif"})
		viper.Set("fileTypes.customBinaryExtensions", []string{".custom", ".mybin"})
		viper.Set("fileTypes.customLanguages", map[string]string{
			".zig": "zig",
			".v":   "vlang",
		})
		viper.Set("fileTypes.disabledImageExtensions", []string{".gif", ".bmp"})
		viper.Set("fileTypes.disabledBinaryExtensions", []string{".exe", ".dll"})
		viper.Set("fileTypes.disabledLanguageExtensions", []string{".rb", ".pl"})

		// Test getter functions
		if GetFileTypesEnabled() {
			t.Error("Expected file types to be disabled")
		}

		customImages := GetCustomImageExtensions()
		expectedImages := []string{".webp", ".avif"}
		if len(customImages) != len(expectedImages) {
			t.Errorf("Expected %d custom image extensions, got %d", len(expectedImages), len(customImages))
		}
		for i, ext := range expectedImages {
			if customImages[i] != ext {
				t.Errorf("Expected custom image extension %s, got %s", ext, customImages[i])
			}
		}

		customBinary := GetCustomBinaryExtensions()
		expectedBinary := []string{".custom", ".mybin"}
		if len(customBinary) != len(expectedBinary) {
			t.Errorf("Expected %d custom binary extensions, got %d", len(expectedBinary), len(customBinary))
		}
		for i, ext := range expectedBinary {
			if customBinary[i] != ext {
				t.Errorf("Expected custom binary extension %s, got %s", ext, customBinary[i])
			}
		}

		customLangs := GetCustomLanguages()
		expectedLangs := map[string]string{
			".zig": "zig",
			".v":   "vlang",
		}
		if len(customLangs) != len(expectedLangs) {
			t.Errorf("Expected %d custom languages, got %d", len(expectedLangs), len(customLangs))
		}
		for ext, lang := range expectedLangs {
			if customLangs[ext] != lang {
				t.Errorf("Expected custom language %s -> %s, got %s", ext, lang, customLangs[ext])
			}
		}

		disabledImages := GetDisabledImageExtensions()
		expectedDisabledImages := []string{".gif", ".bmp"}
		if len(disabledImages) != len(expectedDisabledImages) {
			t.Errorf("Expected %d disabled image extensions, got %d", len(expectedDisabledImages), len(disabledImages))
		}

		disabledBinary := GetDisabledBinaryExtensions()
		expectedDisabledBinary := []string{".exe", ".dll"}
		if len(disabledBinary) != len(expectedDisabledBinary) {
			t.Errorf("Expected %d disabled binary extensions, got %d", len(expectedDisabledBinary), len(disabledBinary))
		}

		disabledLangs := GetDisabledLanguageExtensions()
		expectedDisabledLangs := []string{".rb", ".pl"}
		if len(disabledLangs) != len(expectedDisabledLangs) {
			t.Errorf("Expected %d disabled language extensions, got %d", len(expectedDisabledLangs), len(disabledLangs))
		}
	})

	// Test validation
	t.Run("ValidationSuccess", func(t *testing.T) {
		viper.Reset()
		setDefaultConfig()

		// Set valid configuration
		viper.Set("fileTypes.customImageExtensions", []string{".webp", ".avif"})
		viper.Set("fileTypes.customBinaryExtensions", []string{".custom"})
		viper.Set("fileTypes.customLanguages", map[string]string{
			".zig": "zig",
			".v":   "vlang",
		})

		err := ValidateConfig()
		if err != nil {
			t.Errorf("Expected validation to pass with valid config, got error: %v", err)
		}
	})

	t.Run("ValidationFailure", func(t *testing.T) {
		// Test invalid custom image extensions
		viper.Reset()
		setDefaultConfig()
		viper.Set("fileTypes.customImageExtensions", []string{"", "webp"}) // Empty and missing dot

		err := ValidateConfig()
		if err == nil {
			t.Error("Expected validation to fail with invalid custom image extensions")
		}

		// Test invalid custom binary extensions
		viper.Reset()
		setDefaultConfig()
		viper.Set("fileTypes.customBinaryExtensions", []string{"custom"}) // Missing dot

		err = ValidateConfig()
		if err == nil {
			t.Error("Expected validation to fail with invalid custom binary extensions")
		}

		// Test invalid custom languages
		viper.Reset()
		setDefaultConfig()
		viper.Set("fileTypes.customLanguages", map[string]string{
			"zig": "zig", // Missing dot in extension
			".v":  "",    // Empty language
		})

		err = ValidateConfig()
		if err == nil {
			t.Error("Expected validation to fail with invalid custom languages")
		}
	})
}
