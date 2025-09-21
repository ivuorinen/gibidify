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
	"github.com/ivuorinen/gibidify/utils"
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

// ProcessFile reads the file at filePath and sends a formatted output to outCh.
// It automatically chooses between loading the entire file or streaming based on file size.
func ProcessFile(filePath string, outCh chan<- WriteRequest, rootPath string) {
	processor := NewFileProcessor(rootPath)
	ctx := context.Background()
	if err := processor.ProcessWithContext(ctx, filePath, outCh); err != nil {
		utils.LogErrorf(err, "Failed to process file: %s", filePath)
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
	processor := NewFileProcessorWithMonitor(rootPath, monitor)

	return processor.ProcessWithContext(ctx, filePath, outCh)
}

// Process handles file processing with the configured settings.
func (p *FileProcessor) Process(filePath string, outCh chan<- WriteRequest) {
	ctx := context.Background()
	if err := p.ProcessWithContext(ctx, filePath, outCh); err != nil {
		utils.LogErrorf(err, "Failed to process file: %s", filePath)
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
			structErr := utils.NewStructuredError(
				utils.ErrorTypeValidation,
				utils.CodeResourceLimitTimeout,
				"file processing timeout during rate limiting",
				filePath,
				nil,
			)
			utils.LogErrorf(structErr, "File processing timeout during rate limiting: %s", filePath)

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
			structErr := utils.NewStructuredError(
				utils.ErrorTypeValidation,
				utils.CodeResourceLimitTimeout,
				"file processing timeout waiting for read slot",
				filePath,
				nil,
			)
			utils.LogErrorf(structErr, "File processing timeout waiting for read slot: %s", filePath)

			return structErr
		}

		return err
	}
	defer p.resourceMonitor.ReleaseReadSlot()

	// Check hard memory limits before processing
	if err := p.resourceMonitor.CheckHardMemoryLimit(); err != nil {
		utils.LogErrorf(err, "Hard memory limit check failed for file: %s", filePath)

		return err
	}

	// Get relative path
	relPath := p.getRelativePath(filePath)

	// Process file with timeout
	processStart := time.Now()
	defer func() {
		// Record successful processing
		p.resourceMonitor.RecordFileProcessed(fileInfo.Size())
		logger := utils.GetLogger()
		logger.Debugf("File processed in %v: %s", time.Since(processStart), filePath)
	}()

	// Choose processing strategy based on file size
	if fileInfo.Size() <= StreamThreshold {
		p.processInMemoryWithContext(fileCtx, filePath, relPath, outCh)
	} else {
		p.processStreamingWithContext(fileCtx, filePath, relPath, outCh)
	}

	return nil
}

