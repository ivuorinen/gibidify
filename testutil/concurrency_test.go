package testutil

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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