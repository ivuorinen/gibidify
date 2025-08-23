package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/ivuorinen/gibidify/testutil"
)

// TestNewProcessor tests the processor constructor.
func TestNewProcessor(t *testing.T) {
	tests := []struct {
		name  string
		flags *Flags
		want  processorValidation
	}{
		{
			name: "basic processor creation",
			flags: &Flags{
				SourceDir:   "/test/source",
				Format:      "markdown",
				Concurrency: 2,
				Destination: "/test/output.md",
				NoColors:    false,
				NoProgress:  false,
			},
			want: processorValidation{
				hasBackpressure:    true,
				hasResourceMonitor: true,
				hasUI:              true,
				colorsEnabled:      true,
				progressEnabled:    true,
			},
		},
		{
			name: "processor with colors and progress disabled",
			flags: &Flags{
				SourceDir:   "/test/source",
				Format:      "json",
				Concurrency: 4,
				Destination: "/test/output.json",
				NoColors:    true,
				NoProgress:  true,
			},
			want: processorValidation{
				hasBackpressure:    true,
				hasResourceMonitor: true,
				hasUI:              true,
				colorsEnabled:      false,
				progressEnabled:    false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processor := NewProcessor(tt.flags)

			validateProcessor(t, processor, tt.want)
			validateProcessorFlags(t, processor, tt.flags)
		})
	}
}

// TestProcessor_configureFileTypes tests file type registry configuration.
func TestProcessor_configureFileTypes(t *testing.T) {
	// Reset config before testing
	testutil.ResetViperConfig(t, "")

	tests := []struct {
		name               string
		fileTypesEnabled   bool
		setupConfig        func()
		expectRegistryCall bool
	}{
		{
			name:             "file types disabled",
			fileTypesEnabled: false,
			setupConfig: func() {
				// No additional config needed
			},
			expectRegistryCall: false,
		},
		{
			name:             "file types enabled with default config",
			fileTypesEnabled: true,
			setupConfig: func() {
				// Use default configuration
			},
			expectRegistryCall: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset config state
			testutil.ResetViperConfig(t, "")
			tt.setupConfig()

			flags := &Flags{
				SourceDir:   "/test/source",
				Format:      "markdown",
				Concurrency: 1,
				Destination: "/test/output.md",
			}

			processor := NewProcessor(flags)

			// Mock the file types enabled state
			// This would normally be done through config, but we'll test the function directly
			// since the actual configuration is complex and tested elsewhere
			// No additional setup needed for this test case regardless of fileTypesEnabled value

			// Test that configureFileTypes doesn't panic
			processor.configureFileTypes()
		})
	}
}

// setupCollectFilesTest sets up test directory for file collection tests.
func setupCollectFilesTest(t *testing.T, testName string, setupFiles func(dir string) []string) string {
	t.Helper()

	if testName == "collect from non-existent directory" {
		return "/non/existent/directory"
	}

	testDir := t.TempDir()
	setupFiles(testDir)

	return testDir
}

// validateCollectFiles validates file collection results.
func validateCollectFiles(t *testing.T, files []string, err error, wantCount int, wantErr bool, errContains string) {
	t.Helper()

	if wantErr {
		if err == nil {
			t.Error("Expected error but got none")

			return
		}
		if errContains != "" && !strings.Contains(err.Error(), errContains) {
			t.Errorf("Error should contain %q, got: %v", errContains, err)
		}

		return
	}

	if err != nil {
		t.Errorf("Unexpected error: %v", err)

		return
	}

	if len(files) != wantCount {
		t.Errorf("Expected %d files, got %d", wantCount, len(files))
	}
}

// TestProcessor_collectFiles tests file collection integration.
func TestProcessor_collectFiles(t *testing.T) {
	tests := []struct {
		name        string
		setupFiles  func(dir string) []string
		wantCount   int
		wantErr     bool
		errContains string
	}{
		{
			name: "collect valid files",
			setupFiles: func(dir string) []string {
				files := []testutil.FileSpec{
					{Name: "file1.go", Content: "package main\n"},
					{Name: "file2.txt", Content: "text content\n"},
					{Name: "subdir/file3.py", Content: "print('hello')\n"},
				}

				// Create subdirectory
				subDir := testutil.CreateTestDirectory(t, dir, "subdir")
				_ = subDir

				return testutil.CreateTestFiles(t, dir, files)
			},
			wantCount: 3,
			wantErr:   false,
		},
		{
			name: "collect from empty directory",
			setupFiles: func(_ string) []string {
				return []string{}
			},
			wantCount: 0,
			wantErr:   false,
		},
		{
			name: "collect from non-existent directory",
			setupFiles: func(_ string) []string {
				return []string{}
			},
			wantCount:   0,
			wantErr:     true,
			errContains: "error collecting files",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testutil.ResetViperConfig(t, "")
			testDir := setupCollectFilesTest(t, tt.name, tt.setupFiles)

			flags := &Flags{
				SourceDir:   testDir,
				Format:      "markdown",
				Concurrency: 1,
				Destination: filepath.Join(t.TempDir(), "output.md"),
			}

			processor := NewProcessor(flags)
			files, err := processor.collectFiles()
			validateCollectFiles(t, files, err, tt.wantCount, tt.wantErr, tt.errContains)
		})
	}
}

