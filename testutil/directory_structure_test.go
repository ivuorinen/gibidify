package testutil

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ivuorinen/gibidify/shared"
)

// verifySingleDirectoryFiles verifies single directory with files test case.
func verifySingleDirectoryFiles(t *testing.T, rootDir string, _ []string) {
	t.Helper()

	srcDir := filepath.Join(rootDir, "src")
	if _, err := os.Stat(srcDir); err != nil {
		t.Errorf("Directory %s should exist", srcDir)
	}

	mainFile := filepath.Join(srcDir, shared.TestFileMainGo)
	content, err := os.ReadFile(mainFile)
	if err != nil {
		t.Errorf("Failed to read %s: %v", shared.TestFileMainGo, err)
	}
	if string(content) != shared.LiteralPackageMain {
		t.Errorf("%s content = %q, want %q", shared.TestFileMainGo, content, shared.LiteralPackageMain)
	}

	utilsFile := filepath.Join(srcDir, shared.TestFileSharedGo)
	content, err = os.ReadFile(utilsFile)
	if err != nil {
		t.Errorf("Failed to read shared.go: %v", err)
	}
	if string(content) != shared.TestSharedGoContent {
		t.Errorf("shared.go content = %q, want %q", content, shared.TestSharedGoContent)
	}
}

// verifyMultipleDirectories verifies multiple directories with nested structure.
func verifyMultipleDirectories(t *testing.T, rootDir string, _ []string) {
	t.Helper()

	expectedDirs := []string{
		filepath.Join(rootDir, "src"),
		filepath.Join(rootDir, "src", "handlers"),
		filepath.Join(rootDir, "test"),
	}
	for _, dir := range expectedDirs {
		if info, err := os.Stat(dir); err != nil {
			t.Errorf(shared.TestFmtDirectoryShouldExist, dir, err)
		} else if !info.IsDir() {
			t.Errorf(shared.TestFmtPathShouldBeDirectory, dir)
		}
	}

	handlerFile := filepath.Join(rootDir, "src", "handlers", "handler.go")
	content, err := os.ReadFile(handlerFile)
	if err != nil {
		t.Errorf("Failed to read handler.go: %v", err)
	}
	if string(content) != shared.TestContentPackageHandlers {
		t.Errorf("handler.go content = %q, want 'package handlers'", content)
	}
}

// verifyEmptyDirectory verifies directory with no files.
func verifyEmptyDirectory(t *testing.T, rootDir string, _ []string) {
	t.Helper()

	emptyDir := filepath.Join(rootDir, "empty")
	if info, err := os.Stat(emptyDir); err != nil {
		t.Errorf(shared.TestFmtDirectoryShouldExist, emptyDir, err)
	} else if !info.IsDir() {
		t.Errorf(shared.TestFmtPathShouldBeDirectory, emptyDir)
	}
}

// verifySpecialCharacters verifies directories with special characters.
func verifySpecialCharacters(t *testing.T, rootDir string, _ []string) {
	t.Helper()

	specialDir := filepath.Join(rootDir, "special-dir_123")
	if _, err := os.Stat(specialDir); err != nil {
		t.Errorf("Special directory should exist: %v", err)
	}

	spacedDir := filepath.Join(rootDir, "dir with spaces")
	if _, err := os.Stat(spacedDir); err != nil {
		t.Errorf("Spaced directory should exist: %v", err)
	}

	spacedFile := filepath.Join(spacedDir, "file with spaces.txt")
	content, err := os.ReadFile(spacedFile)
	if err != nil {
		t.Errorf("Failed to read spaced file: %v", err)
	}
	if string(content) != "spaced content" {
		t.Errorf("Spaced file content = %q, want 'spaced content'", content)
	}
}

