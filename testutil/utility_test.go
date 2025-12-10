package testutil

import (
	"path/filepath"
	"testing"
)

// TestGetBaseName tests the GetBaseName utility function.
func TestBaseName(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "simple filename",
			path:     "test.txt",
			expected: "test.txt",
		},
		{
			name:     "absolute path",
			path:     "/path/to/file.go",
			expected: "file.go",
		},
		{
			name:     "relative path",
			path:     "src/main.go",
			expected: "main.go",
		},
		{
			name:     "nested path",
			path:     "/deep/nested/path/to/file.json",
			expected: "file.json",
		},
		{
			name:     "path with trailing slash",
			path:     "/path/to/dir/",
			expected: "dir",
		},
		{
			name:     "empty path",
			path:     "",
			expected: ".",
		},
		{
			name:     "root path",
			path:     "/",
			expected: "/",
		},
		{
			name:     "current directory",
			path:     ".",
			expected: ".",
		},
		{
			name:     "parent directory",
			path:     "..",
			expected: "..",
		},
		{
			name:     "hidden file",
			path:     "/path/to/.hidden",
			expected: ".hidden",
		},
		{
			name:     "file with multiple dots",
			path:     "/path/file.test.go",
			expected: "file.test.go",
		},
		{
			name:     "windows-style path",
			path:     "C:\\Windows\\System32\\file.dll",
			expected: filepath.Base("C:\\Windows\\System32\\file.dll"), // Platform-specific result
		},
		{
			name:     "mixed path separators",
			path:     "/path\\to/file.txt",
			expected: "file.txt",
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				result := BaseName(tt.path)
				if result != tt.expected {
					t.Errorf("BaseName(%q) = %q, want %q", tt.path, result, tt.expected)
				}

				// Also verify against Go's filepath.Base for consistency
				expected := filepath.Base(tt.path)
				if result != expected {
					t.Errorf(
						"BaseName(%q) = %q, filepath.Base = %q, should be consistent",
						tt.path, result, expected,
					)
				}
			},
		)
	}
}

// BenchmarkGetBaseName benchmarks the GetBaseName function.
func BenchmarkBaseName(b *testing.B) {
	testPaths := []string{
		"simple.txt",
		"/path/to/file.go",
		"/very/deep/nested/path/to/some/file.json",
		"../relative/path.txt",
		"",
		"/",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		path := testPaths[i%len(testPaths)]
		_ = BaseName(path)
	}
}
