package fileproc

import (
	"fmt"
	"os"

	"github.com/ivuorinen/gibidify/utils"
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
			return utils.WrapError(err, utils.ErrorTypeIO, utils.CodeIOWrite, "failed to write prefix")
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
			return utils.WrapError(err, utils.ErrorTypeIO, utils.CodeIOWrite, "failed to write suffix")
		}
	}

	return nil
}

// writeStreaming writes a large file in streaming chunks.
func (w *MarkdownWriter) writeStreaming(req WriteRequest) error {
	defer utils.SafeCloseReader(req.Reader, req.Path)

	language := detectLanguage(req.Path)

	// Write file header
	if _, err := fmt.Fprintf(w.outFile, "## File: `%s`\n```%s\n", req.Path, language); err != nil {
		return utils.WrapError(
			err,
			utils.ErrorTypeIO,
			utils.CodeIOWrite,
			"failed to write file header",
		).WithFilePath(req.Path)
	}

	// Stream file content in chunks
	if err := utils.StreamContent(req.Reader, w.outFile, StreamChunkSize, req.Path, nil); err != nil {
		return fmt.Errorf("streaming content for markdown file: %w", err)
	}

	// Write file footer
	if _, err := w.outFile.WriteString("\n```\n\n"); err != nil {
		return utils.WrapError(
			err,
			utils.ErrorTypeIO,
			utils.CodeIOWrite,
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
		return utils.WrapError(
			err,
			utils.ErrorTypeIO,
			utils.CodeIOWrite,
			"failed to write inline content",
		).WithFilePath(req.Path)
	}

	return nil
}

// startMarkdownWriter handles Markdown format output with streaming support.
func startMarkdownWriter(outFile *os.File, writeCh <-chan WriteRequest, done chan<- struct{}, prefix, suffix string) {
	defer close(done)

	writer := NewMarkdownWriter(outFile)

	// Start writing
	if err := writer.Start(prefix, suffix); err != nil {
		utils.LogError("Failed to write markdown prefix", err)

		return
	}

	// Process files
	for req := range writeCh {
		if err := writer.WriteFile(req); err != nil {
			utils.LogError("Failed to write markdown file", err)
		}
	}

	// Close writer
	if err := writer.Close(); err != nil {
		utils.LogError("Failed to write markdown suffix", err)
	}
}
