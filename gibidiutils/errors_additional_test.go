package gibidiutils

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrorTypeString(t *testing.T) {
	tests := []struct {
		name     string
		errType  ErrorType
		expected string
	}{
		{
			name:     "CLI error type",
			errType:  ErrorTypeCLI,
			expected: "CLI",
		},
		{
			name:     "FileSystem error type",
			errType:  ErrorTypeFileSystem,
			expected: "FileSystem",
		},
		{
			name:     "Processing error type",
			errType:  ErrorTypeProcessing,
			expected: "Processing",
		},
		{
			name:     "Configuration error type",
			errType:  ErrorTypeConfiguration,
			expected: "Configuration",
		},
		{
			name:     "IO error type",
			errType:  ErrorTypeIO,
			expected: "IO",
		},
		{
			name:     "Validation error type",
			errType:  ErrorTypeValidation,
			expected: "Validation",
		},
		{
			name:     "Unknown error type",
			errType:  ErrorTypeUnknown,
			expected: "Unknown",
		},
		{
			name:     "Invalid error type",
			errType:  ErrorType(999),
			expected: "Unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.errType.String()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestStructuredErrorMethods(t *testing.T) {
	t.Run("Error method", func(t *testing.T) {
		err := &StructuredError{
			Type:    ErrorTypeValidation,
			Code:    CodeValidationRequired,
			Message: "field is required",
		}
		expected := "Validation [REQUIRED]: field is required"
		assert.Equal(t, expected, err.Error())
	})

	t.Run("Error method with context", func(t *testing.T) {
		err := &StructuredError{
			Type:    ErrorTypeFileSystem,
			Code:    CodeFSNotFound,
			Message: testErrFileNotFound,
			Context: map[string]interface{}{
				"path": "/test/file.txt",
			},
		}
		errStr := err.Error()
		assert.Contains(t, errStr, "FileSystem")
		assert.Contains(t, errStr, "NOT_FOUND")
		assert.Contains(t, errStr, testErrFileNotFound)
		assert.Contains(t, errStr, "/test/file.txt")
		assert.Contains(t, errStr, "path")
	})

	t.Run("Unwrap method", func(t *testing.T) {
		innerErr := errors.New("inner error")
		err := &StructuredError{
			Type:    ErrorTypeIO,
			Code:    CodeIOFileWrite,
			Message: testErrWriteFailed,
			Cause:   innerErr,
		}
		assert.Equal(t, innerErr, err.Unwrap())
	})

	t.Run("Unwrap with nil cause", func(t *testing.T) {
		err := &StructuredError{
			Type:    ErrorTypeIO,
			Code:    CodeIOFileWrite,
			Message: testErrWriteFailed,
		}
		assert.Nil(t, err.Unwrap())
	})
}

func TestWithContextMethods(t *testing.T) {
	t.Run("WithContext", func(t *testing.T) {
		err := &StructuredError{
			Type:    ErrorTypeValidation,
			Code:    CodeValidationFormat,
			Message: testErrInvalidFormat,
		}

		err = err.WithContext("format", "xml")
		err = err.WithContext("expected", "json")

		assert.NotNil(t, err.Context)
		assert.Equal(t, "xml", err.Context["format"])
		assert.Equal(t, "json", err.Context["expected"])
	})

	t.Run("WithFilePath", func(t *testing.T) {
		err := &StructuredError{
			Type:    ErrorTypeFileSystem,
			Code:    CodeFSPermission,
			Message: "permission denied",
		}

		err = err.WithFilePath("/etc/passwd")

		assert.Equal(t, "/etc/passwd", err.FilePath)
	})

	t.Run("WithLine", func(t *testing.T) {
		err := &StructuredError{
			Type:    ErrorTypeProcessing,
			Code:    CodeProcessingFileRead,
			Message: "read error",
		}

		err = err.WithLine(42)

		assert.Equal(t, 42, err.Line)
	})
}

func TestNewStructuredError(t *testing.T) {
	tests := []struct {
		name     string
		errType  ErrorType
		code     string
		message  string
		filePath string
		context  map[string]interface{}
	}{
		{
			name:     "basic error",
			errType:  ErrorTypeValidation,
			code:     CodeValidationRequired,
			message:  "field is required",
			filePath: "",
			context:  nil,
		},
		{
			name:     "error with file path",
			errType:  ErrorTypeFileSystem,
			code:     CodeFSNotFound,
			message:  testErrFileNotFound,
			filePath: "/test/missing.txt",
			context:  nil,
		},
		{
			name:    "error with context",
			errType: ErrorTypeIO,
			code:    CodeIOFileWrite,
			message: testErrWriteFailed,
			context: map[string]interface{}{
				"size":  1024,
				"error": "disk full",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewStructuredError(tt.errType, tt.code, tt.message, tt.filePath, tt.context)

			assert.NotNil(t, err)
			assert.Equal(t, tt.errType, err.Type)
			assert.Equal(t, tt.code, err.Code)
			assert.Equal(t, tt.message, err.Message)
			assert.Equal(t, tt.filePath, err.FilePath)
			assert.Equal(t, tt.context, err.Context)
		})
	}
}

func TestNewStructuredErrorf(t *testing.T) {
	err := NewStructuredErrorf(
		ErrorTypeValidation,
		CodeValidationSize,
		"file size %d exceeds limit %d",
		2048, 1024,
	)

	assert.NotNil(t, err)
	assert.Equal(t, ErrorTypeValidation, err.Type)
	assert.Equal(t, CodeValidationSize, err.Code)
	assert.Equal(t, "file size 2048 exceeds limit 1024", err.Message)
}

func TestWrapError(t *testing.T) {
	innerErr := errors.New("original error")
	wrappedErr := WrapError(
		innerErr,
		ErrorTypeProcessing,
		CodeProcessingFileRead,
		"failed to process file",
	)

	assert.NotNil(t, wrappedErr)
	assert.Equal(t, ErrorTypeProcessing, wrappedErr.Type)
	assert.Equal(t, CodeProcessingFileRead, wrappedErr.Code)
	assert.Equal(t, "failed to process file", wrappedErr.Message)
	assert.Equal(t, innerErr, wrappedErr.Cause)
}

func TestWrapErrorf(t *testing.T) {
	innerErr := errors.New("original error")
	wrappedErr := WrapErrorf(
		innerErr,
		ErrorTypeIO,
		CodeIOFileCreate,
		"failed to create %s in %s",
		"output.txt", "/tmp",
	)

	assert.NotNil(t, wrappedErr)
	assert.Equal(t, ErrorTypeIO, wrappedErr.Type)
	assert.Equal(t, CodeIOFileCreate, wrappedErr.Code)
	assert.Equal(t, "failed to create output.txt in /tmp", wrappedErr.Message)
	assert.Equal(t, innerErr, wrappedErr.Cause)
}

func TestSpecificErrorConstructors(t *testing.T) {
	t.Run("NewMissingSourceError", func(t *testing.T) {
		err := NewMissingSourceError()
		assert.NotNil(t, err)
		assert.Equal(t, ErrorTypeCLI, err.Type)
		assert.Equal(t, CodeCLIMissingSource, err.Code)
		assert.Contains(t, err.Message, "source")
	})

	t.Run("NewFileSystemError", func(t *testing.T) {
		err := NewFileSystemError(CodeFSPermission, "access denied")
		assert.NotNil(t, err)
		assert.Equal(t, ErrorTypeFileSystem, err.Type)
		assert.Equal(t, CodeFSPermission, err.Code)
		assert.Equal(t, "access denied", err.Message)
	})

	t.Run("NewProcessingError", func(t *testing.T) {
		err := NewProcessingError(CodeProcessingCollection, "collection failed")
		assert.NotNil(t, err)
		assert.Equal(t, ErrorTypeProcessing, err.Type)
		assert.Equal(t, CodeProcessingCollection, err.Code)
		assert.Equal(t, "collection failed", err.Message)
	})

	t.Run("NewIOError", func(t *testing.T) {
		err := NewIOError(CodeIOFileWrite, testErrWriteFailed)
		assert.NotNil(t, err)
		assert.Equal(t, ErrorTypeIO, err.Type)
		assert.Equal(t, CodeIOFileWrite, err.Code)
		assert.Equal(t, testErrWriteFailed, err.Message)
	})

	t.Run("NewValidationError", func(t *testing.T) {
		err := NewValidationError(CodeValidationFormat, testErrInvalidFormat)
		assert.NotNil(t, err)
		assert.Equal(t, ErrorTypeValidation, err.Type)
		assert.Equal(t, CodeValidationFormat, err.Code)
		assert.Equal(t, testErrInvalidFormat, err.Message)
	})
}

// TestLogErrorf is already covered in errors_test.go

func TestStructuredErrorChaining(t *testing.T) {
	// Test method chaining
	err := NewStructuredError(
		ErrorTypeFileSystem,
		CodeFSNotFound,
		testErrFileNotFound,
		"",
		nil,
	).WithFilePath("/test.txt").WithLine(10).WithContext("operation", "read")

	assert.Equal(t, "/test.txt", err.FilePath)
	assert.Equal(t, 10, err.Line)
	assert.Equal(t, "read", err.Context["operation"])
}

func TestErrorCodes(t *testing.T) {
	// Test that all error codes are defined
	codes := []string{
		CodeCLIMissingSource,
		CodeCLIInvalidArgs,
		CodeFSPathResolution,
		CodeFSPermission,
		CodeFSNotFound,
		CodeFSAccess,
		CodeProcessingFileRead,
		CodeProcessingCollection,
		CodeProcessingTraversal,
		CodeProcessingEncode,
		CodeConfigValidation,
		CodeConfigMissing,
		CodeIOFileCreate,
		CodeIOFileWrite,
		CodeIOEncoding,
		CodeIOWrite,
		CodeIOFileRead,
		CodeIOClose,
		CodeValidationRequired,
		CodeValidationFormat,
		CodeValidationSize,
		CodeValidationPath,
		CodeResourceLimitFiles,
		CodeResourceLimitTotalSize,
		CodeResourceLimitMemory,
		CodeResourceLimitTimeout,
	}

	// All codes should be non-empty strings
	for _, code := range codes {
		assert.NotEmpty(t, code, "Error code should not be empty")
		assert.NotEqual(t, "", code, "Error code should be defined")
	}
}

func TestErrorUnwrapChain(t *testing.T) {
	// Test unwrapping through multiple levels
	innermost := errors.New("innermost error")
	middle := WrapError(innermost, ErrorTypeIO, CodeIOFileRead, "read failed")
	outer := WrapError(middle, ErrorTypeProcessing, CodeProcessingFileRead, "processing failed")

	// Test unwrapping
	assert.Equal(t, middle, outer.Unwrap())
	assert.Equal(t, innermost, middle.Unwrap())

	// innermost is a plain error, doesn't have Unwrap() method
	// No need to test it

	// Test error chain messages
	assert.Contains(t, outer.Error(), "Processing")
	assert.Contains(t, middle.Error(), "IO")
}