// validateFileWithLimits checks if the file can be processed with resource limits.
func (p *FileProcessor) validateFileWithLimits(ctx context.Context, filePath string) (os.FileInfo, error) {
	// Check context cancellation
	if err := utils.CheckContextCancellation(ctx, "file validation"); err != nil {
		return nil, fmt.Errorf("context check during file validation: %w", err)
	}

	fileInfo, err := os.Stat(filePath)
	if err != nil {
		structErr := utils.WrapError(
			err,
			utils.ErrorTypeFileSystem,
			utils.CodeFSAccess,
			"failed to stat file",
		).WithFilePath(filePath)
		utils.LogErrorf(structErr, "Failed to stat file %s", filePath)

		return nil, fmt.Errorf("stat file: %w", err)
	}

	// Check traditional size limit
	if fileInfo.Size() > p.sizeLimit {
		c := map[string]any{
			"file_size":  fileInfo.Size(),
			"size_limit": p.sizeLimit,
		}
		structErr := utils.NewStructuredError(
			utils.ErrorTypeValidation,
			utils.CodeValidationSize,
			fmt.Sprintf("file size (%d bytes) exceeds limit (%d bytes)", fileInfo.Size(), p.sizeLimit),
			filePath,
			c,
		)
		utils.LogErrorf(structErr, "Skipping large file %s", filePath)

		return nil, structErr
	}

	// Check resource limits
	if err := p.resourceMonitor.ValidateFileProcessing(filePath, fileInfo.Size()); err != nil {
		utils.LogErrorf(err, "Resource limit validation failed for file: %s", filePath)

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
	select {
	case <-ctx.Done():
		utils.LogErrorf(
			utils.NewStructuredError(
				utils.ErrorTypeValidation,
				utils.CodeResourceLimitTimeout,
				"file processing canceled",
				filePath,
				nil,
			),
			"File processing canceled: %s", filePath,
		)

		return
	default:
	}

	content, err := os.ReadFile(filePath) // #nosec G304 - filePath is validated by walker
	if err != nil {
		structErr := utils.WrapError(
			err,
			utils.ErrorTypeProcessing,
			utils.CodeProcessingFileRead,
			"failed to read file",
		).WithFilePath(filePath)
		utils.LogErrorf(structErr, "Failed to read file %s", filePath)

		return
	}

	// Check context again after reading
	select {
	case <-ctx.Done():
		utils.LogErrorf(
			utils.NewStructuredError(
				utils.ErrorTypeValidation,
				utils.CodeResourceLimitTimeout,
				"file processing canceled after read",
				filePath,
				nil,
			),
			"File processing canceled after read: %s", filePath,
		)

		return
	default:
	}

	// Try to send the result, but respect context cancellation
	select {
	case <-ctx.Done():
		utils.LogErrorf(
			utils.NewStructuredError(
				utils.ErrorTypeValidation,
				utils.CodeResourceLimitTimeout,
				"file processing canceled before output",
				filePath,
				nil,
			),
			"File processing canceled before output: %s", filePath,
		)

		return
	case outCh <- WriteRequest{
		Path:     relPath,
		Content:  p.formatContent(relPath, string(content)),
		IsStream: false,
		Size:     int64(len(content)),
	}:
	}
}

// processStreamingWithContext creates a streaming reader for large files with context awareness.
func (p *FileProcessor) processStreamingWithContext(
	ctx context.Context,
	filePath, relPath string,
	outCh chan<- WriteRequest,
) {
	// Check context before creating reader
	select {
	case <-ctx.Done():
		utils.LogErrorf(
			utils.NewStructuredError(
				utils.ErrorTypeValidation,
				utils.CodeResourceLimitTimeout,
				"streaming processing canceled",
				filePath,
				nil,
			),
			"Streaming processing canceled: %s", filePath,
		)

		return
	default:
	}

	// Get file info for size
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		utils.LogErrorf(
			utils.WrapError(err, utils.ErrorTypeIO, utils.CodeIORead,
				"failed to stat file for streaming").WithFilePath(filePath),
			"Failed to stat file for streaming: %s", filePath,
		)
		return
	}

	reader := p.createStreamReaderWithContext(ctx, filePath, relPath)
	if reader == nil {
		return // Error already logged
	}

	// Try to send the result, but respect context cancellation
	select {
	case <-ctx.Done():
		utils.LogErrorf(
			utils.NewStructuredError(
				utils.ErrorTypeValidation,
				utils.CodeResourceLimitTimeout,
				"streaming processing canceled before output",
				filePath,
				nil,
			),
			"Streaming processing canceled before output: %s", filePath,
		)

		return
	case outCh <- WriteRequest{
		Path:     relPath,
		Content:  "", // Empty since content is in Reader
		IsStream: true,
		Reader:   reader,
		Size:     fileInfo.Size(),
	}:
	}
}

// createStreamReaderWithContext creates a reader that combines header and file content with context awareness.
func (p *FileProcessor) createStreamReaderWithContext(ctx context.Context, filePath, relPath string) io.Reader {
	// Check context before opening file
	select {
	case <-ctx.Done():
		return nil
	default:
	}

	file, err := os.Open(filePath) // #nosec G304 - filePath is validated by walker
	if err != nil {
		structErr := utils.WrapError(
			err,
			utils.ErrorTypeProcessing,
			utils.CodeProcessingFileRead,
			"failed to open file for streaming",
		).WithFilePath(filePath)
		utils.LogErrorf(structErr, "Failed to open file for streaming %s", filePath)

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
		return n, utils.WrapError(err, utils.ErrorTypeIO, utils.CodeIORead, "failed to read from header file reader")
	}

	return n, nil
}

// closeFile closes the file once.
func (r *headerFileReader) closeFile() {
	r.mu.Lock()
	defer r.mu.Unlock()
	if !r.closed && r.file != nil {
		if err := r.file.Close(); err != nil {
			utils.LogError("Failed to close file", err)
		}
		r.closed = true
	}
}
