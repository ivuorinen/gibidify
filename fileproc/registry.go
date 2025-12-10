// Package fileproc provides file processing utilities.
package fileproc

import (
	"path/filepath"
	"strings"
	"sync"

	"github.com/ivuorinen/gibidify/shared"
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
		extCache:     make(map[string]string, shared.FileTypeRegistryMaxCacheSize),
		resultCache:  make(map[string]FileTypeResult, shared.FileTypeRegistryMaxCacheSize),
		maxCacheSize: shared.FileTypeRegistryMaxCacheSize,
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

// Stats returns a copy of the current registry statistics.
func (r *FileTypeRegistry) Stats() RegistryStats {
	r.cacheMutex.RLock()
	defer r.cacheMutex.RUnlock()

	return r.stats
}

// CacheInfo returns current cache size information.
func (r *FileTypeRegistry) CacheInfo() (extCacheSize, resultCacheSize, maxCacheSize int) {
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
