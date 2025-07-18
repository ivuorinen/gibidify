package fileproc

import "strings"

// ApplyCustomExtensions applies custom extensions from configuration.
func (r *FileTypeRegistry) ApplyCustomExtensions(customImages, customBinary []string, customLanguages map[string]string) {
	// Add custom image extensions
	r.addExtensions(customImages, r.AddImageExtension)

	// Add custom binary extensions
	r.addExtensions(customBinary, r.AddBinaryExtension)

	// Add custom language mappings
	for ext, lang := range customLanguages {
		if ext != "" && lang != "" {
			r.AddLanguageMapping(strings.ToLower(ext), lang)
		}
	}
}

// addExtensions is a helper to add multiple extensions.
func (r *FileTypeRegistry) addExtensions(extensions []string, adder func(string)) {
	for _, ext := range extensions {
		if ext != "" {
			adder(strings.ToLower(ext))
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
	registry := GetDefaultRegistry()
	registry.ApplyCustomExtensions(customImages, customBinary, customLanguages)
	registry.DisableExtensions(disabledImages, disabledBinary, disabledLanguages)
}
