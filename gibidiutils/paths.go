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

// checkPathTraversal checks for path traversal patterns and returns an error if found.
func checkPathTraversal(path, context string) error {
	// Normalize separators without cleaning (to preserve ..)
	normalized := filepath.ToSlash(path)

	// Split into components
	components := strings.Split(normalized, "/")

	// Check each component for exact ".." match
	for _, component := range components {
		if component == ".." {
			return NewStructuredError(
				ErrorTypeValidation,
				CodeValidationPath,
				fmt.Sprintf("path traversal attempt detected in %s", context),
				path,
				map[string]interface{}{
					"original_path": path,
				},
			)
		}
	}
	return nil
}

// cleanAndResolveAbsPath cleans a path and resolves it to an absolute path.
func cleanAndResolveAbsPath(path, context string) (string, error) {
	cleaned := filepath.Clean(path)
	abs, err := filepath.Abs(cleaned)
	if err != nil {
		return "", NewStructuredError(
			ErrorTypeFileSystem,
			CodeFSPathResolution,
			fmt.Sprintf("cannot resolve %s", context),
			path,
			map[string]interface{}{
				"error": err.Error(),
			},
		)
	}
	return abs, nil
}

// validateWorkingDirectoryBoundary checks if the given absolute path escapes the working directory.
func validateWorkingDirectoryBoundary(abs, path string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return NewStructuredError(
			ErrorTypeFileSystem,
			CodeFSPathResolution,
			"cannot get current working directory",
			path,
			map[string]interface{}{
				"error": err.Error(),
			},
		)
	}

	cwdAbs, err := filepath.Abs(cwd)
	if err != nil {
		return NewStructuredError(
			ErrorTypeFileSystem,
			CodeFSPathResolution,
			"cannot resolve current working directory",
			path,
			map[string]interface{}{
				"error": err.Error(),
			},
		)
	}

	// Use filepath.Rel to check if path escapes working directory
	rel, err := filepath.Rel(cwdAbs, abs)
	if err != nil {
		return NewStructuredError(
			ErrorTypeValidation,
			CodeValidationPath,
			"cannot determine relative path",
			path,
			map[string]interface{}{
				"resolved_path": abs,
				"working_dir":   cwdAbs,
				"error":         err.Error(),
			},
		)
	}

	// Check if the relative path tries to escape the working directory
	if rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return NewStructuredError(
			ErrorTypeValidation,
			CodeValidationPath,
			"source path attempts to access directories outside current working directory",
			path,
			map[string]interface{}{
				"resolved_path": abs,
				"working_dir":   cwdAbs,
				"relative_path": rel,
			},
		)
	}

	return nil
}

// ValidateSourcePath validates a source directory path for security.
// It ensures the path exists, is a directory, and doesn't contain path traversal attempts.
//
//revive:disable-next-line:function-length
func ValidateSourcePath(path string) error {
	if path == "" {
		return NewValidationError(CodeValidationRequired, "source path is required")
	}

	// Check for path traversal patterns before cleaning
	if err := checkPathTraversal(path, "source path"); err != nil {
		return err
	}

	// Clean and get absolute path
	abs, err := cleanAndResolveAbsPath(path, "source path")
	if err != nil {
		return err
	}
	cleaned := filepath.Clean(path)

	// Ensure the resolved path is within or below the current working directory for relative paths
	if !filepath.IsAbs(path) {
		if err := validateWorkingDirectoryBoundary(abs, path); err != nil {
			return err
		}
	}

	// Check if path exists and is a directory
	info, err := os.Stat(cleaned)
	if err != nil {
		if os.IsNotExist(err) {
			return NewFileSystemError(CodeFSNotFound, "source directory does not exist").WithFilePath(path)
		}
		return NewStructuredError(
			ErrorTypeFileSystem,
			CodeFSAccess,
			"cannot access source directory",
			path,
			map[string]interface{}{
				"error": err.Error(),
			},
		)
	}

	if !info.IsDir() {
		return NewStructuredError(
			ErrorTypeValidation,
			CodeValidationPath,
			"source path must be a directory",
			path,
			map[string]interface{}{
				"is_file": true,
			},
		)
	}

	return nil
}

// ValidateDestinationPath validates a destination file path for security.
// It ensures the path doesn't contain path traversal attempts and the parent directory exists.
func ValidateDestinationPath(path string) error {
	if path == "" {
		return NewValidationError(CodeValidationRequired, "destination path is required")
	}

	// Check for path traversal patterns before cleaning
	if err := checkPathTraversal(path, "destination path"); err != nil {
		return err
	}

	// Get absolute path to ensure it's not trying to escape current working directory
	abs, err := cleanAndResolveAbsPath(path, "destination path")
	if err != nil {
		return err
	}

	// Ensure the destination is not a directory
	if info, err := os.Stat(abs); err == nil && info.IsDir() {
		return NewStructuredError(
			ErrorTypeValidation,
			CodeValidationPath,
			"destination cannot be a directory",
			path,
			map[string]interface{}{
				"is_directory": true,
			},
		)
	}

	// Check if parent directory exists and is writable
	parentDir := filepath.Dir(abs)
	if parentInfo, err := os.Stat(parentDir); err != nil {
		if os.IsNotExist(err) {
			return NewStructuredError(
				ErrorTypeFileSystem,
				CodeFSNotFound,
				"destination parent directory does not exist",
				path,
				map[string]interface{}{
					"parent_dir": parentDir,
				},
			)
		}
		return NewStructuredError(
			ErrorTypeFileSystem,
			CodeFSAccess,
			"cannot access destination parent directory",
			path,
			map[string]interface{}{
				"parent_dir": parentDir,
				"error":      err.Error(),
			},
		)
	} else if !parentInfo.IsDir() {
		return NewStructuredError(
			ErrorTypeValidation,
			CodeValidationPath,
			"destination parent is not a directory",
			path,
			map[string]interface{}{
				"parent_dir": parentDir,
			},
		)
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
	return checkPathTraversal(path, "config path")
}
