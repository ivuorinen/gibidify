package fileproc

import (
	"fmt"
	"sync"
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

		// Test case insensitive addition
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
		if !registry.IsBinary("test.custom") {
			t.Errorf("Expected .custom to be recognized as binary after adding")
		}

		// Test case insensitive addition
		registry.AddBinaryExtension(".NEWBIN")
		if !registry.IsBinary("test.newbin") {
			t.Errorf("Expected .newbin to be recognized as binary after adding .NEWBIN")
		}
		if !registry.IsBinary("test.NEWBIN") {
			t.Errorf("Expected .NEWBIN to be recognized as binary")
		}

		// Test overwriting existing extension
		registry.AddBinaryExtension(".custom")
		if !registry.IsBinary("test.custom") {
			t.Errorf("Expected .custom to still be recognized as binary after re-adding")
		}
	})

	// Test AddLanguageMapping
	t.Run("AddLanguageMapping", func(t *testing.T) {
		// Add a new language mapping
		registry.AddLanguageMapping(".zig", "zig")
		if registry.GetLanguage("test.zig") != "zig" {
			t.Errorf("Expected .zig to map to 'zig', got '%s'", registry.GetLanguage("test.zig"))
		}

		// Test case insensitive addition
		registry.AddLanguageMapping(".V", "vlang")
		if registry.GetLanguage("test.v") != "vlang" {
			t.Errorf("Expected .v to map to 'vlang' after adding .V, got '%s'", registry.GetLanguage("test.v"))
		}
		if registry.GetLanguage("test.V") != "vlang" {
			t.Errorf("Expected .V to map to 'vlang', got '%s'", registry.GetLanguage("test.V"))
		}

		// Test overwriting existing mapping
		registry.AddLanguageMapping(".zig", "ziglang")
		if registry.GetLanguage("test.zig") != "ziglang" {
			t.Errorf("Expected .zig to map to 'ziglang' after update, got '%s'", registry.GetLanguage("test.zig"))
		}

		// Test empty language
		registry.AddLanguageMapping(".empty", "")
		if registry.GetLanguage("test.empty") != "" {
			t.Errorf("Expected .empty to map to empty string, got '%s'", registry.GetLanguage("test.empty"))
		}
	})
}

// TestFileTypeRegistry_LanguageDetection tests the language detection functionality.
func TestFileTypeRegistry_LanguageDetection(t *testing.T) {
	registry := GetDefaultRegistry()

	tests := []struct {
		filename string
		expected string
	}{
		// Programming languages
		{"main.go", "go"},
		{"script.py", "python"},
		{"app.js", "javascript"},
		{"component.tsx", "typescript"},
		{"service.ts", "typescript"},
		{"App.java", "java"},
		{"program.c", "c"},
		{"program.cpp", "cpp"},
		{"header.h", "c"},
		{"header.hpp", "cpp"},
		{"main.rs", "rust"},
		{"script.rb", "ruby"},
		{"index.php", "php"},
		{"app.swift", "swift"},
		{"MainActivity.kt", "kotlin"},
		{"Main.scala", "scala"},
		{"analysis.r", "r"},
		{"ViewController.m", "objc"},
		{"ViewController.mm", "objcpp"},
		{"Program.cs", "csharp"},
		{"Module.vb", "vbnet"},
		{"program.fs", "fsharp"},
		{"script.lua", "lua"},
		{"script.pl", "perl"},

		// Shell scripts
		{"script.sh", "bash"},
		{"script.bash", "bash"},
		{"script.zsh", "zsh"},
		{"script.fish", "fish"},
		{"script.ps1", "powershell"},
		{"script.bat", "batch"},
		{"script.cmd", "batch"},

		// Data and markup
		{"query.sql", "sql"},
		{"index.html", "html"},
		{"page.htm", "html"},
		{"data.xml", "xml"},
		{"style.css", "css"},
		{"style.scss", "scss"},
		{"style.sass", "sass"},
		{"style.less", "less"},
		{"data.json", "json"},
		{"config.yaml", "yaml"},
		{"config.yml", "yaml"},
		{"config.toml", "toml"},
		{"README.md", "markdown"},
		{"doc.rst", "rst"},
		{"paper.tex", "latex"},

		// Modern languages
		{"main.dart", "dart"},
		{"Main.elm", "elm"},
		{"core.clj", "clojure"},
		{"server.ex", "elixir"},
		{"test.exs", "elixir"},
		{"server.erl", "erlang"},
		{"header.hrl", "erlang"},
		{"main.hs", "haskell"},
		{"module.ml", "ocaml"},
		{"interface.mli", "ocaml"},
		{"main.nim", "nim"},
		{"config.nims", "nim"},

		// Web frameworks
		{"Component.vue", "vue"},
		{"Component.jsx", "javascript"},

		// Case sensitivity tests
		{"MAIN.GO", "go"},
		{"Script.PY", "python"},
		{"APP.JS", "javascript"},

		// Edge cases
		{"", ""},             // Empty filename
		{"a", ""},            // Too short (less than minExtensionLength)
		{"noext", ""},        // No extension
		{".hidden", ""},      // Hidden file with no name
		{"file.", ""},        // Extension is just a dot
		{"file.unknown", ""}, // Unknown extension
		{"file.123", ""},     // Numeric extension
		{"a.b", ""},          // Very short filename and extension
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			result := registry.GetLanguage(tt.filename)
			if result != tt.expected {
				t.Errorf("GetLanguage(%q) = %q, expected %q", tt.filename, result, tt.expected)
			}
		})
	}
}

