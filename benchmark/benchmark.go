// Package benchmark provides benchmarking infrastructure for gibidify.
package benchmark

import (
	"context"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/ivuorinen/gibidify/config"
	"github.com/ivuorinen/gibidify/fileproc"
	"github.com/ivuorinen/gibidify/utils"
)

// BenchmarkResult represents the results of a benchmark run.
type BenchmarkResult struct {
	Name           string
	Duration       time.Duration
	FilesProcessed int
	BytesProcessed int64
	FilesPerSecond float64
	BytesPerSecond float64
	MemoryUsage    MemoryStats
	CPUUsage       CPUStats
}

// MemoryStats represents memory usage statistics.
type MemoryStats struct {
	AllocMB      float64
	SysMB        float64
	NumGC        uint32
	PauseTotalNs uint64
}

// CPUStats represents CPU usage statistics.
type CPUStats struct {
	UserTime   time.Duration
	SystemTime time.Duration
	Goroutines int
}

// BenchmarkSuite represents a collection of benchmarks.
type BenchmarkSuite struct {
	Name    string
	Results []BenchmarkResult
}

// FileCollectionBenchmark benchmarks file collection operations.
func FileCollectionBenchmark(sourceDir string, numFiles int) (*BenchmarkResult, error) {
	// Load configuration to ensure proper file filtering
	config.LoadConfig()

	// Create temporary directory with test files if no source is provided
	var cleanup func()
	if sourceDir == "" {
		tempDir, cleanupFunc, err := createBenchmarkFiles(numFiles)
		if err != nil {
			return nil, utils.WrapError(err, utils.ErrorTypeFileSystem, utils.CodeFSAccess, "failed to create benchmark files")
		}
		cleanup = cleanupFunc
		defer cleanup()
		sourceDir = tempDir
	}

	// Measure memory before
	var memBefore runtime.MemStats
	runtime.ReadMemStats(&memBefore)

	startTime := time.Now()

	// Run the file collection benchmark
	files, err := fileproc.CollectFiles(sourceDir)
	if err != nil {
		return nil, utils.WrapError(err, utils.ErrorTypeProcessing, utils.CodeProcessingCollection, "benchmark file collection failed")
	}

	duration := time.Since(startTime)

	// Measure memory after
	var memAfter runtime.MemStats
	runtime.ReadMemStats(&memAfter)

	// Calculate total bytes processed
	var totalBytes int64
	for _, file := range files {
		if info, err := os.Stat(file); err == nil {
			totalBytes += info.Size()
		}
	}

	result := &BenchmarkResult{
		Name:           "FileCollection",
		Duration:       duration,
		FilesProcessed: len(files),
		BytesProcessed: totalBytes,
		FilesPerSecond: float64(len(files)) / duration.Seconds(),
		BytesPerSecond: float64(totalBytes) / duration.Seconds(),
		MemoryUsage: MemoryStats{
			AllocMB:      float64(memAfter.Alloc-memBefore.Alloc) / 1024 / 1024,
			SysMB:        float64(memAfter.Sys-memBefore.Sys) / 1024 / 1024,
			NumGC:        memAfter.NumGC - memBefore.NumGC,
			PauseTotalNs: memAfter.PauseTotalNs - memBefore.PauseTotalNs,
		},
		CPUUsage: CPUStats{
			Goroutines: runtime.NumGoroutine(),
		},
	}

	return result, nil
}

