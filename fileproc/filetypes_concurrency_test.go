package fileproc

import (
	"fmt"
	"sync"
	"testing"
)

const (
	numGoroutines             = 100
	numOperationsPerGoroutine = 100
)

// TestFileTypeRegistry_ConcurrentReads tests concurrent read operations.
func TestFileTypeRegistry_ConcurrentReads(_ *testing.T) {
	var wg sync.WaitGroup

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			performConcurrentReads()
		}()
	}
	wg.Wait()
}

// TestFileTypeRegistry_ConcurrentRegistryAccess tests concurrent registry access.
func TestFileTypeRegistry_ConcurrentRegistryAccess(t *testing.T) {
	// Reset the registry to test concurrent initialization
	ResetRegistryForTesting()

	registries := make([]*FileTypeRegistry, numGoroutines)
	var wg sync.WaitGroup

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			registries[id] = GetDefaultRegistry()
		}(i)
	}
	wg.Wait()

	verifySameRegistryInstance(t, registries)
}

// TestFileTypeRegistry_ConcurrentModifications tests concurrent modifications.
func TestFileTypeRegistry_ConcurrentModifications(t *testing.T) {
	var wg sync.WaitGroup

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			performConcurrentModifications(t, id)
		}(i)
	}
	wg.Wait()
}

// performConcurrentReads performs concurrent read operations on the registry.
func performConcurrentReads() {
	registry := GetDefaultRegistry()

	for j := 0; j < numOperationsPerGoroutine; j++ {
		// Test various file detection operations
		_ = registry.IsImage("test.png")
		_ = registry.IsBinary("test.exe")
		_ = registry.GetLanguage("test.go")

		// Test global functions too
		_ = IsImage("image.jpg")
		_ = IsBinary("binary.dll")
		_ = GetLanguage("script.py")
	}
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
		imageExts:   make(map[string]bool),
		binaryExts:  make(map[string]bool),
		languageMap: make(map[string]string),
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
	if registry.GetLanguage("test.lang"+extSuffix) != "lang"+extSuffix {
		t.Errorf("Failed to add language mapping .lang%s", extSuffix)
	}
}
