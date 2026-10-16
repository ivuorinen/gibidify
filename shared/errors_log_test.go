package shared

import (
	"fmt"
	"strings"
	"testing"
)

// TestLogErrorStructuredFields verifies a StructuredError, also when wrapped,
// is logged with its type, code and file path as fields.
func TestLogErrorStructuredFields(t *testing.T) {
	structErr := NewStructuredError(
		ErrorTypeFileSystem, CodeFSAccess, TestErrAccessDenied, TestPathTestFileTXT, map[string]any{"attempt": 2},
	)

	tests := []struct {
		name string
		err  error
	}{
		{name: "structured error", err: structErr},
		{name: "wrapped structured error", err: fmt.Errorf("outer: %w", structErr)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := captureLogOutput(func() {
				LogError("reading %s", tt.err, "input")
			})

			for _, want := range []string{
				"level=error",
				"reading input",
				TestErrAccessDenied,
				"error_type=" + ErrorTypeFileSystem.String(),
				"error_code=" + CodeFSAccess,
				"file_path=" + TestPathTestFileTXT,
				"attempt:2",
			} {
				if !strings.Contains(output, want) {
					t.Errorf("LogError() output = %q, want to contain %q", output, want)
				}
			}
		})
	}
}
