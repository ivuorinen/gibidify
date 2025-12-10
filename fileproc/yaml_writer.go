// Package fileproc handles file processing, collection, and output formatting.
package fileproc

import (
	"fmt"
	"os"
	"strings"

	"github.com/ivuorinen/gibidify/shared"
)

// YAMLWriter handles YAML format output with streaming support.
type YAMLWriter struct {
	outFile *os.File
}

// NewYAMLWriter creates a new YAML writer.
func NewYAMLWriter(outFile *os.File) *YAMLWriter {
	return &YAMLWriter{outFile: outFile}
}

// Start writes the YAML header.
func (w *YAMLWriter) Start(prefix, suffix string) error {
	// Write YAML header
	if _, err := fmt.Fprintf(
		w.outFile,
		"prefix: %s\nsuffix: %s\nfiles:\n",
		shared.EscapeForYAML(prefix),
		shared.EscapeForYAML(suffix),
	); err != nil {
		return shared.WrapError(err, shared.ErrorTypeIO, shared.CodeIOWrite, "failed to write YAML header")
	}

	return nil
}

// WriteFile writes a file entry in YAML format.
func (w *YAMLWriter) WriteFile(req WriteRequest) error {
	if req.IsStream {
		return w.writeStreaming(req)
	}

	return w.writeInline(req)
}

// Close writes the YAML footer (no footer needed for YAML).
func (w *YAMLWriter) Close() error {
	return nil
}

// writeStreaming writes a large file as YAML in streaming chunks.
func (w *YAMLWriter) writeStreaming(req WriteRequest) error {
	defer shared.SafeCloseReader(req.Reader, req.Path)

	language := detectLanguage(req.Path)

	// Write YAML file entry start
	if _, err := fmt.Fprintf(
		w.outFile,
		shared.YAMLFmtFileEntry,
		shared.EscapeForYAML(req.Path),
		language,
	); err != nil {
		return shared.WrapError(
			err,
			shared.ErrorTypeIO,
			shared.CodeIOWrite,
			"failed to write YAML file start",
		).WithFilePath(req.Path)
	}

	// Stream content with YAML indentation
	if err := shared.StreamLines(
		req.Reader, w.outFile, req.Path, func(line string) string {
			return "      " + line
		},
	); err != nil {
		return shared.WrapError(err, shared.ErrorTypeIO, shared.CodeIOWrite, "streaming YAML content")
	}

	return nil
}

// writeInline writes a small file directly as YAML.
func (w *YAMLWriter) writeInline(req WriteRequest) error {
	language := detectLanguage(req.Path)
	fileData := FileData{
		Path:     req.Path,
		Content:  req.Content,
		Language: language,
	}

	// Write YAML entry
	if _, err := fmt.Fprintf(
		w.outFile,
		shared.YAMLFmtFileEntry,
		shared.EscapeForYAML(fileData.Path),
		fileData.Language,
	); err != nil {
		return shared.WrapError(
			err,
			shared.ErrorTypeIO,
			shared.CodeIOWrite,
			"failed to write YAML entry start",
		).WithFilePath(req.Path)
	}

	// Write indented content
	lines := strings.Split(fileData.Content, "\n")
	for _, line := range lines {
		if _, err := fmt.Fprintf(w.outFile, "      %s\n", line); err != nil {
			return shared.WrapError(
				err,
				shared.ErrorTypeIO,
				shared.CodeIOWrite,
				"failed to write YAML content line",
			).WithFilePath(req.Path)
		}
	}

	return nil
}

// startYAMLWriter handles YAML format output with streaming support.
func startYAMLWriter(outFile *os.File, writeCh <-chan WriteRequest, done chan<- struct{}, prefix, suffix string) {
	startFormatWriter(outFile, writeCh, done, prefix, suffix, func(f *os.File) FormatWriter {
		return NewYAMLWriter(f)
	})
}
