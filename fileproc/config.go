// Package fileproc handles file processing, collection, and output formatting.
package fileproc

import "strings"

// normalizeConfiguredExtension trims and lowercases an extension from the
// config file. Validation accepts " .foo" after trimming, so storing it
// untrimmed registered a key no file extension can equal.
func normalizeConfiguredExtension(ext string) string {
	return strings.ToLower(strings.TrimSpace(ext))
}

// ApplyCustomExtensions applies custom extensions from configuration.
func (r *FileTypeRegistry) ApplyCustomExtensions(
	customImages, customBinary []string,
	customLanguages map[string]string,
) {
	// Add custom image extensions
	r.addExtensions(customImages, r.AddImageExtension)

	// Add custom binary extensions
	r.addExtensions(customBinary, r.AddBinaryExtension)

	// Add custom language mappings
	for ext, lang := range customLanguages {
		ext, lang = normalizeConfiguredExtension(ext), strings.TrimSpace(lang)
		if ext != "" && lang != "" {
			r.AddLanguageMapping(ext, lang)
		}
	}
}

// addExtensions is a helper to add multiple extensions.
func (r *FileTypeRegistry) addExtensions(extensions []string, adder func(string)) {
	for _, ext := range extensions {
		if ext = normalizeConfiguredExtension(ext); ext != "" {
			adder(ext)
		}
	}
}

// ConfigureFromSettings applies configuration settings to the registry.
// This function is called from main.go after config is loaded to avoid circular imports.
func ConfigureFromSettings(
	customImages, customBinary []string,
	customLanguages map[string]string,
	disabledImages, disabledBinary, disabledLanguages []string,
) {
	registry := DefaultRegistry()
	registry.ApplyCustomExtensions(customImages, customBinary, customLanguages)
	registry.DisableExtensions(disabledImages, disabledBinary, disabledLanguages)
}
