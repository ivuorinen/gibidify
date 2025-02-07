// Package main for the gibidify CLI application.
// Repository: github.com/ivuorinen/gibidify
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"runtime"
	"sync"

	"github.com/ivuorinen/gibidify/config"
	"github.com/ivuorinen/gibidify/fileproc"
	"github.com/schollz/progressbar/v3"
	"github.com/sirupsen/logrus"
)

var (
	sourceDir   string
	destination string
	prefix      string
	suffix      string
	concurrency int
)

func init() {
	flag.StringVar(&sourceDir, "source", "", "Source directory to scan recursively")
	flag.StringVar(&destination, "destination", "", "Output file to write aggregated code")
	flag.StringVar(&prefix, "prefix", "", "Text to add at the beginning of the output file")
	flag.StringVar(&suffix, "suffix", "", "Text to add at the end of the output file")
	flag.IntVar(&concurrency, "concurrency", runtime.NumCPU(), "Number of concurrent workers (default: number of CPU cores)")
}

// Run executes the main logic of the CLI application using the provided context.
func Run(ctx context.Context) error {
	flag.Parse()

	if sourceDir == "" || destination == "" {
		return fmt.Errorf(
			"usage: gibidify " +
				"-source <source_directory> " +
				"-destination <output_file> " +
				"[--prefix=\"...\"] " +
				"[--suffix=\"...\"] " +
				"[-concurrency=<num>]",
		)
	}

	// Load configuration using Viper.
	config.LoadConfig()

	logrus.Infof(
		"Starting gibidify. Source: %s, Destination: %s, Workers: %d",
		sourceDir,
		destination,
		concurrency,
	)

	// 1. Collect files using the file walker (ProdWalker).
	files, err := fileproc.CollectFiles(sourceDir)
	if err != nil {
		return fmt.Errorf("error collecting files: %w", err)
	}
	logrus.Infof("Found %d files to process", len(files))

	// 2. Open the destination file and write the header.
	outFile, err := os.Create(destination)
	if err != nil {
		return fmt.Errorf("failed to create output file %s: %w", destination, err)
	}
	defer func(outFile *os.File) {
		err := outFile.Close()
		if err != nil {
			logrus.Errorf("failed to close output file %s: %v", destination, err)
		}
	}(outFile)

	header := prefix + "\n" +
		"The following text is a Git repository with code. " +
		"The structure of the text are sections that begin with ----, " +
		"followed by a single line containing the file path and file name, " +
		"followed by a variable amount of lines containing the file contents. " +
		"The text representing the Git repository ends when the symbols --END-- are encountered.\n"

	if _, err := outFile.WriteString(header); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	// 3. Set up channels and a worker pool for processing files.
	fileCh := make(chan string)
	writeCh := make(chan fileproc.WriteRequest)
	var wg sync.WaitGroup

	// Start the writer goroutine.
	writerDone := make(chan struct{})
	go fileproc.StartWriter(outFile, writeCh, writerDone)

	// Start worker goroutines.
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case fp, ok := <-fileCh:
					if !ok {
						return
					}
					// Process the file.
					fileproc.ProcessFile(fp, writeCh, nil)
				case <-ctx.Done():
					return
				}
			}
		}()
	}

	// Feed file paths to the worker pool with progress bar feedback.
	bar := progressbar.Default(int64(len(files)))
loop:
	for _, fp := range files {
		select {
		case fileCh <- fp:
			_ = bar.Add(1)
		case <-ctx.Done():
			close(fileCh)
			break loop
		}
	}
	close(fileCh)

	// Wait for all workers to finish.
	wg.Wait()
	close(writeCh)
	<-writerDone

	// Check for context cancellation.
	if err := ctx.Err(); err != nil {
		return err
	}

	// 4. Write footer.
	footer := "--END--\n" + suffix
	if _, err := outFile.WriteString(footer); err != nil {
		return fmt.Errorf("failed to write footer: %w", err)
	}

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
