// Package config handles application configuration management.
package config

import (
	"fmt"
	"strings"
)

// validateEmptyElement checks if an element in a slice is empty after trimming whitespace.
// Returns a formatted error message if empty, or empty string if valid.
func validateEmptyElement(fieldPath, value string, index int) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fmt.Sprintf("%s[%d] is empty", fieldPath, index)
	}

	return ""
}

// validateDotPrefix ensures an extension starts with a dot.
// Returns a formatted error message if missing dot prefix, or empty string if valid.
func validateDotPrefix(fieldPath, value string, index int) string {
	value = strings.TrimSpace(value)
	if !strings.HasPrefix(value, ".") {
		return fmt.Sprintf("%s[%d] (%s) must start with a dot", fieldPath, index, value)
	}

	return ""
}

// validateDotPrefixMap ensures a map key (extension) starts with a dot.
// Returns a formatted error message if missing dot prefix, or empty string if valid.
func validateDotPrefixMap(fieldPath, key string) string {
	key = strings.TrimSpace(key)
	if !strings.HasPrefix(key, ".") {
		return fmt.Sprintf("%s extension (%s) must start with a dot", fieldPath, key)
	}

	return ""
}

// validateEmptyMapValue checks if a map value is empty after trimming whitespace.
// Returns a formatted error message if empty, or empty string if valid.
func validateEmptyMapValue(fieldPath, key, value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fmt.Sprintf("%s[%s] has empty language value", fieldPath, key)
	}

	return ""
}