// TestFileTypeRegistry_ImageDetection tests the image detection functionality.
func TestFileTypeRegistry_ImageDetection(t *testing.T) {
	registry := GetDefaultRegistry()

	tests := []struct {
		filename string
		expected bool
	}{
		// Common image formats
		{"photo.png", true},
		{"image.jpg", true},
		{"picture.jpeg", true},
		{"animation.gif", true},
		{"bitmap.bmp", true},
		{"image.tiff", true},
		{"scan.tif", true},
		{"vector.svg", true},
		{"modern.webp", true},
		{"favicon.ico", true},

		// Case sensitivity tests
		{"PHOTO.PNG", true},
		{"IMAGE.JPG", true},
		{"PICTURE.JPEG", true},

		// Non-image files
		{"document.txt", false},
		{"script.js", false},
		{"data.json", false},
		{"archive.zip", false},
		{"executable.exe", false},

		// Edge cases
		{"", false},              // Empty filename
		{"image", false},         // No extension
		{".png", true},           // Just extension
		{"file.png.bak", false},  // Multiple extensions
		{"image.unknown", false}, // Unknown extension
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			result := registry.IsImage(tt.filename)
			if result != tt.expected {
				t.Errorf("IsImage(%q) = %t, expected %t", tt.filename, result, tt.expected)
			}
		})
	}
}

// TestFileTypeRegistry_BinaryDetection tests the binary detection functionality.
func TestFileTypeRegistry_BinaryDetection(t *testing.T) {
	registry := GetDefaultRegistry()

	tests := []struct {
		filename string
		expected bool
	}{
		// Executable files
		{"program.exe", true},
		{"library.dll", true},
		{"libfoo.so", true},
		{"framework.dylib", true},
		{"data.bin", true},

		// Object and library files
		{"object.o", true},
		{"archive.a", true},
		{"library.lib", true},
		{"application.jar", true},
		{"bytecode.class", true},
		{"compiled.pyc", true},
		{"optimized.pyo", true},

		// System files
		{".DS_Store", true},

		// Document files (treated as binary)
		{"document.pdf", true},

		// Archive files
		{"archive.zip", true},
		{"backup.tar", true},
		{"compressed.gz", true},
		{"data.bz2", true},
		{"package.xz", true},
		{"archive.7z", true},
		{"backup.rar", true},

		// Font files
		{"font.ttf", true},
		{"font.otf", true},
		{"font.woff", true},
		{"font.woff2", true},

		// Media files
		{"song.mp3", true},
		{"video.mp4", true},
		{"movie.avi", true},
		{"clip.mov", true},
		{"video.wmv", true},
		{"animation.flv", true},
		{"modern.webm", true},
		{"audio.ogg", true},
		{"sound.wav", true},
		{"music.flac", true},

		// Database files
		{"data.dat", true},
		{"database.db", true},
		{"app.sqlite", true},

		// Case sensitivity tests
		{"PROGRAM.EXE", true},
		{"LIBRARY.DLL", true},

		// Non-binary files
		{"document.txt", false},
		{"script.js", false},
		{"data.json", false},
		{"style.css", false},
		{"page.html", false},

		// Edge cases
		{"", false},             // Empty filename
		{"binary", false},       // No extension
		{".exe", true},          // Just extension
		{"file.exe.bak", false}, // Multiple extensions
		{"file.unknown", false}, // Unknown extension
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			result := registry.IsBinary(tt.filename)
			if result != tt.expected {
				t.Errorf("IsBinary(%q) = %t, expected %t", tt.filename, result, tt.expected)
			}
		})
	}
}