// runCreateDirectoryTest runs a single create directory structure test.
func runCreateDirectoryTest(
	t *testing.T,
	dirSpecs []DirSpec,
	wantPaths int,
	verifyFunc func(t *testing.T, rootDir string, createdPaths []string),
) {
	t.Helper()

	rootDir := t.TempDir()
	createdPaths := CreateTestDirectoryStructure(t, rootDir, dirSpecs)

	if len(createdPaths) != wantPaths {
		t.Errorf("Created %d paths, want %d", len(createdPaths), wantPaths)
	}

	for _, path := range createdPaths {
		if _, err := os.Stat(path); err != nil {
			t.Errorf("Created path %s should exist: %v", path, err)
		}
	}

	verifyFunc(t, rootDir, createdPaths)
}

// TestCreateTestDirectoryStructure tests the CreateTestDirectoryStructure function.
func TestCreateTestDirectoryStructure(t *testing.T) {
	tests := []struct {
		name       string
		dirSpecs   []DirSpec
		wantPaths  int
		verifyFunc func(t *testing.T, rootDir string, createdPaths []string)
	}{
		{
			name: "single directory with files",
			dirSpecs: []DirSpec{
				{
					Path: "src",
					Files: []FileSpec{
						{Name: shared.TestFileMainGo, Content: shared.LiteralPackageMain},
						{Name: shared.TestFileSharedGo, Content: shared.TestSharedGoContent},
					},
				},
			},
			wantPaths:  3, // 1 directory + 2 files
			verifyFunc: verifySingleDirectoryFiles,
		},
		{
			name: "multiple directories with nested structure",
			dirSpecs: []DirSpec{
				{
					Path: "src",
					Files: []FileSpec{
						{Name: shared.TestFileMainGo, Content: shared.LiteralPackageMain},
					},
				},
				{
					Path: "src/handlers",
					Files: []FileSpec{
						{Name: "handler.go", Content: shared.TestContentPackageHandlers},
						{Name: "middleware.go", Content: "package handlers\n\ntype Middleware struct {}"},
					},
				},
				{
					Path: "test",
					Files: []FileSpec{
						{Name: "main_test.go", Content: "package main\n\nimport \"testing\""},
					},
				},
			},
			wantPaths:  7, // 3 directories + 4 files
			verifyFunc: verifyMultipleDirectories,
		},
		{
			name: "directory with no files",
			dirSpecs: []DirSpec{
				{
					Path:  "empty",
					Files: []FileSpec{},
				},
			},
			wantPaths:  1, // 1 directory only
			verifyFunc: verifyEmptyDirectory,
		},
		{
			name:      "empty directory specs",
			dirSpecs:  []DirSpec{},
			wantPaths: 0,
			verifyFunc: func(t *testing.T, _ string, _ []string) {
				t.Helper()
				// Nothing to verify for empty specs
			},
		},
		{
			name: "directories with special characters and edge cases",
			dirSpecs: []DirSpec{
				{
					Path: "special-dir_123",
					Files: []FileSpec{
						{Name: "file-with-dashes.txt", Content: "content"},
						{Name: "file_with_underscores.go", Content: "package main"},
					},
				},
				{
					Path: "dir with spaces",
					Files: []FileSpec{
						{Name: "file with spaces.txt", Content: "spaced content"},
					},
				},
			},
			wantPaths:  5, // 2 directories + 3 files
			verifyFunc: verifySpecialCharacters,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				runCreateDirectoryTest(t, tt.dirSpecs, tt.wantPaths, tt.verifyFunc)
			},
		)
	}
}

