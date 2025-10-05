package fileproc

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/ivuorinen/gibidify/gibidiutils"
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
	if _, err := fmt.Fprintf(w.outFile, "prefix: %s\nsuffix: %s\nfiles:\n", yamlQuoteString(prefix), yamlQuoteString(suffix)); err != nil {
		return gibidiutils.WrapError(err, gibidiutils.ErrorTypeIO, gibidiutils.CodeIOWrite, "failed to write YAML header")
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
	defer w.closeReader(req.Reader, req.Path)

	language := detectLanguage(req.Path)

	// Write YAML file entry start
	if _, err := fmt.Fprintf(w.outFile, "  - path: %s\n    language: %s\n    content: |\n", yamlQuoteString(req.Path), language); err != nil {
		return gibidiutils.WrapError(err, gibidiutils.ErrorTypeIO, gibidiutils.CodeIOWrite, "failed to write YAML file start").WithFilePath(req.Path)
	}

	// Stream content with YAML indentation
	return w.streamYAMLContent(req.Reader, req.Path)
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
	if _, err := fmt.Fprintf(w.outFile, "  - path: %s\n    language: %s\n    content: |\n", yamlQuoteString(fileData.Path), fileData.Language); err != nil {
		return gibidiutils.WrapError(err, gibidiutils.ErrorTypeIO, gibidiutils.CodeIOWrite, "failed to write YAML entry start").WithFilePath(req.Path)
	}

	// Write indented content
	lines := strings.Split(fileData.Content, "\n")
	for _, line := range lines {
		if _, err := fmt.Fprintf(w.outFile, "      %s\n", line); err != nil {
			return gibidiutils.WrapError(err, gibidiutils.ErrorTypeIO, gibidiutils.CodeIOWrite, "failed to write YAML content line").WithFilePath(req.Path)
		}
	}

	return nil
}

// streamYAMLContent streams content with YAML indentation.
func (w *YAMLWriter) streamYAMLContent(reader io.Reader, path string) error {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		if _, err := fmt.Fprintf(w.outFile, "      %s\n", line); err != nil {
			return gibidiutils.WrapError(err, gibidiutils.ErrorTypeIO, gibidiutils.CodeIOWrite, "failed to write YAML line").WithFilePath(path)
		}
	}

	if err := scanner.Err(); err != nil {
		return gibidiutils.WrapError(err, gibidiutils.ErrorTypeIO, gibidiutils.CodeIORead, "failed to scan YAML content").WithFilePath(path)
	}
	return nil
}

// closeReader safely closes a reader if it implements io.Closer.
func (w *YAMLWriter) closeReader(reader io.Reader, path string) {
	if closer, ok := reader.(io.Closer); ok {
		if err := closer.Close(); err != nil {
			gibidiutils.LogError(
				"Failed to close file reader",
				gibidiutils.WrapError(err, gibidiutils.ErrorTypeIO, gibidiutils.CodeIOClose, "failed to close file reader").WithFilePath(path),
			)
		}
	}
}

// yamlQuoteString quotes a string for YAML output if needed.
func yamlQuoteString(s string) string {
	if s == "" {
		return `""`
	}
	// Simple YAML quoting - use double quotes if string contains special characters
	if strings.ContainsAny(s, "\n\r\t:\"'\\") {
		return fmt.Sprintf(`"%s"`, strings.ReplaceAll(s, `"`, `\"`))
	}
	return s
}

// startYAMLWriter handles YAML format output with streaming support.
func startYAMLWriter(outFile *os.File, writeCh <-chan WriteRequest, done chan<- struct{}, prefix, suffix string) {
	defer close(done)

	writer := NewYAMLWriter(outFile)

	// Start writing
	if err := writer.Start(prefix, suffix); err != nil {
		gibidiutils.LogError("Failed to write YAML header", err)
		return
	}

	// Process files
	for req := range writeCh {
		if err := writer.WriteFile(req); err != nil {
			gibidiutils.LogError("Failed to write YAML file", err)
		}
	}

	// Close writer
	if err := writer.Close(); err != nil {
		gibidiutils.LogError("Failed to write YAML end", err)
	}
}
