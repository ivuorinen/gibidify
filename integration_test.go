package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestIntegrationFullCLI simulates a full run of the CLI application using adaptive concurrency.
func TestIntegrationFullCLI(t *testing.T) {
	// Create a temporary source directory and populate it with test files.
	srcDir, err := ioutil.TempDir("", "gibidify_src")
	if err != nil {
		t.Fatalf("Failed to create temp source directory: %v", err)
	}
	defer os.RemoveAll(srcDir)

	// Create two test files.
	file1 := filepath.Join(srcDir, "file1.txt")
	if err := ioutil.WriteFile(file1, []byte("Hello World"), 0644); err != nil {
		t.Fatalf("Failed to write file1: %v", err)
	}
	file2 := filepath.Join(srcDir, "file2.go")
	if err := ioutil.WriteFile(file2, []byte("package main\nfunc main() {}"), 0644); err != nil {
		t.Fatalf("Failed to write file2: %v", err)
	}

	// Create a temporary output file.
	outFile, err := ioutil.TempFile("", "gibidify_output.txt")
	if err != nil {
		t.Fatalf("Failed to create temp output file: %v", err)
	}
	outFilePath := outFile.Name()
	outFile.Close()
	defer os.Remove(outFilePath)

	// Set up CLI arguments.
	os.Args = []string{
		"gibidify",
		"-source", srcDir,
		"-destination", outFilePath,
		"-prefix", "PREFIX",
		"-suffix", "SUFFIX",
		"-concurrency", "2", // For testing, set concurrency to 2.
	}

	// Run the application with a background context.
	ctx := context.Background()
	if err := Run(ctx); err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	// Verify the output file contains the expected prefix, file contents, and suffix.
	data, err := ioutil.ReadFile(outFilePath)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}
	output := string(data)
	if !strings.Contains(output, "PREFIX") {
		t.Error("Output missing prefix")
	}
	if !strings.Contains(output, "Hello World") {
		t.Error("Output missing content from file1.txt")
	}
	if !strings.Contains(output, "SUFFIX") {
		t.Error("Output missing suffix")
	}
}

// TestIntegrationCancellation verifies that the application correctly cancels processing when the context times out.
func TestIntegrationCancellation(t *testing.T) {
	// Create a temporary source directory with many files to simulate a long-running process.
	srcDir, err := ioutil.TempDir("", "gibidify_src_long")
	if err != nil {
		t.Fatalf("Failed to create temp source directory: %v", err)
	}
	defer os.RemoveAll(srcDir)

	// Create a large number of small files.
	for i := 0; i < 1000; i++ {
		filePath := filepath.Join(srcDir, fmt.Sprintf("file%d.txt", i))
		if err := ioutil.WriteFile(filePath, []byte("Content"), 0644); err != nil {
			t.Fatalf("Failed to write %s: %v", filePath, err)
		}
	}

	// Create a temporary output file.
	outFile, err := ioutil.TempFile("", "gibidify_output.txt")
	if err != nil {
		t.Fatalf("Failed to create temp output file: %v", err)
	}
	outFilePath := outFile.Name()
	outFile.Close()
	defer os.Remove(outFilePath)

	// Set up CLI arguments.
	os.Args = []string{
		"gibidify",
		"-source", srcDir,
		"-destination", outFilePath,
		"-prefix", "PREFIX",
		"-suffix", "SUFFIX",
		"-concurrency", "2",
	}

	// Create a context with a very short timeout to force cancellation.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	// Run the application; we expect an error due to cancellation.
	err = Run(ctx)
	if err == nil {
		t.Error("Expected Run to fail due to cancellation, but it succeeded")
	}
}