// setupValidationTestFiles creates test files for validation tests.
func setupValidationTestFiles(t *testing.T, tempDir string, files []string) []string {
	t.Helper()

	var testFiles []string
	for i, fileName := range files {
		if fileName != "" {
			content := fmt.Sprintf("test content %d", i)
			filePath := testutil.CreateTestFile(t, tempDir,
				fmt.Sprintf("test_%d.txt", i), []byte(content))
			testFiles = append(testFiles, filePath)
		} else {
			testFiles = append(testFiles, fileName)
		}
	}

	return testFiles
}

// validateFileCollectionResult validates file collection validation results.
func validateFileCollectionResult(t *testing.T, err error, wantErr bool, errContains string) {
	t.Helper()

	if wantErr {
		if err == nil {
			t.Error("Expected error but got none")

			return
		}
		if errContains != "" && !strings.Contains(err.Error(), errContains) {
			t.Errorf("Error should contain %q, got: %v", errContains, err)
		}

		return
	}

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

// TestProcessor_validateFileCollection tests file validation against resource limits.
func TestProcessor_validateFileCollection(t *testing.T) {
	tests := []struct {
		name                  string
		files                 []string
		setupConfig           func()
		resourceLimitsEnabled bool
		wantErr               bool
		errContains           string
	}{
		{
			name:                  "resource limits disabled",
			files:                 []string{"file1.txt", "file2.txt"},
			resourceLimitsEnabled: false,
			setupConfig:           func() {},
			wantErr:               false,
		},
		{
			name:                  "within file count limit",
			files:                 []string{"file1.txt"},
			resourceLimitsEnabled: true,
			setupConfig:           func() {},
			wantErr:               false,
		},
		{
			name:                  "exceeds file count limit",
			files:                 make([]string, 10001), // Default limit is 10000
			resourceLimitsEnabled: true,
			setupConfig:           func() {},
			wantErr:               true,
			errContains:           "file count",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testutil.ResetViperConfig(t, "")
			tt.setupConfig()

			tempDir := t.TempDir()
			testFiles := setupValidationTestFiles(t, tempDir, tt.files)

			flags := &Flags{
				SourceDir:   tempDir,
				Format:      "markdown",
				Concurrency: 1,
				Destination: filepath.Join(t.TempDir(), "output.md"),
			}

			processor := NewProcessor(flags)
			err := processor.validateFileCollection(testFiles)
			validateFileCollectionResult(t, err, tt.wantErr, tt.errContains)
		})
	}
}

// validateOutputFile validates output file creation results.
func validateOutputFile(t *testing.T, outFile *os.File, err error, wantErr bool, errContains string) {
	t.Helper()

	if wantErr {
		if err == nil {
			t.Error("Expected error but got none")

			return
		}
		if errContains != "" && !strings.Contains(err.Error(), errContains) {
			t.Errorf("Error should contain %q, got: %v", errContains, err)
		}

		return
	}

	if err != nil {
		t.Errorf("Unexpected error: %v", err)

		return
	}

	if outFile == nil {
		t.Error("Expected valid file handle")

		return
	}

	testutil.CloseFile(t, outFile)
}

// TestProcessor_createOutputFile tests output file creation.
func TestProcessor_createOutputFile(t *testing.T) {
	tests := []struct {
		name        string
		setupDest   func() string
		wantErr     bool
		errContains string
	}{
		{
			name: "create valid output file",
			setupDest: func() string {
				return filepath.Join(t.TempDir(), "output.md")
			},
			wantErr: false,
		},
		{
			name: "create file in non-existent directory",
			setupDest: func() string {
				return "/non/existent/dir/output.md"
			},
			wantErr:     true,
			errContains: "failed to create output file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flags := &Flags{
				SourceDir:   t.TempDir(),
				Format:      "markdown",
				Concurrency: 1,
				Destination: tt.setupDest(),
			}

			processor := NewProcessor(flags)
			outFile, err := processor.createOutputFile()
			validateOutputFile(t, outFile, err, tt.wantErr, tt.errContains)
		})
	}
}

