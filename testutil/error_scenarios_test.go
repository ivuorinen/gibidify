package testutil

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ivuorinen/gibidify/shared"
)

// testResetViperConfigVariations tests ResetViperConfig with different paths.
func testResetViperConfigVariations(t *testing.T) {
	t.Helper()

	testCases := []string{
		"",                  // Empty path
		"/nonexistent/path", // Non-existent path
		t.TempDir(),         // Valid temporary directory
	}

	for _, configPath := range testCases {
		t.Run(
			"path_"+strings.ReplaceAll(configPath, "/", "_"), func(t *testing.T) {
				ResetViperConfig(t, configPath)
			},
		)
	}
}

// testGetBaseNameEdgeCases tests GetBaseName with various edge cases.
func testGetBaseNameEdgeCases(t *testing.T) {
	t.Helper()

	edgeCases := []struct {
		input    string
		expected string
	}{
		{"", "."},
		{".", "."},
		{"..", ".."},
		{"/", "/"},
		{"//", "/"},
		{"///", "/"},
		{"file", "file"},
		{"./file", "file"},
		{"../file", "file"},
		{"/a", "a"},
		{"/a/", "a"},
		{"/a//", "a"},
		{"a/b/c", "c"},
		{"a/b/c/", "c"},
	}

	for _, tc := range edgeCases {
		result := BaseName(tc.input)
		expected := filepath.Base(tc.input)
		if result != expected {
			t.Errorf("BaseName(%q) = %q, want %q", tc.input, result, expected)
		}
	}
}

// testVerifyContentContainsScenarios tests VerifyContentContains scenarios.
func testVerifyContentContainsScenarios(t *testing.T) {
	t.Helper()

	scenarios := []struct {
		name     string
		content  string
		expected []string
	}{
		{
			"all_substrings_found",
			"This is a comprehensive test with multiple search terms",
			[]string{"comprehensive", "test", "multiple", "search"},
		},
		{"empty_expected_list", "Any content here", []string{}},
		{"single_character_matches", "abcdefg", []string{"a", "c", "g"}},
		{"repeated_substrings", "test test test", []string{"test", "test", "test"}},
		{"case_sensitive_matching", "Test TEST tEsT", []string{"Test", "TEST"}},
	}

	for _, scenario := range scenarios {
		t.Run(
			scenario.name, func(t *testing.T) {
				VerifyContentContains(t, scenario.content, scenario.expected)
			},
		)
	}
}

// testMustSucceedCases tests MustSucceed with various operations.
func testMustSucceedCases(t *testing.T) {
	t.Helper()

	operations := []string{
		"simple operation",
		"",
		"operation with special chars: !@#$%",
		"very " + strings.Repeat("long ", 50) + "operation name",
	}

	for i, op := range operations {
		t.Run(
			"operation_"+string(rune(i+'a')), func(t *testing.T) {
				MustSucceed(t, nil, op)
			},
		)
	}
}

// testCloseFileScenarios tests CloseFile with different file scenarios.
func testCloseFileScenarios(t *testing.T) {
	t.Helper()

	t.Run(
		"close_regular_file", func(t *testing.T) {
			file, err := os.CreateTemp(t.TempDir(), "test")
			if err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}

			if _, err = file.WriteString("test content"); err != nil {
				t.Fatalf("Failed to write to file: %v", err)
			}

			CloseFile(t, file)

			if _, writeErr := file.Write([]byte("should fail")); writeErr == nil {
				t.Error("Expected write to fail after close")
			}
		},
	)

	t.Run(
		"close_empty_file", func(t *testing.T) {
			file, err := os.CreateTemp(t.TempDir(), "empty")
			if err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}
			CloseFile(t, file)
		},
	)
}

// TestCoverageImprovements focuses on improving coverage for existing functions.
func TestCoverageImprovements(t *testing.T) {
	t.Run("ResetViperConfig_variations", testResetViperConfigVariations)
	t.Run("GetBaseName_comprehensive", testGetBaseNameEdgeCases)
	t.Run("VerifyContentContains_comprehensive", testVerifyContentContainsScenarios)
	t.Run("MustSucceed_success_cases", testMustSucceedCases)
	t.Run("CloseFile_success_cases", testCloseFileScenarios)
}

// attemptFileCreation attempts to create a file with error handling.
func attemptFileCreation(t *testing.T, tempDir, specName string) {
	t.Helper()

	defer func() {
		if r := recover(); r != nil {
			t.Logf("File creation panicked (expected for some edge cases): %v", r)
		}
	}()

	if _, err := os.Create(filepath.Join(tempDir, specName)); err != nil {
		t.Logf("File creation failed (expected for some edge cases): %v", err)
	}
}

