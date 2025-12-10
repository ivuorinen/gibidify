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

	"github.com/spf13/viper"

	"github.com/ivuorinen/gibidify/config"
	"github.com/ivuorinen/gibidify/fileproc"
	"github.com/ivuorinen/gibidify/shared"
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
				SourceDir:   shared.TestSourcePath,
				Format:      "markdown",
				Concurrency: 2,
				Destination: shared.TestOutputMarkdown,
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
				SourceDir:   shared.TestSourcePath,
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
		t.Run(
			tt.name, func(t *testing.T) {
				processor := NewProcessor(tt.flags)

				validateProcessor(t, processor, tt.want)
				validateProcessorFlags(t, processor, tt.flags)
			},
		)
	}
}

// configureFileTypesTestCase holds test case data for file types configuration.
type configureFileTypesTestCase struct {
	name             string
	fileTypesEnabled bool
	customExtensions []string
	wantCustom       bool
}

// setupFileTypesConfig initializes viper config for file types test.
func setupFileTypesConfig(t *testing.T, tt configureFileTypesTestCase) {
	t.Helper()
	viper.Reset()
	config.SetDefaultConfig()
	viper.Set(shared.ConfigKeyFileTypesEnabled, tt.fileTypesEnabled)
	if len(tt.customExtensions) > 0 {
		viper.Set("fileTypes.customImageExtensions", tt.customExtensions)
	}
}

// verifyDefaultExtensions checks that default extensions are recognized.
func verifyDefaultExtensions(t *testing.T, registry *fileproc.FileTypeRegistry) {
	t.Helper()
	if !registry.IsImage(shared.TestFilePNG) {
		t.Error("expected .png to be recognized as image (default extension)")
	}
	if !registry.IsImage(shared.TestFileJPG) {
		t.Error("expected .jpg to be recognized as image (default extension)")
	}
	if registry.Language(shared.TestFileGo) == "" {
		t.Error("expected .go to have language mapping (default extension)")
	}
}

// verifyCustomExtensions checks that custom extensions are recognized when expected.
func verifyCustomExtensions(t *testing.T, registry *fileproc.FileTypeRegistry, tt configureFileTypesTestCase) {
	t.Helper()
	if !tt.wantCustom || len(tt.customExtensions) == 0 {
		return
	}
	testFile := "test" + tt.customExtensions[0]
	if !registry.IsImage(testFile) {
		t.Errorf("expected %s to be recognized as image (custom extension)", testFile)
	}
}

// verifyRegistryState checks registry has reasonable state.
func verifyRegistryState(t *testing.T, registry *fileproc.FileTypeRegistry) {
	t.Helper()
	_, _, maxCache := registry.CacheInfo()
	if maxCache <= 0 {
		t.Errorf("expected positive maxCacheSize, got %d", maxCache)
	}
}

// TestProcessorConfigureFileTypes tests file type registry configuration.
func TestProcessorConfigureFileTypes(t *testing.T) {
	tests := []configureFileTypesTestCase{
		{
			name:             "file types disabled - no custom extensions applied",
			fileTypesEnabled: false,
			customExtensions: []string{".testcustom"},
			wantCustom:       false,
		},
		{
			name:             "file types enabled - custom extensions applied",
			fileTypesEnabled: true,
			customExtensions: []string{".mycustomext"},
			wantCustom:       true,
		},
		{
			name:             "file types enabled with defaults",
			fileTypesEnabled: true,
			customExtensions: nil,
			wantCustom:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupFileTypesConfig(t, tt)

			if got := config.FileTypesEnabled(); got != tt.fileTypesEnabled {
				t.Errorf("FileTypesEnabled() = %v, want %v", got, tt.fileTypesEnabled)
			}

			flags := &Flags{
				SourceDir:   shared.TestSourcePath,
				Format:      shared.FormatMarkdown,
				Concurrency: 1,
				Destination: shared.TestOutputMarkdown,
			}
			processor := NewProcessor(flags)
			processor.configureFileTypes()

			registry := fileproc.DefaultRegistry()
			verifyDefaultExtensions(t, registry)
			verifyCustomExtensions(t, registry, tt)
			verifyRegistryState(t, registry)
		})
	}
}

