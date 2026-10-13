package testutil

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

// TestConcurrentOperations creates files and directories from concurrent
// goroutines and checks every one exists afterwards. The goroutines report
// errors instead of calling the t.Fatalf-based helpers: FailNow must run on the
// test goroutine, and from another one it exited only that goroutine, so a
// failure hung the test until the go test timeout.
func TestConcurrentOperations(t *testing.T) {
	tempDir := t.TempDir()
	const perKind = 5
	var paths []string
	for i := range perKind {
		paths = append(paths,
			filepath.Join(tempDir, fmt.Sprintf("%c.txt", 'a'+i)),
			filepath.Join(tempDir, fmt.Sprintf("dir%d", i)))
	}

	errs := make(chan error, len(paths))
	var wg sync.WaitGroup
	for _, p := range paths {
		wg.Go(func() {
			if strings.HasSuffix(p, ".txt") {
				errs <- os.WriteFile(p, []byte("content"), TestFilePermission)
			} else {
				errs <- os.Mkdir(p, TestDirPermission)
			}
		})
	}
	wg.Wait()
	close(errs)

	for err := range errs {
		if err != nil {
			t.Errorf("concurrent create: %v", err)
		}
	}
	for _, p := range paths {
		if _, err := os.Stat(p); err != nil {
			t.Errorf("expected %s to exist: %v", p, err)
		}
	}
}

// Benchmarks.
func BenchmarkCreateTestFile(b *testing.B) {
	tempDir := b.TempDir()
	content := []byte("benchmark content")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Use a unique filename for each iteration to avoid conflicts
		filename := "bench" + string(rune(i%26+'a')) + ".txt"
		filePath := filepath.Join(tempDir, filename)
		if err := os.WriteFile(filePath, content, TestFilePermission); err != nil {
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
			if err := os.WriteFile(filePath, []byte(spec.Content), TestFilePermission); err != nil {
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
