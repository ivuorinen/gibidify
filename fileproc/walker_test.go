package fileproc

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ivuorinen/gibidify/config"
	"github.com/spf13/viper"
)

func TestProdWalkerWithIgnore(t *testing.T) {
	// Create a temporary directory structure.
	rootDir, err := os.MkdirTemp("", "walker_test_root")
	if err != nil {
		t.Fatalf("Failed to create temp root directory: %v", err)
	}
	defer os.RemoveAll(rootDir)

	subDir := filepath.Join(rootDir, "vendor")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subDir: %v", err)
	}

	// Write sample files
	filePaths := []string{
		filepath.Join(rootDir, "file1.go"),
		filepath.Join(rootDir, "file2.txt"),
		filepath.Join(subDir, "file_in_vendor.txt"), // should be ignored
	}
	for _, fp := range filePaths {
		if err := os.WriteFile(fp, []byte("content"), 0644); err != nil {
			t.Fatalf("Failed to write file %s: %v", fp, err)
		}
	}

	// .gitignore that ignores *.txt and itself
	gitignoreContent := `*.txt
.gitignore
`
	gitignorePath := filepath.Join(rootDir, ".gitignore")
	if err := os.WriteFile(gitignorePath, []byte(gitignoreContent), 0644); err != nil {
		t.Fatalf("Failed to write .gitignore: %v", err)
	}

	// Initialize config to ignore "vendor" directory
	viper.Reset()
	config.LoadConfig()
	viper.Set("ignoreDirectories", []string{"vendor"})

	// Run walker
	var w Walker = ProdWalker{}
	found, err := w.Walk(rootDir)
	if err != nil {
		t.Fatalf("Walk returned error: %v", err)
	}

	// We expect only file1.go to appear
	if len(found) != 1 {
		t.Errorf("Expected 1 file to pass filters, got %d: %v", len(found), found)
	}
	if len(found) == 1 && filepath.Base(found[0]) != "file1.go" {
		t.Errorf("Expected file1.go, got %s", found[0])
	}
}

func TestProdWalkerBinaryCheck(t *testing.T) {
	rootDir, err := os.MkdirTemp("", "walker_test_bincheck")
	if err != nil {
		t.Fatalf("Failed to create temp root directory: %v", err)
	}
	defer os.RemoveAll(rootDir)

	// Create a mock binary file
	binFile := filepath.Join(rootDir, "somefile.exe")
	if err := os.WriteFile(binFile, []byte("fake-binary-content"), 0644); err != nil {
		t.Fatalf("Failed to write file %s: %v", binFile, err)
	}

	// Create a normal file
	normalFile := filepath.Join(rootDir, "keep.go")
	if err := os.WriteFile(normalFile, []byte("package main"), 0644); err != nil {
		t.Fatalf("Failed to write file %s: %v", normalFile, err)
	}

	// Reset and load default config
	viper.Reset()
	config.LoadConfig()

	// Run walker
	var w Walker = ProdWalker{}
	found, err := w.Walk(rootDir)
	if err != nil {
		t.Fatalf("Walk returned error: %v", err)
	}

	// Only "keep.go" should be returned
	if len(found) != 1 {
		t.Errorf("Expected 1 file, got %d: %v", len(found), found)
	}
	if len(found) == 1 && filepath.Base(found[0]) != "keep.go" {
		t.Errorf("Expected keep.go in results, got %s", found[0])
	}
}

func TestProdWalkerSizeLimit(t *testing.T) {
	rootDir, err := os.MkdirTemp("", "walker_test_sizelimit")
	if err != nil {
		t.Fatalf("Failed to create temp root directory: %v", err)
	}
	defer os.RemoveAll(rootDir)

	// Create a file exceeding the size limit
	largeFilePath := filepath.Join(rootDir, "largefile.txt")
	largeFileData := make([]byte, 6*1024*1024) // 6 MB
	if err := os.WriteFile(largeFilePath, largeFileData, 0644); err != nil {
		t.Fatalf("Failed to write large file: %v", err)
	}

	// Create a small file
	smallFilePath := filepath.Join(rootDir, "smallfile.go")
	if err := os.WriteFile(smallFilePath, []byte("package main"), 0644); err != nil {
		t.Fatalf("Failed to write small file: %v", err)
	}

	// Reset and load default config, which sets size limit to 5 MB
	viper.Reset()
	config.LoadConfig()

	var w Walker = ProdWalker{}
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
