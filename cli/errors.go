// Package cli provides command-line interface utilities for gibidify.
package cli

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/ivuorinen/gibidify/gibidiutils"
)

// ErrorFormatter handles CLI-friendly error formatting with suggestions.
type ErrorFormatter struct {
	ui *UIManager
}

// NewErrorFormatter creates a new error formatter.
func NewErrorFormatter(ui *UIManager) *ErrorFormatter {
	return &ErrorFormatter{ui: ui}
}

// FormatError formats an error with context and suggestions.
func (ef *ErrorFormatter) FormatError(err error) {
	if err == nil {
		return
	}

	// Handle structured errors
	var structErr *gibidiutils.StructuredError
	if errors.As(err, &structErr) {
		ef.formatStructuredError(structErr)
		return
	}

	// Handle common error types
	ef.formatGenericError(err)
}

// formatStructuredError formats a structured error with context and suggestions.
func (ef *ErrorFormatter) formatStructuredError(err *gibidiutils.StructuredError) {
	// Print main error
	ef.ui.PrintError("Error: %s", err.Message)

	// Print error type and code
	if err.Type != gibidiutils.ErrorTypeUnknown || err.Code != "" {
		ef.ui.PrintInfo("Type: %s, Code: %s", err.Type.String(), err.Code)
	}

	// Print file path if available
	if err.FilePath != "" {
		ef.ui.PrintInfo("File: %s", err.FilePath)
	}

	// Print context if available
	if len(err.Context) > 0 {
		ef.ui.PrintInfo("Context:")
		for key, value := range err.Context {
			ef.ui.printf("  %s: %v\n", key, value)
		}
	}

	// Provide suggestions based on error type
	ef.provideSuggestions(err)
}

// formatGenericError formats a generic error.
func (ef *ErrorFormatter) formatGenericError(err error) {
	ef.ui.PrintError("Error: %s", err.Error())
	ef.provideGenericSuggestions(err)
}

// provideSuggestions provides helpful suggestions based on the error.
func (ef *ErrorFormatter) provideSuggestions(err *gibidiutils.StructuredError) {
	switch err.Type {
	case gibidiutils.ErrorTypeFileSystem:
		ef.provideFileSystemSuggestions(err)
	case gibidiutils.ErrorTypeValidation:
		ef.provideValidationSuggestions(err)
	case gibidiutils.ErrorTypeProcessing:
		ef.provideProcessingSuggestions(err)
	case gibidiutils.ErrorTypeIO:
		ef.provideIOSuggestions(err)
	default:
		ef.provideDefaultSuggestions()
	}
}

// provideFileSystemSuggestions provides suggestions for file system errors.
func (ef *ErrorFormatter) provideFileSystemSuggestions(err *gibidiutils.StructuredError) {
	filePath := err.FilePath

	ef.ui.PrintWarning("Suggestions:")

	switch err.Code {
	case gibidiutils.CodeFSAccess:
		ef.suggestFileAccess(filePath)
	case gibidiutils.CodeFSPathResolution:
		ef.suggestPathResolution(filePath)
	case gibidiutils.CodeFSNotFound:
		ef.suggestFileNotFound(filePath)
	default:
		ef.suggestFileSystemGeneral(filePath)
	}
}

// provideValidationSuggestions provides suggestions for validation errors.
func (ef *ErrorFormatter) provideValidationSuggestions(err *gibidiutils.StructuredError) {
	ef.ui.PrintWarning("Suggestions:")

	switch err.Code {
	case gibidiutils.CodeValidationFormat:
		ef.ui.printf("  %s Use a supported format: markdown, json, yaml\n", gibidiutils.IconBullet)
		ef.ui.printf("  %s Example: -format markdown\n", gibidiutils.IconBullet)
	case gibidiutils.CodeValidationSize:
		ef.ui.printf("  %s Increase file size limit in config.yaml\n", gibidiutils.IconBullet)
		ef.ui.printf("  %s Use smaller files or exclude large files\n", gibidiutils.IconBullet)
	default:
		ef.ui.printf("  %s Check your command line arguments\n", gibidiutils.IconBullet)
		ef.ui.printf("  %s Run with --help for usage information\n", gibidiutils.IconBullet)
	}
}

// provideProcessingSuggestions provides suggestions for processing errors.
func (ef *ErrorFormatter) provideProcessingSuggestions(err *gibidiutils.StructuredError) {
	ef.ui.PrintWarning("Suggestions:")

	switch err.Code {
	case gibidiutils.CodeProcessingCollection:
		ef.ui.printf("  %s Check if the source directory exists and is readable\n", gibidiutils.IconBullet)
		ef.ui.printf("  %s Verify directory permissions\n", gibidiutils.IconBullet)
	case gibidiutils.CodeProcessingFileRead:
		ef.ui.printf("  %s Check file permissions\n", gibidiutils.IconBullet)
		ef.ui.printf("  %s Verify the file is not corrupted\n", gibidiutils.IconBullet)
	default:
		ef.ui.printf("  %s Try reducing concurrency: -concurrency 1\n", gibidiutils.IconBullet)
		ef.ui.printf("  %s Check available system resources\n", gibidiutils.IconBullet)
	}
}

