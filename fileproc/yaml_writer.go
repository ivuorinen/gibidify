// Package fileproc handles file processing, collection, and output formatting.
package fileproc

import (
	"io"

	"gopkg.in/yaml.v3"

	"github.com/ivuorinen/gibidify/shared"
)

// yamlIndent is the indentation of nested YAML nodes.
const yamlIndent = 2

// YAMLWriter handles YAML format output.
//
// Every scalar goes through the yaml.v3 encoder. The previous hand-written
// block scalars copied raw bytes, so one file with an ESC, a form feed or a
// bare CR made the whole document unparseable, and CRLF content lost its CRs;
// the encoder picks literal style for printable text and an escaped
// double-quoted scalar otherwise, so every string round-trips.
type YAMLWriter struct {
	out io.Writer
}

// yamlHeader is the document's leading mapping. Fields keep declaration order.
type yamlHeader struct {
	Prefix string `yaml:"prefix"`
	Suffix string `yaml:"suffix"`
}

// NewYAMLWriter creates a new YAML writer.
func NewYAMLWriter(out io.Writer) *YAMLWriter {
	return &YAMLWriter{out: out}
}

// Start writes the prefix and suffix keys and opens the files sequence.
func (w *YAMLWriter) Start(prefix, suffix string) error {
	if err := w.encode(yamlHeader{Prefix: prefix, Suffix: suffix}); err != nil {
		return shared.WrapError(err, shared.ErrorTypeIO, shared.CodeIOWrite, "failed to write YAML header")
	}
	if _, err := io.WriteString(w.out, "files:\n"); err != nil {
		return shared.WrapError(err, shared.ErrorTypeIO, shared.CodeIOWrite, "failed to write YAML header")
	}

	return nil
}

// WriteFile writes one entry of the files sequence. A one-element sequence
// encodes as "- path: …" at column zero, which YAML accepts as the value of the
// preceding "files:" key.
func (w *YAMLWriter) WriteFile(req WriteRequest) error {
	entry := []FileData{{Path: req.Path, Content: req.Content, Language: detectLanguage(req.Path)}}
	if err := w.encode(entry); err != nil {
		return shared.WrapError(
			err,
			shared.ErrorTypeIO,
			shared.CodeIOWrite,
			"failed to write YAML entry",
		).WithFilePath(req.Path)
	}

	return nil
}

// Close writes the YAML footer (no footer needed for YAML).
func (w *YAMLWriter) Close() error {
	return nil
}

// encode writes v as one YAML document fragment. A fresh encoder per value
// emits no "---" separator.
func (w *YAMLWriter) encode(v any) error {
	enc := yaml.NewEncoder(w.out)
	enc.SetIndent(yamlIndent)
	if err := enc.Encode(v); err != nil {
		return err
	}

	return enc.Close()
}
