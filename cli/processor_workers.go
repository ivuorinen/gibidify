package cli

import (
	"context"
	"sync"

	"github.com/sirupsen/logrus"

	"github.com/ivuorinen/gibidify/fileproc"
	"github.com/ivuorinen/gibidify/gibidiutils"
)

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

	absRoot, err := gibidiutils.GetAbsolutePath(p.flags.SourceDir)
	if err != nil {
		gibidiutils.LogError("Failed to get absolute path", err)
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
