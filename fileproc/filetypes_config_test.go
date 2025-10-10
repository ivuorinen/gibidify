package fileproc

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestFileTypeRegistry_Configuration tests the configuration functionality.
func TestFileTypeRegistry_Configuration(t *testing.T) {
	// Create a new registry instance for testing
	registry := &FileTypeRegistry{
		imageExts:   make(map[string]bool),
		binaryExts:  make(map[string]bool),
		languageMap: make(map[string]string),
	}

	// Test ApplyCustomExtensions
	t.Run("ApplyCustomExtensions", func(t *testing.T) {
		customImages := []string{".webp", ".avif", ".heic"}
		customBinary := []string{".custom", ".mybin"}
		customLanguages := map[string]string{
			".zig":  "zig",
			".odin": "odin",
			".v":    "vlang",
		}

		registry.ApplyCustomExtensions(customImages, customBinary, customLanguages)

		// Test custom image extensions
		for _, ext := range customImages {
			if !registry.IsImage("test" + ext) {
				t.Errorf("Expected %s to be recognized as image", ext)
			}
		}

		// Test custom binary extensions
		for _, ext := range customBinary {
			if !registry.IsBinary("test" + ext) {
				t.Errorf("Expected %s to be recognized as binary", ext)
			}
		}

		// Test custom language mappings
		for ext, expectedLang := range customLanguages {
			if lang := registry.GetLanguage("test" + ext); lang != expectedLang {
				t.Errorf("Expected %s to map to %s, got %s", ext, expectedLang, lang)
			}
		}
	})

	// Test DisableExtensions
	t.Run("DisableExtensions", func(t *testing.T) {
		// Add some extensions first
		registry.AddImageExtension(".png")
		registry.AddImageExtension(".jpg")
		registry.AddBinaryExtension(".exe")
		registry.AddBinaryExtension(".dll")
		registry.AddLanguageMapping(".go", "go")
		registry.AddLanguageMapping(".py", "python")

		// Verify they work
		if !registry.IsImage("test.png") {
			t.Error("Expected .png to be image before disabling")
		}
		if !registry.IsBinary("test.exe") {
			t.Error("Expected .exe to be binary before disabling")
		}
		if registry.GetLanguage("test.go") != "go" {
			t.Error("Expected .go to map to go before disabling")
		}

		// Disable some extensions
		disabledImages := []string{".png"}
		disabledBinary := []string{".exe"}
		disabledLanguages := []string{".go"}

		registry.DisableExtensions(disabledImages, disabledBinary, disabledLanguages)

		// Test that disabled extensions no longer work
		if registry.IsImage("test.png") {
			t.Error("Expected .png to not be image after disabling")
		}
		if registry.IsBinary("test.exe") {
			t.Error("Expected .exe to not be binary after disabling")
		}
		if registry.GetLanguage("test.go") != "" {
			t.Error("Expected .go to not map to language after disabling")
		}

		// Test that non-disabled extensions still work
		if !registry.IsImage("test.jpg") {
			t.Error("Expected .jpg to still be image after disabling .png")
		}
		if !registry.IsBinary("test.dll") {
			t.Error("Expected .dll to still be binary after disabling .exe")
		}
		if registry.GetLanguage("test.py") != "python" {
			t.Error("Expected .py to still map to python after disabling .go")
		}
	})

	// Test empty values handling
	t.Run("EmptyValuesHandling", func(t *testing.T) {
		registry := &FileTypeRegistry{
			imageExts:   make(map[string]bool),
			binaryExts:  make(map[string]bool),
			languageMap: make(map[string]string),
		}

		// Test with empty values
		customImages := []string{"", ".valid", ""}
		customBinary := []string{"", ".valid"}
		customLanguages := map[string]string{
			"":       "invalid",
			".valid": "",
			".good":  "good",
		}

		registry.ApplyCustomExtensions(customImages, customBinary, customLanguages)

		// Only valid entries should be added
		if registry.IsImage("test.") {
			t.Error("Expected empty extension to not be added as image")
		}
		if !registry.IsImage("test.valid") {
			t.Error("Expected .valid to be added as image")
		}
		if registry.IsBinary("test.") {
			t.Error("Expected empty extension to not be added as binary")
		}
		if !registry.IsBinary("test.valid") {
			t.Error("Expected .valid to be added as binary")
		}
		if registry.GetLanguage("test.") != "" {
			t.Error("Expected empty extension to not be added as language")
		}
		if registry.GetLanguage("test.valid") != "" {
			t.Error("Expected .valid with empty language to not be added")
		}
		if registry.GetLanguage("test.good") != "good" {
			t.Error("Expected .good to map to good")
		}
	})

	// Test case-insensitive handling
	t.Run("CaseInsensitiveHandling", func(t *testing.T) {
		registry := &FileTypeRegistry{
			imageExts:   make(map[string]bool),
			binaryExts:  make(map[string]bool),
			languageMap: make(map[string]string),
		}

		customImages := []string{".WEBP", ".Avif"}
		customBinary := []string{".CUSTOM", ".MyBin"}
		customLanguages := map[string]string{
			".ZIG":  "zig",
			".Odin": "odin",
		}

		registry.ApplyCustomExtensions(customImages, customBinary, customLanguages)

		// Test that both upper and lower case work
		if !registry.IsImage("test.webp") {
			t.Error("Expected .webp (lowercase) to work after adding .WEBP")
		}
		if !registry.IsImage("test.WEBP") {
			t.Error("Expected .WEBP (uppercase) to work")
		}
		if !registry.IsBinary("test.custom") {
			t.Error("Expected .custom (lowercase) to work after adding .CUSTOM")
		}
		if !registry.IsBinary("test.CUSTOM") {
			t.Error("Expected .CUSTOM (uppercase) to work")
		}
		if registry.GetLanguage("test.zig") != "zig" {
			t.Error("Expected .zig (lowercase) to work after adding .ZIG")
		}
		if registry.GetLanguage("test.ZIG") != "zig" {
			t.Error("Expected .ZIG (uppercase) to work")
		}
	})
}

