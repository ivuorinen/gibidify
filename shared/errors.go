// Package shared provides common utility functions.
package shared

import (
	"errors"
	"fmt"
)

// ErrorType represents the category of error.
type ErrorType int

const (
	// ErrorTypeUnknown represents an unknown error type.
	ErrorTypeUnknown ErrorType = iota
	// ErrorTypeCLI represents command-line interface errors.
	ErrorTypeCLI
	// ErrorTypeFileSystem represents file system operation errors.
	ErrorTypeFileSystem
	// ErrorTypeProcessing represents file processing errors.
	ErrorTypeProcessing
	// ErrorTypeConfiguration represents configuration errors.
	ErrorTypeConfiguration
	// ErrorTypeIO represents input/output errors.
	ErrorTypeIO
	// ErrorTypeValidation represents validation errors.
	ErrorTypeValidation
)

// String returns the string representation of the error type.
func (e ErrorType) String() string {
	switch e {
	case ErrorTypeCLI:
		return "CLI"
	case ErrorTypeFileSystem:
		return "FileSystem"
	case ErrorTypeProcessing:
		return "Processing"
	case ErrorTypeConfiguration:
		return "Configuration"
	case ErrorTypeIO:
		return "IO"
	case ErrorTypeValidation:
		return "Validation"
	default:
		return "Unknown"
	}
}

// StructuredError represents a structured error with type, code, and context.
type StructuredError struct {
	Type     ErrorType
	Code     string
	Message  string
	Cause    error
	Context  map[string]any
	FilePath string
}

// Error implements the error interface.
func (e *StructuredError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s [%s]: %s: %v", e.Type, e.Code, e.Message, e.Cause)
	}

	return fmt.Sprintf("%s [%s]: %s", e.Type, e.Code, e.Message)
}

// Unwrap returns the underlying cause error.
func (e *StructuredError) Unwrap() error {
	return e.Cause
}

// WithFilePath adds file path information to the error.
func (e *StructuredError) WithFilePath(filePath string) *StructuredError {
	e.FilePath = filePath

	return e
}

// NewStructuredError creates a new structured error.
func NewStructuredError(errorType ErrorType, code, message, filePath string, context map[string]any) *StructuredError {
	return &StructuredError{
		Type:     errorType,
		Code:     code,
		Message:  message,
		FilePath: filePath,
		Context:  context,
	}
}

// WrapError wraps an existing error with structured error information.
func WrapError(err error, errorType ErrorType, code, message string) *StructuredError {
	return &StructuredError{
		Type:    errorType,
		Code:    code,
		Message: message,
		Cause:   err,
	}
}

// Common error codes for each type.
const (
	// CodeFSPathResolution FileSystem Error Codes.
	CodeFSPathResolution = "PATH_RESOLUTION"
	CodeFSNotFound       = "NOT_FOUND"
	CodeFSAccess         = "ACCESS_DENIED"

	// CodeProcessingFileRead Processing Error Codes.
	CodeProcessingFileRead   = "FILE_READ"
	CodeProcessingCollection = "COLLECTION"
	CodeProcessingTraversal  = "TRAVERSAL"
	CodeProcessingEncode     = "ENCODE"

	// CodeConfigValidation Configuration Error Codes.
	CodeConfigValidation = "VALIDATION"

	// CodeIOFileCreate IO Error Codes.
	CodeIOFileCreate = "FILE_CREATE"
	CodeIOWrite      = "WRITE"

	// Validation Error Codes.
	CodeValidationFormat   = "FORMAT"
	CodeValidationSize     = "SIZE_LIMIT"
	CodeValidationRequired = "REQUIRED"
	CodeValidationPath     = "PATH_TRAVERSAL"

	// Resource Limit Error Codes.
	CodeResourceLimitFiles     = "FILE_COUNT_LIMIT"
	CodeResourceLimitTotalSize = "TOTAL_SIZE_LIMIT"
)

// LogError logs an error with a consistent format if the error is not nil.
// The operation parameter describes what was being attempted.
// Additional context can be provided via the args parameter.
func LogError(operation string, err error, args ...any) {
	if err != nil {
		msg := operation
		if len(args) > 0 {
			// Format the operation string with the provided arguments
			msg = fmt.Sprintf(operation, args...)
		}

		logger := GetLogger()
		// Check if it's a structured error and log with additional context
		structErr := &StructuredError{}
		if errors.As(err, &structErr) {
			fields := map[string]any{
				"error_type": structErr.Type.String(),
				"error_code": structErr.Code,
				"context":    structErr.Context,
				"file_path":  structErr.FilePath,
			}
			logger.WithFields(fields).Errorf(ErrorFmtWithCause, msg, err)
		} else {
			logger.Errorf(ErrorFmtWithCause, msg, err)
		}
	}
}

// LogErrorf logs an error with a formatted message if the error is not nil.
// This is a convenience wrapper around LogError for cases where formatting is needed.
func LogErrorf(err error, format string, args ...any) {
	if err != nil {
		LogError(format, err, args...)
	}
}

// Test error variables.
var (
	// ErrTestError is a generic test error.
	ErrTestError = errors.New(TestErrTestErrorMsg)
)
