package gibidiutils

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetBaseName(t *testing.T) {
	tests := []struct {
		name     string
		absPath  string
		expected string
	}{
		{
			name:     "normal path",
			absPath:  "/home/user/project",
			expected: "project",
		},
		{
			name:     "path with trailing slash",
			absPath:  "/home/user/project/",
			expected: "project",
		},
		{
			name:     "root path",
			absPath:  "/",
			expected: "/",
		},
		{
			name:     "current directory",
			absPath:  ".",
			expected: "output",
		},
		{
			name:     testEmptyPath,
			absPath:  "",
			expected: "output",
		},
		{
			name:     "file path",
			absPath:  "/home/user/file.txt",
			expected: "file.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetBaseName(tt.absPath)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValidateSourcePath(t *testing.T) {
	// Create a temp directory for testing
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "test.txt")
	require.NoError(t, os.WriteFile(tempFile, []byte("test"), 0o644))

	tests := []struct {
		name          string
		path          string
		expectedError string
	}{
		{
			name:          testEmptyPath,
			path:          "",
			expectedError: "source path is required",
		},
		{
			name:          testPathTraversalAttempt,
			path:          "../../../etc/passwd",
			expectedError: testPathTraversalDetected,
		},
		{
			name:          "path with double dots",
			path:          "/home/../etc/passwd",
			expectedError: testPathTraversalDetected,
		},
		{
			name:          "non-existent path",
			path:          "/definitely/does/not/exist",
			expectedError: "does not exist",
		},
		{
			name:          "file instead of directory",
			path:          tempFile,
			expectedError: "must be a directory",
		},
		{
			name:          "valid directory",
			path:          tempDir,
			expectedError: "",
		},
		{
			name:          "valid relative path",
			path:          ".",
			expectedError: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSourcePath(tt.path)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)

				// Check if it's a StructuredError
				var structErr *StructuredError
				if errors.As(err, &structErr) {
					assert.NotEmpty(t, structErr.Code)
					assert.NotEqual(t, ErrorTypeUnknown, structErr.Type)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateDestinationPath(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name          string
		path          string
		expectedError string
	}{
		{
			name:          testEmptyPath,
			path:          "",
			expectedError: "destination path is required",
		},
		{
			name:          testPathTraversalAttempt,
			path:          "../../etc/passwd",
			expectedError: testPathTraversalDetected,
		},
		{
			name:          "absolute path traversal",
			path:          "/home/../../../etc/passwd",
			expectedError: testPathTraversalDetected,
		},
		{
			name:          "valid new file",
			path:          filepath.Join(tempDir, "newfile.txt"),
			expectedError: "",
		},
		{
			name:          "valid relative path",
			path:          "output.txt",
			expectedError: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDestinationPath(tt.path)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateConfigPath(t *testing.T) {
	tempDir := t.TempDir()
	validConfig := filepath.Join(tempDir, "config.yaml")
	require.NoError(t, os.WriteFile(validConfig, []byte("key: value"), 0o644))

	tests := []struct {
		name          string
		path          string
		expectedError string
	}{
		{
			name:          testEmptyPath,
			path:          "",
			expectedError: "", // Empty config path is allowed
		},
		{
			name:          testPathTraversalAttempt,
			path:          "../../../etc/config.yaml",
			expectedError: testPathTraversalDetected,
		},
		// ValidateConfigPath doesn't check if file exists or is regular file
		// It only checks for path traversal
		{
			name:          "valid config file",
			path:          validConfig,
			expectedError: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateConfigPath(tt.path)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestGetAbsolutePath is already covered in paths_test.go

func TestValidationErrorTypes(t *testing.T) {
	t.Run("source path validation errors", func(t *testing.T) {
		// Test empty source
		err := ValidateSourcePath("")
		assert.Error(t, err)
		var structErrEmptyPath *StructuredError
		if errors.As(err, &structErrEmptyPath) {
			assert.Equal(t, ErrorTypeValidation, structErrEmptyPath.Type)
			assert.Equal(t, CodeValidationRequired, structErrEmptyPath.Code)
		}

		// Test path traversal
		err = ValidateSourcePath("../../../etc")
		assert.Error(t, err)
		var structErrTraversal *StructuredError
		if errors.As(err, &structErrTraversal) {
			assert.Equal(t, ErrorTypeValidation, structErrTraversal.Type)
			assert.Equal(t, CodeValidationPath, structErrTraversal.Code)
		}
	})

	t.Run("destination path validation errors", func(t *testing.T) {
		// Test empty destination
		err := ValidateDestinationPath("")
		assert.Error(t, err)
		var structErrEmptyDest *StructuredError
		if errors.As(err, &structErrEmptyDest) {
			assert.Equal(t, ErrorTypeValidation, structErrEmptyDest.Type)
			assert.Equal(t, CodeValidationRequired, structErrEmptyDest.Code)
		}
	})

	t.Run("config path validation errors", func(t *testing.T) {
		// Test path traversal in config
		err := ValidateConfigPath("../../etc/config.yaml")
		assert.Error(t, err)
		var structErrTraversalInConfig *StructuredError
		if errors.As(err, &structErrTraversalInConfig) {
			assert.Equal(t, ErrorTypeValidation, structErrTraversalInConfig.Type)
			assert.Equal(t, CodeValidationPath, structErrTraversalInConfig.Code)
		}
	})
}

func TestPathSecurityChecks(t *testing.T) {
	// Test various path traversal attempts
	traversalPaths := []string{
		"../etc/passwd",
		"../../root/.ssh/id_rsa",
		"/home/../../../etc/shadow",
		"./../../sensitive/data",
		"foo/../../../bar",
	}

	for _, path := range traversalPaths {
		t.Run("source_"+path, func(t *testing.T) {
			err := ValidateSourcePath(path)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), testPathTraversal)
		})

		t.Run("dest_"+path, func(t *testing.T) {
			err := ValidateDestinationPath(path)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), testPathTraversal)
		})

		t.Run("config_"+path, func(t *testing.T) {
			err := ValidateConfigPath(path)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), testPathTraversal)
		})
	}
}

func TestSpecialPaths(t *testing.T) {
	t.Run("GetBaseName with special paths", func(t *testing.T) {
		specialPaths := map[string]string{
			"/":   "/",
			"":    "output",
			".":   "output",
			"..":  "..",
			"/.":  "output", // filepath.Base("/.") returns "." which matches the output condition
			"/..": "..",
			"//":  "/",
			"///": "/",
		}

		for path, expected := range specialPaths {
			result := GetBaseName(path)
			assert.Equal(t, expected, result, "Path: %s", path)
		}
	})
}

func TestPathNormalization(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("source path normalization", func(t *testing.T) {
		// Create nested directory
		nestedDir := filepath.Join(tempDir, "a", "b", "c")
		require.NoError(t, os.MkdirAll(nestedDir, 0o755))

		// Test path with redundant separators
		redundantPath := tempDir + string(
			os.PathSeparator,
		) + string(
			os.PathSeparator,
		) + "a" + string(
			os.PathSeparator,
		) + "b" + string(
			os.PathSeparator,
		) + "c"
		err := ValidateSourcePath(redundantPath)
		assert.NoError(t, err)
	})
}

func TestPathValidationConcurrency(t *testing.T) {
	tempDir := t.TempDir()

	// Test concurrent path validation
	paths := []string{
		tempDir,
		".",
		"/tmp",
	}

	errChan := make(chan error, len(paths)*3)

	for _, path := range paths {
		go func(p string) {
			errChan <- ValidateSourcePath(p)
		}(path)

		go func(p string) {
			errChan <- ValidateDestinationPath(p + "/output.txt")
		}(path)
	}

	// Collect results
	for i := 0; i < len(paths)*2; i++ {
		<-errChan
	}

	// No assertions needed - test passes if no panic/race
}
