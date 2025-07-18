package fileproc

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/ivuorinen/gibidify/utils"
)

// JSONWriter handles JSON format output with streaming support.
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
		return utils.WrapError(err, utils.ErrorTypeIO, utils.CodeIOWrite, "failed to write JSON start")
	}

	// Write escaped prefix
	escapedPrefix := escapeJSONString(prefix)
	if _, err := w.outFile.WriteString(escapedPrefix); err != nil {
		return utils.WrapError(err, utils.ErrorTypeIO, utils.CodeIOWrite, "failed to write JSON prefix")
	}

	if _, err := w.outFile.WriteString(`","suffix":"`); err != nil {
		return utils.WrapError(err, utils.ErrorTypeIO, utils.CodeIOWrite, "failed to write JSON middle")
	}

	// Write escaped suffix
	escapedSuffix := escapeJSONString(suffix)
	if _, err := w.outFile.WriteString(escapedSuffix); err != nil {
		return utils.WrapError(err, utils.ErrorTypeIO, utils.CodeIOWrite, "failed to write JSON suffix")
	}

	if _, err := w.outFile.WriteString(`","files":[`); err != nil {
		return utils.WrapError(err, utils.ErrorTypeIO, utils.CodeIOWrite, "failed to write JSON files start")
	}

	return nil
}

// WriteFile writes a file entry in JSON format.
func (w *JSONWriter) WriteFile(req WriteRequest) error {
	if !w.firstFile {
		if _, err := w.outFile.WriteString(","); err != nil {
			return utils.WrapError(err, utils.ErrorTypeIO, utils.CodeIOWrite, "failed to write JSON separator")
		}
	}
	w.firstFile = false

	if req.IsStream {
		return w.writeStreaming(req)
	}
	return w.writeInline(req)
}

// Close writes the JSON footer.
func (w *JSONWriter) Close() error {
	// Close JSON structure
	if _, err := w.outFile.WriteString("]}"); err != nil {
		return utils.WrapError(err, utils.ErrorTypeIO, utils.CodeIOWrite, "failed to write JSON end")
	}
	return nil
}

// writeStreaming writes a large file as JSON in streaming chunks.
func (w *JSONWriter) writeStreaming(req WriteRequest) error {
	defer w.closeReader(req.Reader, req.Path)

	language := detectLanguage(req.Path)

	// Write file start
	escapedPath := escapeJSONString(req.Path)
	if _, err := fmt.Fprintf(w.outFile, `{"path":"%s","language":"%s","content":"`, escapedPath, language); err != nil {
		return utils.WrapError(err, utils.ErrorTypeIO, utils.CodeIOWrite, "failed to write JSON file start").WithFilePath(req.Path)
	}

	// Stream content with JSON escaping
	if err := w.streamJSONContent(req.Reader, req.Path); err != nil {
		return err
	}

	// Write file end
	if _, err := w.outFile.WriteString(`"}`); err != nil {
		return utils.WrapError(err, utils.ErrorTypeIO, utils.CodeIOWrite, "failed to write JSON file end").WithFilePath(req.Path)
	}

	return nil
}

// writeInline writes a small file directly as JSON.
func (w *JSONWriter) writeInline(req WriteRequest) error {
	language := detectLanguage(req.Path)
	fileData := FileData{
		Path:     req.Path,
		Content:  req.Content,
		Language: language,
	}

	encoded, err := json.Marshal(fileData)
	if err != nil {
		return utils.WrapError(err, utils.ErrorTypeProcessing, utils.CodeProcessingEncode, "failed to marshal JSON").WithFilePath(req.Path)
	}

	if _, err := w.outFile.Write(encoded); err != nil {
		return utils.WrapError(err, utils.ErrorTypeIO, utils.CodeIOWrite, "failed to write JSON file").WithFilePath(req.Path)
	}
	return nil
}

// streamJSONContent streams content with JSON escaping.
func (w *JSONWriter) streamJSONContent(reader io.Reader, path string) error {
	buf := make([]byte, StreamChunkSize)
	for {
		n, err := reader.Read(buf)
		if n > 0 {
			escaped := escapeJSONString(string(buf[:n]))
			if _, writeErr := w.outFile.WriteString(escaped); writeErr != nil {
				return utils.WrapError(writeErr, utils.ErrorTypeIO, utils.CodeIOWrite, "failed to write JSON chunk").WithFilePath(path)
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return utils.WrapError(err, utils.ErrorTypeIO, utils.CodeIORead, "failed to read JSON chunk").WithFilePath(path)
		}
	}
	return nil
}

// closeReader safely closes a reader if it implements io.Closer.
func (w *JSONWriter) closeReader(reader io.Reader, path string) {
	if closer, ok := reader.(io.Closer); ok {
		if err := closer.Close(); err != nil {
			utils.LogError(
				"Failed to close file reader",
				utils.WrapError(err, utils.ErrorTypeIO, utils.CodeIOClose, "failed to close file reader").WithFilePath(path),
			)
		}
	}
}

// escapeJSONString escapes a string for JSON output.
func escapeJSONString(s string) string {
	// Use json.Marshal to properly escape the string, then remove the quotes
	escaped, _ := json.Marshal(s)
	return string(escaped[1 : len(escaped)-1]) // Remove surrounding quotes
}

// startJSONWriter handles JSON format output with streaming support.
func startJSONWriter(outFile *os.File, writeCh <-chan WriteRequest, done chan<- struct{}, prefix, suffix string) {
	defer close(done)

	writer := NewJSONWriter(outFile)

	// Start writing
	if err := writer.Start(prefix, suffix); err != nil {
		utils.LogError("Failed to write JSON start", err)
		return
	}

	// Process files
	for req := range writeCh {
		if err := writer.WriteFile(req); err != nil {
			utils.LogError("Failed to write JSON file", err)
		}
	}

	// Close writer
	if err := writer.Close(); err != nil {
		utils.LogError("Failed to write JSON end", err)
	}
}