// FileProcessingBenchmark benchmarks full file processing pipeline.
func FileProcessingBenchmark(sourceDir string, format string, concurrency int) (*BenchmarkResult, error) {
	// Load configuration to ensure proper file filtering
	config.LoadConfig()

	var cleanup func()
	if sourceDir == "" {
		// Create temporary directory with test files
		tempDir, cleanupFunc, err := createBenchmarkFiles(100)
		if err != nil {
			return nil, utils.WrapError(err, utils.ErrorTypeFileSystem, utils.CodeFSAccess, "failed to create benchmark files")
		}
		cleanup = cleanupFunc
		defer cleanup()
		sourceDir = tempDir
	}

	// Create temporary output file
	outputFile, err := os.CreateTemp("", "benchmark_output_*."+format)
	if err != nil {
		return nil, utils.WrapError(err, utils.ErrorTypeIO, utils.CodeIOFileCreate, "failed to create benchmark output file")
	}
	defer func() {
		if err := outputFile.Close(); err != nil {
			// Log error but don't fail the benchmark
			fmt.Printf("Warning: failed to close benchmark output file: %v\n", err)
		}
		if err := os.Remove(outputFile.Name()); err != nil {
			// Log error but don't fail the benchmark
			fmt.Printf("Warning: failed to remove benchmark output file: %v\n", err)
		}
	}()

	// Measure memory before
	var memBefore runtime.MemStats
	runtime.ReadMemStats(&memBefore)

	startTime := time.Now()

	// Run the full processing pipeline
	files, err := fileproc.CollectFiles(sourceDir)
	if err != nil {
		return nil, utils.WrapError(err, utils.ErrorTypeProcessing, utils.CodeProcessingCollection, "benchmark file collection failed")
	}

	// Process files with concurrency
	err = runProcessingPipeline(context.Background(), files, outputFile, format, concurrency, sourceDir)
	if err != nil {
		return nil, utils.WrapError(err, utils.ErrorTypeProcessing, utils.CodeProcessingFileRead, "benchmark processing pipeline failed")
	}

	duration := time.Since(startTime)

	// Measure memory after
	var memAfter runtime.MemStats
	runtime.ReadMemStats(&memAfter)

	// Calculate total bytes processed
	var totalBytes int64
	for _, file := range files {
		if info, err := os.Stat(file); err == nil {
			totalBytes += info.Size()
		}
	}

	result := &BenchmarkResult{
		Name:           fmt.Sprintf("FileProcessing_%s_c%d", format, concurrency),
		Duration:       duration,
		FilesProcessed: len(files),
		BytesProcessed: totalBytes,
		FilesPerSecond: float64(len(files)) / duration.Seconds(),
		BytesPerSecond: float64(totalBytes) / duration.Seconds(),
		MemoryUsage: MemoryStats{
			AllocMB:      float64(memAfter.Alloc-memBefore.Alloc) / 1024 / 1024,
			SysMB:        float64(memAfter.Sys-memBefore.Sys) / 1024 / 1024,
			NumGC:        memAfter.NumGC - memBefore.NumGC,
			PauseTotalNs: memAfter.PauseTotalNs - memBefore.PauseTotalNs,
		},
		CPUUsage: CPUStats{
			Goroutines: runtime.NumGoroutine(),
		},
	}

	return result, nil
}

// ConcurrencyBenchmark benchmarks different concurrency levels.
func ConcurrencyBenchmark(sourceDir string, format string, concurrencyLevels []int) (*BenchmarkSuite, error) {
	suite := &BenchmarkSuite{
		Name:    "ConcurrencyBenchmark",
		Results: make([]BenchmarkResult, 0, len(concurrencyLevels)),
	}

	for _, concurrency := range concurrencyLevels {
		result, err := FileProcessingBenchmark(sourceDir, format, concurrency)
		if err != nil {
			return nil, utils.WrapErrorf(err, utils.ErrorTypeProcessing, utils.CodeProcessingCollection, "concurrency benchmark failed for level %d", concurrency)
		}
		suite.Results = append(suite.Results, *result)
	}

	return suite, nil
}

// FormatBenchmark benchmarks different output formats.
func FormatBenchmark(sourceDir string, formats []string) (*BenchmarkSuite, error) {
	suite := &BenchmarkSuite{
		Name:    "FormatBenchmark",
		Results: make([]BenchmarkResult, 0, len(formats)),
	}

	for _, format := range formats {
		result, err := FileProcessingBenchmark(sourceDir, format, runtime.NumCPU())
		if err != nil {
			return nil, utils.WrapErrorf(err, utils.ErrorTypeProcessing, utils.CodeProcessingCollection, "format benchmark failed for format %s", format)
		}
		suite.Results = append(suite.Results, *result)
	}

	return suite, nil
}

