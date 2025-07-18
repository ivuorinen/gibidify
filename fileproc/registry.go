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
type FileTypeRegistry struct {
	imageExts   map[string]bool
	binaryExts  map[string]bool
	languageMap map[string]string

	// Cache for frequent lookups to avoid repeated string operations
	extCache     map[string]string         // filename -> normalized extension
	resultCache  map[string]FileTypeResult // extension -> cached result
	cacheMutex   sync.RWMutex
	maxCacheSize int

	// Performance statistics
	stats RegistryStats
}

// RegistryStats tracks performance metrics for the registry.
type RegistryStats struct {
	TotalLookups   uint64
	CacheHits      uint64
	CacheMisses    uint64
	CacheEvictions uint64
}

// FileTypeResult represents cached file type detection results.
type FileTypeResult struct {
	IsImage   bool
	IsBinary  bool
	Language  string
	Extension string
}

// initRegistry initializes the default file type registry with common extensions.
func initRegistry() *FileTypeRegistry {
	return &FileTypeRegistry{
		imageExts:    getImageExtensions(),
		binaryExts:   getBinaryExtensions(),
		languageMap:  getLanguageMap(),
		extCache:     make(map[string]string, 1000),        // Cache for extension normalization
		resultCache:  make(map[string]FileTypeResult, 500), // Cache for type results
		maxCacheSize: 500,
	}
}

// getRegistry returns the singleton file type registry, creating it if necessary.
func getRegistry() *FileTypeRegistry {
	registryOnce.Do(func() {
		registry = initRegistry()
	})
	return registry
}

// GetDefaultRegistry returns the default file type registry.
func GetDefaultRegistry() *FileTypeRegistry {
	return getRegistry()
}

// GetStats returns a copy of the current registry statistics.
func (r *FileTypeRegistry) GetStats() RegistryStats {
	r.cacheMutex.RLock()
	defer r.cacheMutex.RUnlock()
	return r.stats
}

// GetCacheInfo returns current cache size information.
func (r *FileTypeRegistry) GetCacheInfo() (extCacheSize, resultCacheSize, maxCacheSize int) {
	r.cacheMutex.RLock()
	defer r.cacheMutex.RUnlock()
	return len(r.extCache), len(r.resultCache), r.maxCacheSize
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
