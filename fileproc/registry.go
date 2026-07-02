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

// FileTypeResult represents a file type detection result.
type FileTypeResult struct {
	IsImage   bool
	IsBinary  bool
	Language  string
	Extension string
}

// initRegistry initializes the default file type registry with common extensions.
func initRegistry() *FileTypeRegistry {
	return &FileTypeRegistry{
		imageExts:   getImageExtensions(),
		binaryExts:  getBinaryExtensions(),
		languageMap: getLanguageMap(),
	}
}

// getRegistry returns the singleton file type registry, creating it if necessary.
func getRegistry() *FileTypeRegistry {
	registryOnce.Do(func() {
		registry = initRegistry()
	})

	return registry
}

// DefaultRegistry returns the default file type registry.
func DefaultRegistry() *FileTypeRegistry {
	return getRegistry()
}

// getFileTypeResult detects the file type for filename from its extension.
func (r *FileTypeRegistry) getFileTypeResult(filename string) FileTypeResult {
	ext := normalizeExtension(filename)
	result := FileTypeResult{
		Extension: ext,
		IsImage:   r.imageExts[ext],
		IsBinary:  r.binaryExts[ext],
		Language:  r.languageMap[ext],
	}

	// Handle special cases for binary detection (like .DS_Store).
	if !result.IsBinary && isSpecialFile(filename, r.binaryExts) {
		result.IsBinary = true
	}

	return result
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