// TestConfigureFromSettings tests the global configuration function.
func TestConfigureFromSettings(t *testing.T) {
	// Reset registry to ensure clean state
	ResetRegistryForTesting()
	// Ensure cleanup runs even if test fails
	t.Cleanup(ResetRegistryForTesting)

	// Test configuration application
	customImages := []string{".webp", ".avif"}
	customBinary := []string{".custom"}
	customLanguages := map[string]string{".zig": "zig"}
	disabledImages := []string{".gif"}   // Disable default extension
	disabledBinary := []string{".exe"}   // Disable default extension
	disabledLanguages := []string{".rb"} // Disable default extension

	err := ConfigureFromSettings(RegistryConfig{
		CustomImages:      customImages,
		CustomBinary:      customBinary,
		CustomLanguages:   customLanguages,
		DisabledImages:    disabledImages,
		DisabledBinary:    disabledBinary,
		DisabledLanguages: disabledLanguages,
	})
	require.NoError(t, err)

	// Test that custom extensions work
	if !IsImage("test.webp") {
		t.Error("Expected custom image extension .webp to work")
	}
	if !IsBinary("test.custom") {
		t.Error("Expected custom binary extension .custom to work")
	}
	if GetLanguage("test.zig") != "zig" {
		t.Error("Expected custom language .zig to work")
	}

	// Test that disabled extensions don't work
	if IsImage("test.gif") {
		t.Error("Expected disabled image extension .gif to not work")
	}
	if IsBinary("test.exe") {
		t.Error("Expected disabled binary extension .exe to not work")
	}
	if GetLanguage("test.rb") != "" {
		t.Error("Expected disabled language extension .rb to not work")
	}

	// Test that non-disabled defaults still work
	if !IsImage("test.png") {
		t.Error("Expected non-disabled image extension .png to still work")
	}
	if !IsBinary("test.dll") {
		t.Error("Expected non-disabled binary extension .dll to still work")
	}
	if GetLanguage("test.go") != "go" {
		t.Error("Expected non-disabled language extension .go to still work")
	}

	// Test multiple calls don't override previous configuration
	err = ConfigureFromSettings(RegistryConfig{
		CustomImages:      []string{".extra"},
		CustomBinary:      []string{},
		CustomLanguages:   map[string]string{},
		DisabledImages:    []string{},
		DisabledBinary:    []string{},
		DisabledLanguages: []string{},
	})
	require.NoError(t, err)

	// Previous configuration should still work
	if !IsImage("test.webp") {
		t.Error("Expected previous configuration to persist")
	}
	// New configuration should also work
	if !IsImage("test.extra") {
		t.Error("Expected new configuration to be applied")
	}
}