// createBenchmarkFiles creates temporary files for benchmarking.
func createBenchmarkFiles(numFiles int) (string, func(), error) {
	tempDir, err := os.MkdirTemp("", "gibidify_benchmark_*")
	if err != nil {
		return "", nil, utils.WrapError(err, utils.ErrorTypeFileSystem, utils.CodeFSAccess, "failed to create temp directory")
	}

	cleanup := func() {
		if err := os.RemoveAll(tempDir); err != nil {
			// Log error but don't fail the benchmark
			fmt.Printf("Warning: failed to remove benchmark temp directory: %v\n", err)
		}
	}

	// Create various file types
	fileTypes := []struct {
		ext     string
		content string
	}{
		{".go", "package main\n\nfunc main() {\n\tprintln(\"Hello, World!\")\n}"},
		{".js", "console.log('Hello, World!');"},
		{".py", "print('Hello, World!')"},
		{".java", "public class Hello {\n\tpublic static void main(String[] args) {\n\t\tSystem.out.println(\"Hello, World!\");\n\t}\n}"},
		{".cpp", "#include <iostream>\n\nint main() {\n\tstd::cout << \"Hello, World!\" << std::endl;\n\treturn 0;\n}"},
		{".rs", "fn main() {\n\tprintln!(\"Hello, World!\");\n}"},
		{".rb", "puts 'Hello, World!'"},
		{".php", "<?php\necho 'Hello, World!';\n?>"},
		{".sh", "#!/bin/bash\necho 'Hello, World!'"},
		{".md", "# Hello, World!\n\nThis is a markdown file."},
	}

	for i := 0; i < numFiles; i++ {
		fileType := fileTypes[i%len(fileTypes)]
		filename := fmt.Sprintf("file_%d%s", i, fileType.ext)

		// Create subdirectories for some files
		if i%10 == 0 {
			subdir := filepath.Join(tempDir, fmt.Sprintf("subdir_%d", i/10))
			if err := os.MkdirAll(subdir, 0o750); err != nil {
				cleanup()
				return "", nil, utils.WrapError(err, utils.ErrorTypeFileSystem, utils.CodeFSAccess, "failed to create subdirectory")
			}
			filename = filepath.Join(subdir, filename)
		} else {
			filename = filepath.Join(tempDir, filename)
		}

		// Create file with repeated content to make it larger
		content := ""
		for j := 0; j < 10; j++ {
			content += fmt.Sprintf("// Line %d\n%s\n", j, fileType.content)
		}

		if err := os.WriteFile(filename, []byte(content), 0o600); err != nil {
			cleanup()
			return "", nil, utils.WrapError(err, utils.ErrorTypeIO, utils.CodeIOFileWrite, "failed to write benchmark file")
		}
	}

	return tempDir, cleanup, nil
}

// runProcessingPipeline runs the processing pipeline similar to main.go.
func runProcessingPipeline(ctx context.Context, files []string, outputFile *os.File, format string, concurrency int, sourceDir string) error {
	fileCh := make(chan string, concurrency)
	writeCh := make(chan fileproc.WriteRequest, concurrency)
	writerDone := make(chan struct{})

	// Start writer
	go fileproc.StartWriter(outputFile, writeCh, writerDone, format, "", "")

	// Get absolute path once
	absRoot, err := utils.GetAbsolutePath(sourceDir)
	if err != nil {
		return utils.WrapError(err, utils.ErrorTypeFileSystem, utils.CodeFSPathResolution, "failed to get absolute path for source directory")
	}

	// Start workers with proper synchronization
	var workersDone sync.WaitGroup
	for i := 0; i < concurrency; i++ {
		workersDone.Add(1)
		go func() {
			defer workersDone.Done()
			for filePath := range fileCh {
				fileproc.ProcessFile(filePath, writeCh, absRoot)
			}
		}()
	}

	// Send files to workers
	for _, file := range files {
		select {
		case <-ctx.Done():
			close(fileCh)
			workersDone.Wait() // Wait for workers to finish
			close(writeCh)
			<-writerDone
			return ctx.Err()
		case fileCh <- file:
		}
	}

	// Close file channel and wait for workers to finish
	close(fileCh)
	workersDone.Wait()

	// Now it's safe to close the write channel
	close(writeCh)
	<-writerDone

	return nil
}

