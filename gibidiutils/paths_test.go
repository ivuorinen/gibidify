// Package gibidiutils provides common utility functions for gibidify.
package gibidiutils

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestGetAbsolutePath(t *testing.T) {
	// Get current working directory for tests
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	tests := []struct {
		name        string
		path        string
		wantPrefix  string
		wantErr     bool
		wantErrMsg  string
		skipWindows bool
	}{
		{
			name:       "absolute path unchanged",
			path:       cwd,
			wantPrefix: cwd,
			wantErr:    false,
		},
		{
			name:       "relative path current directory",
			path:       ".",
			wantPrefix: cwd,
			wantErr:    false,
		},
		{
			name:       "relative path parent directory",
			path:       "..",
			wantPrefix: filepath.Dir(cwd),
			wantErr:    false,
		},
		{
			name:       "relative path with file",
			path:       "test.txt",
			wantPrefix: filepath.Join(cwd, "test.txt"),
			wantErr:    false,
		},
		{
			name:       "relative path with subdirectory",
			path:       "subdir/file.go",
			wantPrefix: filepath.Join(cwd, "subdir", "file.go"),
			wantErr:    false,
		},
		{
			name:       "empty path",
			path:       "",
			wantPrefix: cwd,
			wantErr:    false,
		},
		{
			name:        "path with tilde",
			path:        "~/test",
			wantPrefix:  filepath.Join(cwd, "~", "test"),
			wantErr:     false,
			skipWindows: false,
		},
		{
			name:       "path with multiple separators",
			path:       "path//to///file",
			wantPrefix: filepath.Join(cwd, "path", "to", "file"),
			wantErr:    false,
		},
		{
			name:       "path with trailing separator",
			path:       "path/",
			wantPrefix: filepath.Join(cwd, "path"),
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.skipWindows && runtime.GOOS == "windows" {
				t.Skip("Skipping test on Windows")
			}

			got, err := GetAbsolutePath(tt.path)

			if tt.wantErr {
				if err == nil {
					t.Errorf("GetAbsolutePath() error = nil, wantErr %v", tt.wantErr)
					return
				}
				if tt.wantErrMsg != "" && !strings.Contains(err.Error(), tt.wantErrMsg) {
					t.Errorf("GetAbsolutePath() error = %v, want error containing %v", err, tt.wantErrMsg)
				}
				return
			}

			if err != nil {
				t.Errorf("GetAbsolutePath() unexpected error = %v", err)
				return
			}

			// Clean the expected path for comparison
			wantClean := filepath.Clean(tt.wantPrefix)
			gotClean := filepath.Clean(got)

			if gotClean != wantClean {
				t.Errorf("GetAbsolutePath() = %v, want %v", gotClean, wantClean)
			}

			// Verify the result is actually absolute
			if !filepath.IsAbs(got) {
				t.Errorf("GetAbsolutePath() returned non-absolute path: %v", got)
			}
		})
	}
}

func TestGetAbsolutePathSpecialCases(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping Unix-specific tests on Windows")
	}

	tests := []struct {
		name    string
		setup   func() (string, func())
		path    string
		wantErr bool
	}{
		{
			name: "symlink to directory",
			setup: func() (string, func()) {
				tmpDir := t.TempDir()
				target := filepath.Join(tmpDir, "target")
				link := filepath.Join(tmpDir, "link")

				if err := os.Mkdir(target, 0o750); err != nil {
					t.Fatalf("Failed to create target directory: %v", err)
				}
				if err := os.Symlink(target, link); err != nil {
					t.Fatalf("Failed to create symlink: %v", err)
				}

				return link, func() {}
			},
			path:    "",
			wantErr: false,
		},
		{
			name: "broken symlink",
			setup: func() (string, func()) {
				tmpDir := t.TempDir()
				link := filepath.Join(tmpDir, "broken_link")

				if err := os.Symlink("/nonexistent/path", link); err != nil {
					t.Fatalf("Failed to create broken symlink: %v", err)
				}

				return link, func() {}
			},
			path:    "",
			wantErr: false, // filepath.Abs still works with broken symlinks
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path, cleanup := tt.setup()
			defer cleanup()

			if tt.path == "" {
				tt.path = path
			}

			got, err := GetAbsolutePath(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAbsolutePath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil && !filepath.IsAbs(got) {
				t.Errorf("GetAbsolutePath() returned non-absolute path: %v", got)
			}
		})
	}
}

func TestGetAbsolutePathConcurrency(_ *testing.T) {
	// Test that GetAbsolutePath is safe for concurrent use
	paths := []string{".", "..", "test.go", "subdir/file.txt", "/tmp/test"}
	done := make(chan bool)

	for _, p := range paths {
		go func(path string) {
			_, _ = GetAbsolutePath(path)
			done <- true
		}(p)
	}

	// Wait for all goroutines to complete
	for range paths {
		<-done
	}
}

func TestGetAbsolutePathErrorFormatting(t *testing.T) {
	// This test verifies error message formatting
	// We need to trigger an actual error from filepath.Abs
	// On Unix systems, we can't easily trigger filepath.Abs errors
	// so we'll just verify the error wrapping works correctly

	// Create a test that would fail if filepath.Abs returns an error
	path := "test/path"
	got, err := GetAbsolutePath(path)
	if err != nil {
		// If we somehow get an error, verify it's properly formatted
		if !strings.Contains(err.Error(), "failed to get absolute path for") {
			t.Errorf("Error message format incorrect: %v", err)
		}
		if !strings.Contains(err.Error(), path) {
			t.Errorf("Error message should contain original path: %v", err)
		}
	} else {
		// Normal case - just verify we got a valid absolute path
		if !filepath.IsAbs(got) {
			t.Errorf("Expected absolute path, got: %v", got)
		}
	}
}

// BenchmarkGetAbsolutePath benchmarks the GetAbsolutePath function
func BenchmarkGetAbsolutePath(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = GetAbsolutePath("test/path/file.go")
	}
}

// BenchmarkGetAbsolutePathAbs benchmarks with already absolute path
func BenchmarkGetAbsolutePathAbs(b *testing.B) {
	absPath := "/home/user/test/file.go"
	if runtime.GOOS == "windows" {
		absPath = "C:\\Users\\test\\file.go"
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = GetAbsolutePath(absPath)
	}
}

// BenchmarkGetAbsolutePathCurrent benchmarks with current directory
func BenchmarkGetAbsolutePathCurrent(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = GetAbsolutePath(".")
	}
}
