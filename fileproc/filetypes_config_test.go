package fileproc

import (
	"testing"

	"github.com/ivuorinen/gibidify/shared"
)

const (
	zigLang = "zig"
)

// TestFileTypeRegistry_ApplyCustomExtensions tests applying custom extensions.
func TestFileTypeRegistryApplyCustomExtensions(t *testing.T) {
	registry := createEmptyTestRegistry()

	customImages := []string{".webp", ".avif", ".heic"}
	customBinary := []string{".custom", ".mybin"}
	customLanguages := map[string]string{
		".zig":  zigLang,
		".odin": "odin",
		".v":    "vlang",
	}

	registry.ApplyCustomExtensions(customImages, customBinary, customLanguages)

	verifyCustomExtensions(t, registry, customImages, customBinary, customLanguages)
}

// TestFileTypeRegistry_DisableExtensions tests disabling extensions.
func TestFileTypeRegistryDisableExtensions(t *testing.T) {
	registry := createEmptyTestRegistry()

	// Add some extensions first
	setupRegistryExtensions(registry)

	// Verify they work before disabling
	verifyExtensionsEnabled(t, registry)

	// Disable some extensions
	disabledImages := []string{".png"}
	disabledBinary := []string{".exe"}
	disabledLanguages := []string{".go"}

	registry.DisableExtensions(disabledImages, disabledBinary, disabledLanguages)

	// Verify disabled and remaining extensions
	verifyExtensionsDisabled(t, registry)
	verifyRemainingExtensions(t, registry)
}

// TestFileTypeRegistry_EmptyValuesHandling tests handling of empty values.
func TestFileTypeRegistryEmptyValuesHandling(t *testing.T) {
	registry := createEmptyTestRegistry()

	customImages := []string{"", shared.TestExtensionValid, ""}
	customBinary := []string{"", shared.TestExtensionValid}
	customLanguages := map[string]string{
		"":                        "invalid",
		shared.TestExtensionValid: "",
		".good":                   "good",
	}

	registry.ApplyCustomExtensions(customImages, customBinary, customLanguages)

	verifyEmptyValueHandling(t, registry)
}

// TestFileTypeRegistry_CaseInsensitiveHandling tests case insensitive handling.
func TestFileTypeRegistryCaseInsensitiveHandling(t *testing.T) {
	registry := createEmptyTestRegistry()

	customImages := []string{".WEBP", ".Avif"}
	customBinary := []string{".CUSTOM", ".MyBin"}
	customLanguages := map[string]string{
		".ZIG":  zigLang,
		".Odin": "odin",
	}

	registry.ApplyCustomExtensions(customImages, customBinary, customLanguages)

	verifyCaseInsensitiveHandling(t, registry)
}

// createEmptyTestRegistry creates a new empty test registry instance for config testing.
func createEmptyTestRegistry() *FileTypeRegistry {
	return &FileTypeRegistry{
		imageExts:    make(map[string]bool),
		binaryExts:   make(map[string]bool),
		languageMap:  make(map[string]string),
		extCache:     make(map[string]string, shared.FileTypeRegistryMaxCacheSize),
		resultCache:  make(map[string]FileTypeResult, shared.FileTypeRegistryMaxCacheSize),
		maxCacheSize: shared.FileTypeRegistryMaxCacheSize,
	}
}

// verifyCustomExtensions verifies that custom extensions are applied correctly.
func verifyCustomExtensions(
	t *testing.T,
	registry *FileTypeRegistry,
	customImages, customBinary []string,
	customLanguages map[string]string,
) {
	t.Helper()

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
		if lang := registry.Language("test" + ext); lang != expectedLang {
			t.Errorf("Expected %s to map to %s, got %s", ext, expectedLang, lang)
		}
	}
}

// setupRegistryExtensions adds test extensions to the registry.
func setupRegistryExtensions(registry *FileTypeRegistry) {
	registry.AddImageExtension(".png")
	registry.AddImageExtension(".jpg")
	registry.AddBinaryExtension(".exe")
	registry.AddBinaryExtension(".dll")
	registry.AddLanguageMapping(".go", "go")
	registry.AddLanguageMapping(".py", "python")
}

// verifyExtensionsEnabled verifies that extensions are enabled before disabling.
func verifyExtensionsEnabled(t *testing.T, registry *FileTypeRegistry) {
	t.Helper()

	if !registry.IsImage(shared.TestFilePNG) {
		t.Error("Expected .png to be image before disabling")
	}
	if !registry.IsBinary(shared.TestFileEXE) {
		t.Error("Expected .exe to be binary before disabling")
	}
	if registry.Language(shared.TestFileGo) != "go" {
		t.Error("Expected .go to map to go before disabling")
	}
}