// setupCollectFilesTest sets up test directory for file collection tests.
func setupCollectFilesTest(t *testing.T, useNonExistent bool, setupFiles func(dir string) []string) string {
	t.Helper()

	if useNonExistent {
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
			t.Error(shared.TestMsgExpectedError)

			return
		}
		if errContains != "" && !strings.Contains(err.Error(), errContains) {
			t.Errorf(shared.TestMsgErrorShouldContain, errContains, err)
		}

		return
	}

	if err != nil {
		t.Errorf(shared.TestMsgUnexpectedError, err)

		return
	}

	if len(files) != wantCount {
		t.Errorf("Expected %d files, got %d", wantCount, len(files))
	}
}

// TestProcessor_collectFiles tests file collection integration.
func TestProcessorCollectFiles(t *testing.T) {
	tests := []struct {
		name           string
		setupFiles     func(dir string) []string
		useNonExistent bool
		wantCount      int
		wantErr        bool
		errContains    string
	}{
		{
			name: "collect valid files",
			setupFiles: func(dir string) []string {
				files := []testutil.FileSpec{
					{Name: "file1.go", Content: shared.LiteralPackageMain + "\n"},
					{Name: shared.TestFile2, Content: "text content\n"},
					{Name: "subdir/file3.py", Content: "print('hello')\n"},
				}

				// Create subdirectory
				testutil.CreateTestDirectory(t, dir, "subdir")

				return testutil.CreateTestFiles(t, dir, files)
			},
			useNonExistent: false,
			wantCount:      3,
			wantErr:        false,
		},
		{
			name: "collect from empty directory",
			setupFiles: func(_ string) []string {
				return []string{}
			},
			useNonExistent: false,
			wantCount:      0,
			wantErr:        false,
		},
		{
			name: "collect from non-existent directory",
			setupFiles: func(_ string) []string {
				return []string{}
			},
			useNonExistent: true,
			wantCount:      0,
			wantErr:        true,
			errContains:    "error collecting files",
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				testutil.ResetViperConfig(t, "")
				testDir := setupCollectFilesTest(t, tt.useNonExistent, tt.setupFiles)

				flags := &Flags{
					SourceDir:   testDir,
					Format:      "markdown",
					Concurrency: 1,
					Destination: filepath.Join(t.TempDir(), shared.TestOutputMD),
				}

				processor := NewProcessor(flags)
				files, err := processor.collectFiles()
				validateCollectFiles(t, files, err, tt.wantCount, tt.wantErr, tt.errContains)
			},
		)
	}
}

// setupValidationTestFiles creates test files for validation tests.
func setupValidationTestFiles(t *testing.T, tempDir string, files []string) []string {
	t.Helper()

	var testFiles []string
	for i, fileName := range files {
		if fileName != "" {
			content := fmt.Sprintf("test content %d", i)
			filePath := testutil.CreateTestFile(
				t, tempDir,
				fmt.Sprintf("test_%d.txt", i), []byte(content),
			)
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
			t.Error(shared.TestMsgExpectedError)

			return
		}
		if errContains != "" && !strings.Contains(err.Error(), errContains) {
			t.Errorf(shared.TestMsgErrorShouldContain, errContains, err)
		}

		return
	}

	if err != nil {
		t.Errorf(shared.TestMsgUnexpectedError, err)
	}
}

// TestProcessor_validateFileCollection tests file validation against resource limits.
func TestProcessorvalidateFileCollection(t *testing.T) {
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
			files:                 []string{shared.TestFile1, shared.TestFile2},
			resourceLimitsEnabled: false,
			setupConfig: func() {
				// No configuration needed for this test case
			},
			wantErr: false,
		},
		{
			name:                  "within file count limit",
			files:                 []string{shared.TestFile1},
			resourceLimitsEnabled: true,
			setupConfig: func() {
				// Default configuration is sufficient for this test case
			},
			wantErr: false,
		},
		{
			name:                  "exceeds file count limit",
			files:                 make([]string, 10001), // Default limit is 10000
			resourceLimitsEnabled: true,
			setupConfig: func() {
				// Default configuration is sufficient for this test case
			},
			wantErr:     true,
			errContains: "file count",
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				testutil.ResetViperConfig(t, "")
				tt.setupConfig()

				tempDir := t.TempDir()
				testFiles := setupValidationTestFiles(t, tempDir, tt.files)

				flags := &Flags{
					SourceDir:   tempDir,
					Format:      "markdown",
					Concurrency: 1,
					Destination: filepath.Join(t.TempDir(), shared.TestOutputMD),
				}

				processor := NewProcessor(flags)
				err := processor.validateFileCollection(testFiles)
				validateFileCollectionResult(t, err, tt.wantErr, tt.errContains)
			},
		)
	}
}

