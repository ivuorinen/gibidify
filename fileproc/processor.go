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
	"sync"
	"time"

	"github.com/ivuorinen/gibidify/config"
	"github.com/ivuorinen/gibidify/shared"
)

// WriteRequest represents the content to be written.
type WriteRequest struct {
	Path     string
	Content  string
	IsStream bool
	Reader   io.Reader
	Size     int64 // File size for streaming files
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
		sizeLimit:       config.FileSizeLimit(),
		resourceMonitor: NewResourceMonitor(),
	}
}

// NewFileProcessorWithMonitor creates a new file processor with a shared resource monitor.
func NewFileProcessorWithMonitor(rootPath string, monitor *ResourceMonitor) *FileProcessor {
	return &FileProcessor{
		rootPath:        rootPath,
		sizeLimit:       config.FileSizeLimit(),
		resourceMonitor: monitor,
	}
}

// ProcessFile reads the file at filePath and sends a formatted output to outCh.
// It automatically chooses between loading the entire file or streaming based on file size.
func ProcessFile(filePath string, outCh chan<- WriteRequest, rootPath string) {
	processor := NewFileProcessor(rootPath)
	ctx := context.Background()
	if err := processor.ProcessWithContext(ctx, filePath, outCh); err != nil {
		shared.LogErrorf(err, shared.FileProcessingMsgFailedToProcess, filePath)
	}
}

// ProcessFileWithMonitor processes a file using a shared resource monitor.
func ProcessFileWithMonitor(
	ctx context.Context,
	filePath string,
	outCh chan<- WriteRequest,
	rootPath string,
	monitor *ResourceMonitor,
) error {
	if monitor == nil {
		monitor = NewResourceMonitor()
	}
	processor := NewFileProcessorWithMonitor(rootPath, monitor)

	return processor.ProcessWithContext(ctx, filePath, outCh)
}

// Process handles file processing with the configured settings.
func (p *FileProcessor) Process(filePath string, outCh chan<- WriteRequest) {
	ctx := context.Background()
	if err := p.ProcessWithContext(ctx, filePath, outCh); err != nil {
		shared.LogErrorf(err, shared.FileProcessingMsgFailedToProcess, filePath)
	}
}

// ProcessWithContext handles file processing with context and resource monitoring.
func (p *FileProcessor) ProcessWithContext(ctx context.Context, filePath string, outCh chan<- WriteRequest) error {
	// Create file processing context with timeout
	fileCtx, fileCancel := p.resourceMonitor.CreateFileProcessingContext(ctx)
	defer fileCancel()

	// Wait for rate limiting
	if err := p.resourceMonitor.WaitForRateLimit(fileCtx); err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			structErr := shared.NewStructuredError(
				shared.ErrorTypeValidation,
				shared.CodeResourceLimitTimeout,
				"file processing timeout during rate limiting",
				filePath,
				nil,
			)
			shared.LogErrorf(structErr, "File processing timeout during rate limiting: %s", filePath)

			return structErr
		}

		return err
	}

	// Validate file and check resource limits
	fileInfo, err := p.validateFileWithLimits(fileCtx, filePath)
	if err != nil {
		return err // Error already logged
	}

	// Acquire read slot for concurrent processing
	if err := p.resourceMonitor.AcquireReadSlot(fileCtx); err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			structErr := shared.NewStructuredError(
				shared.ErrorTypeValidation,
				shared.CodeResourceLimitTimeout,
				"file processing timeout waiting for read slot",
				filePath,
				nil,
			)
			shared.LogErrorf(structErr, "File processing timeout waiting for read slot: %s", filePath)

			return structErr
		}

		return err
	}
	defer p.resourceMonitor.ReleaseReadSlot()

	// Check hard memory limits before processing
	if err := p.resourceMonitor.CheckHardMemoryLimit(); err != nil {
		shared.LogErrorf(err, "Hard memory limit check failed for file: %s", filePath)

		return err
	}

	// Get relative path
	relPath := p.getRelativePath(filePath)

	// Process file with timeout
	processStart := time.Now()

	// Choose processing strategy based on file size
	if fileInfo.Size() <= shared.FileProcessingStreamThreshold {
		err = p.processInMemoryWithContext(fileCtx, filePath, relPath, outCh)
	} else {
		err = p.processStreamingWithContext(fileCtx, filePath, relPath, outCh, fileInfo.Size())
	}

	// Only record success if processing completed without error
	if err != nil {
		return err
	}

	// Record successful processing only on success path
	p.resourceMonitor.RecordFileProcessed(fileInfo.Size())
	logger := shared.GetLogger()
	logger.Debugf("File processed in %v: %s", time.Since(processStart), filePath)

	return nil
}

