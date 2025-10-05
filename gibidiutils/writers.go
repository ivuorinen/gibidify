// Package gibidiutils provides common utility functions for gibidify.
package gibidiutils

import (
	"encoding/json"
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
func StreamContent(reader io.Reader, writer io.Writer, chunkSize int, filePath string, processChunk func([]byte) []byte) error {
	buf := make([]byte, chunkSize)
	for {
		n, err := reader.Read(buf)
		if n > 0 {
			processed := buf[:n]
			if processChunk != nil {
				processed = processChunk(processed)
			}
			if _, writeErr := writer.Write(processed); writeErr != nil {
				wrappedErr := WrapError(writeErr, ErrorTypeIO, CodeIOWrite, "failed to write content chunk")
				if filePath != "" {
					wrappedErr = wrappedErr.WithFilePath(filePath)
				}
				return wrappedErr
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			wrappedErr := WrapError(err, ErrorTypeIO, CodeIORead, "failed to read content chunk")
			if filePath != "" {
				wrappedErr = wrappedErr.WithFilePath(filePath)
			}
			return wrappedErr
		}
	}
	return nil
}

// EscapeForJSON escapes content for JSON output using the standard library.
// This replaces the custom escapeJSONString function with a more robust implementation.
func EscapeForJSON(content string) string {
	// Use the standard library's JSON marshaling for proper escaping
	jsonBytes, _ := json.Marshal(content)
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
		content == "true" || content == "false" ||
		content == "null" || content == "~"

	if needsQuotes {
		// Use double quotes and escape internal quotes
		escaped := strings.ReplaceAll(content, "\\", "\\\\")
		escaped = strings.ReplaceAll(escaped, "\"", "\\\"")
		return "\"" + escaped + "\""
	}
	return content
}

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
	return int64(value)
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
