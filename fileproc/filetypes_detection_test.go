package fileproc

import (
	"testing"

	"github.com/ivuorinen/gibidify/shared"
)

// createTestRegistry creates a fresh FileTypeRegistry instance for testing.
// This helper reduces code duplication and ensures consistent registry initialization.
func createTestRegistry() *FileTypeRegistry {
	return &FileTypeRegistry{
		imageExts:    getImageExtensions(),
		binaryExts:   getBinaryExtensions(),
		languageMap:  getLanguageMap(),
		extCache:     make(map[string]string, shared.FileTypeRegistryMaxCacheSize),
		resultCache:  make(map[string]FileTypeResult, shared.FileTypeRegistryMaxCacheSize),
		maxCacheSize: shared.FileTypeRegistryMaxCacheSize,
	}
}

// TestFileTypeRegistry_LanguageDetection tests the language detection functionality.
func TestFileTypeRegistryLanguageDetection(t *testing.T) {
	registry := createTestRegistry()

	tests := []struct {
		filename string
		expected string
	}{
		// Programming languages
		{shared.TestFileMainGo, "go"},
		{shared.TestFileScriptPy, "python"},
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
		{"config.json", "json"},
		{"config.yaml", "yaml"},
		{"config.yml", "yaml"},
		{"data.toml", "toml"},
		{"page.md", "markdown"},
		{"readme.markdown", ""},
		{"doc.rst", "rst"},
		{"book.tex", "latex"},

		// Configuration files
		{"Dockerfile", ""},
		{"Makefile", ""},
		{"GNUmakefile", ""},

		// Case sensitivity tests
		{"MAIN.GO", "go"},
		{"SCRIPT.PY", "python"},
		{"APP.JS", "javascript"},

		// Unknown extensions
		{"unknown.xyz", ""},
		{"file.unknown", ""},
		{"noextension", ""},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			result := registry.Language(tt.filename)
			if result != tt.expected {
				t.Errorf("Language(%q) = %q, expected %q", tt.filename, result, tt.expected)
			}
		})
	}
}

// TestFileTypeRegistry_ImageDetection tests the image detection functionality.
func TestFileTypeRegistryImageDetection(t *testing.T) {
	registry := createTestRegistry()

	tests := []struct {
		filename string
		expected bool
	}{
		// Common image formats
		{"photo.png", true},
		{shared.TestFileImageJPG, true},
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
func TestFileTypeRegistryBinaryDetection(t *testing.T) {
	registry := createTestRegistry()

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

		// Media files (video/audio)
		{"video.mp4", true},
		{"movie.avi", true},
		{"clip.mov", true},
		{"song.mp3", true},
		{"audio.wav", true},
		{"music.flac", true},

		// Case sensitivity tests
		{"PROGRAM.EXE", true},
		{"LIBRARY.DLL", true},
		{"ARCHIVE.ZIP", true},

		// Non-binary files
		{"document.txt", false},
		{shared.TestFileScriptPy, false},
		{"config.json", false},
		{"style.css", false},
		{"page.html", false},

		// Edge cases
		{"", false},             // Empty filename
		{"binary", false},       // No extension
		{".exe", true},          // Just extension
		{"file.exe.txt", false}, // Multiple extensions
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