// validateFileWithLimits checks if the file can be processed with resource limits.
func (p *FileProcessor) validateFileWithLimits(ctx context.Context, filePath string) (os.FileInfo, error) {
	// Check context cancellation
	if err := shared.CheckContextCancellation(ctx, "file validation"); err != nil {
		return nil, fmt.Errorf("context check during file validation: %w", err)
	}

	fileInfo, err := os.Stat(filePath)
	if err != nil {
		structErr := shared.WrapError(
			err,
			shared.ErrorTypeFileSystem,
			shared.CodeFSAccess,
			"failed to stat file",
		).WithFilePath(filePath)
		shared.LogErrorf(structErr, "Failed to stat file %s", filePath)

		return nil, structErr
	}

	// Check traditional size limit
	if fileInfo.Size() > p.sizeLimit {
		c := map[string]any{
			"file_size":  fileInfo.Size(),
			"size_limit": p.sizeLimit,
		}
		structErr := shared.NewStructuredError(
			shared.ErrorTypeValidation,
			shared.CodeValidationSize,
			fmt.Sprintf(shared.FileProcessingMsgSizeExceeds, fileInfo.Size(), p.sizeLimit),
			filePath,
			c,
		)
		shared.LogErrorf(structErr, "Skipping large file %s", filePath)

		return nil, structErr
	}

	// Check resource limits
	if err := p.resourceMonitor.ValidateFileProcessing(filePath, fileInfo.Size()); err != nil {
		shared.LogErrorf(err, "Resource limit validation failed for file: %s", filePath)

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
) error {
	// Check context before reading
	select {
	case <-ctx.Done():
		structErr := shared.NewStructuredError(
			shared.ErrorTypeValidation,
			shared.CodeResourceLimitTimeout,
			"file processing canceled",
			filePath,
			nil,
		)
		shared.LogErrorf(structErr, "File processing canceled: %s", filePath)

		return structErr
	default:
	}

	content, err := os.ReadFile(filePath) // #nosec G304 - filePath is validated by walker
	if err != nil {
		structErr := shared.WrapError(
			err,
			shared.ErrorTypeProcessing,
			shared.CodeProcessingFileRead,
			"failed to read file",
		).WithFilePath(filePath)
		shared.LogErrorf(structErr, "Failed to read file %s", filePath)

		return structErr
	}

	// Check context again after reading
	select {
	case <-ctx.Done():
		structErr := shared.NewStructuredError(
			shared.ErrorTypeValidation,
			shared.CodeResourceLimitTimeout,
			"file processing canceled after read",
			filePath,
			nil,
		)
		shared.LogErrorf(structErr, "File processing canceled after read: %s", filePath)

		return structErr
	default:
	}

	// Try to send the result, but respect context cancellation
	select {
	case <-ctx.Done():
		structErr := shared.NewStructuredError(
			shared.ErrorTypeValidation,
			shared.CodeResourceLimitTimeout,
			"file processing canceled before output",
			filePath,
			nil,
		)
		shared.LogErrorf(structErr, "File processing canceled before output: %s", filePath)

		return structErr
	case outCh <- WriteRequest{
		Path:     relPath,
		Content:  p.formatContent(relPath, string(content)),
		IsStream: false,
		Size:     int64(len(content)),
	}:
	}

	return nil
}

// processStreamingWithContext creates a streaming reader for large files with context awareness.
func (p *FileProcessor) processStreamingWithContext(
	ctx context.Context,
	filePath, relPath string,
	outCh chan<- WriteRequest,
	size int64,
) error {
	// Check context before creating reader
	select {
	case <-ctx.Done():
		structErr := shared.NewStructuredError(
			shared.ErrorTypeValidation,
			shared.CodeResourceLimitTimeout,
			"streaming processing canceled",
			filePath,
			nil,
		)
		shared.LogErrorf(structErr, "Streaming processing canceled: %s", filePath)

		return structErr
	default:
	}

	reader := p.createStreamReaderWithContext(ctx, filePath, relPath)
	if reader == nil {
		// Error already logged, create and return error
		return shared.NewStructuredError(
			shared.ErrorTypeProcessing,
			shared.CodeProcessingFileRead,
			"failed to create stream reader",
			filePath,
			nil,
		)
	}

	// Try to send the result, but respect context cancellation
	select {
	case <-ctx.Done():
		structErr := shared.NewStructuredError(
			shared.ErrorTypeValidation,
			shared.CodeResourceLimitTimeout,
			"streaming processing canceled before output",
			filePath,
			nil,
		)
		shared.LogErrorf(structErr, "Streaming processing canceled before output: %s", filePath)

		return structErr
	case outCh <- WriteRequest{
		Path:     relPath,
		Content:  "", // Empty since content is in Reader
		IsStream: true,
		Reader:   reader,
		Size:     size,
	}:
	}

	return nil
}

// createStreamReaderWithContext creates a reader that combines header and file content with context awareness.
func (p *FileProcessor) createStreamReaderWithContext(
	ctx context.Context, filePath, relPath string,
) io.Reader {
	// Check context before opening file
	select {
	case <-ctx.Done():
		return nil
	default:
	}

	file, err := os.Open(filePath) // #nosec G304 - filePath is validated by walker
	if err != nil {
		structErr := shared.WrapError(
			err,
			shared.ErrorTypeProcessing,
			shared.CodeProcessingFileRead,
			"failed to open file for streaming",
		).WithFilePath(filePath)
		shared.LogErrorf(structErr, "Failed to open file for streaming %s", filePath)

		return nil
	}
	header := p.formatHeader(relPath)

	return newHeaderFileReader(header, file)
}

// formatContent formats the file content with header.
func (p *FileProcessor) formatContent(relPath, content string) string {
	return fmt.Sprintf("\n---\n%s\n%s\n", relPath, content)
}

// formatHeader creates a reader for the file header.
func (p *FileProcessor) formatHeader(relPath string) io.Reader {
	return strings.NewReader(fmt.Sprintf("\n---\n%s\n", relPath))
}

// headerFileReader wraps a MultiReader and closes the file when EOF is reached.
type headerFileReader struct {
	reader io.Reader
	file   *os.File
	mu     sync.Mutex
	closed bool
}

// newHeaderFileReader creates a new headerFileReader.
func newHeaderFileReader(header io.Reader, file *os.File) *headerFileReader {
	return &headerFileReader{
		reader: io.MultiReader(header, file),
		file:   file,
	}
}

// Read implements io.Reader and closes the file on EOF.
func (r *headerFileReader) Read(p []byte) (n int, err error) {
	n, err = r.reader.Read(p)
	if err == io.EOF {
		r.closeFile()
		// EOF is a sentinel value that must be passed through unchanged for io.Reader interface
		return n, err //nolint:wrapcheck // EOF must not be wrapped
	}
	if err != nil {
		return n, shared.WrapError(
			err, shared.ErrorTypeIO, shared.CodeIORead,
			"failed to read from header file reader",
		)
	}

	return n, nil
}

// closeFile closes the file once.
func (r *headerFileReader) closeFile() {
	r.mu.Lock()
	defer r.mu.Unlock()
	if !r.closed && r.file != nil {
		if err := r.file.Close(); err != nil {
			shared.LogError("Failed to close file", err)
		}
		r.closed = true
	}
}

// Close implements io.Closer and ensures the underlying file is closed.
// This allows explicit cleanup when consumers stop reading before EOF.
func (r *headerFileReader) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.closed || r.file == nil {
		return nil
	}
	err := r.file.Close()
	if err != nil {
		shared.LogError("Failed to close file", err)
	}
	r.closed = true

	return err
}
