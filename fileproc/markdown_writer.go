// Package fileproc handles file processing, collection, and output formatting.
package fileproc

import (
	"fmt"
	"os"
	"strings"

	"github.com/ivuorinen/gibidify/shared"
)

// MarkdownWriter handles Markdown format output.
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

// writeInline writes a file directly from content.
func (w *MarkdownWriter) writeInline(req WriteRequest) error {
	language := detectLanguage(req.Path)
	fence := codeFence(req.Content)
	formatted := fmt.Sprintf("## File: `%s`\n%s%s\n%s\n%s\n\n", req.Path, fence, language, req.Content, fence)

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

// codeFence returns a backtick fence long enough that content cannot close it
// early. It uses one more backtick than the longest backtick run in content,
// with a minimum of three (the standard Markdown fence).
func codeFence(content string) string {
	longest, run := 0, 0
	for _, r := range content {
		if r == '`' {
			run++
			if run > longest {
				longest = run
			}
			continue
		}
		run = 0
	}

	n := longest + 1
	if n < 3 {
		n = 3
	}

	return strings.Repeat("`", n)
}

// startMarkdownWriter handles Markdown format output.
func startMarkdownWriter(outFile *os.File, writeCh <-chan WriteRequest, done chan<- struct{}, prefix, suffix string) {
	startFormatWriter(outFile, writeCh, done, prefix, suffix, func(f *os.File) FormatWriter {
		return NewMarkdownWriter(f)
	})
}
