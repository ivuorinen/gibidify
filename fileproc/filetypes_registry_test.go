package fileproc

import (
	"testing"
)

// TestFileTypeRegistry_AddImageExtension tests adding image extensions.
func TestFileTypeRegistry_AddImageExtension(t *testing.T) {
	registry := createModificationTestRegistry()

	testImageExtensionModifications(t, registry)
}

// TestFileTypeRegistry_AddBinaryExtension tests adding binary extensions.
func TestFileTypeRegistry_AddBinaryExtension(t *testing.T) {
	registry := createModificationTestRegistry()

	testBinaryExtensionModifications(t, registry)
}

// TestFileTypeRegistry_AddLanguageMapping tests adding language mappings.
func TestFileTypeRegistry_AddLanguageMapping(t *testing.T) {
	registry := createModificationTestRegistry()

	testLanguageMappingModifications(t, registry)
}

// createModificationTestRegistry creates a registry for modification tests.
func createModificationTestRegistry() *FileTypeRegistry {
	return &FileTypeRegistry{
		imageExts:   make(map[string]bool),
		binaryExts:  make(map[string]bool),
		languageMap: make(map[string]string),
	}
}

// testImageExtensionModifications tests image extension modifications.
func testImageExtensionModifications(t *testing.T, registry *FileTypeRegistry) {
	t.Helper()

	// Add a new image extension
	registry.AddImageExtension(".webp")
	verifyImageExtension(t, registry, ".webp", "test.webp", true)

	// Test case-insensitive addition
	registry.AddImageExtension(".AVIF")
	verifyImageExtension(t, registry, ".AVIF", "test.avif", true)
	verifyImageExtension(t, registry, ".AVIF", "test.AVIF", true)

	// Test with dot prefix
	registry.AddImageExtension("heic")
	verifyImageExtension(t, registry, "heic", "test.heic", false)

	// Test with proper dot prefix
	registry.AddImageExtension(".heic")
	verifyImageExtension(t, registry, ".heic", "test.heic", true)
}

// testBinaryExtensionModifications tests binary extension modifications.
func testBinaryExtensionModifications(t *testing.T, registry *FileTypeRegistry) {
	t.Helper()

	// Add a new binary extension
	registry.AddBinaryExtension(".custom")
	verifyBinaryExtension(t, registry, ".custom", "file.custom", true)

	// Test case-insensitive addition
	registry.AddBinaryExtension(".SPECIAL")
	verifyBinaryExtension(t, registry, ".SPECIAL", "file.special", true)
	verifyBinaryExtension(t, registry, ".SPECIAL", "file.SPECIAL", true)

	// Test with dot prefix
	registry.AddBinaryExtension("bin")
	verifyBinaryExtension(t, registry, "bin", "file.bin", false)

	// Test with proper dot prefix
	registry.AddBinaryExtension(".bin")
	verifyBinaryExtension(t, registry, ".bin", "file.bin", true)
}

// testLanguageMappingModifications tests language mapping modifications.
func testLanguageMappingModifications(t *testing.T, registry *FileTypeRegistry) {
	t.Helper()

	// Add a new language mapping
	registry.AddLanguageMapping(".xyz", "CustomLang")
	verifyLanguageMapping(t, registry, "file.xyz", "CustomLang")

	// Test case-insensitive addition
	registry.AddLanguageMapping(".ABC", "UpperLang")
	verifyLanguageMapping(t, registry, "file.abc", "UpperLang")
	verifyLanguageMapping(t, registry, "file.ABC", "UpperLang")

	// Test with dot prefix (should not work)
	registry.AddLanguageMapping("nolang", "NoLang")
	verifyLanguageMappingAbsent(t, registry, "nolang", "file.nolang")

	// Test with proper dot prefix
	registry.AddLanguageMapping(".nolang", "NoLang")
	verifyLanguageMapping(t, registry, "file.nolang", "NoLang")

	// Test overriding existing mapping
	registry.AddLanguageMapping(".xyz", "NewCustomLang")
	verifyLanguageMapping(t, registry, "file.xyz", "NewCustomLang")
}

// verifyImageExtension verifies image extension behavior.
func verifyImageExtension(t *testing.T, registry *FileTypeRegistry, ext, filename string, expected bool) {
	t.Helper()

	result := registry.IsImage(filename)
	if result != expected {
		if expected {
			t.Errorf("Expected %s to be recognized as image after adding %s", filename, ext)
		} else {
			t.Errorf("Expected extension %s without dot to not work", ext)
		}
	}
}