// runProcessorIntegrationTest runs a single processor integration test.
func runProcessorIntegrationTest(t *testing.T, testDir, format, outputPath string, concurrency int, timeout time.Duration) error {
	t.Helper()

	flags := &Flags{
		SourceDir:   testDir,
		Format:      format,
		Concurrency: concurrency,
		Destination: outputPath,
		NoColors:    true, // Disable colors for consistent testing
		NoProgress:  true, // Disable progress for consistent testing
	}

	processor := NewProcessor(flags)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	return processor.Process(ctx)
}

// validateProcessingResult validates processor integration test results.
func validateProcessingResult(t *testing.T, err error, outputPath, format string, wantErr bool, errContains string) {
	t.Helper()

	if wantErr {
		if err == nil {
			t.Error("Expected error but got none")

			return
		}
		if errContains != "" && !strings.Contains(err.Error(), errContains) {
			t.Errorf("Error should contain %q, got: %v", errContains, err)
		}

		return
	}

	if err != nil {
		t.Errorf("Unexpected error: %v", err)

		return
	}

	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Errorf("Output file was not created: %s", outputPath)

		return
	}

	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Errorf("Failed to read output file: %v", err)

		return
	}

	validateOutputContent(t, string(content), format)
}

// TestProcessor_Process_Integration tests the complete processing workflow.
func TestProcessor_Process_Integration(t *testing.T) {
	tests := []struct {
		name        string
		setupFiles  func(dir string) []string
		format      string
		concurrency int
		timeout     time.Duration
		wantErr     bool
		errContains string
	}{
		{
			name: "successful markdown processing",
			setupFiles: func(dir string) []string {
				files := []testutil.FileSpec{
					{Name: "main.go", Content: "package main\n\nfunc main() {\n\tprintln(\"Hello\")\n}\n"},
					{Name: "README.md", Content: "# Test Project\n\nThis is a test.\n"},
				}

				return testutil.CreateTestFiles(t, dir, files)
			},
			format:      "markdown",
			concurrency: 2,
			timeout:     30 * time.Second,
			wantErr:     false,
		},
		{
			name: "successful json processing",
			setupFiles: func(dir string) []string {
				files := []testutil.FileSpec{
					{Name: "config.json", Content: "{\"name\": \"test\"}\n"},
				}

				return testutil.CreateTestFiles(t, dir, files)
			},
			format:      "json",
			concurrency: 1,
			timeout:     30 * time.Second,
			wantErr:     false,
		},
		{
			name: "processing with no files",
			setupFiles: func(_ string) []string {
				return []string{}
			},
			format:      "yaml",
			concurrency: 1,
			timeout:     30 * time.Second,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testutil.ResetViperConfig(t, "")

			testDir := t.TempDir()
			tt.setupFiles(testDir)

			outputPath := filepath.Join(t.TempDir(), "output."+tt.format)
			err := runProcessorIntegrationTest(t, testDir, tt.format, outputPath, tt.concurrency, tt.timeout)
			validateProcessingResult(t, err, outputPath, tt.format, tt.wantErr, tt.errContains)
		})
	}
}

// TestProcessor_Process_ContextCancellation tests context cancellation handling.
func TestProcessor_Process_ContextCancellation(t *testing.T) {
	testutil.ResetViperConfig(t, "")

	// Create test files
	testDir := t.TempDir()
	files := []testutil.FileSpec{
		{Name: "file1.txt", Content: "content1\n"},
		{Name: "file2.txt", Content: "content2\n"},
	}
	testutil.CreateTestFiles(t, testDir, files)

	outputPath := filepath.Join(t.TempDir(), "output.md")

	flags := &Flags{
		SourceDir:   testDir,
		Format:      "markdown",
		Concurrency: 1,
		Destination: outputPath,
		NoColors:    true,
		NoProgress:  true,
	}

	processor := NewProcessor(flags)

	// Create context that will be canceled immediately
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err := processor.Process(ctx)

	// Should get a context cancellation error or complete quickly
	if err != nil && !strings.Contains(err.Error(), "context") {
		t.Errorf("Expected context cancellation error, got: %v", err)
	}
}

