// Package main provides a CLI for running gibidify benchmarks.
package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/ivuorinen/gibidify/benchmark"
	"github.com/ivuorinen/gibidify/shared"
)

var (
	sourceDir = flag.String(
		shared.CLIArgSource, "", "Source directory to benchmark (uses temp files if empty)",
	)
	benchmarkType = flag.String(
		"type", shared.CLIArgAll, "Benchmark type: all, collection, processing, concurrency, format",
	)
	format = flag.String(
		shared.CLIArgFormat, shared.FormatJSON, "Output format for processing benchmarks",
	)
	concurrency = flag.Int(
		shared.CLIArgConcurrency, runtime.NumCPU(), "Concurrency level for processing benchmarks",
	)
	concurrencyList = flag.String(
		"concurrency-list", shared.TestConcurrencyList, "Comma-separated list of concurrency levels",
	)
	formatList = flag.String(
		"format-list", shared.TestFormatList, "Comma-separated list of formats",
	)
	numFiles = flag.Int("files", shared.BenchmarkDefaultFileCount, "Number of files to create for benchmarks")
)

func main() {
	flag.Parse()

	if err := runBenchmarks(); err != nil {
		//goland:noinspection GoUnhandledErrorResult
		_, _ = fmt.Fprintf(os.Stderr, "Benchmark failed: %v\n", err)
		os.Exit(1)
	}
}

func runBenchmarks() error {
	//nolint:errcheck // Benchmark informational output, errors don't affect benchmark results
	_, _ = fmt.Println("Running gibidify benchmarks...")
	//nolint:errcheck // Benchmark informational output, errors don't affect benchmark results
	_, _ = fmt.Printf("Source: %s\n", getSourceDescription())
	//nolint:errcheck // Benchmark informational output, errors don't affect benchmark results
	_, _ = fmt.Printf("Type: %s\n", *benchmarkType)
	//nolint:errcheck // Benchmark informational output, errors don't affect benchmark results
	_, _ = fmt.Printf("CPU cores: %d\n", runtime.NumCPU())
	//nolint:errcheck // Benchmark informational output, errors don't affect benchmark results
	_, _ = fmt.Println()

	switch *benchmarkType {
	case shared.CLIArgAll:
		if err := benchmark.RunAllBenchmarks(*sourceDir); err != nil {
			return fmt.Errorf("benchmark failed: %w", err)
		}

		return nil
	case "collection":
		return runCollectionBenchmark()
	case "processing":
		return runProcessingBenchmark()
	case "concurrency":
		return runConcurrencyBenchmark()
	case "format":
		return runFormatBenchmark()
	default:
		return shared.NewValidationError(shared.CodeValidationFormat, "invalid benchmark type: "+*benchmarkType)
	}
}

func runCollectionBenchmark() error {
	//nolint:errcheck // Benchmark status message, errors don't affect benchmark results
	_, _ = fmt.Println(shared.BenchmarkMsgRunningCollection)
	result, err := benchmark.FileCollectionBenchmark(*sourceDir, *numFiles)
	if err != nil {
		return shared.WrapError(
			err,
			shared.ErrorTypeProcessing,
			shared.CodeProcessingCollection,
			shared.BenchmarkMsgFileCollectionFailed,
		)
	}
	benchmark.PrintResult(result)

	return nil
}

func runProcessingBenchmark() error {
	//nolint:errcheck // Benchmark status message, errors don't affect benchmark results
	_, _ = fmt.Printf("Running file processing benchmark (format: %s, concurrency: %d)...\n", *format, *concurrency)
	result, err := benchmark.FileProcessingBenchmark(*sourceDir, *format, *concurrency)
	if err != nil {
		return shared.WrapError(
			err,
			shared.ErrorTypeProcessing,
			shared.CodeProcessingCollection,
			"file processing benchmark failed",
		)
	}
	benchmark.PrintResult(result)

	return nil
}

func runConcurrencyBenchmark() error {
	concurrencyLevels, err := parseConcurrencyList(*concurrencyList)
	if err != nil {
		return shared.WrapError(
			err, shared.ErrorTypeValidation, shared.CodeValidationFormat, "invalid concurrency list")
	}

	//nolint:errcheck // Benchmark status message, errors don't affect benchmark results
	_, _ = fmt.Printf("Running concurrency benchmark (format: %s, levels: %v)...\n", *format, concurrencyLevels)
	suite, err := benchmark.ConcurrencyBenchmark(*sourceDir, *format, concurrencyLevels)
	if err != nil {
		return shared.WrapError(
			err,
			shared.ErrorTypeProcessing,
			shared.CodeProcessingCollection,
			shared.BenchmarkMsgConcurrencyFailed,
		)
	}
	benchmark.PrintSuite(suite)

	return nil
}

func runFormatBenchmark() error {
	formats := parseFormatList(*formatList)
	//nolint:errcheck // Benchmark status message, errors don't affect benchmark results
	_, _ = fmt.Printf("Running format benchmark (formats: %v)...\n", formats)
	suite, err := benchmark.FormatBenchmark(*sourceDir, formats)
	if err != nil {
		return shared.WrapError(
			err, shared.ErrorTypeProcessing, shared.CodeProcessingCollection, shared.BenchmarkMsgFormatFailed,
		)
	}
	benchmark.PrintSuite(suite)

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
			return nil, shared.WrapErrorf(
				err,
				shared.ErrorTypeValidation,
				shared.CodeValidationFormat,
				"invalid concurrency level: %s",
				part,
			)
		}
		if level <= 0 {
			return nil, shared.NewValidationError(
				shared.CodeValidationFormat, "concurrency level must be positive: "+part,
			)
		}
		levels = append(levels, level)
	}

	if len(levels) == 0 {
		return nil, shared.NewValidationError(shared.CodeValidationFormat, "no valid concurrency levels found")
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
