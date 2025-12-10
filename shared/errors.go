<<<<<<<< HEAD:gibidiutils/errors.go
// Package gibidiutils provides common utility functions for gibidify.
package gibidiutils
|||||||| parent of e21c976 (refactor: rename utils to shared and deduplicate code):utils/errors.go
// Package utils provides common utility functions.
package utils
========
// Package shared provides common utility functions.
package shared
>>>>>>>> e21c976 (refactor: rename utils to shared and deduplicate code):shared/errors.go

import (
	"errors"
	"fmt"
<<<<<<<< HEAD:gibidiutils/errors.go
	"sort"
	"strings"

	"github.com/sirupsen/logrus"
|||||||| parent of e21c976 (refactor: rename utils to shared and deduplicate code):utils/errors.go

	"github.com/sirupsen/logrus"
========
>>>>>>>> e21c976 (refactor: rename utils to shared and deduplicate code):shared/errors.go
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

// Error formatting templates.
const (
	errorFormatWithCause = "%s: %v"
)

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
	base := fmt.Sprintf("%s [%s]: %s", e.Type, e.Code, e.Message)
	if len(e.Context) > 0 {
		// Sort keys for deterministic output
		keys := make([]string, 0, len(e.Context))
		for k := range e.Context {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		ctxPairs := make([]string, 0, len(e.Context))
		for _, k := range keys {
			ctxPairs = append(ctxPairs, fmt.Sprintf("%s=%v", k, e.Context[k]))
		}
		base = fmt.Sprintf("%s | context: %s", base, strings.Join(ctxPairs, ", "))
	}
<<<<<<<< HEAD:gibidiutils/errors.go
	if e.Cause != nil {
		return fmt.Sprintf(errorFormatWithCause, base, e.Cause)
	}
	return base
|||||||| parent of e21c976 (refactor: rename utils to shared and deduplicate code):utils/errors.go
	return fmt.Sprintf("%s [%s]: %s", e.Type, e.Code, e.Message)
========

	return fmt.Sprintf("%s [%s]: %s", e.Type, e.Code, e.Message)
>>>>>>>> e21c976 (refactor: rename utils to shared and deduplicate code):shared/errors.go
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
<<<<<<<< HEAD:gibidiutils/errors.go
func NewStructuredError(
	errorType ErrorType,
	code, message, filePath string,
	context map[string]any,
) *StructuredError {
|||||||| parent of e21c976 (refactor: rename utils to shared and deduplicate code):utils/errors.go
func NewStructuredError(errorType ErrorType, code, message, filePath string, context map[string]interface{}) *StructuredError {
========
func NewStructuredError(errorType ErrorType, code, message, filePath string, context map[string]any) *StructuredError {
>>>>>>>> e21c976 (refactor: rename utils to shared and deduplicate code):shared/errors.go
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

// Common error codes for each type.
const (
<<<<<<<< HEAD:gibidiutils/errors.go
	// CLI Error Codes

|||||||| parent of e21c976 (refactor: rename utils to shared and deduplicate code):utils/errors.go
	// CLI Error Codes
========
	// CodeCLIMissingSource CLI Error Codes.
>>>>>>>> e21c976 (refactor: rename utils to shared and deduplicate code):shared/errors.go
	CodeCLIMissingSource = "MISSING_SOURCE"
	CodeCLIInvalidArgs   = "INVALID_ARGS"

<<<<<<<< HEAD:gibidiutils/errors.go
	// FileSystem Error Codes

|||||||| parent of e21c976 (refactor: rename utils to shared and deduplicate code):utils/errors.go
	// FileSystem Error Codes
========
	// CodeFSPathResolution FileSystem Error Codes.
>>>>>>>> e21c976 (refactor: rename utils to shared and deduplicate code):shared/errors.go
	CodeFSPathResolution = "PATH_RESOLUTION"
	CodeFSPermission     = "PERMISSION_DENIED"
	CodeFSNotFound       = "NOT_FOUND"
	CodeFSAccess         = "ACCESS_DENIED"

<<<<<<<< HEAD:gibidiutils/errors.go
	// Processing Error Codes

|||||||| parent of e21c976 (refactor: rename utils to shared and deduplicate code):utils/errors.go
	// Processing Error Codes
========
	// CodeProcessingFileRead Processing Error Codes.
>>>>>>>> e21c976 (refactor: rename utils to shared and deduplicate code):shared/errors.go
	CodeProcessingFileRead   = "FILE_READ"
	CodeProcessingCollection = "COLLECTION"
	CodeProcessingTraversal  = "TRAVERSAL"
	CodeProcessingEncode     = "ENCODE"

<<<<<<<< HEAD:gibidiutils/errors.go
	// Configuration Error Codes

|||||||| parent of e21c976 (refactor: rename utils to shared and deduplicate code):utils/errors.go
	// Configuration Error Codes
========
	// CodeConfigValidation Configuration Error Codes.
>>>>>>>> e21c976 (refactor: rename utils to shared and deduplicate code):shared/errors.go
	CodeConfigValidation = "VALIDATION"
	CodeConfigMissing    = "MISSING"

<<<<<<<< HEAD:gibidiutils/errors.go
	// IO Error Codes

|||||||| parent of e21c976 (refactor: rename utils to shared and deduplicate code):utils/errors.go
	// IO Error Codes
========
	// CodeIOFileCreate IO Error Codes.
>>>>>>>> e21c976 (refactor: rename utils to shared and deduplicate code):shared/errors.go
	CodeIOFileCreate = "FILE_CREATE"
	CodeIOFileWrite  = "FILE_WRITE"
	CodeIOEncoding   = "ENCODING"
	CodeIOWrite      = "WRITE"
	CodeIOFileRead   = "FILE_READ"
	CodeIOClose      = "CLOSE"

<<<<<<<< HEAD:gibidiutils/errors.go
	// Validation Error Codes

|||||||| parent of e21c976 (refactor: rename utils to shared and deduplicate code):utils/errors.go
	// Validation Error Codes
========
	// Validation Error Codes.
>>>>>>>> e21c976 (refactor: rename utils to shared and deduplicate code):shared/errors.go
	CodeValidationFormat   = "FORMAT"
	CodeValidationFileType = "FILE_TYPE"
	CodeValidationSize     = "SIZE_LIMIT"
	CodeValidationRequired = "REQUIRED"
	CodeValidationPath     = "PATH_TRAVERSAL"

<<<<<<<< HEAD:gibidiutils/errors.go
	// Resource Limit Error Codes

|||||||| parent of e21c976 (refactor: rename utils to shared and deduplicate code):utils/errors.go
	// Resource Limit Error Codes
========
	// Resource Limit Error Codes.
>>>>>>>> e21c976 (refactor: rename utils to shared and deduplicate code):shared/errors.go
	CodeResourceLimitFiles       = "FILE_COUNT_LIMIT"
	CodeResourceLimitTotalSize   = "TOTAL_SIZE_LIMIT"
	CodeResourceLimitTimeout     = "TIMEOUT"
	CodeResourceLimitMemory      = "MEMORY_LIMIT"
	CodeResourceLimitConcurrency = "CONCURRENCY_LIMIT"
	CodeResourceLimitRate        = "RATE_LIMIT"
)

// Predefined error constructors for common error scenarios

<<<<<<<< HEAD:gibidiutils/errors.go
// NewMissingSourceError creates a CLI error for missing source argument.
func NewMissingSourceError() *StructuredError {
	return NewStructuredError(
		ErrorTypeCLI,
		CodeCLIMissingSource,
		"usage: gibidify -source <source_directory> "+
			"[--destination <output_file>] [--format=json|yaml|markdown]",
		"",
		nil,
	)
|||||||| parent of e21c976 (refactor: rename utils to shared and deduplicate code):utils/errors.go
// NewCLIMissingSourceError creates a CLI error for missing source argument.
func NewCLIMissingSourceError() *StructuredError {
	return NewStructuredError(ErrorTypeCLI, CodeCLIMissingSource, "usage: gibidify -source <source_directory> [--destination <output_file>] [--format=json|yaml|markdown]", "", nil)
========
// NewMissingSourceError creates a CLI error for missing source argument.
func NewMissingSourceError() *StructuredError {
	return NewStructuredError(
		ErrorTypeCLI,
		CodeCLIMissingSource,
		"usage: gibidify -source <source_directory> [--destination <output_file>] "+
			"[--format=json|yaml|markdown (default: json)]",
		"",
		nil,
	)
>>>>>>>> e21c976 (refactor: rename utils to shared and deduplicate code):shared/errors.go
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

		logger := GetLogger()
		// Check if it's a structured error and log with additional context
<<<<<<<< HEAD:gibidiutils/errors.go
		var structErr *StructuredError
		if errors.As(err, &structErr) {
			logrus.WithFields(logrus.Fields{
|||||||| parent of e21c976 (refactor: rename utils to shared and deduplicate code):utils/errors.go
		if structErr, ok := err.(*StructuredError); ok {
			logrus.WithFields(logrus.Fields{
========
		structErr := &StructuredError{}
		if errors.As(err, &structErr) {
			fields := map[string]any{
>>>>>>>> e21c976 (refactor: rename utils to shared and deduplicate code):shared/errors.go
				"error_type": structErr.Type.String(),
				"error_code": structErr.Code,
				"context":    structErr.Context,
				"file_path":  structErr.FilePath,
				"line":       structErr.Line,
<<<<<<<< HEAD:gibidiutils/errors.go
			}).Errorf(errorFormatWithCause, msg, err)
|||||||| parent of e21c976 (refactor: rename utils to shared and deduplicate code):utils/errors.go
			}).Errorf("%s: %v", msg, err)
========
			}
			logger.WithFields(fields).Errorf(ErrorFmtWithCause, msg, err)
>>>>>>>> e21c976 (refactor: rename utils to shared and deduplicate code):shared/errors.go
		} else {
<<<<<<<< HEAD:gibidiutils/errors.go
			// Log regular errors without structured fields
			logrus.Errorf(errorFormatWithCause, msg, err)
|||||||| parent of e21c976 (refactor: rename utils to shared and deduplicate code):utils/errors.go
			logrus.Errorf("%s: %v", msg, err)
========
			logger.Errorf(ErrorFmtWithCause, msg, err)
>>>>>>>> e21c976 (refactor: rename utils to shared and deduplicate code):shared/errors.go
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
