// Package cli provides command-line interface functionality for gibidify.
package cli

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/ivuorinen/gibidify/fileproc"
	"github.com/ivuorinen/gibidify/shared"
)

// fileJob is one file handed to a worker. The worker sends the file's
// WriteRequest on out, or closes out without a value when the file was skipped,
// failed, or the run was canceled. Every job's out is closed exactly once, so the
// sequencer waiting on it can never block forever.
type fileJob struct {
	path string
	out  chan fileproc.WriteRequest
}

// runStats counts per-file outcomes across workers.
type runStats struct {
	skipped atomic.Int64
	failed  atomic.Int64
}

// startWorkers starts the worker goroutines.
func (p *Processor) startWorkers(
	ctx context.Context,
	wg *sync.WaitGroup,
	absRoot string,
	jobs <-chan fileJob,
	stats *runStats,
) {
	for range p.flags.Concurrency {
		wg.Go(func() {
			// Drain jobs until the channel closes, even after cancellation:
			// a job left unreceived would leave its out channel open and
			// stall the sequencer.
			for job := range jobs {
				p.processFile(ctx, absRoot, job, stats)
			}
		})
	}
}

// processFile processes a single file, records its outcome and reports progress.
func (p *Processor) processFile(ctx context.Context, absRoot string, job fileJob, stats *runStats) {
	defer close(job.out)
	defer p.ui.UpdateProgress(1)

	if ctx.Err() != nil {
		return
	}

	err := fileproc.ProcessFileContext(ctx, job.path, job.out, absRoot)
	switch {
	case err == nil:
	case errors.Is(err, fileproc.ErrSkipped):
		stats.skipped.Add(1)
		shared.GetLogger().Infof("Skipping %v", err)
	case ctx.Err() != nil:
		// Canceled: reported once by processFiles, not per file.
	default:
		stats.failed.Add(1)
		shared.LogErrorf(err, "processing file %s", job.path)
	}
}

// sendFiles hands each file to the workers and queues its result channel in
// collection order. It closes both channels when it returns.
//
// The job is sent before its out channel is queued: if cancellation lands in
// between, the queued order never references a job no worker received.
func (p *Processor) sendFiles(
	ctx context.Context,
	files []string,
	jobs chan<- fileJob,
	order chan<- chan fileproc.WriteRequest,
) error {
	defer close(jobs)
	defer close(order)

	for _, fp := range files {
		// select picks randomly among ready cases, so a canceled context
		// would not stop the loop while the jobs buffer has room.
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("processing canceled: %w", err)
		}

		job := fileJob{path: fp, out: make(chan fileproc.WriteRequest, 1)}

		select {
		case jobs <- job:
		case <-ctx.Done():
			return fmt.Errorf("processing canceled: %w", ctx.Err())
		}

		select {
		case order <- job.out:
		case <-ctx.Done():
			return fmt.Errorf("processing canceled: %w", ctx.Err())
		}
	}

	return nil
}

// sequence forwards each file's result to writeCh in collection order, so the
// output is identical across runs regardless of which worker finishes first.
// Memory stays bounded by the order buffer plus in-flight workers.
func sequence(order <-chan chan fileproc.WriteRequest, writeCh chan<- fileproc.WriteRequest) {
	defer close(writeCh)

	for out := range order {
		if req, ok := <-out; ok {
			writeCh <- req
		}
	}
}