// TestFileTypeRegistry_DefaultRegistryConsistency tests that the default registry is consistent.
func TestFileTypeRegistry_DefaultRegistryConsistency(t *testing.T) {
	// Get registry multiple times and ensure it's the same instance
	registry1 := GetDefaultRegistry()
	registry2 := GetDefaultRegistry()
	registry3 := getRegistry()

	if registry1 != registry2 {
		t.Error("GetDefaultRegistry() should return the same instance")
	}
	if registry1 != registry3 {
		t.Error("getRegistry() should return the same instance as GetDefaultRegistry()")
	}

	// Test that global functions use the same registry
	filename := "test.go"
	if IsImage(filename) != registry1.IsImage(filename) {
		t.Error("IsImage() global function should match registry method")
	}
	if IsBinary(filename) != registry1.IsBinary(filename) {
		t.Error("IsBinary() global function should match registry method")
	}
	if GetLanguage(filename) != registry1.GetLanguage(filename) {
		t.Error("GetLanguage() global function should match registry method")
	}
}

// TestFileTypeRegistry_ThreadSafety tests the thread safety of the FileTypeRegistry.
func TestFileTypeRegistry_ThreadSafety(t *testing.T) {
	const numGoroutines = 100
	const numOperationsPerGoroutine = 100

	var wg sync.WaitGroup

	// Test concurrent read operations
	t.Run("ConcurrentReads", func(t *testing.T) {
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				registry := GetDefaultRegistry()

				for j := 0; j < numOperationsPerGoroutine; j++ {
					// Test various file detection operations
					_ = registry.IsImage("test.png")
					_ = registry.IsBinary("test.exe")
					_ = registry.GetLanguage("test.go")

					// Test global functions too
					_ = IsImage("image.jpg")
					_ = IsBinary("binary.dll")
					_ = GetLanguage("script.py")
				}
			}(i)
		}
		wg.Wait()
	})

	// Test concurrent registry access (singleton creation)
	t.Run("ConcurrentRegistryAccess", func(t *testing.T) {
		// Reset the registry to test concurrent initialization
		// Note: This is not safe in a real application, but needed for testing
		registryOnce = sync.Once{}
		registry = nil

		registries := make([]*FileTypeRegistry, numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				registries[id] = GetDefaultRegistry()
			}(i)
		}
		wg.Wait()

		// Verify all goroutines got the same registry instance
		firstRegistry := registries[0]
		for i := 1; i < numGoroutines; i++ {
			if registries[i] != firstRegistry {
				t.Errorf("Registry %d is different from registry 0", i)
			}
		}
	})

	// Test concurrent modifications on separate registry instances
	t.Run("ConcurrentModifications", func(t *testing.T) {
		// Create separate registry instances for each goroutine to test modification thread safety
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()

				// Create a new registry instance for this goroutine
				registry := &FileTypeRegistry{
					imageExts:   make(map[string]bool),
					binaryExts:  make(map[string]bool),
					languageMap: make(map[string]string),
				}

				for j := 0; j < numOperationsPerGoroutine; j++ {
					// Add unique extensions for this goroutine
					extSuffix := fmt.Sprintf("_%d_%d", id, j)

					registry.AddImageExtension(".img" + extSuffix)
					registry.AddBinaryExtension(".bin" + extSuffix)
					registry.AddLanguageMapping(".lang"+extSuffix, "lang"+extSuffix)

					// Verify the additions worked
					if !registry.IsImage("test.img" + extSuffix) {
						t.Errorf("Failed to add image extension .img%s", extSuffix)
					}
					if !registry.IsBinary("test.bin" + extSuffix) {
						t.Errorf("Failed to add binary extension .bin%s", extSuffix)
					}
					if registry.GetLanguage("test.lang"+extSuffix) != "lang"+extSuffix {
						t.Errorf("Failed to add language mapping .lang%s", extSuffix)
					}
				}
			}(i)
		}
		wg.Wait()
	})
}

