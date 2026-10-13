// Package fileproc provides functions for processing files.
package fileproc

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/ivuorinen/gibidify/config"
	"github.com/ivuorinen/gibidify/shared"
)

// binarySniffLen is how many leading bytes are checked for a NUL byte, the same
// window git uses to classify a file as binary.
const binarySniffLen = 8000

// ErrSkipped marks a file that was deliberately left out of the output (binary
// content, not a regular file). Callers count it as skipped, not failed.
var ErrSkipped = errors.New("file skipped")

// errSymlink is returned by openNoFollow when the path is a symlink.
var errSymlink = errors.New("path is a symlink")

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

// ProcessFileContext processes a file, honoring context cancellation.
func ProcessFileContext(ctx context.Context, filePath string, outCh chan<- WriteRequest, rootPath string) error {
	return NewFileProcessor(rootPath).ProcessWithContext(ctx, filePath, outCh)
}

// ProcessWithContext handles file processing with context cancellation.
// Files are always buffered whole; the per-file memory ceiling is
// config.FileSizeLimit (default 5MB), enforced during the read. Streaming would
// only be worth reintroducing if that cap were raised high enough that
// concurrent buffering became a memory problem.
// Errors are returned, not logged: the caller owns reporting so each failure
// is logged once.
func (p *FileProcessor) ProcessWithContext(ctx context.Context, filePath string, outCh chan<- WriteRequest) error {
	if err := shared.CheckContextCancellation(ctx, "file validation"); err != nil {
		return fmt.Errorf("context check during file validation: %w", err)
	}

	f, err := openNoFollow(filePath)
	if errors.Is(err, errSymlink) {
		return fmt.Errorf("%w: %s is a symlink", ErrSkipped, filePath)
	}
	if err != nil {
		return shared.WrapError(
			err,
			shared.ErrorTypeFileSystem,
			shared.CodeFSAccess,
			"failed to open file",
		).WithFilePath(filePath)
	}
	defer func() {
		if cerr := f.Close(); cerr != nil {
			shared.LogErrorf(cerr, "Failed to close file %s", filePath)
		}
	}()

	if err := p.validateFile(f, filePath); err != nil {
		return err
	}

	return p.processInMemoryWithContext(ctx, f, filePath, p.getRelativePath(filePath), outCh)
}

// validateFile enforces that the opened file is a regular file within the size
// limit. It checks the open descriptor, not the path: a separate Lstat followed
// by an open by name left a window in which a path swapped for a symlink was
// followed out of the source tree, or one swapped for a FIFO blocked forever.
func (p *FileProcessor) validateFile(f *os.File, filePath string) error {
	fileInfo, err := f.Stat()
	if err != nil {
		return shared.WrapError(
			err,
			shared.ErrorTypeFileSystem,
			shared.CodeFSAccess,
			"failed to stat file",
		).WithFilePath(filePath)
	}

	if !fileInfo.Mode().IsRegular() {
		return fmt.Errorf("%w: %s is not a regular file", ErrSkipped, filePath)
	}

	if fileInfo.Size() > p.sizeLimit {
		return shared.NewStructuredError(
			shared.ErrorTypeValidation,
			shared.CodeValidationSize,
			fmt.Sprintf(shared.FileProcessingMsgSizeExceeds, fileInfo.Size(), p.sizeLimit),
			filePath,
			map[string]any{"file_size": fileInfo.Size(), "size_limit": p.sizeLimit},
		)
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
	f *os.File,
	filePath, relPath string,
	outCh chan<- WriteRequest,
) error {
	if err := shared.CheckContextCancellation(ctx, "file read"); err != nil {
		return fmt.Errorf("context check before read: %w", err)
	}

	content, err := p.readCapped(f, filePath)
	if err != nil {
		return err
	}

	// Extension lists cannot recognize extensionless binaries; without this
	// check they reach the output as megabytes of U+FFFD replacement runes.
	if bytes.IndexByte(content[:min(len(content), binarySniffLen)], 0) >= 0 {
		return fmt.Errorf("%w: %s has binary content", ErrSkipped, filePath)
	}

	// Every output format is UTF-8 text. JSON replaced invalid bytes silently
	// and YAML emitted an unparseable document, so the replacement happens once
	// here, visibly, and identically for every format.
	text := string(content)
	if !utf8.ValidString(text) {
		shared.GetLogger().Warnf("%s is not valid UTF-8; invalid bytes are replaced with U+FFFD", filePath)
		text = strings.ToValidUTF8(text, string(utf8.RuneError))
	}

	select {
	case <-ctx.Done():
		return fmt.Errorf("file processing canceled before output: %w", ctx.Err())
	case outCh <- WriteRequest{Path: relPath, Content: text}:
	}

	return nil
}

// readCapped reads the whole file but never buffers more than sizeLimit bytes,
// so a file that grows after the size check cannot blow past the per-file
// memory ceiling. Exceeding the cap is reported as a size error.
func (p *FileProcessor) readCapped(f *os.File, filePath string) ([]byte, error) {
	// Read up to sizeLimit+1 so an over-limit file is detectable.
	content, err := io.ReadAll(io.LimitReader(f, p.sizeLimit+1))
	if err != nil {
		return nil, p.wrapReadError(err, filePath)
	}
	if int64(len(content)) > p.sizeLimit {
		return nil, shared.NewStructuredError(
			shared.ErrorTypeValidation,
			shared.CodeValidationSize,
			fmt.Sprintf(shared.FileProcessingMsgSizeExceeds, len(content), p.sizeLimit),
			filePath,
			map[string]any{"size_limit": p.sizeLimit},
		)
	}

	return content, nil
}

// wrapReadError wraps a file-read failure in a structured error.
func (p *FileProcessor) wrapReadError(err error, filePath string) error {
	return shared.WrapError(
		err,
		shared.ErrorTypeProcessing,
		shared.CodeProcessingFileRead,
		"failed to read file",
	).WithFilePath(filePath)
}
