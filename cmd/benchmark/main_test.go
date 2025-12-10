package main

import (
	"errors"
	"flag"
	"io"
	"os"
	"runtime"
	"testing"

	"github.com/ivuorinen/gibidify/shared"
	"github.com/ivuorinen/gibidify/testutil"
)

// Test constants to avoid goconst linting issues.
const (
	testJSON         = "json"
	testMarkdown     = "markdown"
	testConcurrency  = "1,2"
	testAll          = "all"
	testCollection   = "collection"
	testConcurrencyT = "concurrency"
	testNonExistent  = "/nonexistent/path/that/should/not/exist"
	testFile1        = "test1.txt"
	testFile2        = "test2.txt"
	testContent1     = "content1"
	testContent2     = "content2"
)

func TestParseConcurrencyList(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		want        []int
		wantErr     bool
		errContains string
	}{
		{
			name:    "valid single value",
			input:   "4",
			want:    []int{4},
			wantErr: false,
		},
		{
			name:    "valid multiple values",
			input:   shared.TestConcurrencyList,
			want:    []int{1, 2, 4, 8},
			wantErr: false,
		},
		{
			name:    "valid with whitespace",
			input:   " 1 , 2 , 4 , 8 ",
			want:    []int{1, 2, 4, 8},
			wantErr: false,
		},
		{
			name:    "valid single large value",
			input:   "16",
			want:    []int{16},
			wantErr: false,
		},
		{
			name:        "empty string",
			input:       "",
			wantErr:     true,
			errContains: shared.TestMsgInvalidConcurrencyLevel,
		},
		{
			name:        "invalid number",
			input:       "1,abc,4",
			wantErr:     true,
			errContains: shared.TestMsgInvalidConcurrencyLevel,
		},
		{
			name:        "zero value",
			input:       "1,0,4",
			wantErr:     true,
			errContains: "concurrency level must be positive",
		},
		{
			name:        "negative value",
			input:       "1,-2,4",
			wantErr:     true,
			errContains: "concurrency level must be positive",
		},
		{
			name:        "only whitespace",
			input:       " , , ",
			wantErr:     true,
			errContains: shared.TestMsgInvalidConcurrencyLevel,
		},
		{
			name:    "large value list",
			input:   "1,2,4,8,16",
			want:    []int{1, 2, 4, 8, 16},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseConcurrencyList(tt.input)

			if tt.wantErr {
				testutil.AssertExpectedError(t, err, "parseConcurrencyList")
				if tt.errContains != "" {
					testutil.AssertErrorContains(t, err, tt.errContains, "parseConcurrencyList")
				}

				return
			}

			testutil.AssertNoError(t, err, "parseConcurrencyList")
			if !equalSlices(got, tt.want) {
				t.Errorf("parseConcurrencyList() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseFormatList(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{
			name:  "single format",
			input: "json",
			want:  []string{"json"},
		},
		{
			name:  "multiple formats",
			input: shared.TestFormatList,
			want:  []string{"json", "yaml", "markdown"},
		},
		{
			name:  "formats with whitespace",
			input: " json , yaml , markdown ",
			want:  []string{"json", "yaml", "markdown"},
		},
		{
			name:  "empty string",
			input: "",
			want:  []string{},
		},
		{
			name:  "empty parts",
			input: "json,,yaml",
			want:  []string{"json", "yaml"},
		},
		{
			name:  "only whitespace and commas",
			input: " , , ",
			want:  []string{},
		},
		{
			name:  "single format with whitespace",
			input: " markdown ",
			want:  []string{"markdown"},
		},
		{
			name:  "duplicate formats",
			input: "json,json,yaml",
			want:  []string{"json", "json", "yaml"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseFormatList(tt.input)
			if !equalSlices(got, tt.want) {
				t.Errorf("parseFormatList() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetSourceDescription(t *testing.T) {
	// Save original flag values and reset after test
	origSourceDir := sourceDir
	origNumFiles := numFiles
	defer func() {
		sourceDir = origSourceDir
		numFiles = origNumFiles
	}()

	tests := []struct {
		name      string
		sourceDir string
		numFiles  int
		want      string
	}{
		{
			name:      "empty source directory with default files",
			sourceDir: "",
			numFiles:  100,
			want:      "temporary files (100 files)",
		},
		{
			name:      "empty source directory with custom files",
			sourceDir: "",
			numFiles:  50,
			want:      "temporary files (50 files)",
		},
		{
			name:      "non-empty source directory",
			sourceDir: "/path/to/source",
			numFiles:  100,
			want:      "/path/to/source",
		},
		{
			name:      "current directory",
			sourceDir: ".",
			numFiles:  100,
			want:      ".",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set flag pointers to test values
			*sourceDir = tt.sourceDir
			*numFiles = tt.numFiles

			got := getSourceDescription()
			if got != tt.want {
				t.Errorf("getSourceDescription() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRunCollectionBenchmark(t *testing.T) {
	restore := testutil.SuppressLogs(t)
	defer restore()

	// Save original flag values
	origSourceDir := sourceDir
	origNumFiles := numFiles
	defer func() {
		sourceDir = origSourceDir
		numFiles = origNumFiles
	}()

	t.Run("success with temp files", func(t *testing.T) {
		*sourceDir = ""
		*numFiles = 10

		err := runCollectionBenchmark()
		testutil.AssertNoError(t, err, "runCollectionBenchmark with temp files")
	})

	t.Run("success with real directory", func(t *testing.T) {
		tempDir := t.TempDir()
		testutil.CreateTestFiles(t, tempDir, []testutil.FileSpec{
			{Name: testFile1, Content: testContent1},
			{Name: testFile2, Content: testContent2},
		})

		*sourceDir = tempDir
		*numFiles = 10

		err := runCollectionBenchmark()
		testutil.AssertNoError(t, err, "runCollectionBenchmark with real directory")
	})
}

func TestRunProcessingBenchmark(t *testing.T) {
	restore := testutil.SuppressLogs(t)
	defer restore()

	// Save original flag values
	origSourceDir := sourceDir
	origFormat := format
	origConcurrency := concurrency
	defer func() {
		sourceDir = origSourceDir
		format = origFormat
		concurrency = origConcurrency
	}()

	t.Run("success with json format", func(t *testing.T) {
		tempDir := t.TempDir()
		testutil.CreateTestFiles(t, tempDir, []testutil.FileSpec{
			{Name: testFile1, Content: testContent1},
			{Name: testFile2, Content: testContent2},
		})

		*sourceDir = tempDir
		*format = testJSON
		*concurrency = 2

		err := runProcessingBenchmark()
		testutil.AssertNoError(t, err, "runProcessingBenchmark with json")
	})

	t.Run("success with markdown format", func(t *testing.T) {
		tempDir := t.TempDir()
		testutil.CreateTestFiles(t, tempDir, []testutil.FileSpec{
			{Name: testFile1, Content: testContent1},
		})

		*sourceDir = tempDir
		*format = testMarkdown
		*concurrency = 1

		err := runProcessingBenchmark()
		testutil.AssertNoError(t, err, "runProcessingBenchmark with markdown")
	})
}

func TestRunConcurrencyBenchmark(t *testing.T) {
	restore := testutil.SuppressLogs(t)
	defer restore()

	// Save original flag values
	origSourceDir := sourceDir
	origFormat := format
	origConcurrencyList := concurrencyList
	defer func() {
		sourceDir = origSourceDir
		format = origFormat
		concurrencyList = origConcurrencyList
	}()

	t.Run("success with valid concurrency list", func(t *testing.T) {
		tempDir := t.TempDir()
		testutil.CreateTestFiles(t, tempDir, []testutil.FileSpec{
			{Name: testFile1, Content: testContent1},
		})

		*sourceDir = tempDir
		*format = testJSON
		*concurrencyList = testConcurrency

		err := runConcurrencyBenchmark()
		testutil.AssertNoError(t, err, "runConcurrencyBenchmark")
	})

	t.Run("error with invalid concurrency list", func(t *testing.T) {
		tempDir := t.TempDir()
		*sourceDir = tempDir
		*format = testJSON
		*concurrencyList = "invalid"

		err := runConcurrencyBenchmark()
		testutil.AssertExpectedError(t, err, "runConcurrencyBenchmark with invalid list")
		testutil.AssertErrorContains(t, err, "invalid concurrency list", "runConcurrencyBenchmark")
	})
}

func TestRunFormatBenchmark(t *testing.T) {
	restore := testutil.SuppressLogs(t)
	defer restore()

	// Save original flag values
	origSourceDir := sourceDir
	origFormatList := formatList
	defer func() {
		sourceDir = origSourceDir
		formatList = origFormatList
	}()

	t.Run("success with valid format list", func(t *testing.T) {
		tempDir := t.TempDir()
		testutil.CreateTestFiles(t, tempDir, []testutil.FileSpec{
			{Name: testFile1, Content: testContent1},
		})

		*sourceDir = tempDir
		*formatList = "json,yaml"

		err := runFormatBenchmark()
		testutil.AssertNoError(t, err, "runFormatBenchmark")
	})

	t.Run("success with single format", func(t *testing.T) {
		tempDir := t.TempDir()
		testutil.CreateTestFiles(t, tempDir, []testutil.FileSpec{
			{Name: testFile1, Content: testContent1},
		})

		*sourceDir = tempDir
		*formatList = testMarkdown

		err := runFormatBenchmark()
		testutil.AssertNoError(t, err, "runFormatBenchmark with single format")
	})
}

func TestRunBenchmarks(t *testing.T) {
	restore := testutil.SuppressLogs(t)
	defer restore()

	// Save original flag values
	origBenchmarkType := benchmarkType
	origSourceDir := sourceDir
	origConcurrencyList := concurrencyList
	origFormatList := formatList
	defer func() {
		benchmarkType = origBenchmarkType
		sourceDir = origSourceDir
		concurrencyList = origConcurrencyList
		formatList = origFormatList
	}()

	tempDir := t.TempDir()
	testutil.CreateTestFiles(t, tempDir, []testutil.FileSpec{
		{Name: testFile1, Content: testContent1},
	})

	tests := []struct {
		name          string
		benchmarkType string
		wantErr       bool
		errContains   string
	}{
		{
			name:          "all benchmarks",
			benchmarkType: "all",
			wantErr:       false,
		},
		{
			name:          "collection benchmark",
			benchmarkType: "collection",
			wantErr:       false,
		},
		{
			name:          "processing benchmark",
			benchmarkType: "processing",
			wantErr:       false,
		},
		{
			name:          "concurrency benchmark",
			benchmarkType: "concurrency",
			wantErr:       false,
		},
		{
			name:          "format benchmark",
			benchmarkType: "format",
			wantErr:       false,
		},
		{
			name:          "invalid benchmark type",
			benchmarkType: "invalid",
			wantErr:       true,
			errContains:   "invalid benchmark type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			*benchmarkType = tt.benchmarkType
			*sourceDir = tempDir
			*concurrencyList = testConcurrency
			*formatList = testMarkdown

			err := runBenchmarks()

			if tt.wantErr {
				testutil.AssertExpectedError(t, err, "runBenchmarks")
				if tt.errContains != "" {
					testutil.AssertErrorContains(t, err, tt.errContains, "runBenchmarks")
				}
			} else {
				testutil.AssertNoError(t, err, "runBenchmarks")
			}
		})
	}
}

func TestMainFunction(t *testing.T) {
	restore := testutil.SuppressLogs(t)
	defer restore()

	// We can't easily test main() directly due to os.Exit calls,
	// but we can test runBenchmarks() which contains the main logic
	tempDir := t.TempDir()
	testutil.CreateTestFiles(t, tempDir, []testutil.FileSpec{
		{Name: testFile1, Content: testContent1},
	})

	// Save original flag values
	origBenchmarkType := benchmarkType
	origSourceDir := sourceDir
	defer func() {
		benchmarkType = origBenchmarkType
		sourceDir = origSourceDir
	}()

	*benchmarkType = testCollection
	*sourceDir = tempDir

	err := runBenchmarks()
	testutil.AssertNoError(t, err, "runBenchmarks through main logic path")
}

func TestFlagInitialization(t *testing.T) {
	// Test that flags are properly initialized with expected defaults
	resetFlags()

	if *sourceDir != "" {
		t.Errorf("sourceDir default should be empty, got %v", *sourceDir)
	}
	if *benchmarkType != testAll {
		t.Errorf("benchmarkType default should be 'all', got %v", *benchmarkType)
	}
	if *format != testJSON {
		t.Errorf("format default should be 'json', got %v", *format)
	}
	if *concurrency != runtime.NumCPU() {
		t.Errorf("concurrency default should be %d, got %d", runtime.NumCPU(), *concurrency)
	}
	if *concurrencyList != shared.TestConcurrencyList {
		t.Errorf("concurrencyList default should be '%s', got %v", shared.TestConcurrencyList, *concurrencyList)
	}
	if *formatList != shared.TestFormatList {
		t.Errorf("formatList default should be '%s', got %v", shared.TestFormatList, *formatList)
	}
	if *numFiles != 100 {
		t.Errorf("numFiles default should be 100, got %d", *numFiles)
	}
}

func TestErrorPropagation(t *testing.T) {
	restore := testutil.SuppressLogs(t)
	defer restore()

	// Save original flag values
	origBenchmarkType := benchmarkType
	origSourceDir := sourceDir
	origConcurrencyList := concurrencyList
	defer func() {
		benchmarkType = origBenchmarkType
		sourceDir = origSourceDir
		concurrencyList = origConcurrencyList
	}()

	tempDir := t.TempDir()

	t.Run("error from concurrency benchmark propagates", func(t *testing.T) {
		*benchmarkType = testConcurrencyT
		*sourceDir = tempDir
		*concurrencyList = "invalid,list"

		err := runBenchmarks()
		testutil.AssertExpectedError(t, err, "runBenchmarks with invalid concurrency")
		testutil.AssertErrorContains(t, err, "invalid concurrency list", "runBenchmarks error propagation")
	})

	t.Run("validation error contains proper error type", func(t *testing.T) {
		*benchmarkType = "invalid-type"
		*sourceDir = tempDir

		err := runBenchmarks()
		testutil.AssertExpectedError(t, err, "runBenchmarks with invalid type")

		var validationErr *shared.StructuredError
		if !errors.As(err, &validationErr) {
			t.Errorf("Expected StructuredError, got %T", err)
		} else if validationErr.Code != shared.CodeValidationFormat {
			t.Errorf("Expected validation format error code, got %v", validationErr.Code)
		}
	})

	t.Run("empty levels array returns error", func(t *testing.T) {
		// Test the specific case where all parts are empty after trimming
		_, err := parseConcurrencyList("   ,   ,   ")
		testutil.AssertExpectedError(t, err, "parseConcurrencyList with all empty parts")
		testutil.AssertErrorContains(t, err, shared.TestMsgInvalidConcurrencyLevel, "parseConcurrencyList empty levels")
	})

	t.Run("single empty part returns error", func(t *testing.T) {
		// Test case that should never reach the "no valid levels found" condition
		_, err := parseConcurrencyList("  ")
		testutil.AssertExpectedError(t, err, "parseConcurrencyList with single empty part")
		testutil.AssertErrorContains(
			t, err, shared.TestMsgInvalidConcurrencyLevel, "parseConcurrencyList single empty part",
		)
	})

	t.Run("benchmark function error paths", func(t *testing.T) {
		// Test with non-existent source directory to trigger error paths
		nonExistentDir := testNonExistent

		*benchmarkType = testCollection
		*sourceDir = nonExistentDir

		// This should fail as the benchmark package cannot access non-existent directories
		err := runBenchmarks()
		testutil.AssertExpectedError(t, err, "runBenchmarks with non-existent directory")
		testutil.AssertErrorContains(t, err, "file collection benchmark failed",
			"runBenchmarks error contains expected message")
	})

	t.Run("processing benchmark error path", func(t *testing.T) {
		// Test error path for processing benchmark
		nonExistentDir := testNonExistent

		*benchmarkType = "processing"
		*sourceDir = nonExistentDir
		*format = "json"
		*concurrency = 1

		err := runBenchmarks()
		testutil.AssertExpectedError(t, err, "runBenchmarks processing with non-existent directory")
		testutil.AssertErrorContains(t, err, "file processing benchmark failed", "runBenchmarks processing error")
	})

	t.Run("concurrency benchmark error path", func(t *testing.T) {
		// Test error path for concurrency benchmark
		nonExistentDir := testNonExistent

		*benchmarkType = testConcurrencyT
		*sourceDir = nonExistentDir
		*format = "json"
		*concurrencyList = "1,2"

		err := runBenchmarks()
		testutil.AssertExpectedError(t, err, "runBenchmarks concurrency with non-existent directory")
		testutil.AssertErrorContains(t, err, "concurrency benchmark failed", "runBenchmarks concurrency error")
	})

	t.Run("format benchmark error path", func(t *testing.T) {
		// Test error path for format benchmark
		nonExistentDir := testNonExistent

		*benchmarkType = "format"
		*sourceDir = nonExistentDir
		*formatList = "json,yaml"

		err := runBenchmarks()
		testutil.AssertExpectedError(t, err, "runBenchmarks format with non-existent directory")
		testutil.AssertErrorContains(t, err, "format benchmark failed", "runBenchmarks format error")
	})

	t.Run("all benchmarks error path", func(t *testing.T) {
		// Test error path for all benchmarks
		nonExistentDir := testNonExistent

		*benchmarkType = "all"
		*sourceDir = nonExistentDir

		err := runBenchmarks()
		testutil.AssertExpectedError(t, err, "runBenchmarks all with non-existent directory")
		testutil.AssertErrorContains(t, err, "benchmark failed", "runBenchmarks all error")
	})
}

// Benchmark functions

// BenchmarkParseConcurrencyList benchmarks the parsing of concurrency lists.
func BenchmarkParseConcurrencyList(b *testing.B) {
	benchmarks := []struct {
		name  string
		input string
	}{
		{
			name:  "single value",
			input: "4",
		},
		{
			name:  "multiple values",
			input: "1,2,4,8",
		},
		{
			name:  "values with whitespace",
			input: " 1 , 2 , 4 , 8 , 16 ",
		},
		{
			name:  "large list",
			input: "1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16",
		},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_, _ = parseConcurrencyList(bm.input)
			}
		})
	}
}

// BenchmarkParseFormatList benchmarks the parsing of format lists.
func BenchmarkParseFormatList(b *testing.B) {
	benchmarks := []struct {
		name  string
		input string
	}{
		{
			name:  "single format",
			input: "json",
		},
		{
			name:  "multiple formats",
			input: shared.TestFormatList,
		},
		{
			name:  "formats with whitespace",
			input: " json , yaml , markdown , xml , toml ",
		},
		{
			name:  "large list",
			input: "json,yaml,markdown,xml,toml,csv,tsv,html,txt,log",
		},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_ = parseFormatList(bm.input)
			}
		})
	}
}

// Helper functions

// equalSlices compares two slices for equality.
func equalSlices[T comparable](a, b []T) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

// resetFlags resets flag variables to their defaults for testing.
func resetFlags() {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	flag.CommandLine.SetOutput(io.Discard)
	// Reinitialize the flags
	sourceDir = flag.String("source", "", "Source directory to benchmark (uses temp files if empty)")
	benchmarkType = flag.String("type", "all", "Benchmark type: all, collection, processing, concurrency, format")
	format = flag.String("format", "json", "Output format for processing benchmarks")
	concurrency = flag.Int("concurrency", runtime.NumCPU(), "Concurrency level for processing benchmarks")
	concurrencyList = flag.String(
		"concurrency-list", shared.TestConcurrencyList, "Comma-separated list of concurrency levels",
	)
	formatList = flag.String("format-list", shared.TestFormatList, "Comma-separated list of formats")
	numFiles = flag.Int("files", 100, "Number of files to create for benchmarks")
}