// PrintBenchmarkResult prints a formatted benchmark result.
func PrintBenchmarkResult(result *BenchmarkResult) {
	fmt.Printf("=== %s ===\n", result.Name)
	fmt.Printf("Duration: %v\n", result.Duration)
	fmt.Printf("Files Processed: %d\n", result.FilesProcessed)
	fmt.Printf("Bytes Processed: %d (%.2f MB)\n", result.BytesProcessed, float64(result.BytesProcessed)/1024/1024)
	fmt.Printf("Files/sec: %.2f\n", result.FilesPerSecond)
	fmt.Printf("Bytes/sec: %.2f MB/sec\n", result.BytesPerSecond/1024/1024)
	fmt.Printf("Memory Usage: +%.2f MB (Sys: +%.2f MB)\n", result.MemoryUsage.AllocMB, result.MemoryUsage.SysMB)
	// Safe conversion: cap at MaxInt64 to prevent overflow
	pauseTotalNs := result.MemoryUsage.PauseTotalNs
	if pauseTotalNs > math.MaxInt64 {
		pauseTotalNs = math.MaxInt64
	}
	pauseDuration := time.Duration(int64(pauseTotalNs)) // #nosec G115 -- overflow check above
	fmt.Printf("GC Runs: %d (Pause: %v)\n", result.MemoryUsage.NumGC, pauseDuration)
	fmt.Printf("Goroutines: %d\n", result.CPUUsage.Goroutines)
	fmt.Println()
}

// PrintBenchmarkSuite prints all results in a benchmark suite.
func PrintBenchmarkSuite(suite *BenchmarkSuite) {
	fmt.Printf("=== %s ===\n", suite.Name)
	for _, result := range suite.Results {
		PrintBenchmarkResult(&result)
	}
}

// RunAllBenchmarks runs a comprehensive benchmark suite.
func RunAllBenchmarks(sourceDir string) error {
	fmt.Println("Running gibidify benchmark suite...")

	// Load configuration
	config.LoadConfig()

	// File collection benchmark
	fmt.Println("Running file collection benchmark...")
	result, err := FileCollectionBenchmark(sourceDir, 1000)
	if err != nil {
		return utils.WrapError(err, utils.ErrorTypeProcessing, utils.CodeProcessingCollection, "file collection benchmark failed")
	}
	PrintBenchmarkResult(result)

	// Format benchmarks
	fmt.Println("Running format benchmarks...")
	formatSuite, err := FormatBenchmark(sourceDir, []string{"json", "yaml", "markdown"})
	if err != nil {
		return utils.WrapError(err, utils.ErrorTypeProcessing, utils.CodeProcessingCollection, "format benchmark failed")
	}
	PrintBenchmarkSuite(formatSuite)

	// Concurrency benchmarks
	fmt.Println("Running concurrency benchmarks...")
	concurrencyLevels := []int{1, 2, 4, 8, runtime.NumCPU()}
	concurrencySuite, err := ConcurrencyBenchmark(sourceDir, "json", concurrencyLevels)
	if err != nil {
		return utils.WrapError(err, utils.ErrorTypeProcessing, utils.CodeProcessingCollection, "concurrency benchmark failed")
	}
	PrintBenchmarkSuite(concurrencySuite)

	return nil
}
