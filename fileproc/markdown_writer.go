package fileproc

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/ivuorinen/gibidify/gibidiutils"
)

// MarkdownWriter handles Markdown format output with streaming support.
type MarkdownWriter struct {
	outFile *os.File
}

// NewMarkdownWriter creates a new markdown writer.
func NewMarkdownWriter(outFile *os.File) *MarkdownWriter {
	return &MarkdownWriter{outFile: outFile}
}

// Start writes the markdown header.
func (w *MarkdownWriter) Start(prefix, _ string) error {
	if prefix != "" {
		if _, err := fmt.Fprintf(w.outFile, "# %s\n\n", prefix); err != nil {
			return gibidiutils.WrapError(
				err,
				gibidiutils.ErrorTypeIO,
				gibidiutils.CodeIOWrite,
				"failed to write prefix",
			)
		}
	}
	return nil
}

// WriteFile writes a file entry in Markdown format.
func (w *MarkdownWriter) WriteFile(req WriteRequest) error {
	if req.IsStream {
		return w.writeStreaming(req)
	}
	return w.writeInline(req)
}

// Close writes the markdown footer.
func (w *MarkdownWriter) Close(suffix string) error {
	if suffix != "" {
		if _, err := fmt.Fprintf(w.outFile, "\n# %s\n", suffix); err != nil {
			return gibidiutils.WrapError(
				err,
				gibidiutils.ErrorTypeIO,
				gibidiutils.CodeIOWrite,
				"failed to write suffix",
			)
		}
	}
	return nil
}

// validateMarkdownPath validates a file path for markdown output.
func validateMarkdownPath(path string) error {
	trimmed := strings.TrimSpace(path)
	if trimmed == "" {
		return gibidiutils.NewStructuredError(
			gibidiutils.ErrorTypeValidation,
			gibidiutils.CodeValidationRequired,
			"file path cannot be empty",
			"",
			nil,
		)
	}

	// Reject absolute paths
	if filepath.IsAbs(trimmed) {
		return gibidiutils.NewStructuredError(
			gibidiutils.ErrorTypeValidation,
			gibidiutils.CodeValidationPath,
			"absolute paths are not allowed",
			trimmed,
			map[string]any{"path": trimmed},
		)
	}

	// Clean and validate path components
	cleaned := filepath.Clean(trimmed)
	if filepath.IsAbs(cleaned) || strings.HasPrefix(cleaned, "/") {
		return gibidiutils.NewStructuredError(
			gibidiutils.ErrorTypeValidation,
			gibidiutils.CodeValidationPath,
			"path must be relative",
			trimmed,
			map[string]any{"path": trimmed, "cleaned": cleaned},
		)
	}

	// Check for path traversal in components
	components := strings.Split(filepath.ToSlash(cleaned), "/")
	for _, component := range components {
		if component == ".." {
			return gibidiutils.NewStructuredError(
				gibidiutils.ErrorTypeValidation,
				gibidiutils.CodeValidationPath,
				"path traversal not allowed",
				trimmed,
				map[string]any{"path": trimmed, "cleaned": cleaned},
			)
		}
	}

	return nil
}

// writeStreaming writes a large file in streaming chunks.
func (w *MarkdownWriter) writeStreaming(req WriteRequest) error {
	// Validate path before use
	if err := validateMarkdownPath(req.Path); err != nil {
		return err
	}

	defer w.closeReader(req.Reader, req.Path)

	language := detectLanguage(req.Path)

	// Write file header
	if _, err := fmt.Fprintf(w.outFile, "## File: `%s`\n```%s\n", req.Path, language); err != nil {
		return gibidiutils.WrapError(
			err, gibidiutils.ErrorTypeIO, gibidiutils.CodeIOWrite,
			"failed to write file header",
		).WithFilePath(req.Path)
	}

	// Stream file content in chunks
	if err := w.streamContent(req.Reader, req.Path); err != nil {
		return err
	}

	// Write file footer
	if _, err := w.outFile.WriteString("\n```\n\n"); err != nil {
		return gibidiutils.WrapError(
			err, gibidiutils.ErrorTypeIO, gibidiutils.CodeIOWrite,
			"failed to write file footer",
		).WithFilePath(req.Path)
	}

	return nil
}

// writeInline writes a small file directly from content.
func (w *MarkdownWriter) writeInline(req WriteRequest) error {
	// Validate path before use
	if err := validateMarkdownPath(req.Path); err != nil {
		return err
	}

	language := detectLanguage(req.Path)
	formatted := fmt.Sprintf("## File: `%s`\n```%s\n%s\n```\n\n", req.Path, language, req.Content)

	if _, err := w.outFile.WriteString(formatted); err != nil {
		return gibidiutils.WrapError(
			err, gibidiutils.ErrorTypeIO, gibidiutils.CodeIOWrite,
			"failed to write inline content",
		).WithFilePath(req.Path)
	}
	return nil
}

// streamContent streams file content in chunks.
func (w *MarkdownWriter) streamContent(reader io.Reader, path string) error {
	buf := make([]byte, StreamChunkSize)
	for {
		n, err := reader.Read(buf)
		if n > 0 {
			if _, writeErr := w.outFile.Write(buf[:n]); writeErr != nil {
				return gibidiutils.WrapError(
					writeErr, gibidiutils.ErrorTypeIO, gibidiutils.CodeIOWrite,
					"failed to write chunk",
				).WithFilePath(path)
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return gibidiutils.WrapError(
				err, gibidiutils.ErrorTypeIO, gibidiutils.CodeIOFileRead,
				"failed to read chunk",
			).WithFilePath(path)
		}
	}
	return nil
}

// closeReader safely closes a reader if it implements io.Closer.
func (w *MarkdownWriter) closeReader(reader io.Reader, path string) {
	if closer, ok := reader.(io.Closer); ok {
		if err := closer.Close(); err != nil {
			gibidiutils.LogError(
				"Failed to close file reader",
				gibidiutils.WrapError(
					err, gibidiutils.ErrorTypeIO, gibidiutils.CodeIOClose,
					"failed to close file reader",
				).WithFilePath(path),
			)
		}
	}
}

// startMarkdownWriter handles Markdown format output with streaming support.
func startMarkdownWriter(
	outFile *os.File,
	writeCh <-chan WriteRequest,
	done chan<- struct{},
	prefix, suffix string,
) {
	defer close(done)

	writer := NewMarkdownWriter(outFile)

	// Start writing
	if err := writer.Start(prefix, suffix); err != nil {
		gibidiutils.LogError("Failed to write markdown prefix", err)
		return
	}

	// Process files
	for req := range writeCh {
		if err := writer.WriteFile(req); err != nil {
			gibidiutils.LogError("Failed to write markdown file", err)
		}
	}

	// Close writer
	if err := writer.Close(suffix); err != nil {
		gibidiutils.LogError("Failed to write markdown suffix", err)
	}
}
