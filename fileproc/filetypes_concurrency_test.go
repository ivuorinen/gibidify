package fileproc

import (
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/ivuorinen/gibidify/shared"
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

// TestFileTypeRegistryConcurrentModifications tests concurrent modifications.
func TestFileTypeRegistryConcurrentModifications(t *testing.T) {
	var wg sync.WaitGroup

	for i := 0; i < numGoroutines; i++ {
		id := i // capture for closure
		wg.Go(func() {
			performConcurrentModifications(t, id)
		})
	}
	wg.Wait()
}

// performConcurrentReads performs concurrent read operations on the registry.
// Returns an error if any operation produces unexpected results.
func performConcurrentReads() error {
	registry := DefaultRegistry()

	for j := 0; j < numOperationsPerGoroutine; j++ {
		// Test various file detection operations with expected results
		if !registry.IsImage(shared.TestFilePNG) {
			return errors.New("expected .png to be detected as image")
		}
		if !registry.IsBinary(shared.TestFileEXE) {
			return errors.New("expected .exe to be detected as binary")
		}
		if lang := registry.Language(shared.TestFileGo); lang != "go" {
			return fmt.Errorf("expected .go to have language 'go', got %q", lang)
		}

		// Test global functions with expected results
		if !IsImage(shared.TestFileImageJPG) {
			return errors.New("expected .jpg to be detected as image")
		}
		if !IsBinary(shared.TestFileBinaryDLL) {
			return errors.New("expected .dll to be detected as binary")
		}
		if lang := Language(shared.TestFileScriptPy); lang != "python" {
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

// performConcurrentModifications performs concurrent modifications on separate registry instances.
func performConcurrentModifications(t *testing.T, id int) {
	t.Helper()

	// Create a new registry instance for this goroutine
	registry := createConcurrencyTestRegistry()

	for j := 0; j < numOperationsPerGoroutine; j++ {
		extSuffix := fmt.Sprintf("_%d_%d", id, j)

		addTestExtensions(registry, extSuffix)
		verifyTestExtensions(t, registry, extSuffix)
	}
}

// createConcurrencyTestRegistry creates a new registry instance for concurrency testing.
func createConcurrencyTestRegistry() *FileTypeRegistry {
	return &FileTypeRegistry{
		imageExts:    make(map[string]bool),
		binaryExts:   make(map[string]bool),
		languageMap:  make(map[string]string),
		extCache:     make(map[string]string, shared.FileTypeRegistryMaxCacheSize),
		resultCache:  make(map[string]FileTypeResult, shared.FileTypeRegistryMaxCacheSize),
		maxCacheSize: shared.FileTypeRegistryMaxCacheSize,
	}
}

// addTestExtensions adds test extensions to the registry.
func addTestExtensions(registry *FileTypeRegistry, extSuffix string) {
	registry.AddImageExtension(".img" + extSuffix)
	registry.AddBinaryExtension(".bin" + extSuffix)
	registry.AddLanguageMapping(".lang"+extSuffix, "lang"+extSuffix)
}

// verifyTestExtensions verifies that test extensions were added correctly.
func verifyTestExtensions(t *testing.T, registry *FileTypeRegistry, extSuffix string) {
	t.Helper()

	if !registry.IsImage("test.img" + extSuffix) {
		t.Errorf("Failed to add image extension .img%s", extSuffix)
	}
	if !registry.IsBinary("test.bin" + extSuffix) {
		t.Errorf("Failed to add binary extension .bin%s", extSuffix)
	}
	if registry.Language("test.lang"+extSuffix) != "lang"+extSuffix {
		t.Errorf("Failed to add language mapping .lang%s", extSuffix)
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
