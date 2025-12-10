// Package fileproc handles file processing, collection, and output formatting.
package fileproc

import (
	"fmt"
	"os"

	"github.com/ivuorinen/gibidify/shared"
)

// MarkdownWriter handles Markdown format output with streaming support.
type MarkdownWriter struct {
	outFile *os.File
	suffix  string
}

// NewMarkdownWriter creates a new markdown writer.
func NewMarkdownWriter(outFile *os.File) *MarkdownWriter {
	return &MarkdownWriter{outFile: outFile}
}

// Start writes the markdown header and stores the suffix for later use.
func (w *MarkdownWriter) Start(prefix, suffix string) error {
	// Store suffix for use in Close method
	w.suffix = suffix

	if prefix != "" {
		if _, err := fmt.Fprintf(w.outFile, "# %s\n\n", prefix); err != nil {
			return shared.WrapError(err, shared.ErrorTypeIO, shared.CodeIOWrite, "failed to write prefix")
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

// Close writes the markdown footer using the suffix stored in Start.
func (w *MarkdownWriter) Close() error {
	if w.suffix != "" {
		if _, err := fmt.Fprintf(w.outFile, "\n# %s\n", w.suffix); err != nil {
			return shared.WrapError(err, shared.ErrorTypeIO, shared.CodeIOWrite, "failed to write suffix")
		}
	}

	return nil
}

// writeStreaming writes a large file in streaming chunks.
func (w *MarkdownWriter) writeStreaming(req WriteRequest) error {
	defer shared.SafeCloseReader(req.Reader, req.Path)

	language := detectLanguage(req.Path)

	// Write file header
	if _, err := fmt.Fprintf(w.outFile, "## File: `%s`\n```%s\n", req.Path, language); err != nil {
		return shared.WrapError(
			err,
			shared.ErrorTypeIO,
			shared.CodeIOWrite,
			"failed to write file header",
		).WithFilePath(req.Path)
	}

	// Stream file content in chunks
	chunkSize := shared.FileProcessingStreamChunkSize
	if err := shared.StreamContent(req.Reader, w.outFile, chunkSize, req.Path, nil); err != nil {
		return shared.WrapError(err, shared.ErrorTypeIO, shared.CodeIOWrite, "streaming content for markdown file")
	}

	// Write file footer
	if _, err := w.outFile.WriteString("\n```\n\n"); err != nil {
		return shared.WrapError(
			err,
			shared.ErrorTypeIO,
			shared.CodeIOWrite,
			"failed to write file footer",
		).WithFilePath(req.Path)
	}

	return nil
}

// writeInline writes a small file directly from content.
func (w *MarkdownWriter) writeInline(req WriteRequest) error {
	language := detectLanguage(req.Path)
	formatted := fmt.Sprintf("## File: `%s`\n```%s\n%s\n```\n\n", req.Path, language, req.Content)

	if _, err := w.outFile.WriteString(formatted); err != nil {
		return shared.WrapError(
			err,
			shared.ErrorTypeIO,
			shared.CodeIOWrite,
			"failed to write inline content",
		).WithFilePath(req.Path)
	}

	return nil
}

// startMarkdownWriter handles Markdown format output with streaming support.
func startMarkdownWriter(outFile *os.File, writeCh <-chan WriteRequest, done chan<- struct{}, prefix, suffix string) {
	startFormatWriter(outFile, writeCh, done, prefix, suffix, func(f *os.File) FormatWriter {
		return NewMarkdownWriter(f)
	})
}
