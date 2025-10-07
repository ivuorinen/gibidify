// Package fileproc provides functions for processing files.
package fileproc

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/ivuorinen/gibidify/config"
	"github.com/ivuorinen/gibidify/gibidiutils"
)

const (
	// StreamChunkSize is the size of chunks when streaming large files (64KB).
	StreamChunkSize = 65536
	// StreamThreshold is the file size above which we use streaming (1MB).
	StreamThreshold = 1048576
	// MaxMemoryBuffer is the maximum memory to use for buffering content (10MB).
	MaxMemoryBuffer = 10485760
)

// WriteRequest represents the content to be written.
type WriteRequest struct {
	Path     string
	Content  string
	IsStream bool
	Reader   io.Reader
}

// FileProcessor handles file processing operations.
type FileProcessor struct {
	rootPath        string
	sizeLimit       int64
	resourceMonitor *ResourceMonitor
}

// NewFileProcessor creates a new file processor.
func NewFileProcessor(rootPath string) *FileProcessor {
	return &FileProcessor{
		rootPath:        rootPath,
		sizeLimit:       config.GetFileSizeLimit(),
		resourceMonitor: NewResourceMonitor(),
	}
}

// NewFileProcessorWithMonitor creates a new file processor with a shared resource monitor.
func NewFileProcessorWithMonitor(rootPath string, monitor *ResourceMonitor) *FileProcessor {
	return &FileProcessor{
		rootPath:        rootPath,
		sizeLimit:       config.GetFileSizeLimit(),
		resourceMonitor: monitor,
	}
}

// checkContextCancellation checks if context is cancelled and logs an error if so.
// Returns true if context is cancelled, false otherwise.
func (p *FileProcessor) checkContextCancellation(ctx context.Context, filePath, stage string) bool {
	select {
	case <-ctx.Done():
		// Format stage with leading space if provided
		stageMsg := stage
		if stage != "" {
			stageMsg = " " + stage
		}
		gibidiutils.LogErrorf(
			gibidiutils.NewStructuredError(
				gibidiutils.ErrorTypeValidation,
				gibidiutils.CodeResourceLimitTimeout,
				fmt.Sprintf("file processing cancelled%s", stageMsg),
				filePath,
				nil,
			),
			"File processing cancelled%s: %s",
			stageMsg,
			filePath,
		)
		return true
	default:
		return false
	}
}

// ProcessFile reads the file at filePath and sends a formatted output to outCh.
// It automatically chooses between loading the entire file or streaming based on file size.
func ProcessFile(filePath string, outCh chan<- WriteRequest, rootPath string) {
	processor := NewFileProcessor(rootPath)
	ctx := context.Background()
	processor.ProcessWithContext(ctx, filePath, outCh)
}

// ProcessFileWithMonitor processes a file using a shared resource monitor.
func ProcessFileWithMonitor(
	ctx context.Context,
	filePath string,
	outCh chan<- WriteRequest,
	rootPath string,
	monitor *ResourceMonitor,
) {
	processor := NewFileProcessorWithMonitor(rootPath, monitor)
	processor.ProcessWithContext(ctx, filePath, outCh)
}

// Process handles file processing with the configured settings.
func (p *FileProcessor) Process(filePath string, outCh chan<- WriteRequest) {
	ctx := context.Background()
	p.ProcessWithContext(ctx, filePath, outCh)
}

// ProcessWithContext handles file processing with context and resource monitoring.
func (p *FileProcessor) ProcessWithContext(ctx context.Context, filePath string, outCh chan<- WriteRequest) {
	// Create file processing context with timeout
	fileCtx, fileCancel := p.resourceMonitor.CreateFileProcessingContext(ctx)
	defer fileCancel()

	// Wait for rate limiting
	if err := p.resourceMonitor.WaitForRateLimit(fileCtx); err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			gibidiutils.LogErrorf(
				gibidiutils.NewStructuredError(
					gibidiutils.ErrorTypeValidation,
					gibidiutils.CodeResourceLimitTimeout,
					"file processing timeout during rate limiting",
					filePath,
					nil,
				),
				"File processing timeout during rate limiting: %s",
				filePath,
			)
		}
		return
	}

	// Validate file and check resource limits
	fileInfo, err := p.validateFileWithLimits(fileCtx, filePath)
	if err != nil {
		return // Error already logged
	}

	// Acquire read slot for concurrent processing
	if err := p.resourceMonitor.AcquireReadSlot(fileCtx); err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			gibidiutils.LogErrorf(
				gibidiutils.NewStructuredError(
					gibidiutils.ErrorTypeValidation,
					gibidiutils.CodeResourceLimitTimeout,
					"file processing timeout waiting for read slot",
					filePath,
					nil,
				),
				"File processing timeout waiting for read slot: %s",
				filePath,
			)
		}
		return
	}
	defer p.resourceMonitor.ReleaseReadSlot()

	// Check hard memory limits before processing
	if err := p.resourceMonitor.CheckHardMemoryLimit(); err != nil {
		gibidiutils.LogErrorf(err, "Hard memory limit check failed for file: %s", filePath)
		return
	}

	// Get relative path
	relPath := p.getRelativePath(filePath)

	// Process file with timeout
	processStart := time.Now()
	defer func() {
		// Record successful processing
		p.resourceMonitor.RecordFileProcessed(fileInfo.Size())
		logrus.Debugf("File processed in %v: %s", time.Since(processStart), filePath)
	}()

	// Choose processing strategy based on file size
	if fileInfo.Size() <= StreamThreshold {
		p.processInMemoryWithContext(fileCtx, filePath, relPath, outCh)
	} else {
		p.processStreamingWithContext(fileCtx, filePath, relPath, outCh)
	}
}

