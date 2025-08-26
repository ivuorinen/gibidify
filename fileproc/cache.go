// Package fileproc handles file processing, collection, and output formatting.
package fileproc

// getNormalizedExtension efficiently extracts and normalizes the file extension with caching.
func (r *FileTypeRegistry) getNormalizedExtension(filename string) string {
	// Try cache first (read lock)
	r.cacheMutex.RLock()
	if ext, exists := r.extCache[filename]; exists {
		r.cacheMutex.RUnlock()

		return ext
	}
	r.cacheMutex.RUnlock()

	// Compute normalized extension
	ext := normalizeExtension(filename)

	// Cache the result (write lock)
	r.cacheMutex.Lock()
	// Check cache size and clean if needed
	if len(r.extCache) >= r.maxCacheSize*2 {
		r.clearExtCache()
		r.stats.CacheEvictions++
	}
	r.extCache[filename] = ext
	r.cacheMutex.Unlock()

	return ext
}

// getFileTypeResult gets cached file type detection result or computes it.
func (r *FileTypeRegistry) getFileTypeResult(filename string) FileTypeResult {
	ext := r.getNormalizedExtension(filename)

	// Update statistics
	r.updateStats(func() {
		r.stats.TotalLookups++
	})

	// Try cache first (read lock)
	r.cacheMutex.RLock()
	if result, exists := r.resultCache[ext]; exists {
		r.cacheMutex.RUnlock()
		r.updateStats(func() {
			r.stats.CacheHits++
		})

		return result
	}
	r.cacheMutex.RUnlock()

	// Cache miss
	r.updateStats(func() {
		r.stats.CacheMisses++
	})

	// Compute result
	result := FileTypeResult{
		Extension: ext,
		IsImage:   r.imageExts[ext],
		IsBinary:  r.binaryExts[ext],
		Language:  r.languageMap[ext],
	}

	// Handle special cases for binary detection (like .DS_Store)
	if !result.IsBinary && isSpecialFile(filename, r.binaryExts) {
		result.IsBinary = true
	}

	// Cache the result (write lock)
	r.cacheMutex.Lock()
	if len(r.resultCache) >= r.maxCacheSize {
		r.clearResultCache()
		r.stats.CacheEvictions++
	}
	r.resultCache[ext] = result
	r.cacheMutex.Unlock()

	return result
}

// clearExtCache clears half of the extension cache (LRU-like behavior).
func (r *FileTypeRegistry) clearExtCache() {
	r.clearCache(&r.extCache, r.maxCacheSize)
}

// clearResultCache clears half of the result cache.
func (r *FileTypeRegistry) clearResultCache() {
	newCache := make(map[string]FileTypeResult, r.maxCacheSize)
	count := 0
	for k, v := range r.resultCache {
		if count >= r.maxCacheSize/2 {
			break
		}
		newCache[k] = v
		count++
	}
	r.resultCache = newCache
}

// clearCache is a generic cache clearing function.
func (r *FileTypeRegistry) clearCache(cache *map[string]string, maxSize int) {
	newCache := make(map[string]string, maxSize)
	count := 0
	for k, v := range *cache {
		if count >= maxSize/2 {
			break
		}
		newCache[k] = v
		count++
	}
	*cache = newCache
}

// invalidateCache clears both caches when the registry is modified.
func (r *FileTypeRegistry) invalidateCache() {
	r.cacheMutex.Lock()
	defer r.cacheMutex.Unlock()

	r.extCache = make(map[string]string, r.maxCacheSize)
	r.resultCache = make(map[string]FileTypeResult, r.maxCacheSize)
	r.stats.CacheEvictions++
}

// updateStats safely updates statistics.
func (r *FileTypeRegistry) updateStats(fn func()) {
	r.cacheMutex.Lock()
	fn()
	r.cacheMutex.Unlock()
}
