// Package utils provides common utility functions.
package utils

import (
	"fmt"

	"github.com/sirupsen/logrus"
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
	Line     int
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

// WithContext adds context information to the error.
func (e *StructuredError) WithContext(key string, value any) *StructuredError {
	if e.Context == nil {
		e.Context = make(map[string]any)
	}
	e.Context[key] = value
	return e
}

// WithFilePath adds file path information to the error.
func (e *StructuredError) WithFilePath(filePath string) *StructuredError {
	e.FilePath = filePath
	return e
}

// WithLine adds line number information to the error.
func (e *StructuredError) WithLine(line int) *StructuredError {
	e.Line = line
	return e
}

// NewStructuredError creates a new structured error.
func NewStructuredError(errorType ErrorType, code, message, filePath string, context map[string]interface{}) *StructuredError {
	return &StructuredError{
		Type:     errorType,
		Code:     code,
		Message:  message,
		FilePath: filePath,
		Context:  context,
	}
}

// NewStructuredErrorf creates a new structured error with formatted message.
func NewStructuredErrorf(errorType ErrorType, code, format string, args ...any) *StructuredError {
	return &StructuredError{
		Type:    errorType,
		Code:    code,
		Message: fmt.Sprintf(format, args...),
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

// WrapErrorf wraps an existing error with formatted message.
func WrapErrorf(err error, errorType ErrorType, code, format string, args ...any) *StructuredError {
	return &StructuredError{
		Type:    errorType,
		Code:    code,
		Message: fmt.Sprintf(format, args...),
		Cause:   err,
	}
}

// Common error codes for each type
const (
	// CLI Error Codes
	CodeCLIMissingSource = "MISSING_SOURCE"
	CodeCLIInvalidArgs   = "INVALID_ARGS"

	// FileSystem Error Codes
	CodeFSPathResolution = "PATH_RESOLUTION"
	CodeFSPermission     = "PERMISSION_DENIED"
	CodeFSNotFound       = "NOT_FOUND"
	CodeFSAccess         = "ACCESS_DENIED"

	// Processing Error Codes
	CodeProcessingFileRead   = "FILE_READ"
	CodeProcessingCollection = "COLLECTION"
	CodeProcessingTraversal  = "TRAVERSAL"
	CodeProcessingEncode     = "ENCODE"

	// Configuration Error Codes
	CodeConfigValidation = "VALIDATION"
	CodeConfigMissing    = "MISSING"

	// IO Error Codes
	CodeIOFileCreate = "FILE_CREATE"
	CodeIOFileWrite  = "FILE_WRITE"
	CodeIOEncoding   = "ENCODING"
	CodeIOWrite      = "WRITE"
	CodeIORead       = "READ"
	CodeIOClose      = "CLOSE"

	// Validation Error Codes
	CodeValidationFormat   = "FORMAT"
	CodeValidationFileType = "FILE_TYPE"
	CodeValidationSize     = "SIZE_LIMIT"
	CodeValidationRequired = "REQUIRED"
	CodeValidationPath     = "PATH_TRAVERSAL"

	// Resource Limit Error Codes
	CodeResourceLimitFiles       = "FILE_COUNT_LIMIT"
	CodeResourceLimitTotalSize   = "TOTAL_SIZE_LIMIT"
	CodeResourceLimitTimeout     = "TIMEOUT"
	CodeResourceLimitMemory      = "MEMORY_LIMIT"
	CodeResourceLimitConcurrency = "CONCURRENCY_LIMIT"
	CodeResourceLimitRate        = "RATE_LIMIT"
)

// Predefined error constructors for common error scenarios

// NewCLIMissingSourceError creates a CLI error for missing source argument.
func NewCLIMissingSourceError() *StructuredError {
	return NewStructuredError(ErrorTypeCLI, CodeCLIMissingSource, "usage: gibidify -source <source_directory> [--destination <output_file>] [--format=json|yaml|markdown]", "", nil)
}

// NewFileSystemError creates a file system error.
func NewFileSystemError(code, message string) *StructuredError {
	return NewStructuredError(ErrorTypeFileSystem, code, message, "", nil)
}

// NewProcessingError creates a processing error.
func NewProcessingError(code, message string) *StructuredError {
	return NewStructuredError(ErrorTypeProcessing, code, message, "", nil)
}

// NewIOError creates an IO error.
func NewIOError(code, message string) *StructuredError {
	return NewStructuredError(ErrorTypeIO, code, message, "", nil)
}

// NewValidationError creates a validation error.
func NewValidationError(code, message string) *StructuredError {
	return NewStructuredError(ErrorTypeValidation, code, message, "", nil)
}

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

		// Check if it's a structured error and log with additional context
		if structErr, ok := err.(*StructuredError); ok {
			logrus.WithFields(logrus.Fields{
				"error_type": structErr.Type.String(),
				"error_code": structErr.Code,
				"context":    structErr.Context,
				"file_path":  structErr.FilePath,
				"line":       structErr.Line,
			}).Errorf("%s: %v", msg, err)
		} else {
			logrus.Errorf("%s: %v", msg, err)
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
