<<<<<<<< HEAD:gibidiutils/writers.go
// Package gibidiutils provides common utility functions for gibidify.
package gibidiutils
|||||||| parent of e21c976 (refactor: rename utils to shared and deduplicate code):utils/writers.go
package utils
========
// Package shared provides common utility functions.
package shared
>>>>>>>> e21c976 (refactor: rename utils to shared and deduplicate code):shared/writers.go

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
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
<<<<<<<< HEAD:gibidiutils/writers.go
//
//revive:disable-next-line:cognitive-complexity
func StreamContent(
	reader io.Reader,
	writer io.Writer,
	chunkSize int,
	filePath string,
	processChunk func([]byte) []byte,
) error {
|||||||| parent of e21c976 (refactor: rename utils to shared and deduplicate code):utils/writers.go
func StreamContent(reader io.Reader, writer io.Writer, chunkSize int, filePath string, processChunk func([]byte) []byte) error {
========
func StreamContent(
	reader io.Reader,
	writer io.Writer,
	chunkSize int,
	filePath string,
	processChunk func([]byte) []byte,
) error {
>>>>>>>> e21c976 (refactor: rename utils to shared and deduplicate code):shared/writers.go
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
<<<<<<<< HEAD:gibidiutils/writers.go
			wrappedErr := WrapError(err, ErrorTypeIO, CodeIOFileRead, "failed to read content chunk")
			if filePath != "" {
				wrappedErr = wrappedErr.WithFilePath(filePath)
			}
			return wrappedErr
|||||||| parent of e21c976 (refactor: rename utils to shared and deduplicate code):utils/writers.go
			wrappedErr := WrapError(err, ErrorTypeIO, CodeIORead, "failed to read content chunk")
			if filePath != "" {
				wrappedErr = wrappedErr.WithFilePath(filePath)
			}
			return wrappedErr
========
			return wrapReadError(err, filePath)
>>>>>>>> e21c976 (refactor: rename utils to shared and deduplicate code):shared/writers.go
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

<<<<<<<< HEAD:gibidiutils/writers.go
// SafeUint64ToInt64WithDefault safely converts uint64 to int64, returning a default value if overflow would occur.
// When defaultValue is 0 (the safe default), clamps to MaxInt64 on overflow to keep guardrails active.
// This prevents overflow from making monitors think memory usage is zero when it's actually maxed out.
func SafeUint64ToInt64WithDefault(value uint64, defaultValue int64) int64 {
	if value > math.MaxInt64 {
		// When caller uses 0 as "safe" default, clamp to max so overflow still trips guardrails
		if defaultValue == 0 {
			return math.MaxInt64
		}
		return defaultValue
	}
	return int64(value) //#nosec G115 -- Safe: value <= MaxInt64 checked above
}

|||||||| parent of e21c976 (refactor: rename utils to shared and deduplicate code):utils/writers.go
========
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

>>>>>>>> e21c976 (refactor: rename utils to shared and deduplicate code):shared/writers.go
// StreamLines provides line-based streaming for YAML content.
// This provides an alternative streaming approach for YAML writers.
func StreamLines(reader io.Reader, writer io.Writer, filePath string, lineProcessor func(string) string) error {
	// Read all content first (for small files this is fine)
	content, err := io.ReadAll(reader)
	if err != nil {
		wrappedErr := WrapError(err, ErrorTypeIO, CodeIOFileRead, "failed to read content for line processing")
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