// TestFileTypeRegistry_EdgeCases tests edge cases and boundary conditions.
func TestFileTypeRegistry_EdgeCases(t *testing.T) {
	registry := GetDefaultRegistry()

	// Test various edge cases for filename handling
	edgeCases := []struct {
		name     string
		filename string
		desc     string
	}{
		{"empty", "", "empty filename"},
		{"single_char", "a", "single character filename"},
		{"just_dot", ".", "just a dot"},
		{"double_dot", "..", "double dot"},
		{"hidden_file", ".hidden", "hidden file"},
		{"hidden_with_ext", ".hidden.txt", "hidden file with extension"},
		{"multiple_dots", "file.tar.gz", "multiple extensions"},
		{"trailing_dot", "file.", "trailing dot"},
		{"unicode", "файл.txt", "unicode filename"},
		{"spaces", "my file.txt", "filename with spaces"},
		{"special_chars", "file@#$.txt", "filename with special characters"},
		{"very_long", "very_long_filename_with_many_characters_in_it.extension", "very long filename"},
		{"no_basename", ".gitignore", "dotfile with no basename"},
		{"case_mixed", "FiLe.ExT", "mixed case"},
	}

	for _, tc := range edgeCases {
		t.Run(tc.name, func(t *testing.T) {
			// These should not panic
			_ = registry.IsImage(tc.filename)
			_ = registry.IsBinary(tc.filename)
			_ = registry.GetLanguage(tc.filename)

			// Global functions should also not panic
			_ = IsImage(tc.filename)
			_ = IsBinary(tc.filename)
			_ = GetLanguage(tc.filename)
		})
	}
}

// TestFileTypeRegistry_MinimumExtensionLength tests the minimum extension length requirement.
func TestFileTypeRegistry_MinimumExtensionLength(t *testing.T) {
	registry := GetDefaultRegistry()

	tests := []struct {
		filename string
		expected string
	}{
		{"", ""},            // Empty filename
		{"a", ""},           // Single character (less than minExtensionLength)
		{"ab", ""},          // Two characters, no extension
		{"a.b", ""},         // Extension too short, but filename too short anyway
		{"ab.c", "c"},       // Valid: filename >= minExtensionLength and .c is valid extension
		{"a.go", "go"},      // Valid extension
		{"ab.py", "python"}, // Valid extension
		{"a.unknown", ""},   // Valid length but unknown extension
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			result := registry.GetLanguage(tt.filename)
			if result != tt.expected {
				t.Errorf("GetLanguage(%q) = %q, expected %q", tt.filename, result, tt.expected)
			}
		})
	}
}

// BenchmarkFileTypeRegistry tests performance of the registry operations.
func BenchmarkFileTypeRegistry_IsImage(b *testing.B) {
	registry := GetDefaultRegistry()
	filename := "test.png"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = registry.IsImage(filename)
	}
}

func BenchmarkFileTypeRegistry_IsBinary(b *testing.B) {
	registry := GetDefaultRegistry()
	filename := "test.exe"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = registry.IsBinary(filename)
	}
}

func BenchmarkFileTypeRegistry_GetLanguage(b *testing.B) {
	registry := GetDefaultRegistry()
	filename := "test.go"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = registry.GetLanguage(filename)
	}
}

func BenchmarkFileTypeRegistry_GlobalFunctions(b *testing.B) {
	filename := "test.go"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = IsImage(filename)
		_ = IsBinary(filename)
		_ = GetLanguage(filename)
	}
}

func BenchmarkFileTypeRegistry_ConcurrentAccess(b *testing.B) {
	filename := "test.go"

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = IsImage(filename)
			_ = IsBinary(filename)
			_ = GetLanguage(filename)
		}
	})
}

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

	// Test case insensitive handling
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
	registryOnce = sync.Once{}
	registry = nil

	// Test configuration application
	customImages := []string{".webp", ".avif"}
	customBinary := []string{".custom"}
	customLanguages := map[string]string{".zig": "zig"}
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
}
