// Package cli provides command-line interface functionality for gibidify.
package cli

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/ivuorinen/gibidify/fileproc"
	"github.com/ivuorinen/gibidify/metrics"
	"github.com/ivuorinen/gibidify/shared"
)

// startWorkers starts the worker goroutines.
func (p *Processor) startWorkers(
	ctx context.Context,
	wg *sync.WaitGroup,
	fileCh chan string,
	writeCh chan fileproc.WriteRequest,
) {
	for range p.flags.Concurrency {
		wg.Add(1)
		go p.worker(ctx, wg, fileCh, writeCh)
	}
}

// worker is the worker goroutine function.
func (p *Processor) worker(
	ctx context.Context,
	wg *sync.WaitGroup,
	fileCh chan string,
	writeCh chan fileproc.WriteRequest,
) {
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

// processFile processes a single file with resource monitoring and metrics collection.
func (p *Processor) processFile(ctx context.Context, filePath string, writeCh chan fileproc.WriteRequest) {
	// Create file processing context with timeout
	fileCtx, fileCancel := p.resourceMonitor.CreateFileProcessingContext(ctx)
	defer fileCancel()

	// Track concurrency
	if p.metricsCollector != nil {
		p.metricsCollector.IncrementConcurrency()
		defer p.metricsCollector.DecrementConcurrency()
	}

	// Check for emergency stop
	if p.resourceMonitor != nil && p.resourceMonitor.IsEmergencyStopActive() {
		logger := shared.GetLogger()
		logger.Warnf("Emergency stop active, skipping file: %s", filePath)

		// Record skipped file
		p.recordFileResult(filePath, 0, "", false, true, "emergency stop active", nil)

		return
	}

	absRoot, err := shared.AbsolutePath(p.flags.SourceDir)
	if err != nil {
		shared.LogError("Failed to get absolute path", err)

		// Record error
		p.recordFileResult(filePath, 0, "", false, false, "", err)

		return
	}

	// Use the resource monitor-aware processing with metrics tracking
	fileSize, format, success, processErr := p.processFileWithMetrics(fileCtx, filePath, writeCh, absRoot)

	// Record the processing result (skipped=false, skipReason="" since processFileWithMetrics never skips)
	p.recordFileResult(filePath, fileSize, format, success, false, "", processErr)

	// Update progress bar with metrics
	p.ui.UpdateProgress(1)

	// Show real-time stats in verbose mode
	if p.flags.Verbose && p.metricsCollector != nil {
		currentMetrics := p.metricsCollector.CurrentMetrics()
		if currentMetrics.ProcessedFiles%10 == 0 && p.metricsReporter != nil {
			logger := shared.GetLogger()
			logger.Info(p.metricsReporter.ReportProgress())
		}
	}
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

		if err := shared.CheckContextCancellation(ctx, shared.CLIMsgFileProcessingWorker); err != nil {
			return fmt.Errorf("context check failed: %w", err)
		}

		select {
		case fileCh <- fp:
		case <-ctx.Done():
			if err := shared.CheckContextCancellation(ctx, shared.CLIMsgFileProcessingWorker); err != nil {
				return fmt.Errorf("context cancellation during channel send: %w", err)
			}

			return errors.New("context canceled during channel send")
		}
	}

	return nil
}

// processFileWithMetrics wraps the file processing with detailed metrics collection.
func (p *Processor) processFileWithMetrics(
	ctx context.Context,
	filePath string,
	writeCh chan fileproc.WriteRequest,
	absRoot string,
) (fileSize int64, format string, success bool, err error) {
	// Get file info
	fileInfo, statErr := os.Stat(filePath)
	if statErr != nil {
		return 0, "", false, fmt.Errorf("getting file info for %s: %w", filePath, statErr)
	}

	fileSize = fileInfo.Size()

	// Detect format from file extension
	format = filepath.Ext(filePath)
	if format != "" && format[0] == '.' {
		format = format[1:] // Remove the dot
	}

	// Use the existing resource monitor-aware processing
	err = fileproc.ProcessFileWithMonitor(ctx, filePath, writeCh, absRoot, p.resourceMonitor)

	// Check if processing was successful
	select {
	case <-ctx.Done():
		return fileSize, format, false, fmt.Errorf("file processing worker canceled: %w", ctx.Err())
	default:
		if err != nil {
			return fileSize, format, false, fmt.Errorf("processing file %s: %w", filePath, err)
		}

		return fileSize, format, true, nil
	}
}

// recordFileResult records the result of file processing in metrics.
func (p *Processor) recordFileResult(
	filePath string,
	fileSize int64,
	format string,
	success bool,
	skipped bool,
	skipReason string,
	err error,
) {
	if p.metricsCollector == nil {
		return // No metrics collector, skip recording
	}

	result := metrics.FileProcessingResult{
		FilePath:   filePath,
		FileSize:   fileSize,
		Format:     format,
		Success:    success,
		Error:      err,
		Skipped:    skipped,
		SkipReason: skipReason,
	}

	p.metricsCollector.RecordFileProcessed(result)
}

// waitForCompletion waits for all workers to complete.
func (p *Processor) waitForCompletion(
	wg *sync.WaitGroup,
	writeCh chan fileproc.WriteRequest,
	writerDone chan struct{},
) {
	wg.Wait()
	close(writeCh)
	<-writerDone
}
