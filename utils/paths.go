// Package utils provides common utility functions.
package utils

import (
	"fmt"
	"path/filepath"
)

// GetAbsolutePath returns the absolute path for the given path.
// It wraps filepath.Abs with consistent error handling.
func GetAbsolutePath(path string) (string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path for %s: %w", path, err)
	}
	return abs, nil
}

// GetBaseName returns the base name for the given path, handling special cases.
func GetBaseName(absPath string) string {
	baseName := filepath.Base(absPath)
	if baseName == "." || baseName == "" {
		return "output"
	}
	return baseName
}
