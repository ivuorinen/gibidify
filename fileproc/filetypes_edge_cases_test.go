package fileproc

import (
	"testing"
)

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

// Benchmark tests for performance validation
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