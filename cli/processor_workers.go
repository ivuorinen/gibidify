// Package cli provides command-line interface functionality for gibidify.
package cli

import (
	"context"
	"fmt"
	"sync"

	"github.com/ivuorinen/gibidify/fileproc"
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

// processFile processes a single file and reports progress.
func (p *Processor) processFile(ctx context.Context, filePath string, writeCh chan fileproc.WriteRequest) {
	absRoot, err := shared.AbsolutePath(p.flags.SourceDir)
	if err != nil {
		shared.LogError("Failed to get absolute path", err)
	} else if err := fileproc.ProcessFileContext(ctx, filePath, writeCh, absRoot); err != nil {
		shared.LogErrorf(err, "processing file %s", filePath)
	}

	if p.ui != nil {
		p.ui.UpdateProgress(1)
	}
}

// sendFiles sends files to the worker channel.
func (p *Processor) sendFiles(ctx context.Context, files []string, fileCh chan string) error {
	defer close(fileCh)

	for _, fp := range files {
		if err := shared.CheckContextCancellation(ctx, shared.CLIMsgFileProcessingWorker); err != nil {
			return fmt.Errorf("context check failed: %w", err)
		}

		select {
		case fileCh <- fp:
		case <-ctx.Done():
			return fmt.Errorf("context canceled during channel send: %w", ctx.Err())
		}
	}

	return nil
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
