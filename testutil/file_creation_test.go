package testutil

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ivuorinen/gibidify/shared"
)

func TestCreateTestFile(t *testing.T) {
	tests := createTestFileTestCases()

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				runCreateTestFileTest(t, tt.dir, tt.filename, tt.content)
			},
		)
	}
}

// createTestFileTestCases creates test cases for TestCreateTestFile.
func createTestFileTestCases() []struct {
	name     string
	dir      string
	filename string
	content  []byte
	wantErr  bool
} {
	return []struct {
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
			content:  []byte(shared.LiteralPackageMain),
			wantErr:  false,
		},
	}
}

// runCreateTestFileTest runs a single test case for CreateTestFile.
func runCreateTestFileTest(t *testing.T, dir, filename string, content []byte) {
	t.Helper()

	tempDir := t.TempDir()
	if dir == "" {
		dir = tempDir
	}

	createSubdirectoryIfNeeded(t, dir, filename)
	filePath := CreateTestFile(t, dir, filename, content)
	verifyCreatedFile(t, filePath, content)
}

// createSubdirectoryIfNeeded creates subdirectory if the filename contains a path separator.
func createSubdirectoryIfNeeded(t *testing.T, dir, filename string) {
	t.Helper()

	if strings.ContainsRune(filename, filepath.Separator) {
		subdir := filepath.Join(dir, filepath.Dir(filename))
		if err := os.MkdirAll(subdir, shared.TestDirPermission); err != nil {
			t.Fatalf("Failed to create subdirectory: %v", err)
		}
	}
}

// verifyCreatedFile verifies that the created file has correct properties.
func verifyCreatedFile(t *testing.T, filePath string, expectedContent []byte) {
	t.Helper()

	info := verifyFileExists(t, filePath)
	verifyFileType(t, info)
	verifyFilePermissions(t, info)
	verifyFileContent(t, filePath, expectedContent)
}

// verifyFileExists verifies that the file exists and returns its info.
func verifyFileExists(t *testing.T, filePath string) os.FileInfo {
	t.Helper()

	info, err := os.Stat(filePath)
	if err != nil {
		t.Fatalf("Created file does not exist: %v", err)
	}

	return info
}

// verifyFileType verifies that the file is a regular file.
func verifyFileType(t *testing.T, info os.FileInfo) {
	t.Helper()

	if !info.Mode().IsRegular() {
		t.Error("Created path is not a regular file")
	}
}

// verifyFilePermissions verifies that the file has correct permissions.
func verifyFilePermissions(t *testing.T, info os.FileInfo) {
	t.Helper()

	if info.Mode().Perm() != shared.TestFilePermission {
		t.Errorf("File permissions = %v, want %v", info.Mode().Perm(), shared.TestFilePermission)
	}
}

// verifyFileContent verifies that the file has the expected content.
func verifyFileContent(t *testing.T, filePath string, expectedContent []byte) {
	t.Helper()

	readContent, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read created file: %v", err)
	}
	if string(readContent) != string(expectedContent) {
		t.Errorf("File content = %q, want %q", readContent, expectedContent)
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
		t.Run(
			tt.name, func(t *testing.T) {
				file, path := CreateTempOutputFile(t, tt.pattern)
				defer CloseFile(t, file)

				// Verify file exists
				info, err := os.Stat(path)
				if err != nil {
					t.Fatalf("Temp file does not exist: %v", err)
				}

				// Verify it's a regular file
				if !info.Mode().IsRegular() {
					t.Error("Created path is not a regular file")
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
			},
		)
	}
}

func TestCreateTestDirectory(t *testing.T) {
	tests := createTestDirectoryTestCases()

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				runCreateTestDirectoryTest(t, tt.parent, tt.dir)
			},
		)
	}
}

