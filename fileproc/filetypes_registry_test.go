package fileproc

import (
	"testing"
)

// TestFileTypeRegistry_ModificationMethods tests the modification methods of FileTypeRegistry.
func TestFileTypeRegistry_ModificationMethods(t *testing.T) {
	// Create a new registry instance for testing
	registry := &FileTypeRegistry{
		imageExts:   make(map[string]bool),
		binaryExts:  make(map[string]bool),
		languageMap: make(map[string]string),
	}

	// Test AddImageExtension
	t.Run("AddImageExtension", func(t *testing.T) {
		// Add a new image extension
		registry.AddImageExtension(".webp")
		if !registry.IsImage("test.webp") {
			t.Errorf("Expected .webp to be recognized as image after adding")
		}

		// Test case-insensitive addition
		registry.AddImageExtension(".AVIF")
		if !registry.IsImage("test.avif") {
			t.Errorf("Expected .avif to be recognized as image after adding .AVIF")
		}
		if !registry.IsImage("test.AVIF") {
			t.Errorf("Expected .AVIF to be recognized as image")
		}

		// Test with dot prefix
		registry.AddImageExtension("heic")
		if registry.IsImage("test.heic") {
			t.Errorf("Expected extension without dot to not work")
		}

		// Test with proper dot prefix
		registry.AddImageExtension(".heic")
		if !registry.IsImage("test.heic") {
			t.Errorf("Expected .heic to be recognized as image")
		}
	})

	// Test AddBinaryExtension
	t.Run("AddBinaryExtension", func(t *testing.T) {
		// Add a new binary extension
		registry.AddBinaryExtension(".custom")
		if !registry.IsBinary("file.custom") {
			t.Errorf("Expected .custom to be recognized as binary after adding")
		}

		// Test case-insensitive addition
		registry.AddBinaryExtension(".SPECIAL")
		if !registry.IsBinary("file.special") {
			t.Errorf("Expected .special to be recognized as binary after adding .SPECIAL")
		}
		if !registry.IsBinary("file.SPECIAL") {
			t.Errorf("Expected .SPECIAL to be recognized as binary")
		}

		// Test with dot prefix
		registry.AddBinaryExtension("bin")
		if registry.IsBinary("file.bin") {
			t.Errorf("Expected extension without dot to not work")
		}

		// Test with proper dot prefix
		registry.AddBinaryExtension(".bin")
		if !registry.IsBinary("file.bin") {
			t.Errorf("Expected .bin to be recognized as binary")
		}
	})

	// Test AddLanguageMapping
	t.Run("AddLanguageMapping", func(t *testing.T) {
		// Add a new language mapping
		registry.AddLanguageMapping(".xyz", "CustomLang")
		if lang := registry.GetLanguage("file.xyz"); lang != "CustomLang" {
			t.Errorf("Expected CustomLang, got %s", lang)
		}

		// Test case-insensitive addition
		registry.AddLanguageMapping(".ABC", "UpperLang")
		if lang := registry.GetLanguage("file.abc"); lang != "UpperLang" {
			t.Errorf("Expected UpperLang, got %s", lang)
		}
		if lang := registry.GetLanguage("file.ABC"); lang != "UpperLang" {
			t.Errorf("Expected UpperLang for uppercase, got %s", lang)
		}

		// Test with dot prefix
		registry.AddLanguageMapping("nolang", "NoLang")
		if lang := registry.GetLanguage("file.nolang"); lang == "NoLang" {
			t.Errorf("Expected extension without dot to not work")
		}

		// Test with proper dot prefix
		registry.AddLanguageMapping(".nolang", "NoLang")
		if lang := registry.GetLanguage("file.nolang"); lang != "NoLang" {
			t.Errorf("Expected NoLang, got %s", lang)
		}

		// Test overriding existing mapping
		registry.AddLanguageMapping(".xyz", "NewCustomLang")
		if lang := registry.GetLanguage("file.xyz"); lang != "NewCustomLang" {
			t.Errorf("Expected NewCustomLang after override, got %s", lang)
		}
	})
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
