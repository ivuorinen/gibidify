// Package cli provides command-line interface functionality for gibidify.
package cli

import (
	"context"
	"os"
	"sync"
	"time"

	"github.com/ivuorinen/gibidify/fileproc"
	"github.com/ivuorinen/gibidify/shared"
)

// channelBuffer bounds the file and write channels. A small multiple of the
// worker count keeps producers slightly ahead of consumers without unbounded
// memory growth. ponytail: fixed heuristic, revisit only if profiling shows stalls.
const channelBuffer = 64

// Process executes the main file processing workflow.
func (p *Processor) Process(ctx context.Context) error {
	start := time.Now()

	// Configure file type registry
	p.configureFileTypes()

	// Print startup info with colors
	p.ui.PrintHeader("🚀 Starting gibidify")
	p.ui.PrintInfo("Format: %s", p.flags.Format)
	p.ui.PrintInfo("Source: %s", p.flags.SourceDir)
	p.ui.PrintInfo("Destination: %s", p.flags.Destination)
	p.ui.PrintInfo("Workers: %d", p.flags.Concurrency)

	// Collect files with progress indication
	p.ui.PrintInfo("📁 Collecting files...")
	files, err := p.collectFiles()
	if err != nil {
		return err
	}

	// Show collection results
	p.ui.PrintSuccess(shared.CLIMsgFoundFilesToProcess, len(files))

	// Pre-validate file collection against resource limits
	if err := p.validateFileCollection(files); err != nil {
		return err
	}

	// Process files
	if err := p.processFiles(ctx, files); err != nil {
		return err
	}

	shared.GetLogger().Infof("Processed %d files in %s", len(files), time.Since(start).Round(time.Millisecond))

	return nil
}

// processFiles processes the collected files.
func (p *Processor) processFiles(ctx context.Context, files []string) error {
	outFile, err := p.createOutputFile()
	if err != nil {
		return err
	}
	defer func() {
		shared.LogError("Error closing output file", outFile.Close())
	}()

	// Initialize channels
	p.ui.PrintInfo("⚙️  Initializing processing...")
	fileCh := make(chan string, channelBuffer)
	writeCh := make(chan fileproc.WriteRequest, channelBuffer)
	writerDone := make(chan struct{})

	// Start writer
	go fileproc.StartWriter(outFile, writeCh, writerDone, p.flags.Format, p.flags.Prefix, p.flags.Suffix)

	// Start workers
	var wg sync.WaitGroup
	p.startWorkers(ctx, &wg, fileCh, writeCh)

	// Start progress bar
	p.ui.StartProgress(len(files), "📝 Processing files")

	// Send files to workers
	if err := p.sendFiles(ctx, files, fileCh); err != nil {
		p.ui.FinishProgress()

		return err
	}

	// Wait for completion
	p.waitForCompletion(&wg, writeCh, writerDone)
	p.ui.FinishProgress()

	p.ui.PrintSuccess("Processing completed. Output saved to %s", p.flags.Destination)

	return nil
}

// createOutputFile creates the output file.
func (p *Processor) createOutputFile() (*os.File, error) {
	// Destination path has been validated in CLI flags validation for path traversal attempts
	outFile, err := os.Create(p.flags.Destination) // #nosec G304 - destination is validated in flags.validate()
	if err != nil {
		return nil, shared.WrapError(
			err,
			shared.ErrorTypeIO,
			shared.CodeIOFileCreate,
			"failed to create output file",
		).WithFilePath(p.flags.Destination)
	}

	return outFile, nil
}
