package fileproc

import (
	"fmt"
	"sync"
	"testing"
)

// TestFileTypeRegistry_ThreadSafety tests thread safety of the FileTypeRegistry.
func TestFileTypeRegistry_ThreadSafety(t *testing.T) {
	const numGoroutines = 100
	const numOperationsPerGoroutine = 100

	var wg sync.WaitGroup

	// Test concurrent read operations
	t.Run("ConcurrentReads", func(t *testing.T) {
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
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
			}(i)
		}
		wg.Wait()
	})

	// Test concurrent registry access (singleton creation)
	t.Run("ConcurrentRegistryAccess", func(t *testing.T) {
		// Reset the registry to test concurrent initialization
		// Note: This is not safe in a real application, but needed for testing
		registryOnce = sync.Once{}
		registry = nil

		registries := make([]*FileTypeRegistry, numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				registries[id] = GetDefaultRegistry()
			}(i)
		}
		wg.Wait()

		// Verify all goroutines got the same registry instance
		firstRegistry := registries[0]
		for i := 1; i < numGoroutines; i++ {
			if registries[i] != firstRegistry {
				t.Errorf("Registry %d is different from registry 0", i)
			}
		}
	})

	// Test concurrent modifications on separate registry instances
	t.Run("ConcurrentModifications", func(t *testing.T) {
		// Create separate registry instances for each goroutine to test modification thread safety
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()

				// Create a new registry instance for this goroutine
				registry := &FileTypeRegistry{
					imageExts:   make(map[string]bool),
					binaryExts:  make(map[string]bool),
					languageMap: make(map[string]string),
				}

				for j := 0; j < numOperationsPerGoroutine; j++ {
					// Add unique extensions for this goroutine
					extSuffix := fmt.Sprintf("_%d_%d", id, j)

					registry.AddImageExtension(".img" + extSuffix)
					registry.AddBinaryExtension(".bin" + extSuffix)
					registry.AddLanguageMapping(".lang"+extSuffix, "lang"+extSuffix)

					// Verify the additions worked
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
			}(i)
		}
		wg.Wait()
	})
}