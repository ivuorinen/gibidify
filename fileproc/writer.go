// Package fileproc provides a writer for the output of the file processor.
package fileproc

import (
	"os"

	"github.com/ivuorinen/gibidify/shared"
)

// StartWriter writes the output in the specified format with memory optimization.
func StartWriter(outFile *os.File, writeCh <-chan WriteRequest, done chan<- struct{}, format, prefix, suffix string) {
	switch format {
	case "markdown":
		startMarkdownWriter(outFile, writeCh, done, prefix, suffix)
	case "json":
		startJSONWriter(outFile, writeCh, done, prefix, suffix)
	case "yaml":
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
