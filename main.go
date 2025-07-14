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
	flag.StringVar(&format, "format", "markdown", "Output format (json, markdown, yaml)")
	flag.IntVar(&concurrency, "concurrency", runtime.NumCPU(), "Number of concurrent workers (default: number of CPU cores)")
}

func main() {
	// In production, use a background context.
	if err := run(context.Background()); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}

// Run executes the main logic of the CLI application using the provided context.
func run(ctx context.Context) error {
	flag.Parse()

	if err := validateFlags(); err != nil {
		return err
	}

	if err := setDestination(); err != nil {
		return err
	}

	config.LoadConfig()

	logrus.Infof(
		"Starting gibidify. Format: %s, Source: %s, Destination: %s, Workers: %d",
		format,
		sourceDir,
		destination,
		concurrency,
	)

	files, err := fileproc.CollectFiles(sourceDir)
	if err != nil {
		return fmt.Errorf("error collecting files: %w", err)
	}
	logrus.Infof("Found %d files to process", len(files))

	outFile, err := os.Create(destination)
	if err != nil {
		return fmt.Errorf("failed to create output file %s: %w", destination, err)
	}
	defer func(outFile *os.File) {
		if err := outFile.Close(); err != nil {
			logrus.Errorf("Error closing output file: %v", err)
		}
	}(outFile)

	fileCh := make(chan string)
	writeCh := make(chan fileproc.WriteRequest)
	writerDone := make(chan struct{})

	go fileproc.StartWriter(outFile, writeCh, writerDone, format, prefix, suffix)

	var wg sync.WaitGroup

	startWorkers(ctx, &wg, fileCh, writeCh)

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
func validateFlags() error {
	if sourceDir == "" {
		return fmt.Errorf("usage: gibidify -source <source_directory> [--destination <output_file>] [--format=json|yaml|markdown] ")
	}
	return nil
}

func setDestination() error {
	if destination == "" {
		absRoot, err := filepath.Abs(sourceDir)
		if err != nil {
			return fmt.Errorf("failed to get absolute path for %s: %w", sourceDir, err)
		}
		baseName := filepath.Base(absRoot)
		if baseName == "." || baseName == "" {
			baseName = "output"
		}
		destination = baseName + "." + format
	}
	return nil
}

func startWorkers(ctx context.Context, wg *sync.WaitGroup, fileCh chan string, writeCh chan fileproc.WriteRequest) {
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
}
