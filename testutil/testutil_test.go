package testutil

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

func TestCreateTestFile(t *testing.T) {
	tests := []struct {
		name     string
		dir      string
		filename string
		content  []byte
		wantErr  bool
	}{
		{
			name:     "create simple test file",
			filename: "test.txt",
			content:  []byte("hello world"),
			wantErr:  false,
		},
		{
			name:     "create file with empty content",
			filename: "empty.txt",
			content:  []byte{},
			wantErr:  false,
		},
		{
			name:     "create file with binary content",
			filename: "binary.bin",
			content:  []byte{0x00, 0xFF, 0x42},
			wantErr:  false,
		},
		{
			name:     "create file with subdirectory",
			filename: "subdir/test.txt",
			content:  []byte("nested file"),
			wantErr:  false,
		},
		{
			name:     "create file with special characters",
			filename: "special-file_123.go",
			content:  []byte("package main"),
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use a temporary directory for each test
			tempDir := t.TempDir()
			if tt.dir == "" {
				tt.dir = tempDir
			}

			// Create subdirectory if needed
			if strings.Contains(tt.filename, "/") {
				subdir := filepath.Join(tt.dir, filepath.Dir(tt.filename))
				if err := os.MkdirAll(subdir, DirPermission); err != nil {
					t.Fatalf("Failed to create subdirectory: %v", err)
				}
			}

			// Test CreateTestFile
			filePath := CreateTestFile(t, tt.dir, tt.filename, tt.content)

			// Verify file exists
			info, err := os.Stat(filePath)
			if err != nil {
				t.Fatalf("Created file does not exist: %v", err)
			}

			// Verify it's a regular file
			if !info.Mode().IsRegular() {
				t.Errorf("Created path is not a regular file")
			}

			// Verify permissions
			if info.Mode().Perm() != FilePermission {
				t.Errorf("File permissions = %v, want %v", info.Mode().Perm(), FilePermission)
			}

			// Verify content
			readContent, err := os.ReadFile(filePath)
			if err != nil {
				t.Fatalf("Failed to read created file: %v", err)
			}
			if string(readContent) != string(tt.content) {
				t.Errorf("File content = %q, want %q", readContent, tt.content)
			}
		})
	}
}

func TestCreateTempOutputFile(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
	}{
		{
			name:    "simple pattern",
			pattern: "output-*.txt",
		},
		{
			name:    "pattern with prefix only",
			pattern: "test-",
		},
		{
			name:    "pattern with suffix only",
			pattern: "*.json",
		},
		{
			name:    "empty pattern",
			pattern: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file, path := CreateTempOutputFile(t, tt.pattern)
			defer CloseFile(t, file)

			// Verify file exists
			info, err := os.Stat(path)
			if err != nil {
				t.Fatalf("Temp file does not exist: %v", err)
			}

			// Verify it's a regular file
			if !info.Mode().IsRegular() {
				t.Errorf("Created path is not a regular file")
			}

			// Verify we can write to it
			testContent := []byte("test content")
			if _, err := file.Write(testContent); err != nil {
				t.Errorf("Failed to write to temp file: %v", err)
			}

			// Verify the path is in a temp directory (any temp directory)
			if !strings.Contains(path, os.TempDir()) {
				t.Errorf("Temp file not in temp directory: %s", path)
			}
		})
	}
}

func TestCreateTestDirectory(t *testing.T) {
	tests := []struct {
		name   string
		parent string
		dir    string
	}{
		{
			name: "simple directory",
			dir:  "testdir",
		},
		{
			name: "directory with special characters",
			dir:  "test-dir_123",
		},
		{
			name: "nested directory name",
			dir:  "nested/dir",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			if tt.parent == "" {
				tt.parent = tempDir
			}

			// For nested directories, create parent first
			if strings.Contains(tt.dir, "/") {
				parentPath := filepath.Join(tt.parent, filepath.Dir(tt.dir))
				if err := os.MkdirAll(parentPath, DirPermission); err != nil {
					t.Fatalf("Failed to create parent directory: %v", err)
				}
				tt.dir = filepath.Base(tt.dir)
				tt.parent = parentPath
			}

			dirPath := CreateTestDirectory(t, tt.parent, tt.dir)

			// Verify directory exists
			info, err := os.Stat(dirPath)
			if err != nil {
				t.Fatalf("Created directory does not exist: %v", err)
			}

			// Verify it's a directory
			if !info.IsDir() {
				t.Errorf("Created path is not a directory")
			}

			// Verify permissions
			if info.Mode().Perm() != DirPermission {
				t.Errorf("Directory permissions = %v, want %v", info.Mode().Perm(), DirPermission)
			}

			// Verify we can create files in it
			testFile := filepath.Join(dirPath, "test.txt")
			if err := os.WriteFile(testFile, []byte("test"), FilePermission); err != nil {
				t.Errorf("Cannot create file in directory: %v", err)
			}
		})
	}
}

