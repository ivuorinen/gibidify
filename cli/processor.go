package cli

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/sirupsen/logrus"

	"github.com/ivuorinen/gibidify/config"
	"github.com/ivuorinen/gibidify/fileproc"
	"github.com/ivuorinen/gibidify/utils"
)

// Processor handles the main file processing logic.
type Processor struct {
	flags           *Flags
	backpressure    *fileproc.BackpressureManager
	resourceMonitor *fileproc.ResourceMonitor
	ui              *UIManager
}

// NewProcessor creates a new processor with the given flags.
func NewProcessor(flags *Flags) *Processor {
	ui := NewUIManager()

	// Configure UI based on flags
	ui.SetColorOutput(!flags.NoColors)
	ui.SetProgressOutput(!flags.NoProgress)

	return &Processor{
		flags:           flags,
		backpressure:    fileproc.NewBackpressureManager(),
		resourceMonitor: fileproc.NewResourceMonitor(),
		ui:              ui,
	}
}

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

// configureFileTypes configures the file type registry.
func (p *Processor) configureFileTypes() {
	if config.GetFileTypesEnabled() {
		fileproc.ConfigureFromSettings(
			config.GetCustomImageExtensions(),
			config.GetCustomBinaryExtensions(),
			config.GetCustomLanguages(),
			config.GetDisabledImageExtensions(),
			config.GetDisabledBinaryExtensions(),
			config.GetDisabledLanguageExtensions(),
		)
	}
}

// collectFiles collects all files to be processed.
func (p *Processor) collectFiles() ([]string, error) {
	files, err := fileproc.CollectFiles(p.flags.SourceDir)
	if err != nil {
		return nil, utils.WrapError(err, utils.ErrorTypeProcessing, utils.CodeProcessingCollection, "error collecting files")
	}
	logrus.Infof("Found %d files to process", len(files))
	return files, nil
}

// validateFileCollection validates the collected files against resource limits.
func (p *Processor) validateFileCollection(files []string) error {
	if !config.GetResourceLimitsEnabled() {
		return nil
	}

	// Check file count limit
	maxFiles := config.GetMaxFiles()
	if len(files) > maxFiles {
		return utils.NewStructuredError(
			utils.ErrorTypeValidation,
			utils.CodeResourceLimitFiles,
			fmt.Sprintf("file count (%d) exceeds maximum limit (%d)", len(files), maxFiles),
			"",
			map[string]interface{}{
				"file_count": len(files),
				"max_files":  maxFiles,
			},
		)
	}

	// Check total size limit (estimate)
	maxTotalSize := config.GetMaxTotalSize()
	totalSize := int64(0)
	oversizedFiles := 0

	for _, filePath := range files {
		if fileInfo, err := os.Stat(filePath); err == nil {
			totalSize += fileInfo.Size()
			if totalSize > maxTotalSize {
				return utils.NewStructuredError(
					utils.ErrorTypeValidation,
					utils.CodeResourceLimitTotalSize,
					fmt.Sprintf("total file size (%d bytes) would exceed maximum limit (%d bytes)", totalSize, maxTotalSize),
					"",
					map[string]interface{}{
						"total_size":     totalSize,
						"max_total_size": maxTotalSize,
						"files_checked":  len(files),
					},
				)
			}
		} else {
			oversizedFiles++
		}
	}

	if oversizedFiles > 0 {
		logrus.Warnf("Could not stat %d files during pre-validation", oversizedFiles)
	}

	logrus.Infof("Pre-validation passed: %d files, %d MB total", len(files), totalSize/1024/1024)
	return nil
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

// startWorkers starts the worker goroutines.
func (p *Processor) startWorkers(ctx context.Context, wg *sync.WaitGroup, fileCh chan string, writeCh chan fileproc.WriteRequest) {
	for range p.flags.Concurrency {
		wg.Add(1)
		go p.worker(ctx, wg, fileCh, writeCh)
	}
}

// worker is the worker goroutine function.
func (p *Processor) worker(ctx context.Context, wg *sync.WaitGroup, fileCh chan string, writeCh chan fileproc.WriteRequest) {
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case filePath, ok := <-fileCh:
			if !ok {
				return
			}
			p.processFile(ctx, filePath, writeCh)
		}
	}
}

// processFile processes a single file with resource monitoring.
func (p *Processor) processFile(ctx context.Context, filePath string, writeCh chan fileproc.WriteRequest) {
	// Check for emergency stop
	if p.resourceMonitor.IsEmergencyStopActive() {
		logrus.Warnf("Emergency stop active, skipping file: %s", filePath)
		return
	}

	absRoot, err := utils.GetAbsolutePath(p.flags.SourceDir)
	if err != nil {
		utils.LogError("Failed to get absolute path", err)
		return
	}

	// Use the resource monitor-aware processing
	fileproc.ProcessFileWithMonitor(ctx, filePath, writeCh, absRoot, p.resourceMonitor)

	// Update progress bar
	p.ui.UpdateProgress(1)
}

// sendFiles sends files to the worker channels with back-pressure handling.
func (p *Processor) sendFiles(ctx context.Context, files []string, fileCh chan string) error {
	defer close(fileCh)

	for _, fp := range files {
		// Check if we should apply back-pressure
		if p.backpressure.ShouldApplyBackpressure(ctx) {
			p.backpressure.ApplyBackpressure(ctx)
		}

		// Wait for channel space if needed
		p.backpressure.WaitForChannelSpace(ctx, fileCh, nil)

		select {
		case <-ctx.Done():
			return ctx.Err()
		case fileCh <- fp:
		}
	}
	return nil
}

// waitForCompletion waits for all workers to complete.
func (p *Processor) waitForCompletion(wg *sync.WaitGroup, writeCh chan fileproc.WriteRequest, writerDone chan struct{}) {
	wg.Wait()
	close(writeCh)
	<-writerDone
}

// logFinalStats logs the final back-pressure and resource monitoring statistics.
func (p *Processor) logFinalStats() {
	// Log back-pressure stats
	backpressureStats := p.backpressure.GetStats()
	if backpressureStats.Enabled {
		logrus.Infof("Back-pressure stats: processed=%d files, memory=%dMB/%dMB",
			backpressureStats.FilesProcessed, backpressureStats.CurrentMemoryUsage/1024/1024, backpressureStats.MaxMemoryUsage/1024/1024)
	}

	// Log resource monitoring stats
	resourceStats := p.resourceMonitor.GetMetrics()
	if config.GetResourceLimitsEnabled() {
		logrus.Infof("Resource stats: processed=%d files, totalSize=%dMB, avgFileSize=%.2fKB, rate=%.2f files/sec",
			resourceStats.FilesProcessed, resourceStats.TotalSizeProcessed/1024/1024,
			resourceStats.AverageFileSize/1024, resourceStats.ProcessingRate)

		if len(resourceStats.ViolationsDetected) > 0 {
			logrus.Warnf("Resource violations detected: %v", resourceStats.ViolationsDetected)
		}

		if resourceStats.DegradationActive {
			logrus.Warnf("Processing completed with degradation mode active")
		}

		if resourceStats.EmergencyStopActive {
			logrus.Errorf("Processing completed with emergency stop active")
		}
	}

	// Clean up resource monitor
	p.resourceMonitor.Close()
}
