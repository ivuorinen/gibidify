package cli

import (
	"context"
	"os"
	"sync"

	"github.com/ivuorinen/gibidify/fileproc"
	"github.com/ivuorinen/gibidify/utils"
)

// Process executes the main file processing workflow.
func (p *Processor) Process(ctx context.Context) error {
	// Create overall processing context with timeout
	overallCtx, overallCancel := p.resourceMonitor.CreateOverallProcessingContext(ctx)
	defer overallCancel()

	// Configure file type registry
	p.configureFileTypes()

	// Print startup info with colors
	p.ui.PrintHeader("🚀 Starting gibidify")
	p.ui.PrintInfo("Format: %s", p.flags.Format)
	p.ui.PrintInfo("Source: %s", p.flags.SourceDir)
	p.ui.PrintInfo("Destination: %s", p.flags.Destination)
	p.ui.PrintInfo("Workers: %d", p.flags.Concurrency)

	// Log resource monitoring configuration
	p.resourceMonitor.LogResourceInfo()
	p.backpressure.LogBackpressureInfo()

	// Collect files with progress indication
	p.ui.PrintInfo("📁 Collecting files...")
	files, err := p.collectFiles()
	if err != nil {
		return err
	}

	// Show collection results
	p.ui.PrintSuccess("Found %d files to process", len(files))

	// Pre-validate file collection against resource limits
	if err := p.validateFileCollection(files); err != nil {
		return err
	}

	// Process files with overall timeout
	return p.processFiles(overallCtx, files)
}

// processFiles processes the collected files.
func (p *Processor) processFiles(ctx context.Context, files []string) error {
	outFile, err := p.createOutputFile()
	if err != nil {
		return err
	}
	defer func() {
		utils.LogError("Error closing output file", outFile.Close())
	}()

	// Initialize back-pressure and channels
	p.ui.PrintInfo("⚙️  Initializing processing...")
	p.backpressure.LogBackpressureInfo()
	fileCh, writeCh := p.backpressure.CreateChannels()
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

	p.logFinalStats()
	p.ui.PrintSuccess("Processing completed. Output saved to %s", p.flags.Destination)
	return nil
}

// createOutputFile creates the output file.
func (p *Processor) createOutputFile() (*os.File, error) {
	// Destination path has been validated in CLI flags validation for path traversal attempts
	outFile, err := os.Create(p.flags.Destination) // #nosec G304 - destination is validated in flags.validate()
	if err != nil {
		return nil, utils.WrapError(err, utils.ErrorTypeIO, utils.CodeIOFileCreate, "failed to create output file").WithFilePath(p.flags.Destination)
	}
	return outFile, nil
}
