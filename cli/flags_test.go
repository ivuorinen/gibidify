package cli

import (
	"errors"
	"flag"
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestParseFlags(t *testing.T) {
	// Save original command line args and restore after test
	oldArgs := os.Args
	oldFlagsParsed := flagsParsed
	defer func() {
		os.Args = oldArgs
		flagsParsed = oldFlagsParsed
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	}()

	tests := []struct {
		name          string
		args          []string
		expectedError string
		validate      func(t *testing.T, f *Flags)
		setup         func(t *testing.T)
	}{
		{
			name: "valid flags with all options",
			args: []string{
				"gibidify",
				testFlagSource, "", // will set to tempDir in test body
				"-destination", "output.md",
				"-format", "json",
				testFlagConcurrency, "4",
				"-prefix", "prefix",
				"-suffix", "suffix",
				"-no-colors",
				"-no-progress",
				"-verbose",
			},
			validate: nil, // set in test body using closure
		},
		{
			name:          "missing source directory",
			args:          []string{"gibidify"},
			expectedError: testErrSourceRequired,
		},
		{
			name: "invalid format",
			args: []string{
				"gibidify",
				testFlagSource, "", // will set to tempDir in test body
				"-format", "invalid",
			},
			expectedError: "unsupported output format: invalid",
		},
		{
			name: "invalid concurrency (zero)",
			args: []string{
				"gibidify",
				testFlagSource, "", // will set to tempDir in test body
				testFlagConcurrency, "0",
			},
			expectedError: "concurrency (0) must be at least 1",
		},
		{
			name: "invalid concurrency (too high)",
			args: []string{
				"gibidify",
				testFlagSource, "", // will set to tempDir in test body
				testFlagConcurrency, "200",
			},
			// Set maxConcurrency so the upper bound is enforced
			expectedError: "concurrency (200) exceeds maximum (128)",
			setup: func(t *testing.T) {
				orig := viper.Get("maxConcurrency")
				viper.Set("maxConcurrency", 128)
				t.Cleanup(func() { viper.Set("maxConcurrency", orig) })
			},
		},
		{
			name: "path traversal in source",
			args: []string{
				"gibidify",
				testFlagSource, testPathTraversalPath,
			},
			expectedError: testErrPathTraversal,
		},
		{
			name: "default values",
			args: []string{
				"gibidify",
				testFlagSource, "", // will set to tempDir in test body
			},
			validate: nil, // set in test body using closure
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset flags for each test
			flagsParsed = false
			globalFlags = nil
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

			// Create a local copy of args to avoid corrupting shared test data
			args := append([]string{}, tt.args...)

			// Use t.TempDir for source directory if needed
			tempDir := ""
			for i := range args {
				if i > 0 && args[i-1] == testFlagSource && args[i] == "" {
					tempDir = t.TempDir()
					args[i] = tempDir
				}
			}
			os.Args = args

			// Set validate closure if needed (for tempDir)
			if tt.name == "valid flags with all options" {
				tt.validate = func(t *testing.T, f *Flags) {
					assert.Equal(t, tempDir, f.SourceDir)
					assert.Equal(t, "output.md", f.Destination)
					assert.Equal(t, "json", f.Format)
					assert.Equal(t, 4, f.Concurrency)
					assert.Equal(t, "prefix", f.Prefix)
					assert.Equal(t, "suffix", f.Suffix)
					assert.True(t, f.NoColors)
					assert.True(t, f.NoProgress)
					assert.True(t, f.Verbose)
				}
			}
			if tt.name == "default values" {
				tt.validate = func(t *testing.T, f *Flags) {
					assert.Equal(t, tempDir, f.SourceDir)
					assert.Equal(t, "markdown", f.Format)
					assert.Equal(t, runtime.NumCPU(), f.Concurrency)
					assert.Equal(t, "", f.Prefix)
					assert.Equal(t, "", f.Suffix)
					assert.False(t, f.NoColors)
					assert.False(t, f.NoProgress)
					assert.False(t, f.Verbose)
					// Destination should be set by setDefaultDestination
					assert.NotEmpty(t, f.Destination)
				}
			}

			// Call setup if present (e.g. for maxConcurrency)
			if tt.setup != nil {
				tt.setup(t)
			}

			flags, err := ParseFlags()

			if tt.expectedError != "" {
				if assert.Error(t, err) {
					assert.Contains(t, err.Error(), tt.expectedError)
				}
				assert.Nil(t, flags)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, flags)
				if tt.validate != nil {
					tt.validate(t, flags)
				}
			}
		})
	}
}