// provideIOSuggestions provides suggestions for I/O errors.
func (ef *ErrorFormatter) provideIOSuggestions(err *gibidiutils.StructuredError) {
	ef.ui.PrintWarning("Suggestions:")

	switch err.Code {
	case gibidiutils.CodeIOFileCreate:
		ef.ui.printf("  %s Check if the destination directory exists\n", gibidiutils.IconBullet)
		ef.ui.printf("  %s Verify write permissions for the output file\n", gibidiutils.IconBullet)
		ef.ui.printf("  %s Ensure sufficient disk space\n", gibidiutils.IconBullet)
	case gibidiutils.CodeIOWrite:
		ef.ui.printf("  %s Check available disk space\n", gibidiutils.IconBullet)
		ef.ui.printf("  %s Verify write permissions\n", gibidiutils.IconBullet)
	default:
		ef.ui.printf("  %s Check file/directory permissions\n", gibidiutils.IconBullet)
		ef.ui.printf("  %s Verify available disk space\n", gibidiutils.IconBullet)
	}
}

// Helper methods for specific suggestions
func (ef *ErrorFormatter) suggestFileAccess(filePath string) {
	ef.ui.printf("  %s Check if the path exists: %s\n", gibidiutils.IconBullet, filePath)
	ef.ui.printf("  %s Verify read permissions\n", gibidiutils.IconBullet)
	if filePath != "" {
		if stat, err := os.Stat(filePath); err == nil {
			ef.ui.printf("  %s Path exists but may not be accessible\n", gibidiutils.IconBullet)
			ef.ui.printf("  %s Mode: %s\n", gibidiutils.IconBullet, stat.Mode())
		}
	}
}

func (ef *ErrorFormatter) suggestPathResolution(filePath string) {
	ef.ui.printf("  %s Use an absolute path instead of relative\n", gibidiutils.IconBullet)
	if filePath != "" {
		if abs, err := filepath.Abs(filePath); err == nil {
			ef.ui.printf("  %s Try: %s\n", gibidiutils.IconBullet, abs)
		}
	}
}

func (ef *ErrorFormatter) suggestFileNotFound(filePath string) {
	ef.ui.printf("  %s Check if the file/directory exists: %s\n", gibidiutils.IconBullet, filePath)
	if filePath != "" {
		dir := filepath.Dir(filePath)
		if entries, err := os.ReadDir(dir); err == nil {
			ef.ui.printf("  %s Similar files in %s:\n", gibidiutils.IconBullet, dir)
			count := 0
			for _, entry := range entries {
				if count >= 3 {
					break
				}
				if strings.Contains(entry.Name(), filepath.Base(filePath)) {
					ef.ui.printf("    %s %s\n", gibidiutils.IconBullet, entry.Name())
					count++
				}
			}
		}
	}
}

func (ef *ErrorFormatter) suggestFileSystemGeneral(filePath string) {
	ef.ui.printf("  %s Check file/directory permissions\n", gibidiutils.IconBullet)
	ef.ui.printf("  %s Verify the path is correct\n", gibidiutils.IconBullet)
	if filePath != "" {
		ef.ui.printf("  %s Path: %s\n", gibidiutils.IconBullet, filePath)
	}
}

// provideDefaultSuggestions provides general suggestions.
func (ef *ErrorFormatter) provideDefaultSuggestions() {
	ef.ui.printf("  %s Check your command line arguments\n", gibidiutils.IconBullet)
	ef.ui.printf("  %s Run with --help for usage information\n", gibidiutils.IconBullet)
	ef.ui.printf("  %s Try with -concurrency 1 to reduce resource usage\n", gibidiutils.IconBullet)
}

// provideGenericSuggestions provides suggestions for generic errors.
func (ef *ErrorFormatter) provideGenericSuggestions(err error) {
	errorMsg := err.Error()

	ef.ui.PrintWarning("Suggestions:")

	// Pattern matching for common errors
	switch {
	case strings.Contains(errorMsg, "permission denied"):
		ef.ui.printf("  %s Check file/directory permissions\n", gibidiutils.IconBullet)
		ef.ui.printf("  %s Try running with appropriate privileges\n", gibidiutils.IconBullet)
	case strings.Contains(errorMsg, "no such file or directory"):
		ef.ui.printf("  %s Verify the file/directory path is correct\n", gibidiutils.IconBullet)
		ef.ui.printf("  %s Check if the file exists\n", gibidiutils.IconBullet)
	case strings.Contains(errorMsg, "flag") && strings.Contains(errorMsg, "redefined"):
		ef.ui.printf("  %s This is likely a test environment issue\n", gibidiutils.IconBullet)
		ef.ui.printf("  %s Try running the command directly instead of in tests\n", gibidiutils.IconBullet)
	default:
		ef.provideDefaultSuggestions()
	}
}

// CLI-specific error types

// MissingSourceError represents a missing source directory error.
type MissingSourceError struct{}

func (e MissingSourceError) Error() string {
	return "source directory is required"
}

// NewMissingSourceError creates a new CLI missing source error with suggestions.
func NewMissingSourceError() error {
	return &MissingSourceError{}
}

// IsUserError checks if an error is a user input error that should be handled gracefully.
func IsUserError(err error) bool {
	if err == nil {
		return false
	}

	// Check for specific user error types
	var cliErr *MissingSourceError
	if errors.As(err, &cliErr) {
		return true
	}

	// Check for structured errors that are user-facing
	var structErr *gibidiutils.StructuredError
	if errors.As(err, &structErr) {
		return structErr.Type == gibidiutils.ErrorTypeValidation ||
			structErr.Code == gibidiutils.CodeValidationFormat ||
			structErr.Code == gibidiutils.CodeValidationSize
	}

	// Check error message patterns
	errMsg := err.Error()
	userErrorPatterns := []string{
		"flag",
		"usage",
		"invalid argument",
		"file not found",
		"permission denied",
	}

	for _, pattern := range userErrorPatterns {
		if strings.Contains(strings.ToLower(errMsg), pattern) {
			return true
		}
	}

	return false
}
