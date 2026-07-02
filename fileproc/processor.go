// Package fileproc provides functions for processing files.
package fileproc

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

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
	rootPath  string
	sizeLimit int64
}

// NewFileProcessor creates a new file processor.
func NewFileProcessor(rootPath string) *FileProcessor {
	return &FileProcessor{
		rootPath:  rootPath,
		sizeLimit: config.FileSizeLimit(),
	}
}

// ProcessFile reads the file at filePath and sends a formatted output to outCh.
// It automatically chooses between loading the entire file or streaming based on file size.
func ProcessFile(filePath string, outCh chan<- WriteRequest, rootPath string) {
	if err := ProcessFileContext(context.Background(), filePath, outCh, rootPath); err != nil {
		shared.LogErrorf(err, shared.FileProcessingMsgFailedToProcess, filePath)
	}
}

// ProcessFileContext processes a file, honoring context cancellation.
func ProcessFileContext(ctx context.Context, filePath string, outCh chan<- WriteRequest, rootPath string) error {
	return NewFileProcessor(rootPath).ProcessWithContext(ctx, filePath, outCh)
}

// Process handles file processing with the configured settings.
func (p *FileProcessor) Process(filePath string, outCh chan<- WriteRequest) {
	if err := p.ProcessWithContext(context.Background(), filePath, outCh); err != nil {
		shared.LogErrorf(err, shared.FileProcessingMsgFailedToProcess, filePath)
	}
}

// ProcessWithContext handles file processing with context cancellation.
func (p *FileProcessor) ProcessWithContext(ctx context.Context, filePath string, outCh chan<- WriteRequest) error {
	fileInfo, err := p.validateFile(ctx, filePath)
	if err != nil {
		return err // Error already logged
	}

	relPath := p.getRelativePath(filePath)

	// Choose processing strategy based on file size
	if fileInfo.Size() <= shared.FileProcessingStreamThreshold {
		return p.processInMemoryWithContext(ctx, filePath, relPath, outCh)
	}

	return p.processStreamingWithContext(ctx, filePath, relPath, outCh, fileInfo.Size())
}

// validateFile stats the file and enforces the configured size limit.
func (p *FileProcessor) validateFile(ctx context.Context, filePath string) (os.FileInfo, error) {
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

	if fileInfo.Size() > p.sizeLimit {
		structErr := shared.NewStructuredError(
			shared.ErrorTypeValidation,
			shared.CodeValidationSize,
			fmt.Sprintf(shared.FileProcessingMsgSizeExceeds, fileInfo.Size(), p.sizeLimit),
			filePath,
			map[string]any{"file_size": fileInfo.Size(), "size_limit": p.sizeLimit},
		)
		shared.LogErrorf(structErr, "Skipping large file %s", filePath)

		return nil, structErr
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
	if err := shared.CheckContextCancellation(ctx, "file read"); err != nil {
		return fmt.Errorf("context check before read: %w", err)
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

	select {
	case <-ctx.Done():
		return fmt.Errorf("file processing canceled before output: %w", ctx.Err())
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
	reader := p.createStreamReader(filePath, relPath)
	if reader == nil {
		return shared.NewStructuredError(
			shared.ErrorTypeProcessing,
			shared.CodeProcessingFileRead,
			"failed to create stream reader",
			filePath,
			nil,
		)
	}

	select {
	case <-ctx.Done():
		return fmt.Errorf("streaming processing canceled before output: %w", ctx.Err())
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

// createStreamReader creates a reader that combines header and file content.
func (p *FileProcessor) createStreamReader(filePath, relPath string) io.Reader {
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

	return newHeaderFileReader(p.formatHeader(relPath), file)
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
