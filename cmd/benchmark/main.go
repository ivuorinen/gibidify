// Package main provides a CLI for running gibidify benchmarks.
package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/ivuorinen/gibidify/benchmark"
	"github.com/ivuorinen/gibidify/utils"
)

var (
	sourceDir       = flag.String("source", "", "Source directory to benchmark (uses temp files if empty)")
	benchmarkType   = flag.String("type", "all", "Benchmark type: all, collection, processing, concurrency, format")
	format          = flag.String("format", "json", "Output format for processing benchmarks")
	concurrency     = flag.Int("concurrency", runtime.NumCPU(), "Concurrency level for processing benchmarks")
	concurrencyList = flag.String("concurrency-list", "1,2,4,8", "Comma-separated list of concurrency levels")
	formatList      = flag.String("format-list", "json,yaml,markdown", "Comma-separated list of formats")
	numFiles        = flag.Int("files", 100, "Number of files to create for benchmarks")
)

func main() {
	flag.Parse()

	if err := runBenchmarks(); err != nil {
		fmt.Fprintf(os.Stderr, "Benchmark failed: %v\n", err)
		os.Exit(1)
	}
}

func runBenchmarks() error {
	fmt.Printf("Running gibidify benchmarks...\n")
	fmt.Printf("Source: %s\n", getSourceDescription())
	fmt.Printf("Type: %s\n", *benchmarkType)
	fmt.Printf("CPU cores: %d\n", runtime.NumCPU())
	fmt.Println()

	switch *benchmarkType {
	case "all":
		return benchmark.RunAllBenchmarks(*sourceDir)
	case "collection":
		return runCollectionBenchmark()
	case "processing":
		return runProcessingBenchmark()
	case "concurrency":
		return runConcurrencyBenchmark()
	case "format":
		return runFormatBenchmark()
	default:
		return utils.NewValidationError(utils.CodeValidationFormat, "invalid benchmark type: "+*benchmarkType)
	}
}

func runCollectionBenchmark() error {
	fmt.Println("Running file collection benchmark...")
	result, err := benchmark.FileCollectionBenchmark(*sourceDir, *numFiles)
	if err != nil {
		return utils.WrapError(err, utils.ErrorTypeProcessing, utils.CodeProcessingCollection, "file collection benchmark failed")
	}
	benchmark.PrintBenchmarkResult(result)
	return nil
}

func runProcessingBenchmark() error {
	fmt.Printf("Running file processing benchmark (format: %s, concurrency: %d)...\n", *format, *concurrency)
	result, err := benchmark.FileProcessingBenchmark(*sourceDir, *format, *concurrency)
	if err != nil {
		return utils.WrapError(err, utils.ErrorTypeProcessing, utils.CodeProcessingCollection, "file processing benchmark failed")
	}
	benchmark.PrintBenchmarkResult(result)
	return nil
}

func runConcurrencyBenchmark() error {
	concurrencyLevels, err := parseConcurrencyList(*concurrencyList)
	if err != nil {
		return utils.WrapError(err, utils.ErrorTypeValidation, utils.CodeValidationFormat, "invalid concurrency list")
	}

	fmt.Printf("Running concurrency benchmark (format: %s, levels: %v)...\n", *format, concurrencyLevels)
	suite, err := benchmark.ConcurrencyBenchmark(*sourceDir, *format, concurrencyLevels)
	if err != nil {
		return utils.WrapError(err, utils.ErrorTypeProcessing, utils.CodeProcessingCollection, "concurrency benchmark failed")
	}
	benchmark.PrintBenchmarkSuite(suite)
	return nil
}

func runFormatBenchmark() error {
	formats := parseFormatList(*formatList)
	fmt.Printf("Running format benchmark (formats: %v)...\n", formats)
	suite, err := benchmark.FormatBenchmark(*sourceDir, formats)
	if err != nil {
		return utils.WrapError(err, utils.ErrorTypeProcessing, utils.CodeProcessingCollection, "format benchmark failed")
	}
	benchmark.PrintBenchmarkSuite(suite)
	return nil
}

func getSourceDescription() string {
	if *sourceDir == "" {
		return fmt.Sprintf("temporary files (%d files)", *numFiles)
	}
	return *sourceDir
}

func parseConcurrencyList(list string) ([]int, error) {
	parts := strings.Split(list, ",")
	levels := make([]int, 0, len(parts))

	for _, part := range parts {
		part = strings.TrimSpace(part)
		var level int
		if _, err := fmt.Sscanf(part, "%d", &level); err != nil {
			return nil, utils.WrapErrorf(err, utils.ErrorTypeValidation, utils.CodeValidationFormat, "invalid concurrency level: %s", part)
		}
		if level <= 0 {
			return nil, utils.NewValidationError(utils.CodeValidationFormat, "concurrency level must be positive: "+part)
		}
		levels = append(levels, level)
	}

	if len(levels) == 0 {
		return nil, utils.NewValidationError(utils.CodeValidationFormat, "no valid concurrency levels found")
	}

	return levels, nil
}

func parseFormatList(list string) []string {
	parts := strings.Split(list, ",")
	formats := make([]string, 0, len(parts))

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			formats = append(formats, part)
		}
	}

	return formats
}