// verifyBinaryExtension verifies binary extension behavior.
func verifyBinaryExtension(t *testing.T, registry *FileTypeRegistry, ext, filename string, expected bool) {
	t.Helper()

	result := registry.IsBinary(filename)
	if result != expected {
		if expected {
			t.Errorf("Expected %s to be recognized as binary after adding %s", filename, ext)
		} else {
			t.Errorf("Expected extension %s without dot to not work", ext)
		}
	}
}

// verifyLanguageMapping verifies language mapping behavior.
func verifyLanguageMapping(t *testing.T, registry *FileTypeRegistry, filename, expectedLang string) {
	t.Helper()

	lang := registry.GetLanguage(filename)
	if lang != expectedLang {
		t.Errorf("Expected %s, got %s for %s", expectedLang, lang, filename)
	}
}

// verifyLanguageMappingAbsent verifies that a language mapping is absent.
func verifyLanguageMappingAbsent(t *testing.T, registry *FileTypeRegistry, ext, filename string) {
	t.Helper()

	lang := registry.GetLanguage(filename)
	if lang != "" {
		t.Errorf("Expected extension %s without dot to not work, but got %s", ext, lang)
	}
}

// TestFileTypeRegistry_DefaultRegistryConsistency tests default registry behavior.
func TestFileTypeRegistry_DefaultRegistryConsistency(t *testing.T) {
	registry := GetDefaultRegistry()

	// Test that registry methods work consistently
	if !registry.IsImage("test.png") {
		t.Error("Expected .png to be recognized as image")
	}
	if !registry.IsBinary("test.exe") {
		t.Error("Expected .exe to be recognized as binary")
	}
	if lang := registry.GetLanguage("test.go"); lang != "go" {
		t.Errorf("Expected go, got %s", lang)
	}

	// Test that multiple calls return consistent results
	for i := 0; i < 5; i++ {
		if !registry.IsImage("test.jpg") {
			t.Errorf("Iteration %d: Expected .jpg to be recognized as image", i)
		}
		if registry.IsBinary("test.txt") {
			t.Errorf("Iteration %d: Expected .txt to not be recognized as binary", i)
		}
	}
}

// TestFileTypeRegistry_GetStats tests the GetStats method.
func TestFileTypeRegistry_GetStats(t *testing.T) {
	registry := GetDefaultRegistry()

	// Call some methods to populate cache and update stats
	registry.IsImage("test.png")
	registry.IsBinary("test.exe")
	registry.GetLanguage("test.go")

	// Get stats
	stats := registry.GetStats()

	// Verify stats structure - all values are uint64 and therefore non-negative by definition
	// We can verify they exist and are properly initialized

	// Test that stats include our calls
	if stats.TotalLookups < 3 { // We made at least 3 calls above
		t.Errorf("Expected at least 3 total lookups, got %d", stats.TotalLookups)
	}

	// Total lookups should equal hits + misses
	if stats.TotalLookups != stats.CacheHits+stats.CacheMisses {
		t.Errorf("Total lookups (%d) should equal hits (%d) + misses (%d)",
			stats.TotalLookups, stats.CacheHits, stats.CacheMisses)
	}
}

// TestFileTypeRegistry_GetCacheInfo tests the GetCacheInfo method.
func TestFileTypeRegistry_GetCacheInfo(t *testing.T) {
	registry := GetDefaultRegistry()

	// Call some methods to populate cache
	registry.IsImage("test1.png")
	registry.IsBinary("test2.exe")
	registry.GetLanguage("test3.go")
	registry.IsImage("test4.jpg")
	registry.IsBinary("test5.dll")

	// Get cache info
	extCacheSize, resultCacheSize, maxCacheSize := registry.GetCacheInfo()

	// Verify cache info
	if extCacheSize < 0 {
		t.Error("Expected non-negative extension cache size")
	}
	if resultCacheSize < 0 {
		t.Error("Expected non-negative result cache size")
	}
	if maxCacheSize <= 0 {
		t.Error("Expected positive max cache size")
	}

	// We should have some cache entries from our calls
	totalCacheSize := extCacheSize + resultCacheSize
	if totalCacheSize == 0 {
		t.Error("Expected some cache entries after multiple calls")
	}

	// Max cache size should be reasonable
	if maxCacheSize < 100 || maxCacheSize > 10000 {
		t.Errorf("Max cache size %d seems unreasonable", maxCacheSize)
	}
}