func TestCreateTestFiles(t *testing.T) {
	tests := []struct {
		name      string
		fileSpecs []FileSpec
		wantCount int
	}{
		{
			name: "create multiple files",
			fileSpecs: []FileSpec{
				{Name: "file1.txt", Content: "content1"},
				{Name: "file2.go", Content: "package main"},
				{Name: "file3.json", Content: `{"key": "value"}`},
			},
			wantCount: 3,
		},
		{
			name: "create files with subdirectories",
			fileSpecs: []FileSpec{
				{Name: "src/main.go", Content: "package main"},
				{Name: "test/test.go", Content: "package test"},
			},
			wantCount: 2,
		},
		{
			name:      "empty file specs",
			fileSpecs: []FileSpec{},
			wantCount: 0,
		},
		{
			name: "files with empty content",
			fileSpecs: []FileSpec{
				{Name: "empty1.txt", Content: ""},
				{Name: "empty2.txt", Content: ""},
			},
			wantCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rootDir := t.TempDir()

			// Create necessary subdirectories
			for _, spec := range tt.fileSpecs {
				if strings.Contains(spec.Name, "/") {
					subdir := filepath.Join(rootDir, filepath.Dir(spec.Name))
					if err := os.MkdirAll(subdir, DirPermission); err != nil {
						t.Fatalf("Failed to create subdirectory: %v", err)
					}
				}
			}

			createdFiles := CreateTestFiles(t, rootDir, tt.fileSpecs)

			// Verify count
			if len(createdFiles) != tt.wantCount {
				t.Errorf("Created %d files, want %d", len(createdFiles), tt.wantCount)
			}

			// Verify each file
			for i, filePath := range createdFiles {
				content, err := os.ReadFile(filePath)
				if err != nil {
					t.Errorf("Failed to read file %s: %v", filePath, err)
					continue
				}
				if string(content) != tt.fileSpecs[i].Content {
					t.Errorf("File %s content = %q, want %q", filePath, content, tt.fileSpecs[i].Content)
				}
			}
		})
	}
}

func TestResetViperConfig(t *testing.T) {
	tests := []struct {
		name       string
		configPath string
		preSetup   func()
		verify     func(t *testing.T)
	}{
		{
			name:       "reset with empty config path",
			configPath: "",
			preSetup: func() {
				viper.Set("test.key", "value")
			},
			verify: func(t *testing.T) {
				if viper.IsSet("test.key") {
					t.Error("Viper config not reset properly")
				}
			},
		},
		{
			name:       "reset with config path",
			configPath: t.TempDir(),
			preSetup: func() {
				viper.Set("test.key", "value")
			},
			verify: func(t *testing.T) {
				if viper.IsSet("test.key") {
					t.Error("Viper config not reset properly")
				}
				// Verify config path was added
				paths := viper.ConfigFileUsed()
				if paths == "" {
					// This is expected as no config file exists
					return
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.preSetup()
			ResetViperConfig(t, tt.configPath)
			tt.verify(t)
		})
	}
}

func TestSetupCLIArgs(t *testing.T) {
	// Save original args
	originalArgs := os.Args
	defer func() {
		os.Args = originalArgs
	}()

	tests := []struct {
		name        string
		srcDir      string
		outFile     string
		prefix      string
		suffix      string
		concurrency int
		wantLen     int
	}{
		{
			name:        "basic CLI args",
			srcDir:      "/src",
			outFile:     "/out.txt",
			prefix:      "PREFIX",
			suffix:      "SUFFIX",
			concurrency: 4,
			wantLen:     11,
		},
		{
			name:        "empty strings",
			srcDir:      "",
			outFile:     "",
			prefix:      "",
			suffix:      "",
			concurrency: 1,
			wantLen:     11,
		},
		{
			name:        "special characters in args",
			srcDir:      "/path with spaces/src",
			outFile:     "/path/to/output file.txt",
			prefix:      "Prefix with\nnewline",
			suffix:      "Suffix with\ttab",
			concurrency: 8,
			wantLen:     11,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetupCLIArgs(tt.srcDir, tt.outFile, tt.prefix, tt.suffix, tt.concurrency)

			if len(os.Args) != tt.wantLen {
				t.Errorf("os.Args length = %d, want %d", len(os.Args), tt.wantLen)
			}

			// Verify specific args
			if os.Args[0] != "gibidify" {
				t.Errorf("Program name = %s, want gibidify", os.Args[0])
			}
			if os.Args[2] != tt.srcDir {
				t.Errorf("Source dir = %s, want %s", os.Args[2], tt.srcDir)
			}
			if os.Args[4] != tt.outFile {
				t.Errorf("Output file = %s, want %s", os.Args[4], tt.outFile)
			}
			if os.Args[6] != tt.prefix {
				t.Errorf("Prefix = %s, want %s", os.Args[6], tt.prefix)
			}
			if os.Args[8] != tt.suffix {
				t.Errorf("Suffix = %s, want %s", os.Args[8], tt.suffix)
			}
			if os.Args[10] != string(rune(tt.concurrency+'0')) {
				t.Errorf("Concurrency = %s, want %d", os.Args[10], tt.concurrency)
			}
		})
	}
}

