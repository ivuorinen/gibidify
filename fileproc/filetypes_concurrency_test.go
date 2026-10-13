package fileproc

import (
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/ivuorinen/gibidify/testutil"
)

const (
	numGoroutines             = 100
	numOperationsPerGoroutine = 100
)

// TestFileTypeRegistryConcurrentReads tests concurrent read operations.
// This test verifies thread-safety of registry reads under concurrent access.
// For race condition detection, run with: go test -race
func TestFileTypeRegistryConcurrentReads(t *testing.T) {
	var wg sync.WaitGroup
	errChan := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		wg.Go(func() {
			if err := performConcurrentReads(); err != nil {
				errChan <- err
			}
		})
	}
	wg.Wait()
	close(errChan)

	// Check for any errors from goroutines
	for err := range errChan {
		t.Errorf("Concurrent read operation failed: %v", err)
	}
}

// TestFileTypeRegistryConcurrentRegistryAccess tests concurrent registry access.
func TestFileTypeRegistryConcurrentRegistryAccess(t *testing.T) {
	// Reset the registry to test concurrent initialization
	ResetRegistryForTesting()
	t.Cleanup(func() {
		ResetRegistryForTesting()
	})

	registries := make([]*FileTypeRegistry, numGoroutines)
	var wg sync.WaitGroup

	for i := 0; i < numGoroutines; i++ {
		idx := i // capture for closure
		wg.Go(func() {
			registries[idx] = DefaultRegistry()
		})
	}
	wg.Wait()

	verifySameRegistryInstance(t, registries)
}

// performConcurrentReads performs concurrent read operations on the registry.
// Returns an error if any operation produces unexpected results.
func performConcurrentReads() error {
	registry := DefaultRegistry()

	for j := 0; j < numOperationsPerGoroutine; j++ {
		// Test various file detection operations with expected results
		if !registry.IsImage(testutil.TestFilePNG) {
			return errors.New("expected .png to be detected as image")
		}
		if !registry.IsBinary(testutil.TestFileEXE) {
			return errors.New("expected .exe to be detected as binary")
		}
		if lang := registry.Language(testutil.TestFileGo); lang != "go" {
			return fmt.Errorf("expected .go to have language 'go', got %q", lang)
		}

		// Test global functions with expected results
		if !DefaultRegistry().IsImage(testutil.TestFileImageJPG) {
			return errors.New("expected .jpg to be detected as image")
		}
		if !DefaultRegistry().IsBinary(testutil.TestFileBinaryDLL) {
			return errors.New("expected .dll to be detected as binary")
		}
		if lang := DefaultRegistry().Language(testutil.TestFileScriptPy); lang != "python" {
			return fmt.Errorf("expected .py to have language 'python', got %q", lang)
		}
	}
	return nil
}

// verifySameRegistryInstance verifies all goroutines got the same registry instance.
func verifySameRegistryInstance(t *testing.T, registries []*FileTypeRegistry) {
	t.Helper()

	firstRegistry := registries[0]
	for i := 1; i < numGoroutines; i++ {
		if registries[i] != firstRegistry {
			t.Errorf("Registry %d is different from registry 0", i)
		}
	}
}

// createConcurrencyTestRegistry creates a new registry instance for concurrency testing.
func createConcurrencyTestRegistry() *FileTypeRegistry {
	return &FileTypeRegistry{
		imageExts:   make(map[string]bool),
		binaryExts:  make(map[string]bool),
		languageMap: make(map[string]string),
	}
}

// Benchmarks for concurrency performance

// BenchmarkConcurrentReads benchmarks concurrent read operations on the registry.
func BenchmarkConcurrentReads(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = performConcurrentReads()
		}
	})
}

// BenchmarkConcurrentRegistryAccess benchmarks concurrent registry singleton access.
func BenchmarkConcurrentRegistryAccess(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = DefaultRegistry()
		}
	})
}

// BenchmarkConcurrentModifications benchmarks sequential registry modifications.
// Note: Concurrent modifications to the same registry require external synchronization.
// This benchmark measures the cost of modification operations themselves.
func BenchmarkConcurrentModifications(b *testing.B) {
	for b.Loop() {
		registry := createConcurrencyTestRegistry()
		for i := 0; i < 10; i++ {
			extSuffix := fmt.Sprintf("_bench_%d", i)
			registry.AddImageExtension(".img" + extSuffix)
			registry.AddBinaryExtension(".bin" + extSuffix)
			registry.AddLanguageMapping(".lang"+extSuffix, "lang"+extSuffix)
		}
	}
}
