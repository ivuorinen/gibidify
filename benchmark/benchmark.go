// Package benchmark provides benchmarking infrastructure for gibidify.
package benchmark

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/ivuorinen/gibidify/config"
	"github.com/ivuorinen/gibidify/fileproc"
	"github.com/ivuorinen/gibidify/shared"
)

// Result represents the results of a benchmark run.
type Result struct {
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

// Suite represents a collection of benchmarks.
type Suite struct {
	Name    string
	Results []Result
}

// buildBenchmarkResult constructs a Result with all metrics calculated.
// This eliminates code duplication across benchmark functions.
func buildBenchmarkResult(
	name string,
	files []string,
	totalBytes int64,
	duration time.Duration,
	memBefore, memAfter runtime.MemStats,
) *Result {
	result := &Result{
		Name:           name,
		Duration:       duration,
		FilesProcessed: len(files),
		BytesProcessed: totalBytes,
	}

	// Calculate rates with zero-division guard
	secs := duration.Seconds()
	if secs == 0 {
		result.FilesPerSecond = 0
		result.BytesPerSecond = 0
	} else {
		result.FilesPerSecond = float64(len(files)) / secs
		result.BytesPerSecond = float64(totalBytes) / secs
	}

	result.MemoryUsage = MemoryStats{
		AllocMB:      shared.SafeMemoryDiffMB(memAfter.Alloc, memBefore.Alloc),
		SysMB:        shared.SafeMemoryDiffMB(memAfter.Sys, memBefore.Sys),
		NumGC:        memAfter.NumGC - memBefore.NumGC,
		PauseTotalNs: memAfter.PauseTotalNs - memBefore.PauseTotalNs,
	}

	result.CPUUsage = CPUStats{
		Goroutines: runtime.NumGoroutine(),
	}

	return result
}

// FileCollectionBenchmark benchmarks file collection operations.
func FileCollectionBenchmark(sourceDir string, numFiles int) (*Result, error) {
	// Load configuration to ensure proper file filtering
	config.LoadConfig()

	// Create temporary directory with test files if no source is provided
	var cleanup func()
	if sourceDir == "" {
		tempDir, cleanupFunc, err := createBenchmarkFiles(numFiles)
		if err != nil {
			return nil, shared.WrapError(
				err,
				shared.ErrorTypeFileSystem,
				shared.CodeFSAccess,
				shared.BenchmarkMsgFailedToCreateFiles,
			)
		}
		cleanup = cleanupFunc
		//nolint:errcheck // Benchmark output, errors don't affect results
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
		return nil, shared.WrapError(
			err,
			shared.ErrorTypeProcessing,
			shared.CodeProcessingCollection,
			shared.BenchmarkMsgCollectionFailed,
		)
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

	result := buildBenchmarkResult("FileCollection", files, totalBytes, duration, memBefore, memAfter)
	return result, nil
}

// FileProcessingBenchmark benchmarks full file processing pipeline.
func FileProcessingBenchmark(sourceDir string, format string, concurrency int) (*Result, error) {
	// Load configuration to ensure proper file filtering
	config.LoadConfig()

	var cleanup func()
	if sourceDir == "" {
		// Create temporary directory with test files
		tempDir, cleanupFunc, err := createBenchmarkFiles(shared.BenchmarkDefaultFileCount)
		if err != nil {
			return nil, shared.WrapError(
				err,
				shared.ErrorTypeFileSystem,
				shared.CodeFSAccess,
				shared.BenchmarkMsgFailedToCreateFiles,
			)
		}
		cleanup = cleanupFunc
		//nolint:errcheck // Benchmark output, errors don't affect results
		defer cleanup()
		sourceDir = tempDir
	}

	// Create temporary output file
	outputFile, err := os.CreateTemp("", "benchmark_output_*."+format)
	if err != nil {
		return nil, shared.WrapError(
			err,
			shared.ErrorTypeIO,
			shared.CodeIOFileCreate,
			"failed to create benchmark output file",
		)
	}
	defer func() {
		if err := outputFile.Close(); err != nil {
			//nolint:errcheck // Warning message in defer, failure doesn't affect benchmark
			_, _ = fmt.Printf("Warning: failed to close benchmark output file: %v\n", err)
		}
		if err := os.Remove(outputFile.Name()); err != nil {
			//nolint:errcheck // Warning message in defer, failure doesn't affect benchmark
			_, _ = fmt.Printf("Warning: failed to remove benchmark output file: %v\n", err)
		}
	}()

	// Measure memory before
	var memBefore runtime.MemStats
	runtime.ReadMemStats(&memBefore)

	startTime := time.Now()

	// Run the full processing pipeline
	files, err := fileproc.CollectFiles(sourceDir)
	if err != nil {
		return nil, shared.WrapError(
			err,
			shared.ErrorTypeProcessing,
			shared.CodeProcessingCollection,
			shared.BenchmarkMsgCollectionFailed,
		)
	}

	// Process files with concurrency
	err = runProcessingPipeline(context.Background(), files, outputFile, format, concurrency, sourceDir)
	if err != nil {
		return nil, shared.WrapError(
			err,
			shared.ErrorTypeProcessing,
			shared.CodeProcessingFileRead,
			"benchmark processing pipeline failed",
		)
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

	benchmarkName := fmt.Sprintf("FileProcessing_%s_c%d", format, concurrency)
	result := buildBenchmarkResult(benchmarkName, files, totalBytes, duration, memBefore, memAfter)
	return result, nil
}

// ConcurrencyBenchmark benchmarks different concurrency levels.
func ConcurrencyBenchmark(sourceDir string, format string, concurrencyLevels []int) (*Suite, error) {
	suite := &Suite{
		Name:    "ConcurrencyBenchmark",
		Results: make([]Result, 0, len(concurrencyLevels)),
	}

	for _, concurrency := range concurrencyLevels {
		result, err := FileProcessingBenchmark(sourceDir, format, concurrency)
		if err != nil {
			return nil, shared.WrapErrorf(
				err,
				shared.ErrorTypeProcessing,
				shared.CodeProcessingCollection,
				"concurrency benchmark failed for level %d",
				concurrency,
			)
		}
		suite.Results = append(suite.Results, *result)
	}

	return suite, nil
}

// FormatBenchmark benchmarks different output formats.
func FormatBenchmark(sourceDir string, formats []string) (*Suite, error) {
	suite := &Suite{
		Name:    "FormatBenchmark",
		Results: make([]Result, 0, len(formats)),
	}

	for _, format := range formats {
		result, err := FileProcessingBenchmark(sourceDir, format, runtime.NumCPU())
		if err != nil {
			return nil, shared.WrapErrorf(
				err,
				shared.ErrorTypeProcessing,
				shared.CodeProcessingCollection,
				"format benchmark failed for format %s",
				format,
			)
		}
		suite.Results = append(suite.Results, *result)
	}

	return suite, nil
}

// createBenchmarkFiles creates temporary files for benchmarking.
func createBenchmarkFiles(numFiles int) (string, func(), error) {
	tempDir, err := os.MkdirTemp("", "gibidify_benchmark_*")
	if err != nil {
		return "", nil, shared.WrapError(
			err,
			shared.ErrorTypeFileSystem,
			shared.CodeFSAccess,
			"failed to create temp directory",
		)
	}

	cleanup := func() {
		if err := os.RemoveAll(tempDir); err != nil {
			//nolint:errcheck // Warning message in cleanup, failure doesn't affect benchmark
			_, _ = fmt.Printf("Warning: failed to remove benchmark temp directory: %v\n", err)
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
		{
			".java",
			"public class Hello {\n\tpublic static void main(String[] args) {\n\t" +
				"\tSystem.out.println(\"Hello, World!\");\n\t}\n}",
		},
		{
			".cpp",
			"#include <iostream>\n\n" +
				"int main() {\n\tstd::cout << \"Hello, World!\" << std::endl;\n\treturn 0;\n}",
		},
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

				return "", nil, shared.WrapError(
					err,
					shared.ErrorTypeFileSystem,
					shared.CodeFSAccess,
					"failed to create subdirectory",
				)
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

			return "", nil, shared.WrapError(
				err, shared.ErrorTypeIO, shared.CodeIOFileWrite, "failed to write benchmark file",
			)
		}
	}

	return tempDir, cleanup, nil
}

// runProcessingPipeline runs the processing pipeline similar to main.go.
func runProcessingPipeline(
	ctx context.Context,
	files []string,
	outputFile *os.File,
	format string,
	concurrency int,
	sourceDir string,
) error {
	// Guard against invalid concurrency to prevent deadlocks
	if concurrency < 1 {
		concurrency = 1
	}

	fileCh := make(chan string, concurrency)
	writeCh := make(chan fileproc.WriteRequest, concurrency)
	writerDone := make(chan struct{})

	// Start writer
	go fileproc.StartWriter(outputFile, writeCh, writerDone, format, "", "")

	// Get absolute path once
	absRoot, err := shared.AbsolutePath(sourceDir)
	if err != nil {
		return shared.WrapError(
			err,
			shared.ErrorTypeFileSystem,
			shared.CodeFSPathResolution,
			"failed to get absolute path for source directory",
		)
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

			return fmt.Errorf("context canceled: %w", ctx.Err())
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

// PrintResult prints a formatted benchmark result.
func PrintResult(result *Result) {
	printBenchmarkLine := func(format string, args ...any) {
		if _, err := fmt.Printf(format, args...); err != nil {
			// Stdout write errors are rare (broken pipe, etc.) - log but continue
			shared.LogError("failed to write benchmark output", err)
		}
	}

	printBenchmarkLine(shared.BenchmarkFmtSectionHeader, result.Name)
	printBenchmarkLine("Duration: %v\n", result.Duration)
	printBenchmarkLine("Files Processed: %d\n", result.FilesProcessed)
	printBenchmarkLine("Bytes Processed: %d (%.2f MB)\n", result.BytesProcessed,
		float64(result.BytesProcessed)/float64(shared.BytesPerMB))
	printBenchmarkLine("Files/sec: %.2f\n", result.FilesPerSecond)
	printBenchmarkLine("Bytes/sec: %.2f MB/sec\n", result.BytesPerSecond/float64(shared.BytesPerMB))
	printBenchmarkLine(
		"Memory Usage: +%.2f MB (Sys: +%.2f MB)\n",
		result.MemoryUsage.AllocMB,
		result.MemoryUsage.SysMB,
	)
	//nolint:errcheck // Overflow unlikely for pause duration, result output only
	pauseDuration, _ := shared.SafeUint64ToInt64(result.MemoryUsage.PauseTotalNs)
	printBenchmarkLine("GC Runs: %d (Pause: %v)\n", result.MemoryUsage.NumGC, time.Duration(pauseDuration))
	printBenchmarkLine("Goroutines: %d\n", result.CPUUsage.Goroutines)
	printBenchmarkLine("\n")
}

// PrintSuite prints all results in a benchmark suite.
func PrintSuite(suite *Suite) {
	if _, err := fmt.Printf(shared.BenchmarkFmtSectionHeader, suite.Name); err != nil {
		shared.LogError("failed to write benchmark suite header", err)
	}
	// Iterate by index to avoid taking address of range variable
	for i := range suite.Results {
		PrintResult(&suite.Results[i])
	}
}

// RunAllBenchmarks runs a comprehensive benchmark suite.
func RunAllBenchmarks(sourceDir string) error {
	printBenchmark := func(msg string) {
		if _, err := fmt.Println(msg); err != nil {
			shared.LogError("failed to write benchmark message", err)
		}
	}

	printBenchmark("Running gibidify benchmark suite...")

	// Load configuration
	config.LoadConfig()

	// File collection benchmark
	printBenchmark(shared.BenchmarkMsgRunningCollection)
	result, err := FileCollectionBenchmark(sourceDir, shared.BenchmarkDefaultFileCount)
	if err != nil {
		return shared.WrapError(
			err,
			shared.ErrorTypeProcessing,
			shared.CodeProcessingCollection,
			shared.BenchmarkMsgFileCollectionFailed,
		)
	}
	PrintResult(result)

	// Format benchmarks
	printBenchmark("Running format benchmarks...")
	formats := []string{shared.FormatJSON, shared.FormatYAML, shared.FormatMarkdown}
	formatSuite, err := FormatBenchmark(sourceDir, formats)
	if err != nil {
		return shared.WrapError(
			err,
			shared.ErrorTypeProcessing,
			shared.CodeProcessingCollection,
			shared.BenchmarkMsgFormatFailed,
		)
	}
	PrintSuite(formatSuite)

	// Concurrency benchmarks
	printBenchmark("Running concurrency benchmarks...")
	concurrencyLevels := []int{1, 2, 4, 8, runtime.NumCPU()}
	concurrencySuite, err := ConcurrencyBenchmark(sourceDir, shared.FormatJSON, concurrencyLevels)
	if err != nil {
		return shared.WrapError(
			err,
			shared.ErrorTypeProcessing,
			shared.CodeProcessingCollection,
			shared.BenchmarkMsgConcurrencyFailed,
		)
	}
	PrintSuite(concurrencySuite)

	return nil
}