// verifyBasicDirectoryStructure verifies basic directory structure.
func verifyBasicDirectoryStructure(t *testing.T, rootDir string) {
	t.Helper()

	if !strings.Contains(rootDir, os.TempDir()) {
		t.Errorf("Root directory %s should be in temp directory", rootDir)
	}

	appDir := filepath.Join(rootDir, "app")
	if info, err := os.Stat(appDir); err != nil {
		t.Errorf("App directory should exist: %v", err)
	} else if !info.IsDir() {
		t.Error("App path should be a directory")
	}

	mainFile := filepath.Join(appDir, shared.TestFileMainGo)
	content, err := os.ReadFile(mainFile)
	if err != nil {
		t.Errorf("Failed to read %s: %v", shared.TestFileMainGo, err)
	}
	expectedMain := "package main\n\nfunc main() {}"
	if string(content) != expectedMain {
		t.Errorf("%s content = %q, want %q", shared.TestFileMainGo, content, expectedMain)
	}

	configFile := filepath.Join(appDir, shared.TestFileConfigJSON)
	content, err = os.ReadFile(configFile)
	if err != nil {
		t.Errorf("Failed to read %s: %v", shared.TestFileConfigJSON, err)
	}
	if string(content) != `{"debug": true}` {
		t.Errorf("%s content = %q, want %q", shared.TestFileConfigJSON, content, `{"debug": true}`)
	}

	docsDir := filepath.Join(rootDir, "docs")
	if info, err := os.Stat(docsDir); err != nil {
		t.Errorf("Docs directory should exist: %v", err)
	} else if !info.IsDir() {
		t.Error("Docs path should be a directory")
	}

	readmeFile := filepath.Join(docsDir, shared.TestFileReadmeMD)
	content, err = os.ReadFile(readmeFile)
	if err != nil {
		t.Errorf("Failed to read %s: %v", shared.TestFileReadmeMD, err)
	}
	if string(content) != shared.TestContentDocumentation {
		t.Errorf("%s content = %q, want '# Documentation'", shared.TestFileReadmeMD, content)
	}
}

// verifyEmptyDirectorySpecs verifies empty directory specs.
func verifyEmptyDirectorySpecs(t *testing.T, rootDir string) {
	t.Helper()

	if info, err := os.Stat(rootDir); err != nil {
		t.Errorf("Root directory should exist: %v", err)
	} else if !info.IsDir() {
		t.Error("Root path should be a directory")
	}

	entries, err := os.ReadDir(rootDir)
	if err != nil {
		t.Errorf("Failed to read root directory: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("Root directory should be empty, but has %d entries", len(entries))
	}
}

// verifyComplexNestedStructure verifies complex nested structure.
func verifyComplexNestedStructure(t *testing.T, rootDir string) {
	t.Helper()

	deepPath := filepath.Join(rootDir, "project", "internal", "handlers", "auth.go")
	content, err := os.ReadFile(deepPath)
	if err != nil {
		t.Errorf("Failed to read deep nested file: %v", err)
	}
	expectedContent := "package handlers\n\ntype AuthHandler struct{}"
	if string(content) != expectedContent {
		t.Errorf("Deep nested file content = %q, want %q", content, expectedContent)
	}

	expectedDirs := []string{
		"project",
		"project/cmd",
		"project/cmd/server",
		"project/internal",
		"project/internal/handlers",
		"project/test",
		"project/test/integration",
	}
	for _, dir := range expectedDirs {
		fullPath := filepath.Join(rootDir, dir)
		if info, err := os.Stat(fullPath); err != nil {
			t.Errorf(shared.TestFmtDirectoryShouldExist, fullPath, err)
		} else if !info.IsDir() {
			t.Errorf(shared.TestFmtPathShouldBeDirectory, fullPath)
		}
	}
}

// runSetupTempDirTest runs a single setup temp dir test.
func runSetupTempDirTest(t *testing.T, dirSpecs []DirSpec, verifyFunc func(t *testing.T, rootDir string)) {
	t.Helper()

	rootDir := SetupTempDirWithStructure(t, dirSpecs)

	if info, err := os.Stat(rootDir); err != nil {
		t.Fatalf("Root directory should exist: %v", err)
	} else if !info.IsDir() {
		t.Fatal("Root path should be a directory")
	}

	verifyFunc(t, rootDir)
}

// TestSetupTempDirWithStructure tests the SetupTempDirWithStructure function.
func TestSetupTempDirWithStructure(t *testing.T) {
	tests := []struct {
		name       string
		dirSpecs   []DirSpec
		verifyFunc func(t *testing.T, rootDir string)
	}{
		{
			name: "basic directory structure",
			dirSpecs: []DirSpec{
				{
					Path: "app",
					Files: []FileSpec{
						{Name: shared.TestFileMainGo, Content: "package main\n\nfunc main() {}"},
						{Name: shared.TestFileConfigJSON, Content: `{"debug": true}`},
					},
				},
				{
					Path: "docs",
					Files: []FileSpec{
						{Name: shared.TestFileReadmeMD, Content: shared.TestContentDocumentation},
					},
				},
			},
			verifyFunc: verifyBasicDirectoryStructure,
		},
		{
			name:       "empty directory specs",
			dirSpecs:   []DirSpec{},
			verifyFunc: verifyEmptyDirectorySpecs,
		},
		{
			name: "complex nested structure",
			dirSpecs: []DirSpec{
				{
					Path: "project",
					Files: []FileSpec{
						{Name: "go.mod", Content: "module test\n\ngo 1.21"},
					},
				},
				{
					Path: "project/cmd/server",
					Files: []FileSpec{
						{Name: shared.TestFileMainGo, Content: shared.LiteralPackageMain},
					},
				},
				{
					Path: "project/internal/handlers",
					Files: []FileSpec{
						{Name: "health.go", Content: shared.TestContentPackageHandlers},
						{Name: "auth.go", Content: "package handlers\n\ntype AuthHandler struct{}"},
					},
				},
				{
					Path: "project/test/integration",
					Files: []FileSpec{
						{Name: "server_test.go", Content: "package integration\n\nimport \"testing\""},
					},
				},
			},
			verifyFunc: verifyComplexNestedStructure,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				runSetupTempDirTest(t, tt.dirSpecs, tt.verifyFunc)
			},
		)
	}
}

