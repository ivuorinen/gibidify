// Package fileproc provides a writer for the output of the file processor.
package fileproc

import (
	"fmt"
	"os"

	"github.com/ivuorinen/gibidify/gibidiutils"
)

// WriterConfig holds configuration for the writer.
type WriterConfig struct {
	Format string
	Prefix string
	Suffix string
}

// Validate checks if the WriterConfig is valid.
func (c WriterConfig) Validate() error {
	if c.Format == "" {
		return gibidiutils.NewStructuredError(
			gibidiutils.ErrorTypeValidation,
			gibidiutils.CodeValidationFormat,
			"format cannot be empty",
			"",
			nil,
		)
	}

	switch c.Format {
	case "markdown", "json", "yaml":
		return nil
	default:
		context := map[string]any{
			"format": c.Format,
		}
		return gibidiutils.NewStructuredError(
			gibidiutils.ErrorTypeValidation,
			gibidiutils.CodeValidationFormat,
			fmt.Sprintf("unsupported format: %s", c.Format),
			"",
			context,
		)
	}
}

// StartWriter writes the output in the specified format with memory optimization.
func StartWriter(outFile *os.File, writeCh <-chan WriteRequest, done chan<- struct{}, config WriterConfig) {
	// Validate config
	if err := config.Validate(); err != nil {
		gibidiutils.LogError("Invalid writer configuration", err)
		close(done)
		return
	}

	// Validate outFile is not nil
	if outFile == nil {
		err := gibidiutils.NewStructuredError(
			gibidiutils.ErrorTypeIO,
			gibidiutils.CodeIOFileWrite,
			"output file is nil",
			"",
			nil,
		)
		gibidiutils.LogError("Failed to write output", err)
		close(done)
		return
	}

	// Validate outFile is accessible
	if _, err := outFile.Stat(); err != nil {
		structErr := gibidiutils.WrapError(
			err,
			gibidiutils.ErrorTypeIO,
			gibidiutils.CodeIOFileWrite,
			"failed to stat output file",
		)
		gibidiutils.LogError("Failed to validate output file", structErr)
		close(done)
		return
	}

	switch config.Format {
	case "markdown":
		startMarkdownWriter(outFile, writeCh, done, config.Prefix, config.Suffix)
	case "json":
		startJSONWriter(outFile, writeCh, done, config.Prefix, config.Suffix)
	case "yaml":
		startYAMLWriter(outFile, writeCh, done, config.Prefix, config.Suffix)
	default:
		context := map[string]interface{}{
			"format": config.Format,
		}
		err := gibidiutils.NewStructuredError(
			gibidiutils.ErrorTypeValidation,
			gibidiutils.CodeValidationFormat,
			fmt.Sprintf("unsupported format: %s", config.Format),
			"",
			context,
		)
		gibidiutils.LogError("Failed to encode output", err)
		close(done)
	}
}
