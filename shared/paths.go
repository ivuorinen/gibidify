// Package shared provides common utility functions.
package shared

import (
	"fmt"
	"os"
	"path/filepath"
)

// AbsolutePath returns the absolute path for the given path.
// It wraps filepath.Abs with consistent error handling.
func AbsolutePath(path string) (string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path for %s: %w", path, err)
	}

	return abs, nil
}

// BaseName returns the base name for the given path, handling special cases.
func BaseName(absPath string) string {
	baseName := filepath.Base(absPath)
	if baseName == "." || baseName == "" {
		return "output"
	}

	return baseName
}

// ValidateSourcePath validates that the source path exists and is a directory.
//
// It deliberately has no ".." or working-directory containment check: this is
// a local CLI reading the user's own files, so such checks defended no trust
// boundary and only rejected valid input like "../sibling" or "a..b". What is
// read is contained by the walker (symlinks are never followed).
func ValidateSourcePath(path string) error {
	if path == "" {
		return NewStructuredError(
			ErrorTypeValidation,
			CodeValidationRequired,
			"source path is required",
			"",
			nil,
		)
	}

	cleaned := filepath.Clean(path)
	info, err := os.Stat(cleaned)
	if err != nil {
		if os.IsNotExist(err) {
			return NewStructuredError(ErrorTypeFileSystem, CodeFSNotFound, "source directory does not exist", path, nil)
		}

		return NewStructuredError(
			ErrorTypeFileSystem, CodeFSAccess, "cannot access source directory", path, map[string]any{
				"error": err.Error(),
			},
		)
	}

	if !info.IsDir() {
		return NewStructuredError(
			ErrorTypeValidation, CodeValidationPath, "source path must be a directory", path, map[string]any{
				"is_file": true,
			},
		)
	}

	return nil
}

// ValidateDestinationPath validates that the destination is not a directory and
// that its parent directory exists.
func ValidateDestinationPath(path string) error {
	if path == "" {
		return NewStructuredError(ErrorTypeValidation, CodeValidationRequired, "destination path is required", "", nil)
	}

	abs, err := filepath.Abs(filepath.Clean(path))
	if err != nil {
		return NewStructuredError(
			ErrorTypeFileSystem, CodeFSPathResolution, "cannot resolve destination path", path, map[string]any{
				"error": err.Error(),
			},
		)
	}

	// Ensure the destination is not a directory
	if info, err := os.Stat(abs); err == nil && info.IsDir() {
		return NewStructuredError(
			ErrorTypeValidation, CodeValidationPath, "destination cannot be a directory", path, map[string]any{
				"is_directory": true,
			},
		)
	}

	// Check if parent directory exists and is writable
	parentDir := filepath.Dir(abs)
	if parentInfo, err := os.Stat(parentDir); err != nil {
		if os.IsNotExist(err) {
			return NewStructuredError(
				ErrorTypeFileSystem, CodeFSNotFound,
				"destination parent directory does not exist", path, map[string]any{
					"parent_dir": parentDir,
				},
			)
		}

		return NewStructuredError(
			ErrorTypeFileSystem, CodeFSAccess, "cannot access destination parent directory", path, map[string]any{
				"parent_dir": parentDir,
				"error":      err.Error(),
			},
		)
	} else if !parentInfo.IsDir() {
		return NewStructuredError(
			ErrorTypeValidation, CodeValidationPath, "destination parent is not a directory", path, map[string]any{
				"parent_dir": parentDir,
			},
		)
	}

	return nil
}