// validateOutputFile validates output file creation results.
func validateOutputFile(t *testing.T, outFile *os.File, err error, wantErr bool, errContains string) {
	t.Helper()

	if wantErr {
		if err == nil {
			t.Error(shared.TestMsgExpectedError)

			return
		}
		if errContains != "" && !strings.Contains(err.Error(), errContains) {
			t.Errorf(shared.TestMsgErrorShouldContain, errContains, err)
		}

		return
	}

	if err != nil {
		t.Errorf(shared.TestMsgUnexpectedError, err)

		return
	}

	if outFile == nil {
		t.Error("Expected valid file handle")

		return
	}

	testutil.CloseFile(t, outFile)
}

// TestProcessor_createOutputFile tests output file creation.
func TestProcessorcreateOutputFile(t *testing.T) {
	tests := []struct {
		name        string
		setupDest   func() string
		wantErr     bool
		errContains string
	}{
		{
			name: "create valid output file",
			setupDest: func() string {
				return filepath.Join(t.TempDir(), shared.TestOutputMD)
			},
			wantErr: false,
		},
		{
			name: "create file in non-existent directory",
			setupDest: func() string {
				return "/non/existent/dir/" + shared.TestOutputMD
			},
			wantErr:     true,
			errContains: "failed to create output file",
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				flags := &Flags{
					SourceDir:   t.TempDir(),
					Format:      "markdown",
					Concurrency: 1,
					Destination: tt.setupDest(),
				}

				processor := NewProcessor(flags)
				outFile, err := processor.createOutputFile()
				validateOutputFile(t, outFile, err, tt.wantErr, tt.errContains)
			},
		)
	}
}

// runProcessorIntegrationTest runs a single processor integration test.
func runProcessorIntegrationTest(
	t *testing.T,
	testDir, format, outputPath string,
	concurrency int,
	timeout time.Duration,
) error {
	t.Helper()

	flags := &Flags{
		SourceDir:   testDir,
		Format:      format,
		Concurrency: concurrency,
		Destination: outputPath,
		NoColors:    true, // Disable colors for consistent testing
		NoProgress:  true, // Disable progress for consistent testing
		NoUI:        true, // Disable all UI output for testing
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
			t.Error(shared.TestMsgExpectedError)

			return
		}
		if errContains != "" && !strings.Contains(err.Error(), errContains) {
			t.Errorf(shared.TestMsgErrorShouldContain, errContains, err)
		}

		return
	}

	if err != nil {
		t.Errorf(shared.TestMsgUnexpectedError, err)

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
func TestProcessorProcessIntegration(t *testing.T) {
	// Suppress all output for cleaner test output
	restore := testutil.SuppressAllOutput(t)
	defer restore()

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
					{
						Name:    "main.go",
						Content: shared.LiteralPackageMain + "\n\nfunc main() {\n\tprintln(\"Hello\")\n}\n",
					},
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
		t.Run(
			tt.name, func(t *testing.T) {
				testutil.ResetViperConfig(t, "")

				testDir := t.TempDir()
				tt.setupFiles(testDir)

				outputPath := filepath.Join(t.TempDir(), "output."+tt.format)
				err := runProcessorIntegrationTest(t, testDir, tt.format, outputPath, tt.concurrency, tt.timeout)
				validateProcessingResult(t, err, outputPath, tt.format, tt.wantErr, tt.errContains)
			},
		)
	}
}

// TestProcessor_Process_ContextCancellation tests context cancellation handling.
func TestProcessorProcessContextCancellation(t *testing.T) {
	// Suppress all output for cleaner test output
	restore := testutil.SuppressAllOutput(t)
	defer restore()

	testutil.ResetViperConfig(t, "")

	// Create test files
	testDir := t.TempDir()
	files := []testutil.FileSpec{
		{Name: shared.TestFile1, Content: "content1\n"},
		{Name: shared.TestFile2, Content: "content2\n"},
	}
	testutil.CreateTestFiles(t, testDir, files)

	outputPath := filepath.Join(t.TempDir(), shared.TestOutputMD)

	flags := &Flags{
		SourceDir:   testDir,
		Format:      "markdown",
		Concurrency: 1,
		Destination: outputPath,
		NoColors:    true,
		NoProgress:  true,
		NoUI:        true, // Disable all UI output for testing
	}

	processor := NewProcessor(flags)

	// Create context that will be canceled immediately
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err := processor.Process(ctx)

	// Pre-canceled context must return an error
	if err == nil {
		t.Fatal("Expected error for pre-canceled context, got nil")
	}

	// Verify the error is related to context cancellation
	if !strings.Contains(err.Error(), "context") {
		t.Errorf("Expected error containing 'context', got: %v", err)
	}
}

