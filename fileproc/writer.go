// Package fileproc provides a writer for the output of the file processor.
package fileproc

import (
	"bufio"
	"io"
	"os"

	"github.com/ivuorinen/gibidify/shared"
)

// outputBufferSize is the write buffer between the format writers and the
// output file. Unbuffered, the YAML writer issued one write(2) per content line
// (900k syscalls for a 22 MB tree).
const outputBufferSize = 64 << 10

// runFormatWriter drives a FormatWriter over every request on writeCh and
// returns the first error from Start, WriteFile, Close or the final flush.
//
// After a failure it keeps draining writeCh without writing: producers block on
// sends to writeCh, so returning early would deadlock every worker (a full disk
// at Start hung the process indefinitely).
func runFormatWriter(
	outFile *os.File,
	writeCh <-chan WriteRequest,
	prefix, suffix string,
	writerFactory func(io.Writer) FormatWriter,
) error {
	buffered := bufio.NewWriterSize(outFile, outputBufferSize)
	writer := writerFactory(buffered)
	err := writer.Start(prefix, suffix)

	for req := range writeCh {
		if err == nil {
			err = writer.WriteFile(req)
		}
	}

	if err == nil {
		err = writer.Close()
	}
	if err == nil {
		if flushErr := buffered.Flush(); flushErr != nil {
			err = shared.WrapError(flushErr, shared.ErrorTypeIO, shared.CodeIOWrite, "failed to flush output")
		}
	}

	return err
}

// StartWriter writes every request from writeCh to outFile in the given format
// and returns once writeCh is closed. It returns the first write error; writeCh
// is always drained, even after an error or for an unsupported format.
func StartWriter(outFile *os.File, writeCh <-chan WriteRequest, format, prefix, suffix string) error {
	switch format {
	case shared.FormatMarkdown:
		return runFormatWriter(outFile, writeCh, prefix, suffix, func(w io.Writer) FormatWriter {
			return NewMarkdownWriter(w)
		})
	case shared.FormatJSON:
		return runFormatWriter(outFile, writeCh, prefix, suffix, func(w io.Writer) FormatWriter {
			return NewJSONWriter(w)
		})
	case shared.FormatYAML:
		return runFormatWriter(outFile, writeCh, prefix, suffix, func(w io.Writer) FormatWriter {
			return NewYAMLWriter(w)
		})
	default:
		// Through runFormatWriter so the channel is drained the same way.
		return runFormatWriter(outFile, writeCh, prefix, suffix, func(io.Writer) FormatWriter {
			return unsupportedFormatWriter{format: format}
		})
	}
}

// unsupportedFormatWriter fails at Start, writing nothing.
type unsupportedFormatWriter struct {
	format string
}

// Start reports the unsupported format.
func (w unsupportedFormatWriter) Start(_, _ string) error {
	return shared.NewStructuredError(
		shared.ErrorTypeValidation,
		shared.CodeValidationFormat,
		"unsupported format: "+w.format,
		"",
		map[string]any{"format": w.format},
	)
}

// WriteFile is never called: runFormatWriter stops writing after Start fails.
func (unsupportedFormatWriter) WriteFile(WriteRequest) error { return nil }

// Close is never called: runFormatWriter skips Close after Start fails.
func (unsupportedFormatWriter) Close() error { return nil }
