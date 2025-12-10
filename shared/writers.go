// Package shared provides common utility functions.
package shared

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// SafeCloseReader safely closes a reader if it implements io.Closer.
// This eliminates the duplicated closeReader methods across all writers.
func SafeCloseReader(reader io.Reader, path string) {
	if closer, ok := reader.(io.Closer); ok {
		if err := closer.Close(); err != nil {
			LogError(
				"Failed to close file reader",
				WrapError(err, ErrorTypeIO, CodeIOClose, "failed to close file reader").WithFilePath(path),
			)
		}
	}
}

// WriteWithErrorWrap performs file writing with consistent error handling.
// This centralizes the common pattern of writing strings with error wrapping.
func WriteWithErrorWrap(writer io.Writer, content, errorMsg, filePath string) error {
	if _, err := writer.Write([]byte(content)); err != nil {
		wrappedErr := WrapError(err, ErrorTypeIO, CodeIOWrite, errorMsg)
		if filePath != "" {
			wrappedErr = wrappedErr.WithFilePath(filePath)
		}

		return wrappedErr
	}

	return nil
}

// StreamContent provides a common streaming implementation with chunk processing.
// This eliminates the similar streaming patterns across JSON and Markdown writers.
func StreamContent(
	reader io.Reader,
	writer io.Writer,
	chunkSize int,
	filePath string,
	processChunk func([]byte) []byte,
) error {
	buf := make([]byte, chunkSize)
	for {
		n, err := reader.Read(buf)
		if n > 0 {
			if err := writeProcessedChunk(writer, buf[:n], filePath, processChunk); err != nil {
				return err
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return wrapReadError(err, filePath)
		}
	}

	return nil
}

// writeProcessedChunk processes and writes a chunk of data.
func writeProcessedChunk(writer io.Writer, chunk []byte, filePath string, processChunk func([]byte) []byte) error {
	processed := chunk
	if processChunk != nil {
		processed = processChunk(processed)
	}
	if _, writeErr := writer.Write(processed); writeErr != nil {
		return wrapWriteError(writeErr, filePath)
	}

	return nil
}

// wrapWriteError wraps a write error with context.
func wrapWriteError(err error, filePath string) error {
	wrappedErr := WrapError(err, ErrorTypeIO, CodeIOWrite, "failed to write content chunk")
	if filePath != "" {
		//nolint:errcheck // WithFilePath error doesn't affect wrapped error integrity
		wrappedErr = wrappedErr.WithFilePath(filePath)
	}

	return wrappedErr
}

// wrapReadError wraps a read error with context.
func wrapReadError(err error, filePath string) error {
	wrappedErr := WrapError(err, ErrorTypeIO, CodeIORead, "failed to read content chunk")
	if filePath != "" {
		wrappedErr = wrappedErr.WithFilePath(filePath)
	}

	return wrappedErr
}

// EscapeForJSON escapes content for JSON output using the standard library.
// This replaces the custom escapeJSONString function with a more robust implementation.
func EscapeForJSON(content string) string {
	// Use the standard library's JSON marshaling for proper escaping
	jsonBytes, err := json.Marshal(content)
	if err != nil {
		// If marshaling fails (which is very unlikely for a string), return the original
		return content
	}
	// Remove the surrounding quotes that json.Marshal adds
	jsonStr := string(jsonBytes)
	if len(jsonStr) >= 2 && jsonStr[0] == '"' && jsonStr[len(jsonStr)-1] == '"' {
		return jsonStr[1 : len(jsonStr)-1]
	}

	return jsonStr
}

// EscapeForYAML quotes/escapes content for YAML output if needed.
// This centralizes the YAML string quoting logic.
func EscapeForYAML(content string) string {
	// Quote if contains special characters, spaces, or starts with special chars
	needsQuotes := strings.ContainsAny(content, " \t\n\r:{}[]|>-'\"\\") ||
		strings.HasPrefix(content, "-") ||
		strings.HasPrefix(content, "?") ||
		strings.HasPrefix(content, ":") ||
		content == "" ||
		content == LiteralTrue || content == LiteralFalse ||
		content == LiteralNull || content == "~"

	if needsQuotes {
		// Use double quotes and escape internal quotes
		escaped := strings.ReplaceAll(content, "\\", "\\\\")
		escaped = strings.ReplaceAll(escaped, "\"", "\\\"")

		return "\"" + escaped + "\""
	}

	return content
}

// CheckContextCancellation is a helper function that checks if context is canceled and returns appropriate error.
func CheckContextCancellation(ctx context.Context, operation string) error {
	select {
	case <-ctx.Done():
		return fmt.Errorf("%s canceled: %w", operation, ctx.Err())
	default:
		return nil
	}
}

// WithContextCheck wraps an operation with context cancellation checking.
func WithContextCheck(ctx context.Context, operation string, fn func() error) error {
	if err := CheckContextCancellation(ctx, operation); err != nil {
		return err
	}

	return fn()
}

// StreamLines provides line-based streaming for YAML content.
// This provides an alternative streaming approach for YAML writers.
func StreamLines(reader io.Reader, writer io.Writer, filePath string, lineProcessor func(string) string) error {
	// Read all content first (for small files this is fine)
	content, err := io.ReadAll(reader)
	if err != nil {
		wrappedErr := WrapError(err, ErrorTypeIO, CodeIORead, "failed to read content for line processing")
		if filePath != "" {
			wrappedErr = wrappedErr.WithFilePath(filePath)
		}

		return wrappedErr
	}

	// Split into lines and process each
	lines := strings.Split(string(content), "\n")
	for i, line := range lines {
		processedLine := line
		if lineProcessor != nil {
			processedLine = lineProcessor(line)
		}

		// Write line with proper line ending (except for last empty line)
		lineToWrite := processedLine
		if i < len(lines)-1 || line != "" {
			lineToWrite += "\n"
		}

		if _, writeErr := writer.Write([]byte(lineToWrite)); writeErr != nil {
			wrappedErr := WrapError(writeErr, ErrorTypeIO, CodeIOWrite, "failed to write processed line")
			if filePath != "" {
				wrappedErr = wrappedErr.WithFilePath(filePath)
			}

			return wrappedErr
		}
	}

	return nil
}