// validateFileWithLimits checks if the file can be processed with resource limits.
func (p *FileProcessor) validateFileWithLimits(ctx context.Context, filePath string) (os.FileInfo, error) {
	// Check context cancellation
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	fileInfo, err := os.Stat(filePath)
	if err != nil {
		structErr := gibidiutils.WrapError(
			err, gibidiutils.ErrorTypeFileSystem, gibidiutils.CodeFSAccess,
			"failed to stat file",
		).WithFilePath(filePath)
		gibidiutils.LogErrorf(structErr, "Failed to stat file %s", filePath)
		return nil, structErr
	}

	// Check traditional size limit
	if fileInfo.Size() > p.sizeLimit {
		filesizeContext := map[string]interface{}{
			"file_size":  fileInfo.Size(),
			"size_limit": p.sizeLimit,
		}
		gibidiutils.LogErrorf(
			gibidiutils.NewStructuredError(
				gibidiutils.ErrorTypeValidation,
				gibidiutils.CodeValidationSize,
				fmt.Sprintf("file size (%d bytes) exceeds limit (%d bytes)", fileInfo.Size(), p.sizeLimit),
				filePath,
				filesizeContext,
			),
			"Skipping large file %s", filePath,
		)
		return nil, fmt.Errorf("file too large")
	}

	// Check resource limits
	if err := p.resourceMonitor.ValidateFileProcessing(filePath, fileInfo.Size()); err != nil {
		gibidiutils.LogErrorf(err, "Resource limit validation failed for file: %s", filePath)
		return nil, err
	}

	return fileInfo, nil
}

// getRelativePath computes the path relative to rootPath.
func (p *FileProcessor) getRelativePath(filePath string) string {
	relPath, err := filepath.Rel(p.rootPath, filePath)
	if err != nil {
		return filePath // Fallback
	}
	return relPath
}

// processInMemoryWithContext loads the entire file into memory with context awareness.
func (p *FileProcessor) processInMemoryWithContext(
	ctx context.Context,
	filePath, relPath string,
	outCh chan<- WriteRequest,
) {
	// Check context before reading
	if p.checkContextCancellation(ctx, filePath, "") {
		return
	}

	// #nosec G304 - filePath is validated by walker
	content, err := os.ReadFile(filePath)
	if err != nil {
		structErr := gibidiutils.WrapError(
			err, gibidiutils.ErrorTypeProcessing, gibidiutils.CodeProcessingFileRead,
			"failed to read file",
		).WithFilePath(filePath)
		gibidiutils.LogErrorf(structErr, "Failed to read file %s", filePath)
		return
	}

	// Check context again after reading
	if p.checkContextCancellation(ctx, filePath, "after read") {
		return
	}

	// Check context before sending output
	if p.checkContextCancellation(ctx, filePath, "before output") {
		return
	}

	outCh <- WriteRequest{
		Path:     relPath,
		Content:  p.formatContent(relPath, string(content)),
		IsStream: false,
	}
}

// processStreamingWithContext creates a streaming reader for large files with context awareness.
func (p *FileProcessor) processStreamingWithContext(
	ctx context.Context,
	filePath, relPath string,
	outCh chan<- WriteRequest,
) {
	// Check context before creating reader
	if p.checkContextCancellation(ctx, filePath, "before streaming") {
		return
	}

	reader := p.createStreamReaderWithContext(ctx, filePath, relPath)
	if reader == nil {
		return // Error already logged
	}

	// Check context before sending output
	if p.checkContextCancellation(ctx, filePath, "before streaming output") {
		return
	}

	outCh <- WriteRequest{
		Path:     relPath,
		Content:  "", // Empty since content is in Reader
		IsStream: true,
		Reader:   reader,
	}
}

// createStreamReaderWithContext creates a reader that combines header and file content with context awareness.
func (p *FileProcessor) createStreamReaderWithContext(ctx context.Context, filePath, relPath string) io.Reader {
	// Check context before opening file
	if p.checkContextCancellation(ctx, filePath, "before opening file") {
		return nil
	}

	// #nosec G304 - filePath is validated by walker
	file, err := os.Open(filePath)
	if err != nil {
		structErr := gibidiutils.WrapError(
			err, gibidiutils.ErrorTypeProcessing, gibidiutils.CodeProcessingFileRead,
			"failed to open file for streaming",
		).WithFilePath(filePath)
		gibidiutils.LogErrorf(structErr, "Failed to open file for streaming %s", filePath)
		return nil
	}
	// Note: file will be closed by the writer

	header := p.formatHeader(relPath)
	return io.MultiReader(header, file)
}

// formatContent formats the file content with header.
func (p *FileProcessor) formatContent(relPath, content string) string {
	return fmt.Sprintf("\n---\n%s\n%s\n", relPath, content)
}

// formatHeader creates a reader for the file header.
func (p *FileProcessor) formatHeader(relPath string) io.Reader {
	return strings.NewReader(fmt.Sprintf("\n---\n%s\n", relPath))
}