func TestFlagsValidate(t *testing.T) {
	tests := []struct {
		name          string
		flags         *Flags
		setupFunc     func(t *testing.T, f *Flags)
		expectedError string
	}{
		{
			name:          "missing source directory",
			flags:         &Flags{},
			expectedError: testErrSourceRequired,
		},
		{
			name: "invalid format",
			flags: &Flags{
				Format: "invalid",
			},
			setupFunc: func(t *testing.T, f *Flags) {
				f.SourceDir = t.TempDir()
			},
			expectedError: "unsupported output format: invalid",
		},
		{
			name: "invalid concurrency",
			flags: &Flags{
				Format:      "markdown",
				Concurrency: 0,
			},
			setupFunc: func(t *testing.T, f *Flags) {
				f.SourceDir = t.TempDir()
			},
			expectedError: "concurrency (0) must be at least 1",
		},
		{
			name: "path traversal attempt",
			flags: &Flags{
				SourceDir: testPathTraversalPath,
				Format:    "markdown",
			},
			expectedError: testErrPathTraversal,
		},
		{
			name: "valid flags",
			flags: &Flags{
				Format:      "json",
				Concurrency: 4,
			},
			setupFunc: func(t *testing.T, f *Flags) {
				f.SourceDir = t.TempDir()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupFunc != nil {
				tt.setupFunc(t, tt.flags)
			}

			err := tt.flags.validate()

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSetDefaultDestination(t *testing.T) {
	tests := []struct {
		name          string
		flags         *Flags
		setupFunc     func(t *testing.T, f *Flags)
		expectedDest  string
		expectedError string
	}{
		{
			name: "default destination for directory",
			flags: &Flags{
				Format: "markdown",
			},
			setupFunc: func(t *testing.T, f *Flags) {
				f.SourceDir = t.TempDir()
			},
			expectedDest: "", // will check suffix below
		},
		{
			name: "default destination for json format",
			flags: &Flags{
				Format: "json",
			},
			setupFunc: func(t *testing.T, f *Flags) {
				f.SourceDir = t.TempDir()
			},
			expectedDest: "", // will check suffix below
		},
		{
			name: "provided destination unchanged",
			flags: &Flags{
				Format:      "markdown",
				Destination: "custom-output.txt",
			},
			setupFunc: func(t *testing.T, f *Flags) {
				f.SourceDir = t.TempDir()
			},
			expectedDest: "custom-output.txt",
		},
		{
			name: "path traversal in destination",
			flags: &Flags{
				Format:      "markdown",
				Destination: testPathTraversalPath,
			},
			setupFunc: func(t *testing.T, f *Flags) {
				f.SourceDir = t.TempDir()
			},
			expectedError: testErrPathTraversal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupFunc != nil {
				tt.setupFunc(t, tt.flags)
			}

			err := tt.flags.setDefaultDestination()

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				switch {
				case tt.expectedDest != "":
					assert.Equal(t, tt.expectedDest, tt.flags.Destination)
				case tt.flags.Format == "json":
					assert.True(
						t, strings.HasSuffix(tt.flags.Destination, ".json"),
						"expected %q to have suffix .json", tt.flags.Destination,
					)
				case tt.flags.Format == "markdown":
					assert.True(
						t, strings.HasSuffix(tt.flags.Destination, ".markdown"),
						"expected %q to have suffix .markdown", tt.flags.Destination,
					)
				}
			}
		})
	}
}

func TestFlagsSingleton(t *testing.T) {
	// Save original state
	oldFlagsParsed := flagsParsed
	oldGlobalFlags := globalFlags
	defer func() {
		flagsParsed = oldFlagsParsed
		globalFlags = oldGlobalFlags
	}()

	// Test singleton behavior
	flagsParsed = true
	expectedFlags := &Flags{
		SourceDir:   "/test",
		Format:      "json",
		Concurrency: 2,
	}
	globalFlags = expectedFlags

	// Should return cached flags without parsing
	flags, err := ParseFlags()
	assert.NoError(t, err)
	assert.Equal(t, expectedFlags, flags)
	assert.Same(t, globalFlags, flags)
}

func TestNewMissingSourceError(t *testing.T) {
	err := NewMissingSourceError()

	assert.Error(t, err)
	assert.Equal(t, testErrSourceRequired, err.Error())

	// Check if it's the right type
	var missingSourceError *MissingSourceError
	ok := errors.As(err, &missingSourceError)
	assert.True(t, ok)
}
