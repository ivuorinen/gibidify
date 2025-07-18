package fileproc_test

import (
	"path/filepath"
	"testing"

	"github.com/spf13/viper"

	"github.com/ivuorinen/gibidify/fileproc"
	"github.com/ivuorinen/gibidify/testutil"
)

func TestProdWalkerWithIgnore(t *testing.T) {
	// Create a temporary directory structure.
	rootDir := t.TempDir()

	subDir := testutil.CreateTestDirectory(t, rootDir, "vendor")

	// Write sample files
	testutil.CreateTestFiles(t, rootDir, []testutil.FileSpec{
		{Name: "file1.go", Content: "content"},
		{Name: "file2.txt", Content: "content"},
	})
	testutil.CreateTestFile(t, subDir, "file_in_vendor.txt", []byte("content")) // should be ignored

	// .gitignore that ignores *.txt and itself
	gitignoreContent := `*.txt
.gitignore
`
	testutil.CreateTestFile(t, rootDir, ".gitignore", []byte(gitignoreContent))

	// Initialize config to ignore "vendor" directory
	testutil.ResetViperConfig(t, "")
	viper.Set("ignoreDirectories", []string{"vendor"})

	// Run walker
	w := fileproc.NewProdWalker()
	found, err := w.Walk(rootDir)
	testutil.MustSucceed(t, err, "walking directory")

	// We expect only file1.go to appear
	if len(found) != 1 {
		t.Errorf("Expected 1 file to pass filters, got %d: %v", len(found), found)
	}
	if len(found) == 1 && filepath.Base(found[0]) != "file1.go" {
		t.Errorf("Expected file1.go, got %s", found[0])
	}
}

func TestProdWalkerBinaryCheck(t *testing.T) {
	rootDir := t.TempDir()

	// Create test files
	testutil.CreateTestFiles(t, rootDir, []testutil.FileSpec{
		{Name: "somefile.exe", Content: "fake-binary-content"},
		{Name: "keep.go", Content: "package main"},
	})

	// Reset and load default config
	testutil.ResetViperConfig(t, "")

	// Reset FileTypeRegistry to ensure clean state
	fileproc.ResetRegistryForTesting()

	// Run walker
	w := fileproc.NewProdWalker()
	found, err := w.Walk(rootDir)
	testutil.MustSucceed(t, err, "walking directory")

	// Only "keep.go" should be returned
	if len(found) != 1 {
		t.Errorf("Expected 1 file, got %d: %v", len(found), found)
	}
	if len(found) == 1 && filepath.Base(found[0]) != "keep.go" {
		t.Errorf("Expected keep.go in results, got %s", found[0])
	}
}

func TestProdWalkerSizeLimit(t *testing.T) {
	rootDir := t.TempDir()

	// Create test files
	largeFileData := make([]byte, 6*1024*1024) // 6 MB
	testutil.CreateTestFile(t, rootDir, "largefile.txt", largeFileData)
	testutil.CreateTestFile(t, rootDir, "smallfile.go", []byte("package main"))

	// Reset and load default config, which sets size limit to 5 MB
	testutil.ResetViperConfig(t, "")

	w := fileproc.NewProdWalker()
	found, err := w.Walk(rootDir)
	if err != nil {
		t.Fatalf("Walk returned error: %v", err)
	}

	// We should only get the small file
	if len(found) != 1 {
		t.Errorf("Expected 1 file under size limit, got %d", len(found))
	}
	if len(found) == 1 && filepath.Base(found[0]) != "smallfile.go" {
		t.Errorf("Expected smallfile.go, got %s", found[0])
	}
}
