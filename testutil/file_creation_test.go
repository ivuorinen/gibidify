package testutil

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
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
			readContent, err := os.ReadFile(filePath) // #nosec G304 - test file path is controlled
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
				content, err := os.ReadFile(filePath) // #nosec G304 - test file path is controlled
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