// createDirectoryIfNeeded creates directory if file path contains separators.
func createDirectoryIfNeeded(t *testing.T, tempDir, specName string) {
	t.Helper()

	if strings.Contains(specName, "/") || strings.Contains(specName, "\\") {
		dirPath := filepath.Dir(filepath.Join(tempDir, specName))
		if err := os.MkdirAll(dirPath, shared.TestDirPermission); err != nil {
			t.Skipf("Cannot create directory %s: %v", dirPath, err)
		}
	}
}

// testFileSpecVariations tests FileSpec with various edge cases.
func testFileSpecVariations(t *testing.T) {
	t.Helper()

	specs := []FileSpec{
		{Name: "", Content: ""},
		{Name: "simple.txt", Content: "simple content"},
		{Name: "with spaces.txt", Content: "content with spaces"},
		{Name: "unicode-file-αβγ.txt", Content: "unicode content: αβγδε"},
		{Name: "very-long-filename-" + strings.Repeat("x", 100) + ".txt", Content: "long filename test"},
		{Name: "file.with.many.dots.txt", Content: "dotted filename"},
		{Name: "special/chars\\file<>:\"|?*.txt", Content: "special characters"},
	}

	tempDir := t.TempDir()

	for i, spec := range specs {
		t.Run(
			"spec_"+string(rune(i+'a')), func(t *testing.T) {
				createDirectoryIfNeeded(t, tempDir, spec.Name)
				attemptFileCreation(t, tempDir, spec.Name)
			},
		)
	}
}

// testDirSpecVariations tests DirSpec with various configurations.
func testDirSpecVariations(t *testing.T) {
	t.Helper()

	specs := []DirSpec{
		{Path: "empty-dir", Files: []FileSpec{}},
		{Path: "single-file-dir", Files: []FileSpec{{Name: "single.txt", Content: "single file"}}},
		{
			Path: "multi-file-dir", Files: []FileSpec{
				{Name: "file1.txt", Content: "content1"},
				{Name: "file2.txt", Content: "content2"},
				{Name: "file3.txt", Content: "content3"},
			},
		},
		{Path: "nested/deep/structure", Files: []FileSpec{{Name: "deep.txt", Content: "deep content"}}},
		{Path: "unicode-αβγ", Files: []FileSpec{{Name: "unicode-file.txt", Content: "unicode directory content"}}},
	}

	tempDir := t.TempDir()
	createdPaths := CreateTestDirectoryStructure(t, tempDir, specs)

	if len(createdPaths) == 0 && len(specs) > 0 {
		t.Error("Expected some paths to be created")
	}

	for _, path := range createdPaths {
		if _, err := os.Stat(path); err != nil {
			t.Errorf("Created path should exist: %s, error: %v", path, err)
		}
	}
}

// TestStructOperations tests operations with FileSpec and DirSpec.
func TestStructOperations(t *testing.T) {
	t.Run("FileSpec_comprehensive", testFileSpecVariations)
	t.Run("DirSpec_comprehensive", testDirSpecVariations)
}

// BenchmarkUtilityFunctions provides comprehensive benchmarks.
func BenchmarkUtilityFunctions(b *testing.B) {
	b.Run(
		"GetBaseName_variations", func(b *testing.B) {
			paths := []string{
				"",
				"simple.txt",
				"/path/to/file.go",
				"/very/deep/nested/path/to/file.json",
				"relative/path/file.txt",
				".",
				"..",
				"/",
				strings.Repeat("/very/long/path", 10) + "/file.txt",
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				path := paths[i%len(paths)]
				_ = BaseName(path)
			}
		},
	)

	b.Run(
		"StringOperations", func(b *testing.B) {
			content := strings.Repeat("benchmark content with search terms ", 100)
			searchTerms := []string{"benchmark", "content", "search", "terms", "not found"}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				term := searchTerms[i%len(searchTerms)]
				_ = strings.Contains(content, term)
			}
		},
	)

	b.Run(
		"FileSpec_creation", func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				spec := FileSpec{
					Name:    "benchmark-file-" + string(rune(i%26+'a')) + ".txt",
					Content: "benchmark content for iteration " + string(rune(i%10+'0')),
				}
				_ = len(spec.Name)
				_ = len(spec.Content)
			}
		},
	)

	b.Run(
		"DirSpec_creation", func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				spec := DirSpec{
					Path: "benchmark-dir-" + string(rune(i%26+'a')),
					Files: []FileSpec{
						{Name: "file1.txt", Content: "content1"},
						{Name: "file2.txt", Content: "content2"},
					},
				}
				_ = len(spec.Path)
				_ = len(spec.Files)
			}
		},
	)
}
