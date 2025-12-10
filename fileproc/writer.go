// Package fileproc provides a writer for the output of the file processor.
package fileproc

import (
	"os"

	"github.com/ivuorinen/gibidify/shared"
)

// startFormatWriter handles generic writer orchestration for any format.
// This eliminates code duplication across format-specific writer functions.
// Uses the FormatWriter interface defined in formats.go.
func startFormatWriter(
	outFile *os.File,
	writeCh <-chan WriteRequest,
	done chan<- struct{},
	prefix, suffix string,
	writerFactory func(*os.File) FormatWriter,
) {
	defer close(done)

	writer := writerFactory(outFile)

	// Start writing
	if err := writer.Start(prefix, suffix); err != nil {
		shared.LogError("Failed to start writer", err)

		return
	}

	// Process files
	for req := range writeCh {
		if err := writer.WriteFile(req); err != nil {
			shared.LogError("Failed to write file", err)
		}
	}

	// Close writer
	if err := writer.Close(); err != nil {
		shared.LogError("Failed to close writer", err)
	}
}

// StartWriter writes the output in the specified format with memory optimization.
func StartWriter(outFile *os.File, writeCh <-chan WriteRequest, done chan<- struct{}, format, prefix, suffix string) {
	switch format {
	case shared.FormatMarkdown:
		startMarkdownWriter(outFile, writeCh, done, prefix, suffix)
	case shared.FormatJSON:
		startJSONWriter(outFile, writeCh, done, prefix, suffix)
	case shared.FormatYAML:
		startYAMLWriter(outFile, writeCh, done, prefix, suffix)
	default:
		context := map[string]any{
			"format": format,
		}
		err := shared.NewStructuredError(
			shared.ErrorTypeValidation,
			shared.CodeValidationFormat,
			"unsupported format: "+format,
			"",
			context,
		)
		shared.LogError("Failed to encode output", err)
		close(done)
	}
}
