// Package fileproc provides file processing utilities.
package fileproc

import (
	"path/filepath"
	"strings"
	"sync"
)

const minExtensionLength = 2

var (
	registry     *FileTypeRegistry
	registryOnce sync.Once
)

// FileTypeRegistry manages file type detection and classification.
// Its maps are populated once during ConfigureFromSettings (before any worker
// starts) and only read thereafter, so lookups need no locking.
type FileTypeRegistry struct {
	imageExts   map[string]bool
	binaryExts  map[string]bool
	languageMap map[string]string
}

// initRegistry initializes the default file type registry with common extensions.
func initRegistry() *FileTypeRegistry {
	return &FileTypeRegistry{
		imageExts:   getImageExtensions(),
		binaryExts:  getBinaryExtensions(),
		languageMap: getLanguageMap(),
	}
}

// DefaultRegistry returns the process-wide file type registry, creating it on
// first use.
func DefaultRegistry() *FileTypeRegistry {
	registryOnce.Do(func() {
		registry = initRegistry()
	})

	return registry
}

// ResetRegistryForTesting resets the registry to its initial state.
// This function should only be used in tests.
func ResetRegistryForTesting() {
	registryOnce = sync.Once{}
	registry = nil
}

// normalizeExtension extracts and normalizes the file extension.
func normalizeExtension(filename string) string {
	return strings.ToLower(filepath.Ext(filename))
}

// isSpecialFile checks if the filename matches special cases like .DS_Store.
func isSpecialFile(filename string, extensions map[string]bool) bool {
	if filepath.Ext(filename) == "" {
		basename := strings.ToLower(filepath.Base(filename))

		return extensions[basename]
	}

	return false
}