// TestProcessor_Process_ResourceLimits tests processing with resource limits.
func TestProcessorProcessResourceLimits(t *testing.T) {
	// Suppress all output for cleaner test output
	restore := testutil.SuppressAllOutput(t)
	defer restore()

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
		t.Run(
			tt.name, func(t *testing.T) {
				testutil.ResetViperConfig(t, "")
				tt.setupConfig()

				testDir := t.TempDir()
				tt.setupFiles(testDir)

				outputPath := filepath.Join(t.TempDir(), shared.TestOutputMD)

				flags := &Flags{
					SourceDir:   testDir,
					Format:      "markdown",
					Concurrency: 1,
					Destination: outputPath,
					NoColors:    true,
					NoProgress:  true,
					NoUI:        true, // Disable all UI output for testing
				}

				processor := NewProcessor(flags)
				ctx := context.Background()

				err := processor.Process(ctx)

				if tt.wantErr {
					if err == nil {
						t.Error(shared.TestMsgExpectedError)

						return
					}
					if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
						t.Errorf(shared.TestMsgErrorShouldContain, tt.errContains, err)
					}
				} else if err != nil {
					t.Errorf(shared.TestMsgUnexpectedError, err)
				}
			},
		)
	}
}

// logFinalStatsTestCase holds test case data for log final stats tests.
type logFinalStatsTestCase struct {
	name                 string
	enableBackpressure   bool
	enableResourceLimits bool
	simulateProcessing   bool
	expectedKeywords     []string
	unexpectedKeywords   []string
}

// setupLogStatsConfig initializes config for log stats test.
func setupLogStatsConfig(t *testing.T, tt logFinalStatsTestCase) {
	t.Helper()
	viper.Reset()
	config.SetDefaultConfig()
	viper.Set(shared.ConfigKeyBackpressureEnabled, tt.enableBackpressure)
	viper.Set(shared.ConfigKeyResourceLimitsEnabled, tt.enableResourceLimits)
	shared.GetLogger().SetLevel(shared.LogLevelInfo)
}

// createLogStatsProcessor creates a processor for log stats testing.
func createLogStatsProcessor(t *testing.T) *Processor {
	t.Helper()
	flags := &Flags{
		SourceDir:   t.TempDir(),
		Format:      shared.FormatMarkdown,
		Concurrency: 1,
		Destination: filepath.Join(t.TempDir(), shared.TestOutputMD),
		NoUI:        true,
		NoColors:    true,
		NoProgress:  true,
	}
	return NewProcessor(flags)
}

// simulateProcessing records file processing activity for stats generation.
func simulateProcessing(processor *Processor, simulate bool) {
	if !simulate || processor.resourceMonitor == nil {
		return
	}
	processor.resourceMonitor.RecordFileProcessed(1024)
	processor.resourceMonitor.RecordFileProcessed(2048)
}

// verifyLogKeywords checks expected and unexpected keywords in output.
func verifyLogKeywords(t *testing.T, output string, expected, unexpected []string) {
	t.Helper()
	for _, keyword := range expected {
		if !strings.Contains(output, keyword) {
			t.Errorf("expected output to contain %q, got: %s", keyword, output)
		}
	}
	for _, keyword := range unexpected {
		if strings.Contains(output, keyword) {
			t.Errorf("expected output NOT to contain %q, got: %s", keyword, output)
		}
	}
}

