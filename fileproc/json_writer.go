// Package fileproc handles file processing, collection, and output formatting.
package fileproc

import (
	"bytes"
	"encoding/json"
	"io"

	"github.com/ivuorinen/gibidify/shared"
)

// JSONWriter handles JSON format output.
type JSONWriter struct {
	out       io.Writer
	firstFile bool
}

// NewJSONWriter creates a new JSON writer.
func NewJSONWriter(out io.Writer) *JSONWriter {
	return &JSONWriter{
		out:       out,
		firstFile: true,
	}
}

// Start writes the prefix and suffix and opens the files array.
func (w *JSONWriter) Start(prefix, suffix string) error {
	encodedPrefix, err := encodeJSON(prefix)
	if err != nil {
		return shared.WrapError(err, shared.ErrorTypeProcessing, shared.CodeProcessingEncode, "failed to encode prefix")
	}
	encodedSuffix, err := encodeJSON(suffix)
	if err != nil {
		return shared.WrapError(err, shared.ErrorTypeProcessing, shared.CodeProcessingEncode, "failed to encode suffix")
	}

	header := `{"prefix":` + string(encodedPrefix) + `,"suffix":` + string(encodedSuffix) + `,"files":[`
	if _, err := io.WriteString(w.out, header); err != nil {
		return shared.WrapError(err, shared.ErrorTypeIO, shared.CodeIOWrite, "failed to write JSON start")
	}

	return nil
}

// WriteFile writes a file entry in JSON format.
func (w *JSONWriter) WriteFile(req WriteRequest) error {
	if !w.firstFile {
		if _, err := io.WriteString(w.out, ","); err != nil {
			return shared.WrapError(err, shared.ErrorTypeIO, shared.CodeIOWrite, "failed to write JSON separator")
		}
	}
	w.firstFile = false

	return w.writeInline(req)
}

// Close writes the JSON footer.
func (w *JSONWriter) Close() error {
	if _, err := io.WriteString(w.out, "]}"); err != nil {
		return shared.WrapError(err, shared.ErrorTypeIO, shared.CodeIOWrite, "failed to write JSON end")
	}

	return nil
}

// writeInline writes a file directly as JSON.
func (w *JSONWriter) writeInline(req WriteRequest) error {
	encoded, err := encodeJSON(FileData{
		Path:     req.Path,
		Content:  req.Content,
		Language: detectLanguage(req.Path),
	})
	if err != nil {
		return shared.WrapError(
			err,
			shared.ErrorTypeProcessing,
			shared.CodeProcessingEncode,
			"failed to marshal JSON",
		).WithFilePath(req.Path)
	}

	if _, err := w.out.Write(encoded); err != nil {
		return shared.WrapError(
			err,
			shared.ErrorTypeIO,
			shared.CodeIOWrite,
			"failed to write JSON file",
		).WithFilePath(req.Path)
	}

	return nil
}

// encodeJSON encodes v without HTML escaping. json.Marshal rewrites every <, >
// and & as a six-byte \u escape, which only matters for JSON embedded in HTML;
// this output is read by a model, where the escapes cost tokens and readability.
func encodeJSON(v any) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		return nil, err
	}

	return bytes.TrimSuffix(buf.Bytes(), []byte("\n")), nil
}