// verifyExtensionsDisabled verifies that disabled extensions no longer work.
func verifyExtensionsDisabled(t *testing.T, registry *FileTypeRegistry) {
	t.Helper()

	if registry.IsImage(shared.TestFilePNG) {
		t.Error("Expected .png to not be image after disabling")
	}
	if registry.IsBinary(shared.TestFileEXE) {
		t.Error("Expected .exe to not be binary after disabling")
	}
	if registry.Language(shared.TestFileGo) != "" {
		t.Error("Expected .go to not map to language after disabling")
	}
}

// verifyRemainingExtensions verifies that non-disabled extensions still work.
func verifyRemainingExtensions(t *testing.T, registry *FileTypeRegistry) {
	t.Helper()

	if !registry.IsImage(shared.TestFileJPG) {
		t.Error("Expected .jpg to still be image after disabling .png")
	}
	if !registry.IsBinary(shared.TestFileDLL) {
		t.Error("Expected .dll to still be binary after disabling .exe")
	}
	if registry.Language(shared.TestFilePy) != "python" {
		t.Error("Expected .py to still map to python after disabling .go")
	}
}

// verifyEmptyValueHandling verifies handling of empty values.
func verifyEmptyValueHandling(t *testing.T, registry *FileTypeRegistry) {
	t.Helper()

	if registry.IsImage("test") {
		t.Error("Expected empty extension to not be added as image")
	}
	if !registry.IsImage(shared.TestFileValid) {
		t.Error("Expected .valid to be added as image")
	}
	if registry.IsBinary("test") {
		t.Error("Expected empty extension to not be added as binary")
	}
	if !registry.IsBinary(shared.TestFileValid) {
		t.Error("Expected .valid to be added as binary")
	}
	if registry.Language("test") != "" {
		t.Error("Expected empty extension to not be added as language")
	}
	if registry.Language(shared.TestFileValid) != "" {
		t.Error("Expected .valid with empty language to not be added")
	}
	if registry.Language("test.good") != "good" {
		t.Error("Expected .good to map to good")
	}
}

// verifyCaseInsensitiveHandling verifies case insensitive handling.
func verifyCaseInsensitiveHandling(t *testing.T, registry *FileTypeRegistry) {
	t.Helper()

	if !registry.IsImage(shared.TestFileWebP) {
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
	if registry.Language("test.zig") != zigLang {
		t.Error("Expected .zig (lowercase) to work after adding .ZIG")
	}
	if registry.Language("test.ZIG") != zigLang {
		t.Error("Expected .ZIG (uppercase) to work")
	}
}

// TestConfigureFromSettings tests the global configuration function.
func TestConfigureFromSettings(t *testing.T) {
	// Reset registry to ensure clean state
	ResetRegistryForTesting()

	// Test configuration application
	customImages := []string{".webp", ".avif"}
	customBinary := []string{".custom"}
	customLanguages := map[string]string{".zig": zigLang}
	disabledImages := []string{".gif"}   // Disable default extension
	disabledBinary := []string{".exe"}   // Disable default extension
	disabledLanguages := []string{".rb"} // Disable default extension

	ConfigureFromSettings(
		customImages,
		customBinary,
		customLanguages,
		disabledImages,
		disabledBinary,
		disabledLanguages,
	)

	// Test that custom extensions work
	if !IsImage(shared.TestFileWebP) {
		t.Error("Expected custom image extension .webp to work")
	}
	if !IsBinary("test.custom") {
		t.Error("Expected custom binary extension .custom to work")
	}
	if Language("test.zig") != zigLang {
		t.Error("Expected custom language .zig to work")
	}

	// Test that disabled extensions don't work
	if IsImage("test.gif") {
		t.Error("Expected disabled image extension .gif to not work")
	}
	if IsBinary(shared.TestFileEXE) {
		t.Error("Expected disabled binary extension .exe to not work")
	}
	if Language("test.rb") != "" {
		t.Error("Expected disabled language extension .rb to not work")
	}

	// Test that non-disabled defaults still work
	if !IsImage(shared.TestFilePNG) {
		t.Error("Expected non-disabled image extension .png to still work")
	}
	if !IsBinary(shared.TestFileDLL) {
		t.Error("Expected non-disabled binary extension .dll to still work")
	}
	if Language(shared.TestFileGo) != "go" {
		t.Error("Expected non-disabled language extension .go to still work")
	}

	// Test multiple calls don't override previous configuration
	ConfigureFromSettings(
		[]string{".extra"},
		[]string{},
		map[string]string{},
		[]string{},
		[]string{},
		[]string{},
	)

	// Previous configuration should still work
	if !IsImage(shared.TestFileWebP) {
		t.Error("Expected previous configuration to persist")
	}
	// New configuration should also work
	if !IsImage("test.extra") {
		t.Error("Expected new configuration to be applied")
	}

	// Reset registry after test to avoid affecting other tests
	ResetRegistryForTesting()
}
