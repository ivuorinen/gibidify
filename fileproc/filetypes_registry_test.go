package fileproc

import (
	"testing"

	"github.com/ivuorinen/gibidify/shared"
)

// TestFileTypeRegistry_AddImageExtension tests adding image extensions.
func TestFileTypeRegistryAddImageExtension(t *testing.T) {
	registry := createModificationTestRegistry()

	testImageExtensionModifications(t, registry)
}

// TestFileTypeRegistry_AddBinaryExtension tests adding binary extensions.
func TestFileTypeRegistryAddBinaryExtension(t *testing.T) {
	registry := createModificationTestRegistry()

	testBinaryExtensionModifications(t, registry)
}

// TestFileTypeRegistry_AddLanguageMapping tests adding language mappings.
func TestFileTypeRegistryAddLanguageMapping(t *testing.T) {
	registry := createModificationTestRegistry()

	testLanguageMappingModifications(t, registry)
}

// createModificationTestRegistry creates a registry for modification tests.
func createModificationTestRegistry() *FileTypeRegistry {
	return &FileTypeRegistry{
		imageExts:    make(map[string]bool),
		binaryExts:   make(map[string]bool),
		languageMap:  make(map[string]string),
		extCache:     make(map[string]string, shared.FileTypeRegistryMaxCacheSize),
		resultCache:  make(map[string]FileTypeResult, shared.FileTypeRegistryMaxCacheSize),
		maxCacheSize: shared.FileTypeRegistryMaxCacheSize,
	}
}

// testImageExtensionModifications tests image extension modifications.
func testImageExtensionModifications(t *testing.T, registry *FileTypeRegistry) {
	t.Helper()

	// Add a new image extension
	registry.AddImageExtension(".webp")
	verifyImageExtension(t, registry, ".webp", shared.TestFileWebP, true)

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
	registry.AddBinaryExtension(shared.TestExtensionSpecial)
	verifyBinaryExtension(t, registry, shared.TestExtensionSpecial, "file.special", true)
	verifyBinaryExtension(t, registry, shared.TestExtensionSpecial, "file.SPECIAL", true)

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

	if registry.IsImage(filename) != expected {
		if expected {
			t.Errorf("Expected %s to be recognized as image after adding %s", filename, ext)
		} else {
			t.Errorf(shared.TestMsgExpectedExtensionWithoutDot)
		}
	}
}

// verifyBinaryExtension verifies binary extension behavior.
func verifyBinaryExtension(t *testing.T, registry *FileTypeRegistry, ext, filename string, expected bool) {
	t.Helper()

	if registry.IsBinary(filename) != expected {
		if expected {
			t.Errorf("Expected %s to be recognized as binary after adding %s", filename, ext)
		} else {
			t.Errorf(shared.TestMsgExpectedExtensionWithoutDot)
		}
	}
}

// verifyLanguageMapping verifies language mapping behavior.
func verifyLanguageMapping(t *testing.T, registry *FileTypeRegistry, filename, expectedLang string) {
	t.Helper()

	lang := registry.Language(filename)
	if lang != expectedLang {
		t.Errorf("Expected %s, got %s for %s", expectedLang, lang, filename)
	}
}

// verifyLanguageMappingAbsent verifies that a language mapping is absent.
func verifyLanguageMappingAbsent(t *testing.T, registry *FileTypeRegistry, _ string, filename string) {
	t.Helper()

	lang := registry.Language(filename)
	if lang != "" {
		t.Errorf(shared.TestMsgExpectedExtensionWithoutDot+", but got %s", lang)
	}
}

// TestFileTypeRegistry_DefaultRegistryConsistency tests default registry behavior.
func TestFileTypeRegistryDefaultRegistryConsistency(t *testing.T) {
	registry := DefaultRegistry()

	// Test that registry methods work consistently
	if !registry.IsImage(shared.TestFilePNG) {
		t.Error("Expected .png to be recognized as image")
	}
	if !registry.IsBinary(shared.TestFileEXE) {
		t.Error("Expected .exe to be recognized as binary")
	}
	if lang := registry.Language(shared.TestFileGo); lang != "go" {
		t.Errorf("Expected go, got %s", lang)
	}

	// Test that multiple calls return consistent results
	for i := 0; i < 5; i++ {
		if !registry.IsImage(shared.TestFileJPG) {
			t.Errorf("Iteration %d: Expected .jpg to be recognized as image", i)
		}
		if registry.IsBinary(shared.TestFileTXT) {
			t.Errorf("Iteration %d: Expected .txt to not be recognized as binary", i)
		}
	}
}

// TestFileTypeRegistry_GetStats tests the GetStats method.
func TestFileTypeRegistryGetStats(t *testing.T) {
	// Ensure clean, isolated state
	ResetRegistryForTesting()
	t.Cleanup(ResetRegistryForTesting)
	registry := DefaultRegistry()

	// Call some methods to populate cache and update stats
	registry.IsImage(shared.TestFilePNG)
	registry.IsBinary(shared.TestFileEXE)
	registry.Language(shared.TestFileGo)
	// Repeat to generate cache hits
	registry.IsImage(shared.TestFilePNG)
	registry.IsBinary(shared.TestFileEXE)
	registry.Language(shared.TestFileGo)

	// Get stats
	stats := registry.Stats()

	// Verify stats structure - all values are uint64 and therefore non-negative by definition
	// We can verify they exist and are properly initialized

	// Test that stats include our calls
	if stats.TotalLookups < 6 { // We made at least 6 calls above
		t.Errorf("Expected at least 6 total lookups, got %d", stats.TotalLookups)
	}

	// Total lookups should equal hits + misses
	if stats.TotalLookups != stats.CacheHits+stats.CacheMisses {
		t.Errorf("Total lookups (%d) should equal hits (%d) + misses (%d)",
			stats.TotalLookups, stats.CacheHits, stats.CacheMisses)
	}
	// With repeated lookups we should see some cache hits
	if stats.CacheHits == 0 {
		t.Error("Expected some cache hits after repeated lookups")
	}
}

// TestFileTypeRegistry_GetCacheInfo tests the GetCacheInfo method.
func TestFileTypeRegistryGetCacheInfo(t *testing.T) {
	// Ensure clean, isolated state
	ResetRegistryForTesting()
	t.Cleanup(ResetRegistryForTesting)
	registry := DefaultRegistry()

	// Call some methods to populate cache
	registry.IsImage("test1.png")
	registry.IsBinary("test2.exe")
	registry.Language("test3.go")
	registry.IsImage("test4.jpg")
	registry.IsBinary("test5.dll")

	// Get cache info
	extCacheSize, resultCacheSize, maxCacheSize := registry.CacheInfo()

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

	// Max cache size should be positive
	if maxCacheSize <= 0 {
		t.Errorf("Expected positive max cache size, got %d", maxCacheSize)
	}
}