// benchmarkDirectoryStructure benchmarks creation of a single directory structure.
func benchmarkDirectoryStructure(b *testing.B, dirSpecs []DirSpec) {
	b.Helper()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		rootDir := b.TempDir()
		b.StartTimer()

		for _, dirSpec := range dirSpecs {
			dirPath := filepath.Join(rootDir, dirSpec.Path)
			if err := os.MkdirAll(dirPath, shared.TestDirPermission); err != nil {
				b.Fatalf("Failed to create directory: %v", err)
			}

			for _, fileSpec := range dirSpec.Files {
				filePath := filepath.Join(dirPath, fileSpec.Name)
				if err := os.WriteFile(filePath, []byte(fileSpec.Content), shared.TestFilePermission); err != nil {
					b.Fatalf("Failed to create file: %v", err)
				}
			}
		}
	}
}

// BenchmarkDirectoryCreation benchmarks directory structure creation with different specs.
func BenchmarkDirectoryCreation(b *testing.B) {
	testCases := []struct {
		name     string
		dirSpecs []DirSpec
	}{
		{
			name: "simple_source_structure",
			dirSpecs: []DirSpec{
				{
					Path: "src",
					Files: []FileSpec{
						{Name: shared.TestFileMainGo, Content: shared.LiteralPackageMain},
						{Name: shared.TestFileSharedGo, Content: shared.TestSharedGoContent},
					},
				},
				{
					Path: "test",
					Files: []FileSpec{
						{Name: "main_test.go", Content: "package main\n\nimport \"testing\""},
					},
				},
			},
		},
		{
			name: "application_structure",
			dirSpecs: []DirSpec{
				{
					Path: "app",
					Files: []FileSpec{
						{Name: shared.TestFileMainGo, Content: shared.LiteralPackageMain},
						{Name: shared.TestFileConfigJSON, Content: `{"debug": true}`},
					},
				},
				{
					Path: "docs",
					Files: []FileSpec{
						{Name: shared.TestFileReadmeMD, Content: shared.TestContentDocumentation},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		b.Run(
			tc.name, func(b *testing.B) {
				benchmarkDirectoryStructure(b, tc.dirSpecs)
			},
		)
	}
}
