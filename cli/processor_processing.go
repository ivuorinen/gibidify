// Package cli provides command-line interface functionality for gibidify.
package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/ivuorinen/gibidify/fileproc"
	"github.com/ivuorinen/gibidify/shared"
)

// channelBuffer is the fixed capacity of the job, order and write channels. It
// keeps producers slightly ahead of consumers without unbounded memory growth;
// revisit only if profiling shows workers stalling on the channel.
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
	files, oversized, err := p.collectFiles()
	if err != nil {
		return err
	}

	// Show collection results
	p.ui.PrintSuccess(shared.CLIMsgFoundFilesToProcess, len(files))

	// Pre-validate file collection against resource limits
	if err := p.validateFileCollection(files); err != nil {
		return err
	}

	stats, err := p.processFiles(ctx, files)
	skipped := stats.skipped.Load() + int64(oversized)
	shared.GetLogger().Infof(
		"Processed %d files (%d skipped, %d failed) in %s",
		len(files), skipped, stats.failed.Load(), time.Since(start).Round(time.Millisecond),
	)

	return err
}

// processFiles runs the pipeline: sendFiles → workers → sequence → writer.
//
// Every return path runs the full shutdown (send loop ends, workers drain,
// writer finishes) before the output file is closed, so no goroutine outlives
// the call and nothing writes to a closed file. The output goes to a temporary
// file that replaces the destination only when the writer succeeded and the run
// was not canceled; otherwise it is removed and an existing destination is left
// as it was. Files that failed individually leave a valid output but a non-nil
// error, so the process exits non-zero.
func (p *Processor) processFiles(ctx context.Context, files []string) (*runStats, error) {
	stats := &runStats{}

	absRoot, err := shared.AbsolutePath(p.flags.SourceDir)
	if err != nil {
		return stats, shared.WrapError(
			err, shared.ErrorTypeFileSystem, shared.CodeFSPathResolution, "failed to resolve source path",
		).WithFilePath(p.flags.SourceDir)
	}

	outFile, err := p.createOutputFile()
	if err != nil {
		return stats, err
	}

	p.ui.PrintInfo("⚙️  Initializing processing...")
	jobs := make(chan fileJob, channelBuffer)
	order := make(chan chan fileproc.WriteRequest, channelBuffer)
	writeCh := make(chan fileproc.WriteRequest, channelBuffer)
	writerErr := make(chan error, 1)

	go func() {
		writerErr <- fileproc.StartWriter(outFile.File, writeCh, p.flags.Format, p.flags.Prefix, p.flags.Suffix)
	}()
	go sequence(order, writeCh)

	var wg sync.WaitGroup
	p.startWorkers(ctx, &wg, absRoot, jobs, stats)

	p.ui.StartProgress(len(files), "📝 Processing files")
	sendErr := p.sendFiles(ctx, files, jobs, order)
	wg.Wait()
	writeErr := <-writerErr
	closeErr := outFile.Close()
	p.ui.FinishProgress()

	if writeErr == nil {
		writeErr = closeErr
	}
	// Cancellation after the last send still makes workers drop files.
	if sendErr == nil && ctx.Err() != nil {
		sendErr = fmt.Errorf("processing canceled: %w", ctx.Err())
	}

	return stats, p.finishOutput(outFile, len(files), stats, sendErr, writeErr)
}

// finishOutput turns the pipeline's outcome into the run's error, and moves the
// completed output into place or discards an incomplete one.
func (p *Processor) finishOutput(out *outputFile, total int, stats *runStats, sendErr, writeErr error) error {
	if writeErr == nil && sendErr == nil {
		writeErr = out.commit(p.flags.Destination)
	} else {
		out.discard()
	}

	if writeErr != nil {
		return shared.WrapError(
			writeErr, shared.ErrorTypeIO, shared.CodeIOWrite, "failed to write output",
		).WithFilePath(p.flags.Destination)
	}

	if sendErr != nil {
		return sendErr
	}

	if failed := stats.failed.Load(); failed > 0 {
		return shared.NewStructuredError(
			shared.ErrorTypeProcessing,
			shared.CodeProcessingFileRead,
			fmt.Sprintf("%d of %d files could not be processed (see errors above)", failed, total),
			"",
			map[string]any{"failed": failed, "files": total},
		)
	}

	p.ui.PrintSuccess("Processing completed. Output saved to %s", p.flags.Destination)

	return nil
}

// outputFile is the file the writer fills. tmpPath is set when it is a
// temporary file that replaces the destination on success; it is empty when
// the writer writes to the destination itself (a device such as /dev/stdout).
type outputFile struct {
	*os.File
	tmpPath string
}

// commit moves a completed temporary output over the destination. The file
// must already be closed.
func (o *outputFile) commit(dest string) error {
	if o.tmpPath == "" {
		return nil
	}
	if err := os.Rename(o.tmpPath, filepath.Clean(dest)); err != nil {
		o.discard()

		return fmt.Errorf("replacing %s: %w", dest, err)
	}

	return nil
}

// discard removes an incomplete temporary output. The destination itself is
// never touched, so the previous run's output survives a failed or interrupted
// run (os.Create used to truncate it before processing started).
func (o *outputFile) discard() {
	if o.tmpPath == "" {
		return
	}
	if err := os.Remove(o.tmpPath); err != nil && !os.IsNotExist(err) {
		shared.LogError("Failed to remove partial output", err)
	}
}

// createOutputFile opens where the writer writes. A regular or missing
// destination gets a temporary file in the same directory, so the final rename
// stays on one filesystem; it is created 0600, or with the existing
// destination's permissions. Any other destination (a device, FIFO or symlink)
// is written directly, as there is nothing to replace atomically.
func (p *Processor) createOutputFile() (*outputFile, error) {
	// Clean, as ValidateDestinationPath does, so the file written is the one
	// that was validated.
	dest := filepath.Clean(p.flags.Destination)
	wrap := func(err error) error {
		return shared.WrapError(
			err, shared.ErrorTypeIO, shared.CodeIOFileCreate, "failed to create output file",
		).WithFilePath(p.flags.Destination)
	}

	info, statErr := os.Lstat(dest)
	if statErr == nil && !info.Mode().IsRegular() {
		f, err := os.OpenFile(dest, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, shared.OutputFilePermission)
		if err != nil {
			return nil, wrap(err)
		}

		return &outputFile{File: f}, nil
	}

	f, err := os.CreateTemp(filepath.Dir(dest), "."+filepath.Base(dest)+".*.tmp")
	if err != nil {
		return nil, wrap(err)
	}
	out := &outputFile{File: f, tmpPath: f.Name()}
	if statErr == nil {
		if err := f.Chmod(info.Mode().Perm()); err != nil {
			_ = f.Close()
			out.discard()

			return nil, wrap(err)
		}
	}

	return out, nil
}
