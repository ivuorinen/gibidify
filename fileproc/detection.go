// Package fileproc handles file processing, collection, and output formatting.
package fileproc

// IsImage checks if the file extension indicates an image file.
func (r *FileTypeRegistry) IsImage(filename string) bool {
	return r.imageExts[normalizeExtension(filename)]
}

// IsBinary checks if the file extension indicates a binary file, including
// extensionless names listed as binary (such as .DS_Store's lowercase form).
func (r *FileTypeRegistry) IsBinary(filename string) bool {
	return r.binaryExts[normalizeExtension(filename)] || isSpecialFile(filename, r.binaryExts)
}

// Language returns the language identifier for the given filename based on its extension.
func (r *FileTypeRegistry) Language(filename string) string {
	if len(filename) < minExtensionLength {
		return ""
	}

	return r.languageMap[normalizeExtension(filename)]
}

// AddImageExtension adds a new image extension to the registry.
func (r *FileTypeRegistry) AddImageExtension(ext string) {
	r.imageExts[normalizeConfiguredExtension(ext)] = true
}

// AddBinaryExtension adds a new binary extension to the registry.
func (r *FileTypeRegistry) AddBinaryExtension(ext string) {
	r.binaryExts[normalizeConfiguredExtension(ext)] = true
}

// AddLanguageMapping adds a new language mapping to the registry.
func (r *FileTypeRegistry) AddLanguageMapping(ext, language string) {
	r.languageMap[normalizeConfiguredExtension(ext)] = language
}

// DisableExtensions removes specified extensions from the registry.
func (r *FileTypeRegistry) DisableExtensions(disabledImages, disabledBinary, disabledLanguages []string) {
	for _, ext := range disabledImages {
		delete(r.imageExts, normalizeConfiguredExtension(ext))
	}
	for _, ext := range disabledBinary {
		delete(r.binaryExts, normalizeConfiguredExtension(ext))
	}
	for _, ext := range disabledLanguages {
		delete(r.languageMap, normalizeConfiguredExtension(ext))
	}
}
