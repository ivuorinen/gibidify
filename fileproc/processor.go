// Package fileproc provides functions for processing files.
package fileproc

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

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
		sizeLimit: config.GetFileSizeLimit(),
	}
}

// ProcessFile reads the file at filePath and sends a formatted output to outCh.
// It automatically chooses between loading the entire file or streaming based on file size.
func ProcessFile(filePath string, outCh chan<- WriteRequest, rootPath string) {
	processor := NewFileProcessor(rootPath)
	processor.Process(filePath, outCh)
}

// Process handles file processing with the configured settings.
func (p *FileProcessor) Process(filePath string, outCh chan<- WriteRequest) {
	// Validate file
	fileInfo, err := p.validateFile(filePath)
	if err != nil {
		return // Error already logged
	}

	// Get relative path
	relPath := p.getRelativePath(filePath)

	// Choose processing strategy based on file size
	if fileInfo.Size() <= StreamThreshold {
		p.processInMemory(filePath, relPath, outCh)
	} else {
		p.processStreaming(filePath, relPath, outCh)
	}
}

// validateFile checks if the file can be processed.
func (p *FileProcessor) validateFile(filePath string) (os.FileInfo, error) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		structErr := utils.WrapError(err, utils.ErrorTypeFileSystem, utils.CodeFSAccess, "failed to stat file").WithFilePath(filePath)
		utils.LogErrorf(structErr, "Failed to stat file %s", filePath)
		return nil, err
	}

	// Check size limit
	if fileInfo.Size() > p.sizeLimit {
		utils.LogErrorf(
			utils.NewStructuredError(
				utils.ErrorTypeValidation,
				utils.CodeValidationSize,
				fmt.Sprintf("file size (%d bytes) exceeds limit (%d bytes)", fileInfo.Size(), p.sizeLimit),
			).WithFilePath(filePath).WithContext("file_size", fileInfo.Size()).WithContext("size_limit", p.sizeLimit),
			"Skipping large file %s", filePath,
		)
		return nil, fmt.Errorf("file too large")
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

// processInMemory loads the entire file into memory (for small files).
func (p *FileProcessor) processInMemory(filePath, relPath string, outCh chan<- WriteRequest) {
	content, err := os.ReadFile(filePath) // #nosec G304 - filePath is validated by walker
	if err != nil {
		structErr := utils.WrapError(err, utils.ErrorTypeProcessing, utils.CodeProcessingFileRead, "failed to read file").WithFilePath(filePath)
		utils.LogErrorf(structErr, "Failed to read file %s", filePath)
		return
	}

	outCh <- WriteRequest{
		Path:     relPath,
		Content:  p.formatContent(relPath, string(content)),
		IsStream: false,
	}
}

// processStreaming creates a streaming reader for large files.
func (p *FileProcessor) processStreaming(filePath, relPath string, outCh chan<- WriteRequest) {
	reader := p.createStreamReader(filePath, relPath)
	if reader == nil {
		return // Error already logged
	}

	outCh <- WriteRequest{
		Path:     relPath,
		Content:  "", // Empty since content is in Reader
		IsStream: true,
		Reader:   reader,
	}
}

// createStreamReader creates a reader that combines header and file content.
func (p *FileProcessor) createStreamReader(filePath, relPath string) io.Reader {
	file, err := os.Open(filePath) // #nosec G304 - filePath is validated by walker
	if err != nil {
		structErr := utils.WrapError(err, utils.ErrorTypeProcessing, utils.CodeProcessingFileRead, "failed to open file for streaming").WithFilePath(filePath)
		utils.LogErrorf(structErr, "Failed to open file for streaming %s", filePath)
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
