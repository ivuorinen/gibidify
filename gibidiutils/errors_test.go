// Package gibidiutils provides common utility functions for gibidify.
package gibidiutils

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
)

// captureLogOutput captures logrus output for testing
func captureLogOutput(f func()) string {
	var buf bytes.Buffer
	logrus.SetOutput(&buf)
	defer logrus.SetOutput(logrus.StandardLogger().Out)
	f()
	return buf.String()
}

func TestLogError(t *testing.T) {
	tests := []struct {
		name      string
		operation string
		err       error
		args      []any
		wantLog   string
		wantEmpty bool
	}{
		{
			name:      "nil error should not log",
			operation: "test operation",
			err:       nil,
			args:      nil,
			wantEmpty: true,
		},
		{
			name:      "basic error logging",
			operation: "failed to read file",
			err:       errors.New("permission denied"),
			args:      nil,
			wantLog:   "failed to read file: permission denied",
		},
		{
			name:      "error with formatting args",
			operation: "failed to process file %s",
			err:       errors.New("file too large"),
			args:      []any{"test.txt"},
			wantLog:   "failed to process file test.txt: file too large",
		},
		{
			name:      "error with multiple formatting args",
			operation: "failed to copy from %s to %s",
			err:       errors.New("disk full"),
			args:      []any{"source.txt", "dest.txt"},
			wantLog:   "failed to copy from source.txt to dest.txt: disk full",
		},
		{
			name:      "wrapped error",
			operation: "database operation failed",
			err:       fmt.Errorf("connection error: %w", errors.New("timeout")),
			args:      nil,
			wantLog:   "database operation failed: connection error: timeout",
		},
		{
			name:      "empty operation string",
			operation: "",
			err:       errors.New("some error"),
			args:      nil,
			wantLog:   ": some error",
		},
		{
			name:      "operation with percentage sign",
			operation: "processing 50% complete",
			err:       errors.New("interrupted"),
			args:      nil,
			wantLog:   "processing 50% complete: interrupted",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := captureLogOutput(func() {
				LogError(tt.operation, tt.err, tt.args...)
			})

			if tt.wantEmpty {
				if output != "" {
					t.Errorf("LogError() logged output when error was nil: %q", output)
				}
				return
			}

			if !strings.Contains(output, tt.wantLog) {
				t.Errorf("LogError() output = %q, want to contain %q", output, tt.wantLog)
			}

			// Verify it's logged at ERROR level
			if !strings.Contains(output, "level=error") {
				t.Errorf("LogError() should log at ERROR level, got: %q", output)
			}
		})
	}
}

func TestLogErrorf(t *testing.T) {
	tests := []struct {
		name      string
		err       error
		format    string
		args      []any
		wantLog   string
		wantEmpty bool
	}{
		{
			name:      "nil error should not log",
			err:       nil,
			format:    "operation %s failed",
			args:      []any{"test"},
			wantEmpty: true,
		},
		{
			name:    "basic formatted error",
			err:     errors.New("not found"),
			format:  "file %s not found",
			args:    []any{"config.yaml"},
			wantLog: "file config.yaml not found: not found",
		},
		{
			name:    "multiple format arguments",
			err:     errors.New("invalid range"),
			format:  "value %d is not between %d and %d",
			args:    []any{150, 0, 100},
			wantLog: "value 150 is not between 0 and 100: invalid range",
		},
		{
			name:    "no format arguments",
			err:     errors.New("generic error"),
			format:  "operation failed",
			args:    nil,
			wantLog: "operation failed: generic error",
		},
		{
			name:    "format with different types",
			err:     errors.New("type mismatch"),
			format:  "expected %s but got %d",
			args:    []any{"string", 42},
			wantLog: "expected string but got 42: type mismatch",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := captureLogOutput(func() {
				LogErrorf(tt.err, tt.format, tt.args...)
			})

			if tt.wantEmpty {
				if output != "" {
					t.Errorf("LogErrorf() logged output when error was nil: %q", output)
				}
				return
			}

			if !strings.Contains(output, tt.wantLog) {
				t.Errorf("LogErrorf() output = %q, want to contain %q", output, tt.wantLog)
			}

			// Verify it's logged at ERROR level
			if !strings.Contains(output, "level=error") {
				t.Errorf("LogErrorf() should log at ERROR level, got: %q", output)
			}
		})
	}
}

func TestLogErrorConcurrency(_ *testing.T) {
	// Test that LogError is safe for concurrent use
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(n int) {
			LogError("concurrent operation", fmt.Errorf("error %d", n))
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestLogErrorfConcurrency(_ *testing.T) {
	// Test that LogErrorf is safe for concurrent use
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(n int) {
			LogErrorf(fmt.Errorf("error %d", n), "concurrent operation %d", n)
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}

// BenchmarkLogError benchmarks the LogError function
func BenchmarkLogError(b *testing.B) {
	err := errors.New("benchmark error")
	// Disable output during benchmark
	logrus.SetOutput(bytes.NewBuffer(nil))
	defer logrus.SetOutput(logrus.StandardLogger().Out)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		LogError("benchmark operation", err)
	}
}

// BenchmarkLogErrorf benchmarks the LogErrorf function
func BenchmarkLogErrorf(b *testing.B) {
	err := errors.New("benchmark error")
	// Disable output during benchmark
	logrus.SetOutput(bytes.NewBuffer(nil))
	defer logrus.SetOutput(logrus.StandardLogger().Out)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		LogErrorf(err, "benchmark operation %d", i)
	}
}

// BenchmarkLogErrorNil benchmarks LogError with nil error (no-op case)
func BenchmarkLogErrorNil(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		LogError("benchmark operation", nil)
	}
}