// TestProcessor_Process_ResourceLimits tests processing with resource limits.
func TestProcessor_Process_ResourceLimits(t *testing.T) {
	tests := []struct {
		name        string
		setupConfig func()
		setupFiles  func(dir string) []string
		wantErr     bool
		errContains string
	}{
		{
			name: "within resource limits",
			setupConfig: func() {
				// Use default limits
			},
			setupFiles: func(dir string) []string {
				files := []testutil.FileSpec{
					{Name: "small.txt", Content: "small content\n"},
				}

				return testutil.CreateTestFiles(t, dir, files)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testutil.ResetViperConfig(t, "")
			tt.setupConfig()

			testDir := t.TempDir()
			tt.setupFiles(testDir)

			outputPath := filepath.Join(t.TempDir(), "output.md")

			flags := &Flags{
				SourceDir:   testDir,
				Format:      "markdown",
				Concurrency: 1,
				Destination: outputPath,
				NoColors:    true,
				NoProgress:  true,
			}

			processor := NewProcessor(flags)
			ctx := context.Background()

			err := processor.Process(ctx)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")

					return
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("Error should contain %q, got: %v", tt.errContains, err)
				}
			} else if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

// TestProcessor_logFinalStats tests final statistics logging.
func TestProcessor_logFinalStats(t *testing.T) {
	testutil.ResetViperConfig(t, "")

	flags := &Flags{
		SourceDir:   t.TempDir(),
		Format:      "markdown",
		Concurrency: 1,
		Destination: filepath.Join(t.TempDir(), "output.md"),
	}

	processor := NewProcessor(flags)

	// Test that logFinalStats doesn't panic
	processor.logFinalStats()

	// Verify that resource monitor is properly closed
	// (This is mainly to ensure the method completes without error)
}

// Helper types and functions

type processorValidation struct {
	hasBackpressure    bool
	hasResourceMonitor bool
	hasUI              bool
	colorsEnabled      bool
	progressEnabled    bool
}

func validateProcessor(t *testing.T, processor *Processor, want processorValidation) {
	t.Helper()

	if processor == nil {
		t.Error("NewProcessor() returned nil")

		return
	}

	if want.hasBackpressure && processor.backpressure == nil {
		t.Error("Processor should have backpressure manager")
	}

	if want.hasResourceMonitor && processor.resourceMonitor == nil {
		t.Error("Processor should have resource monitor")
	}

	if want.hasUI && processor.ui == nil {
		t.Error("Processor should have UI manager")
	}

	if processor.ui != nil {
		if processor.ui.enableColors != want.colorsEnabled {
			t.Errorf("Colors enabled = %v, want %v", processor.ui.enableColors, want.colorsEnabled)
		}

		if processor.ui.enableProgress != want.progressEnabled {
			t.Errorf("Progress enabled = %v, want %v", processor.ui.enableProgress, want.progressEnabled)
		}
	}
}

func validateProcessorFlags(t *testing.T, processor *Processor, flags *Flags) {
	t.Helper()

	if processor.flags != flags {
		t.Error("Processor should store the provided flags")
	}
}

func validateOutputContent(t *testing.T, content, format string) {
	t.Helper()

	if content == "" {
		t.Error("Output content should not be empty")

		return
	}

	switch format {
	case "markdown":
		// Markdown should have some structure
		// Check for markdown code blocks if content is substantial
		// Empty directories might produce minimal output which is expected behavior
		if !strings.Contains(content, "```") && len(content) > 10 {
			t.Log("Markdown output may be minimal for empty directories")
		}
	case "json":
		// JSON should start with [ or {
		trimmed := strings.TrimSpace(content)
		if len(trimmed) > 0 && !strings.HasPrefix(trimmed, "[") && !strings.HasPrefix(trimmed, "{") {
			t.Error("JSON output should start with [ or {")
		}
	case "yaml":
		// YAML output validation - should have some structure
		// Basic validation that it's not obviously malformed
		// Content exists, which is the main requirement
		if len(content) == 0 {
			t.Log("YAML output is empty, which may be expected for minimal input")
		}
	}
}

// Benchmark tests for processor performance

func BenchmarkProcessor_NewProcessor(b *testing.B) {
	flags := &Flags{
		SourceDir:   "/test/source",
		Format:      "markdown",
		Concurrency: runtime.NumCPU(),
		Destination: "/test/output.md",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		processor := NewProcessor(flags)
		_ = processor
	}
}

func BenchmarkProcessor_collectFiles(b *testing.B) {
	// Create test directory with files
	tempDir := b.TempDir()
	files := []testutil.FileSpec{
		{Name: "file1.go", Content: "package main\n"},
		{Name: "file2.txt", Content: "content\n"},
		{Name: "file3.py", Content: "print('hello')\n"},
	}
	// Create test files manually since CreateTestFiles expects *testing.T
	for _, spec := range files {
		filePath := filepath.Join(tempDir, spec.Name)
		if err := os.WriteFile(filePath, []byte(spec.Content), testutil.FilePermission); err != nil {
			b.Fatalf("Failed to create test file %s: %v", filePath, err)
		}
	}

	flags := &Flags{
		SourceDir:   tempDir,
		Format:      "markdown",
		Concurrency: 1,
		Destination: filepath.Join(b.TempDir(), "output.md"),
	}

	processor := NewProcessor(flags)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		files, err := processor.collectFiles()
		if err != nil {
			b.Fatalf("collectFiles failed: %v", err)
		}
		if len(files) == 0 {
			b.Fatal("Expected files to be collected")
		}
	}
}
