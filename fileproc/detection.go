package fileproc

import "strings"

// Package-level detection functions

// IsImage checks if the file extension indicates an image file.
func IsImage(filename string) bool {
	return getRegistry().IsImage(filename)
}

// IsBinary checks if the file extension indicates a binary file.
func IsBinary(filename string) bool {
	return getRegistry().IsBinary(filename)
}

// GetLanguage returns the language identifier for the given filename based on its extension.
func GetLanguage(filename string) string {
	return getRegistry().GetLanguage(filename)
}

// Registry methods for detection

// IsImage checks if the file extension indicates an image file.
func (r *FileTypeRegistry) IsImage(filename string) bool {
	result := r.getFileTypeResult(filename)
	return result.IsImage
}

// IsBinary checks if the file extension indicates a binary file.
func (r *FileTypeRegistry) IsBinary(filename string) bool {
	result := r.getFileTypeResult(filename)
	return result.IsBinary
}

// GetLanguage returns the language identifier for the given filename based on its extension.
func (r *FileTypeRegistry) GetLanguage(filename string) string {
	if len(filename) < minExtensionLength {
		return ""
	}
	result := r.getFileTypeResult(filename)
	return result.Language
}

// Extension management methods

// AddImageExtension adds a new image extension to the registry.
func (r *FileTypeRegistry) AddImageExtension(ext string) {
	r.addExtension(ext, r.imageExts)
}

// AddBinaryExtension adds a new binary extension to the registry.
func (r *FileTypeRegistry) AddBinaryExtension(ext string) {
	r.addExtension(ext, r.binaryExts)
}

// AddLanguageMapping adds a new language mapping to the registry.
func (r *FileTypeRegistry) AddLanguageMapping(ext, language string) {
	r.languageMap[strings.ToLower(ext)] = language
	r.invalidateCache()
}

// addExtension is a helper to add extensions to a map.
func (r *FileTypeRegistry) addExtension(ext string, target map[string]bool) {
	target[strings.ToLower(ext)] = true
	r.invalidateCache()
}

// removeExtension is a helper to remove extensions from a map.
func (r *FileTypeRegistry) removeExtension(ext string, target map[string]bool) {
	delete(target, strings.ToLower(ext))
}

// DisableExtensions removes specified extensions from the registry.
func (r *FileTypeRegistry) DisableExtensions(disabledImages, disabledBinary, disabledLanguages []string) {
	// Disable image extensions
	for _, ext := range disabledImages {
		if ext != "" {
			r.removeExtension(ext, r.imageExts)
		}
	}

	// Disable binary extensions
	for _, ext := range disabledBinary {
		if ext != "" {
			r.removeExtension(ext, r.binaryExts)
		}
	}

	// Disable language extensions
	for _, ext := range disabledLanguages {
		if ext != "" {
			delete(r.languageMap, strings.ToLower(ext))
		}
	}

	// Invalidate cache after all modifications
	r.invalidateCache()
}
