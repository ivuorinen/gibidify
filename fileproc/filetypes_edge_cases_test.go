package fileproc

import (
	"testing"

	"github.com/ivuorinen/gibidify/shared"
)

// TestFileTypeRegistry_EdgeCases tests edge cases and boundary conditions.
func TestFileTypeRegistryEdgeCases(t *testing.T) {
	registry := DefaultRegistry()

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
		t.Run(tc.name, func(_ *testing.T) {
			// These should not panic
			_ = registry.IsImage(tc.filename)
			_ = registry.IsBinary(tc.filename)
			_ = registry.Language(tc.filename)

			// Global functions should also not panic
			_ = IsImage(tc.filename)
			_ = IsBinary(tc.filename)
			_ = Language(tc.filename)
		})
	}
}

// TestFileTypeRegistry_MinimumExtensionLength tests the minimum extension length requirement.
func TestFileTypeRegistryMinimumExtensionLength(t *testing.T) {
	registry := DefaultRegistry()

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
			result := registry.Language(tt.filename)
			if result != tt.expected {
				t.Errorf("Language(%q) = %q, expected %q", tt.filename, result, tt.expected)
			}
		})
	}
}

// Benchmark tests for performance validation.
func BenchmarkFileTypeRegistryIsImage(b *testing.B) {
	registry := DefaultRegistry()
	filename := shared.TestFilePNG

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = registry.IsImage(filename)
	}
}

func BenchmarkFileTypeRegistryIsBinary(b *testing.B) {
	registry := DefaultRegistry()
	filename := shared.TestFileEXE

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = registry.IsBinary(filename)
	}
}

func BenchmarkFileTypeRegistryLanguage(b *testing.B) {
	registry := DefaultRegistry()
	filename := shared.TestFileGo

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = registry.Language(filename)
	}
}

func BenchmarkFileTypeRegistryGlobalFunctions(b *testing.B) {
	filename := shared.TestFileGo

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = IsImage(filename)
		_ = IsBinary(filename)
		_ = Language(filename)
	}
}

func BenchmarkFileTypeRegistryConcurrentAccess(b *testing.B) {
	filename := shared.TestFileGo

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = IsImage(filename)
			_ = IsBinary(filename)
			_ = Language(filename)
		}
	})
}
