package fileproc

import (
	"fmt"
	"path/filepath"
	"strings"
)

const (
	// MaxRegistryEntries is the maximum number of entries allowed in registry config slices/maps.
	MaxRegistryEntries = 1000
	// MaxExtensionLength is the maximum length for a single extension string.
	MaxExtensionLength = 100
)

// RegistryConfig holds configuration for file type registry.
// All paths must be relative without path traversal (no ".." or leading "/").
// Extensions in CustomLanguages keys must start with "." or be alphanumeric with underscore/hyphen.
type RegistryConfig struct {
	// CustomImages: file extensions to treat as images (e.g., ".svg", ".webp").
	// Must be relative paths without ".." or leading separators.
	CustomImages []string

	// CustomBinary: file extensions to treat as binary (e.g., ".bin", ".dat").
	// Must be relative paths without ".." or leading separators.
	CustomBinary []string

	// CustomLanguages: maps file extensions to language names (e.g., {".tsx": "TypeScript"}).
	// Keys must start with "." or be alphanumeric with underscore/hyphen.
	CustomLanguages map[string]string

	// DisabledImages: image extensions to disable from default registry.
	DisabledImages []string

	// DisabledBinary: binary extensions to disable from default registry.
	DisabledBinary []string

	// DisabledLanguages: language extensions to disable from default registry.
	DisabledLanguages []string
}

// Validate checks the RegistryConfig for invalid entries and enforces limits.
func (c *RegistryConfig) Validate() error {
	// Validate CustomImages
	if err := validateExtensionSlice(c.CustomImages, "CustomImages"); err != nil {
		return err
	}

	// Validate CustomBinary
	if err := validateExtensionSlice(c.CustomBinary, "CustomBinary"); err != nil {
		return err
	}

	// Validate CustomLanguages
	if len(c.CustomLanguages) > MaxRegistryEntries {
		return fmt.Errorf(
			"CustomLanguages exceeds maximum entries (%d > %d)",
			len(c.CustomLanguages),
			MaxRegistryEntries,
		)
	}
	for ext, lang := range c.CustomLanguages {
		if err := validateExtension(ext, "CustomLanguages key"); err != nil {
			return err
		}
		if len(lang) > MaxExtensionLength {
			return fmt.Errorf(
				"CustomLanguages value %q exceeds maximum length (%d > %d)",
				lang,
				len(lang),
				MaxExtensionLength,
			)
		}
	}

	// Validate Disabled slices
	if err := validateExtensionSlice(c.DisabledImages, "DisabledImages"); err != nil {
		return err
	}
	if err := validateExtensionSlice(c.DisabledBinary, "DisabledBinary"); err != nil {
		return err
	}

	return validateExtensionSlice(c.DisabledLanguages, "DisabledLanguages")
}

// validateExtensionSlice validates a slice of extensions for path safety and limits.
func validateExtensionSlice(slice []string, fieldName string) error {
	if len(slice) > MaxRegistryEntries {
		return fmt.Errorf("%s exceeds maximum entries (%d > %d)", fieldName, len(slice), MaxRegistryEntries)
	}
	for _, ext := range slice {
		if err := validateExtension(ext, fieldName); err != nil {
			return err
		}
	}
	return nil
}

// validateExtension validates a single extension for path safety.
//
//revive:disable-next-line:cyclomatic
func validateExtension(ext, context string) error {
	// Reject empty strings
	if ext == "" {
		return fmt.Errorf("%s entry cannot be empty", context)
	}

	if len(ext) > MaxExtensionLength {
		return fmt.Errorf(
			"%s entry %q exceeds maximum length (%d > %d)",
			context, ext, len(ext), MaxExtensionLength,
		)
	}

	// Reject absolute paths
	if filepath.IsAbs(ext) {
		return fmt.Errorf("%s entry %q is an absolute path (not allowed)", context, ext)
	}

	// Reject path traversal
	if strings.Contains(ext, "..") {
		return fmt.Errorf("%s entry %q contains path traversal (not allowed)", context, ext)
	}

	// For extensions, verify they start with "." or are alphanumeric
	if strings.HasPrefix(ext, ".") {
		// Reject extensions containing path separators
		if strings.ContainsRune(ext, filepath.Separator) || strings.ContainsRune(ext, '/') ||
			strings.ContainsRune(ext, '\\') {
			return fmt.Errorf("%s entry %q contains path separators (not allowed)", context, ext)
		}
		// Valid extension format
		return nil
	}

	// Check if purely alphanumeric (for bare names)
	for _, r := range ext {
		isValid := (r >= 'a' && r <= 'z') ||
			(r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') ||
			r == '_' || r == '-'
		if !isValid {
			return fmt.Errorf(
				"%s entry %q contains invalid characters (must start with '.' or be alphanumeric/_/-)",
				context,
				ext,
			)
		}
	}

	return nil
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
// It validates the configuration before applying it.
func ConfigureFromSettings(config RegistryConfig) error {
	// Validate configuration first
	if err := config.Validate(); err != nil {
		return err
	}

	registry := GetDefaultRegistry()

	// Only apply custom extensions if they are non-empty (len() for nil slices/maps is zero)
	if len(config.CustomImages) > 0 || len(config.CustomBinary) > 0 || len(config.CustomLanguages) > 0 {
		registry.ApplyCustomExtensions(config.CustomImages, config.CustomBinary, config.CustomLanguages)
	}

	// Only disable extensions if they are non-empty
	if len(config.DisabledImages) > 0 || len(config.DisabledBinary) > 0 || len(config.DisabledLanguages) > 0 {
		registry.DisableExtensions(config.DisabledImages, config.DisabledBinary, config.DisabledLanguages)
	}

	return nil
}
