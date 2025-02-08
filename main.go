// Package main for the gibidify CLI application.
// Repository: github.com/ivuorinen/gibidify
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/ivuorinen/gibidify/config"
	"github.com/ivuorinen/gibidify/fileproc"
	"github.com/sirupsen/logrus"
)

var (
	sourceDir   string
	destination string
	prefix      string
	suffix      string
	concurrency int
	format      string
)

func init() {
	flag.StringVar(&sourceDir, "source", "", "Source directory to scan recursively")
	flag.StringVar(&destination, "destination", "", "Output file to write aggregated code")
	flag.StringVar(&prefix, "prefix", "", "Text to add at the beginning of the output file")
	flag.StringVar(&suffix, "suffix", "", "Text to add at the end of the output file")
	flag.StringVar(&format, "format", "json", "Output format (json, markdown, yaml)")
	flag.IntVar(&concurrency, "concurrency", runtime.NumCPU(), "Number of concurrent workers (default: number of CPU cores)")
}

// Run executes the main logic of the CLI application using the provided context.
func Run(ctx context.Context) error {
	flag.Parse()

	// We need at least a source directory
	if sourceDir == "" {
		return fmt.Errorf("usage: gibidify -source <source_directory> [--destination <output_file>] [--format=json|yaml|markdown] ")
	}

	// If destination is not specified, auto-generate it using the base name of sourceDir + "." + format
	if destination == "" {
		absRoot, err := filepath.Abs(sourceDir)
		if err != nil {
			return fmt.Errorf("failed to get absolute path for %s: %w", sourceDir, err)
		}
		baseName := filepath.Base(absRoot)
		// If sourceDir ends with a slash, baseName might be "." so handle that case as needed
		if baseName == "." || baseName == "" {
			baseName = "output"
		}
		destination = baseName + "." + format
	}

	config.LoadConfig()

	logrus.Infof("Starting gibidify. Format: %s, Source: %s, Destination: %s, Workers: %d", format, sourceDir, destination, concurrency)

	// Collect files
	files, err := fileproc.CollectFiles(sourceDir)
	if err != nil {
		return fmt.Errorf("error collecting files: %w", err)
	}
	logrus.Infof("Found %d files to process", len(files))

	// Open output file
	outFile, err := os.Create(destination)
	if err != nil {
		return fmt.Errorf("failed to create output file %s: %w", destination, err)
	}
	defer func(outFile *os.File) {
		if err := outFile.Close(); err != nil {
			logrus.Errorf("Error closing output file: %v", err)
		}
	}(outFile)

	// Create channels
	fileCh := make(chan string)
	writeCh := make(chan fileproc.WriteRequest)
	writerDone := make(chan struct{})

	// Start writer goroutine
	go fileproc.StartWriter(outFile, writeCh, writerDone, format, prefix, suffix)

	var wg sync.WaitGroup

	// Start worker goroutines with context cancellation
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case filePath, ok := <-fileCh:
					if !ok {
						return
					}
					// Pass sourceDir to ProcessFile so it knows the 'root'
					absRoot, err := filepath.Abs(sourceDir)
					if err != nil {
						logrus.Errorf("Failed to get absolute path for %s: %v", sourceDir, err)
						return
					}
					fileproc.ProcessFile(filePath, writeCh, absRoot)
				}
			}
		}()
	}

	// Feed files to worker pool while checking for cancellation
	for _, fp := range files {
		select {
		case <-ctx.Done():
			close(fileCh)
			return ctx.Err()
		case fileCh <- fp:
		}
	}
	close(fileCh)

	wg.Wait()
	close(writeCh)
	<-writerDone

	logrus.Infof("Processing completed. Output saved to %s", destination)
	return nil
}

func main() {
	// In production, use a background context.
	if err := Run(context.Background()); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}