// createTestDirectoryTestCases creates test cases for TestCreateTestDirectory.
func createTestDirectoryTestCases() []struct {
	name   string
	parent string
	dir    string
} {
	return []struct {
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
}

// runCreateTestDirectoryTest runs a single test case for CreateTestDirectory.
func runCreateTestDirectoryTest(t *testing.T, parent, dir string) {
	t.Helper()

	tempDir := t.TempDir()
	if parent == "" {
		parent = tempDir
	}

	parent, dir = prepareNestedDirectoryPath(t, parent, dir)
	dirPath := CreateTestDirectory(t, parent, dir)
	verifyCreatedDirectory(t, dirPath)
}

// prepareNestedDirectoryPath prepares parent and directory paths for nested directories.
func prepareNestedDirectoryPath(t *testing.T, parent, dir string) (parentPath, fullPath string) {
	t.Helper()

	if strings.Contains(dir, "/") {
		parentPath := filepath.Join(parent, filepath.Dir(dir))
		if err := os.MkdirAll(parentPath, shared.TestDirPermission); err != nil {
			t.Fatalf("Failed to create parent directory: %v", err)
		}

		return parentPath, filepath.Base(dir)
	}

	return parent, dir
}

// verifyCreatedDirectory verifies that the created directory has correct properties.
func verifyCreatedDirectory(t *testing.T, dirPath string) {
	t.Helper()

	info := verifyDirectoryExists(t, dirPath)
	verifyIsDirectory(t, info)
	verifyDirectoryPermissions(t, info)
	verifyDirectoryUsability(t, dirPath)
}

// verifyDirectoryExists verifies that the directory exists and returns its info.
func verifyDirectoryExists(t *testing.T, dirPath string) os.FileInfo {
	t.Helper()

	info, err := os.Stat(dirPath)
	if err != nil {
		t.Fatalf("Created directory does not exist: %v", err)
	}

	return info
}

// verifyIsDirectory verifies that the path is a directory.
func verifyIsDirectory(t *testing.T, info os.FileInfo) {
	t.Helper()

	if !info.IsDir() {
		t.Error("Created path is not a directory")
	}
}

// verifyDirectoryPermissions verifies that the directory has correct permissions.
func verifyDirectoryPermissions(t *testing.T, info os.FileInfo) {
	t.Helper()

	if info.Mode().Perm() != shared.TestDirPermission {
		t.Errorf("Directory permissions = %v, want %v", info.Mode().Perm(), shared.TestDirPermission)
	}
}

// verifyDirectoryUsability verifies that files can be created in the directory.
func verifyDirectoryUsability(t *testing.T, dirPath string) {
	t.Helper()

	testFile := filepath.Join(dirPath, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), shared.TestFilePermission); err != nil {
		t.Errorf("Cannot create file in directory: %v", err)
	}
}

func TestCreateTestFiles(t *testing.T) {
	tests := createTestFilesTestCases()

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				runTestFilesTest(t, tt.fileSpecs, tt.wantCount)
			},
		)
	}
}

// createTestFilesTestCases creates test cases for TestCreateTestFiles.
func createTestFilesTestCases() []struct {
	name      string
	fileSpecs []FileSpec
	wantCount int
} {
	return []struct {
		name      string
		fileSpecs []FileSpec
		wantCount int
	}{
		{
			name: "create multiple files",
			fileSpecs: []FileSpec{
				{Name: "file1.txt", Content: "content1"},
				{Name: "file2.go", Content: shared.LiteralPackageMain},
				{Name: "file3.json", Content: `{"key": "value"}`},
			},
			wantCount: 3,
		},
		{
			name: "create files with subdirectories",
			fileSpecs: []FileSpec{
				{Name: "src/main.go", Content: shared.LiteralPackageMain},
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
}

// runTestFilesTest runs a single test case for CreateTestFiles.
func runTestFilesTest(t *testing.T, fileSpecs []FileSpec, wantCount int) {
	t.Helper()

	rootDir := t.TempDir()

	createNecessarySubdirectories(t, rootDir, fileSpecs)
	createdFiles := CreateTestFiles(t, rootDir, fileSpecs)
	verifyCreatedFilesCount(t, createdFiles, wantCount)
	verifyCreatedFilesContent(t, createdFiles, fileSpecs)
}

// createNecessarySubdirectories creates subdirectories for file specs that need them.
func createNecessarySubdirectories(t *testing.T, rootDir string, fileSpecs []FileSpec) {
	t.Helper()

	for _, spec := range fileSpecs {
		if strings.Contains(spec.Name, "/") {
			subdir := filepath.Join(rootDir, filepath.Dir(spec.Name))
			if err := os.MkdirAll(subdir, shared.TestDirPermission); err != nil {
				t.Fatalf("Failed to create subdirectory: %v", err)
			}
		}
	}
}

// verifyCreatedFilesCount verifies the count of created files.
func verifyCreatedFilesCount(t *testing.T, createdFiles []string, wantCount int) {
	t.Helper()

	if len(createdFiles) != wantCount {
		t.Errorf("Created %d files, want %d", len(createdFiles), wantCount)
	}
}

// verifyCreatedFilesContent verifies the content of created files.
func verifyCreatedFilesContent(t *testing.T, createdFiles []string, fileSpecs []FileSpec) {
	t.Helper()

	for i, filePath := range createdFiles {
		content, err := os.ReadFile(filePath)
		if err != nil {
			t.Errorf("Failed to read file %s: %v", filePath, err)

			continue
		}
		if string(content) != fileSpecs[i].Content {
			t.Errorf("File %s content = %q, want %q", filePath, content, fileSpecs[i].Content)
		}
	}
}
