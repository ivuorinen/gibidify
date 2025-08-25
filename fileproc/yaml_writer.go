package fileproc

import (
	"fmt"
	"os"
	"strings"

	"github.com/ivuorinen/gibidify/utils"
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
		utils.EscapeForYAML(prefix),
		utils.EscapeForYAML(suffix),
	); err != nil {
		return utils.WrapError(err, utils.ErrorTypeIO, utils.CodeIOWrite, "failed to write YAML header")
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
	defer utils.SafeCloseReader(req.Reader, req.Path)

	language := detectLanguage(req.Path)

	// Write YAML file entry start
	if _, err := fmt.Fprintf(
		w.outFile,
		"  - path: %s\n    language: %s\n    content: |\n",
		utils.EscapeForYAML(req.Path),
		language,
	); err != nil {
		return utils.WrapError(
			err,
			utils.ErrorTypeIO,
			utils.CodeIOWrite,
			"failed to write YAML file start",
		).WithFilePath(req.Path)
	}

	// Stream content with YAML indentation
	if err := utils.StreamLines(
		req.Reader, w.outFile, req.Path, func(line string) string {
			return "      " + line
		},
	); err != nil {
		return utils.WrapError(err, utils.ErrorTypeIO, utils.CodeIOWrite, "streaming YAML content")
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
		"  - path: %s\n    language: %s\n    content: |\n",
		utils.EscapeForYAML(fileData.Path),
		fileData.Language,
	); err != nil {
		return utils.WrapError(
			err,
			utils.ErrorTypeIO,
			utils.CodeIOWrite,
			"failed to write YAML entry start",
		).WithFilePath(req.Path)
	}

	// Write indented content
	lines := strings.Split(fileData.Content, "\n")
	for _, line := range lines {
		if _, err := fmt.Fprintf(w.outFile, "      %s\n", line); err != nil {
			return utils.WrapError(
				err,
				utils.ErrorTypeIO,
				utils.CodeIOWrite,
				"failed to write YAML content line",
			).WithFilePath(req.Path)
		}
	}

	return nil
}

// startYAMLWriter handles YAML format output with streaming support.
func startYAMLWriter(outFile *os.File, writeCh <-chan WriteRequest, done chan<- struct{}, prefix, suffix string) {
	defer close(done)

	writer := NewYAMLWriter(outFile)

	// Start writing
	if err := writer.Start(prefix, suffix); err != nil {
		utils.LogError("Failed to write YAML header", err)

		return
	}

	// Process files
	for req := range writeCh {
		if err := writer.WriteFile(req); err != nil {
			utils.LogError("Failed to write YAML file", err)
		}
	}

	// Close writer
	if err := writer.Close(); err != nil {
		utils.LogError("Failed to write YAML end", err)
	}
}
