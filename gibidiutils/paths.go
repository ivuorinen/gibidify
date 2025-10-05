// Package gibidiutils provides common utility functions for gibidify.
package gibidiutils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
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

// ValidateSourcePath validates a source directory path for security.
// It ensures the path exists, is a directory, and doesn't contain path traversal attempts.
func ValidateSourcePath(path string) error {
	if path == "" {
		return NewStructuredError(ErrorTypeValidation, CodeValidationRequired, "source path is required", "", nil)
	}

	// Check for path traversal patterns before cleaning
	if strings.Contains(path, "..") {
		return NewStructuredError(ErrorTypeValidation, CodeValidationPath, "path traversal attempt detected in source path", path, map[string]interface{}{
			"original_path": path,
		})
	}

	// Clean and get absolute path
	cleaned := filepath.Clean(path)
	abs, err := filepath.Abs(cleaned)
	if err != nil {
		return NewStructuredError(ErrorTypeFileSystem, CodeFSPathResolution, "cannot resolve source path", path, map[string]interface{}{
			"error": err.Error(),
		})
	}

	// Get current working directory to ensure we're not escaping it for relative paths
	if !filepath.IsAbs(path) {
		cwd, err := os.Getwd()
		if err != nil {
			return NewStructuredError(ErrorTypeFileSystem, CodeFSPathResolution, "cannot get current working directory", path, map[string]interface{}{
				"error": err.Error(),
			})
		}

		// Ensure the resolved path is within or below the current working directory
		cwdAbs, err := filepath.Abs(cwd)
		if err != nil {
			return NewStructuredError(ErrorTypeFileSystem, CodeFSPathResolution, "cannot resolve current working directory", path, map[string]interface{}{
				"error": err.Error(),
			})
		}

		// Check if the absolute path tries to escape the current working directory
		if !strings.HasPrefix(abs, cwdAbs) {
			return NewStructuredError(ErrorTypeValidation, CodeValidationPath, "source path attempts to access directories outside current working directory", path, map[string]interface{}{
				"resolved_path": abs,
				"working_dir":   cwdAbs,
			})
		}
	}

	// Check if path exists and is a directory
	info, err := os.Stat(cleaned)
	if err != nil {
		if os.IsNotExist(err) {
			return NewStructuredError(ErrorTypeFileSystem, CodeFSNotFound, "source directory does not exist", path, nil)
		}
		return NewStructuredError(ErrorTypeFileSystem, CodeFSAccess, "cannot access source directory", path, map[string]interface{}{
			"error": err.Error(),
		})
	}

	if !info.IsDir() {
		return NewStructuredError(ErrorTypeValidation, CodeValidationPath, "source path must be a directory", path, map[string]interface{}{
			"is_file": true,
		})
	}

	return nil
}

// ValidateDestinationPath validates a destination file path for security.
// It ensures the path doesn't contain path traversal attempts and the parent directory exists.
func ValidateDestinationPath(path string) error {
	if path == "" {
		return NewStructuredError(ErrorTypeValidation, CodeValidationRequired, "destination path is required", "", nil)
	}

	// Check for path traversal patterns before cleaning
	if strings.Contains(path, "..") {
		return NewStructuredError(ErrorTypeValidation, CodeValidationPath, "path traversal attempt detected in destination path", path, map[string]interface{}{
			"original_path": path,
		})
	}

	// Clean and validate the path
	cleaned := filepath.Clean(path)

	// Get absolute path to ensure it's not trying to escape current working directory
	abs, err := filepath.Abs(cleaned)
	if err != nil {
		return NewStructuredError(ErrorTypeFileSystem, CodeFSPathResolution, "cannot resolve destination path", path, map[string]interface{}{
			"error": err.Error(),
		})
	}

	// Ensure the destination is not a directory
	if info, err := os.Stat(abs); err == nil && info.IsDir() {
		return NewStructuredError(ErrorTypeValidation, CodeValidationPath, "destination cannot be a directory", path, map[string]interface{}{
			"is_directory": true,
		})
	}

	// Check if parent directory exists and is writable
	parentDir := filepath.Dir(abs)
	if parentInfo, err := os.Stat(parentDir); err != nil {
		if os.IsNotExist(err) {
			return NewStructuredError(ErrorTypeFileSystem, CodeFSNotFound, "destination parent directory does not exist", path, map[string]interface{}{
				"parent_dir": parentDir,
			})
		}
		return NewStructuredError(ErrorTypeFileSystem, CodeFSAccess, "cannot access destination parent directory", path, map[string]interface{}{
			"parent_dir": parentDir,
			"error":      err.Error(),
		})
	} else if !parentInfo.IsDir() {
		return NewStructuredError(ErrorTypeValidation, CodeValidationPath, "destination parent is not a directory", path, map[string]interface{}{
			"parent_dir": parentDir,
		})
	}

	return nil
}

// ValidateConfigPath validates a configuration file path for security.
// It ensures the path doesn't contain path traversal attempts.
func ValidateConfigPath(path string) error {
	if path == "" {
		return nil // Empty path is allowed for config
	}

	// Check for path traversal patterns before cleaning
	if strings.Contains(path, "..") {
		return NewStructuredError(ErrorTypeValidation, CodeValidationPath, "path traversal attempt detected in config path", path, map[string]interface{}{
			"original_path": path,
		})
	}

	return nil
}