// TestProcessorLogFinalStats tests final statistics logging.
func TestProcessorLogFinalStats(t *testing.T) {
	tests := []logFinalStatsTestCase{
		{
			name:                 "basic stats without features enabled",
			enableBackpressure:   false,
			enableResourceLimits: false,
			simulateProcessing:   false,
			expectedKeywords:     []string{},
			unexpectedKeywords:   []string{"Back-pressure stats", "Resource stats"},
		},
		{
			name:                 "with backpressure enabled",
			enableBackpressure:   true,
			enableResourceLimits: false,
			simulateProcessing:   true,
			expectedKeywords:     []string{"Back-pressure stats", "processed", "memory"},
			unexpectedKeywords:   []string{},
		},
		{
			name:                 "with resource limits enabled",
			enableBackpressure:   false,
			enableResourceLimits: true,
			simulateProcessing:   true,
			expectedKeywords:     []string{"Resource stats", "processed", "files"},
			unexpectedKeywords:   []string{},
		},
		{
			name:                 "with all features enabled",
			enableBackpressure:   true,
			enableResourceLimits: true,
			simulateProcessing:   true,
			expectedKeywords:     []string{"processed"},
			unexpectedKeywords:   []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupLogStatsConfig(t, tt)
			_, getStderr, restore := testutil.CaptureOutput(t)

			processor := createLogStatsProcessor(t)
			simulateProcessing(processor, tt.simulateProcessing)
			processor.logFinalStats()

			restore()
			verifyLogKeywords(t, getStderr(), tt.expectedKeywords, tt.unexpectedKeywords)

			if processor.resourceMonitor != nil {
				processor.resourceMonitor.Close()
			}
		})
	}
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
		// Check for Markdown code blocks if content is substantial
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
		// YAML output validation - content existence verified above
		// Could add YAML structure validation if needed
	default:
		// For unknown formats, just log that we have content
		t.Logf("Unknown format %s, content length: %d", format, len(content))
	}
}

// Benchmark tests for processor performance

func BenchmarkProcessorNewProcessor(b *testing.B) {
	flags := &Flags{
		SourceDir:   shared.TestSourcePath,
		Format:      "markdown",
		Concurrency: runtime.NumCPU(),
		Destination: shared.TestOutputMarkdown,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		processor := NewProcessor(flags)
		_ = processor
	}
}

func BenchmarkProcessorCollectFiles(b *testing.B) {
	// Initialize config for file collection
	viper.Reset()
	config.LoadConfig()

	fileSpecs := []testutil.FileSpec{
		{Name: "file1.go", Content: shared.LiteralPackageMain + "\n"},
		{Name: shared.TestFile2, Content: "content\n"},
		{Name: "file3.py", Content: "print('hello')\n"},
	}

	for b.Loop() {
		// Create fresh directories for each iteration
		tempDir := b.TempDir()
		outDir := b.TempDir()

		// Create test files
		for _, spec := range fileSpecs {
			filePath := filepath.Join(tempDir, spec.Name)
			if err := os.WriteFile(filePath, []byte(spec.Content), shared.TestFilePermission); err != nil {
				b.Fatalf("Failed to create test file %s: %v", filePath, err)
			}
		}

		flags := &Flags{
			SourceDir:   tempDir,
			Format:      "markdown",
			Concurrency: 1,
			Destination: filepath.Join(outDir, shared.TestOutputMD),
		}

		processor := NewProcessor(flags)
		files, err := processor.collectFiles()
		if err != nil {
			b.Fatalf("collectFiles failed: %v", err)
		}
		if len(files) == 0 {
			b.Fatal("Expected files to be collected")
		}
	}
}

// BenchmarkProcessor_Process benchmarks the full Process workflow.
// This provides baseline measurements for the complete processing pipeline.
func BenchmarkProcessorProcess(b *testing.B) {
	// Initialize config for file collection and processing
	viper.Reset()
	config.LoadConfig()

	tempDir := b.TempDir()

	// Create a representative set of test files
	for i := 0; i < 10; i++ {
		filePath := filepath.Join(tempDir, fmt.Sprintf("file%d.go", i))
		content := fmt.Sprintf("package main\n\nfunc fn%d() {}\n", i)
		if err := os.WriteFile(filePath, []byte(content), shared.TestFilePermission); err != nil {
			b.Fatalf("Failed to create test file: %v", err)
		}
	}

	outputDir := b.TempDir()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		flags := &Flags{
			SourceDir:   tempDir,
			Format:      "markdown",
			Concurrency: runtime.NumCPU(),
			Destination: filepath.Join(outputDir, fmt.Sprintf("output_%d.md", i)),
			NoUI:        true,
			NoColors:    true,
			NoProgress:  true,
			LogLevel:    "warn",
		}

		processor := NewProcessor(flags)
		_ = processor.Process(context.Background())
	}
}
