// Package fileproc provides a writer for the output of the file processor.
package fileproc

import (
	"fmt"
	"os"

	"github.com/ivuorinen/gibidify/utils"
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
		context := map[string]interface{}{
			"format": format,
		}
		err := utils.NewStructuredError(
			utils.ErrorTypeValidation,
			utils.CodeValidationFormat,
			fmt.Sprintf("unsupported format: %s", format),
			"",
			context,
		)
		utils.LogError("Failed to encode output", err)
		close(done)
	}
}
