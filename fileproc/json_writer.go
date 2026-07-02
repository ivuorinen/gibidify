// Package fileproc handles file processing, collection, and output formatting.
package fileproc

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/ivuorinen/gibidify/shared"
)

// JSONWriter handles JSON format output.
type JSONWriter struct {
	outFile   *os.File
	firstFile bool
}

// NewJSONWriter creates a new JSON writer.
func NewJSONWriter(outFile *os.File) *JSONWriter {
	return &JSONWriter{
		outFile:   outFile,
		firstFile: true,
	}
}

// Start writes the JSON header.
func (w *JSONWriter) Start(prefix, suffix string) error {
	// Start JSON structure
	if _, err := w.outFile.WriteString(`{"prefix":"`); err != nil {
		return shared.WrapError(err, shared.ErrorTypeIO, shared.CodeIOWrite, "failed to write JSON start")
	}

	// Write escaped prefix
	escapedPrefix := shared.EscapeForJSON(prefix)
	if err := shared.WriteWithErrorWrap(w.outFile, escapedPrefix, "failed to write JSON prefix", ""); err != nil {
		return fmt.Errorf("writing JSON prefix: %w", err)
	}

	if _, err := w.outFile.WriteString(`","suffix":"`); err != nil {
		return shared.WrapError(err, shared.ErrorTypeIO, shared.CodeIOWrite, "failed to write JSON middle")
	}

	// Write escaped suffix
	escapedSuffix := shared.EscapeForJSON(suffix)
	if err := shared.WriteWithErrorWrap(w.outFile, escapedSuffix, "failed to write JSON suffix", ""); err != nil {
		return fmt.Errorf("writing JSON suffix: %w", err)
	}

	if _, err := w.outFile.WriteString(`","files":[`); err != nil {
		return shared.WrapError(err, shared.ErrorTypeIO, shared.CodeIOWrite, "failed to write JSON files start")
	}

	return nil
}

// WriteFile writes a file entry in JSON format.
func (w *JSONWriter) WriteFile(req WriteRequest) error {
	if !w.firstFile {
		if _, err := w.outFile.WriteString(","); err != nil {
			return shared.WrapError(err, shared.ErrorTypeIO, shared.CodeIOWrite, "failed to write JSON separator")
		}
	}
	w.firstFile = false

	return w.writeInline(req)
}

// Close writes the JSON footer.
func (w *JSONWriter) Close() error {
	// Close JSON structure
	if _, err := w.outFile.WriteString("]}"); err != nil {
		return shared.WrapError(err, shared.ErrorTypeIO, shared.CodeIOWrite, "failed to write JSON end")
	}

	return nil
}

// writeInline writes a file directly as JSON.
func (w *JSONWriter) writeInline(req WriteRequest) error {
	language := detectLanguage(req.Path)
	fileData := FileData{
		Path:     req.Path,
		Content:  req.Content,
		Language: language,
	}

	encoded, err := json.Marshal(fileData)
	if err != nil {
		return shared.WrapError(
			err,
			shared.ErrorTypeProcessing,
			shared.CodeProcessingEncode,
			"failed to marshal JSON",
		).WithFilePath(req.Path)
	}

	if _, err := w.outFile.Write(encoded); err != nil {
		return shared.WrapError(
			err,
			shared.ErrorTypeIO,
			shared.CodeIOWrite,
			"failed to write JSON file",
		).WithFilePath(req.Path)
	}

	return nil
}

// startJSONWriter handles JSON format output.
func startJSONWriter(outFile *os.File, writeCh <-chan WriteRequest, done chan<- struct{}, prefix, suffix string) {
	startFormatWriter(outFile, writeCh, done, prefix, suffix, func(f *os.File) FormatWriter {
		return NewJSONWriter(f)
	})
}
