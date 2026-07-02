// Package fileproc provides functions for processing files.
package fileproc

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ivuorinen/gibidify/config"
	"github.com/ivuorinen/gibidify/shared"
)

// WriteRequest represents the content to be written.
type WriteRequest struct {
	Path    string
	Content string
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
// ponytail: files are always buffered whole; the memory ceiling per file is
// config.FileSizeLimit (default 5MB). Reintroduce streaming only if the cap is
// raised high enough that concurrent buffering becomes a memory problem.
func (p *FileProcessor) ProcessWithContext(ctx context.Context, filePath string, outCh chan<- WriteRequest) error {
	if err := p.validateFile(ctx, filePath); err != nil {
		return err // Error already logged
	}

	relPath := p.getRelativePath(filePath)

	return p.processInMemoryWithContext(ctx, filePath, relPath, outCh)
}

// validateFile stats the file and enforces the configured size limit.
func (p *FileProcessor) validateFile(ctx context.Context, filePath string) error {
	if err := shared.CheckContextCancellation(ctx, "file validation"); err != nil {
		return fmt.Errorf("context check during file validation: %w", err)
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

		return structErr
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

		return structErr
	}

	return nil
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
		Path:    relPath,
		Content: p.formatContent(relPath, string(content)),
	}:
	}

	return nil
}

// formatContent formats the file content with header.
func (p *FileProcessor) formatContent(relPath, content string) string {
	return fmt.Sprintf("\n---\n%s\n%s\n", relPath, content)
}