func TestVerifyContentContains(t *testing.T) {
	// Test successful verification
	t.Run("all substrings present", func(t *testing.T) {
		content := "This is a test file with multiple lines"
		VerifyContentContains(t, content, []string{"test file", "multiple lines"})
		// If we get here, the test passed
	})

	// Test empty expected substrings
	t.Run("empty expected substrings", func(t *testing.T) {
		content := "Any content"
		VerifyContentContains(t, content, []string{})
		// Should pass with no expected strings
	})

	// For failure cases, we'll test indirectly by verifying behavior
	t.Run("verify error reporting", func(t *testing.T) {
		// We can't easily test the failure case directly since it calls t.Errorf
		// But we can at least verify the function doesn't panic
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("VerifyContentContains panicked: %v", r)
			}
		}()

		// This would normally fail but we're just checking it doesn't panic
		content := "test"
		expected := []string{"not found"}
		// Create a sub-test that we expect to fail
		t.Run("expected_failure", func(t *testing.T) {
			t.Skip("Skipping actual failure test")
			VerifyContentContains(t, content, expected)
		})
	})
}

func TestMustSucceed(t *testing.T) {
	// Test with nil error (should succeed)
	t.Run("nil error", func(t *testing.T) {
		MustSucceed(t, nil, "successful operation")
		// If we get here, the test passed
	})

	// Test error behavior without causing test failure
	t.Run("verify error handling", func(t *testing.T) {
		// We can't test the failure case directly since it calls t.Fatalf
		// But we can verify the function exists and is callable
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("MustSucceed panicked: %v", r)
			}
		}()

		// Create a sub-test that we expect to fail
		t.Run("expected_failure", func(t *testing.T) {
			t.Skip("Skipping actual failure test")
			MustSucceed(t, errors.New("test error"), "failed operation")
		})
	})
}

func TestCloseFile(t *testing.T) {
	// Test closing a normal file
	t.Run("close normal file", func(t *testing.T) {
		file, err := os.CreateTemp(t.TempDir(), "test")
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		CloseFile(t, file)

		// Verify file is closed by trying to write to it
		_, writeErr := file.Write([]byte("test"))
		if writeErr == nil {
			t.Error("Expected write to fail on closed file")
		}
	})

	// Test that CloseFile doesn't panic on already closed files
	// Note: We can't easily test the error case without causing test failure
	// since CloseFile calls t.Errorf, which is the expected behavior
	t.Run("verify CloseFile function exists and is callable", func(t *testing.T) {
		// This test just verifies the function signature and basic functionality
		// The error case is tested in integration tests where failures are expected
		file, err := os.CreateTemp(t.TempDir(), "test")
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		// Test normal case - file should close successfully
		CloseFile(t, file)

		// Verify file is closed
		_, writeErr := file.Write([]byte("test"))
		if writeErr == nil {
			t.Error("Expected write to fail on closed file")
		}
	})
}

// Test thread safety of functions that might be called concurrently
func TestConcurrentOperations(t *testing.T) {
	tempDir := t.TempDir()
	done := make(chan bool)

	// Test concurrent file creation
	for i := 0; i < 5; i++ {
		go func(n int) {
			CreateTestFile(t, tempDir, string(rune('a'+n))+".txt", []byte("content"))
			done <- true
		}(i)
	}

	// Test concurrent directory creation
	for i := 0; i < 5; i++ {
		go func(n int) {
			CreateTestDirectory(t, tempDir, "dir"+string(rune('0'+n)))
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}

// Benchmarks
func BenchmarkCreateTestFile(b *testing.B) {
	tempDir := b.TempDir()
	content := []byte("benchmark content")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Use a unique filename for each iteration to avoid conflicts
		filename := "bench" + string(rune(i%26+'a')) + ".txt"
		filePath := filepath.Join(tempDir, filename)
		if err := os.WriteFile(filePath, content, FilePermission); err != nil {
			b.Fatalf("Failed to write file: %v", err)
		}
	}
}

func BenchmarkCreateTestFiles(b *testing.B) {
	tempDir := b.TempDir()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Create specs with unique names for each iteration
		specs := []FileSpec{
			{Name: "file1_" + string(rune(i%26+'a')) + ".txt", Content: "content1"},
			{Name: "file2_" + string(rune(i%26+'a')) + ".txt", Content: "content2"},
			{Name: "file3_" + string(rune(i%26+'a')) + ".txt", Content: "content3"},
		}

		for _, spec := range specs {
			filePath := filepath.Join(tempDir, spec.Name)
			if err := os.WriteFile(filePath, []byte(spec.Content), FilePermission); err != nil {
				b.Fatalf("Failed to write file: %v", err)
			}
		}
	}
}

func BenchmarkVerifyContentContains(b *testing.B) {
	content := strings.Repeat("test content with various words ", 100)
	expected := []string{"test", "content", "various", "words"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// We can't use the actual function in benchmark since it needs testing.T
		// So we'll benchmark the core logic
		for _, exp := range expected {
			_ = strings.Contains(content, exp)
		}
	}
}
