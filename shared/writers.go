// Package shared provides common utility functions.
package shared

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

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
