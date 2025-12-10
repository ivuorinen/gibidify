package shared

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

func TestGetLogger(t *testing.T) {
	// Test singleton behavior
	logger1 := GetLogger()
	logger2 := GetLogger()

	if logger1 != logger2 {
		t.Error("GetLogger should return the same instance (singleton)")
	}
}

func TestLogServiceLevels(t *testing.T) {
	tests := []struct {
		name     string
		level    LogLevel
		logFunc  func(Logger)
		expected bool
	}{
		{
			name:  "debug level allows debug messages",
			level: LogLevelDebug,
			logFunc: func(l Logger) {
				l.Debug(TestLoggerDebugMsg)
			},
			expected: true,
		},
		{
			name:  "info level blocks debug messages",
			level: LogLevelInfo,
			logFunc: func(l Logger) {
				l.Debug(TestLoggerDebugMsg)
			},
			expected: false,
		},
		{
			name:  "info level allows info messages",
			level: LogLevelInfo,
			logFunc: func(l Logger) {
				l.Info(TestLoggerInfoMsg)
			},
			expected: true,
		},
		{
			name:  "warn level blocks info messages",
			level: LogLevelWarn,
			logFunc: func(l Logger) {
				l.Info(TestLoggerInfoMsg)
			},
			expected: false,
		},
		{
			name:  "warn level allows warn messages",
			level: LogLevelWarn,
			logFunc: func(l Logger) {
				l.Warn(TestLoggerWarnMsg)
			},
			expected: true,
		},
		{
			name:  "error level blocks warn messages",
			level: LogLevelError,
			logFunc: func(l Logger) {
				l.Warn(TestLoggerWarnMsg)
			},
			expected: false,
		},
		{
			name:  "error level allows error messages",
			level: LogLevelError,
			logFunc: func(l Logger) {
				l.Error("error message")
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				var buf bytes.Buffer
				logger := GetLogger()
				logger.SetOutput(&buf)
				logger.SetLevel(tt.level)

				tt.logFunc(logger)

				output := buf.String()
				hasOutput := len(strings.TrimSpace(output)) > 0
				if hasOutput != tt.expected {
					t.Errorf("Expected output: %v, got output: %v, output: %s", tt.expected, hasOutput, output)
				}
			},
		)
	}
}

func TestLogServiceFormattedLogging(t *testing.T) {
	tests := []struct {
		name     string
		level    LogLevel
		logFunc  func(Logger)
		contains string
	}{
		{
			name:  "debugf formats correctly",
			level: LogLevelDebug,
			logFunc: func(l Logger) {
				l.Debugf("debug %s %d", "message", 42)
			},
			contains: "debug message 42",
		},
		{
			name:  "infof formats correctly",
			level: LogLevelInfo,
			logFunc: func(l Logger) {
				l.Infof("info %s %d", "message", 42)
			},
			contains: "info message 42",
		},
		{
			name:  "warnf formats correctly",
			level: LogLevelWarn,
			logFunc: func(l Logger) {
				l.Warnf("warn %s %d", "message", 42)
			},
			contains: "warn message 42",
		},
		{
			name:  "errorf formats correctly",
			level: LogLevelError,
			logFunc: func(l Logger) {
				l.Errorf("error %s %d", "message", 42)
			},
			contains: "error message 42",
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				var buf bytes.Buffer
				logger := GetLogger()
				logger.SetOutput(&buf)
				logger.SetLevel(tt.level)

				tt.logFunc(logger)

				output := buf.String()
				if !strings.Contains(output, tt.contains) {
					t.Errorf("Expected output to contain %q, got: %s", tt.contains, output)
				}
			},
		)
	}
}

func TestLogServiceWithFields(t *testing.T) {
	var buf bytes.Buffer
	logger := GetLogger()
	logger.SetOutput(&buf)
	logger.SetLevel(LogLevelInfo)

	fields := map[string]any{
		"component": "test",
		"count":     42,
		"enabled":   true,
	}

	fieldLogger := logger.WithFields(fields)
	fieldLogger.Info("test message")

	output := buf.String()
	expectedFields := []string{"component=test", "count=42", "enabled=true", "test message"}
	for _, expected := range expectedFields {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected output to contain %q, got: %s", expected, output)
		}
	}
}

func TestLogServiceSetOutput(t *testing.T) {
	var buf1, buf2 bytes.Buffer
	logger := GetLogger()

	// Set initial output
	logger.SetOutput(&buf1)
	logger.SetLevel(LogLevelInfo)
	logger.Info("message1")

	// Change output
	logger.SetOutput(&buf2)
	logger.Info("message2")

	// Verify messages went to correct outputs
	if !strings.Contains(buf1.String(), "message1") {
		t.Error("First message should be in first buffer")
	}
	if strings.Contains(buf1.String(), "message2") {
		t.Error("Second message should not be in first buffer")
	}
	if !strings.Contains(buf2.String(), "message2") {
		t.Error("Second message should be in second buffer")
	}
}

func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected LogLevel
	}{
		{"debug", LogLevelDebug},
		{"info", LogLevelInfo},
		{"warn", LogLevelWarn},
		{"warning", LogLevelWarn},
		{"error", LogLevelError},
		{"invalid", LogLevelWarn}, // default
		{"", LogLevelWarn},        // default
	}

	for _, tt := range tests {
		t.Run(
			tt.input, func(t *testing.T) {
				result := ParseLogLevel(tt.input)
				if result != tt.expected {
					t.Errorf("ParseLogLevel(%q) = %v, want %v", tt.input, result, tt.expected)
				}
			},
		)
	}
}

func TestValidateLogLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"debug", true},
		{"info", true},
		{"warn", true},
		{"warning", true},
		{"error", true},
		{"invalid", false},
		{"", false},
		{"DEBUG", false}, // case-sensitive
		{"INFO", false},  // case-sensitive
	}

	for _, tt := range tests {
		t.Run(
			tt.input, func(t *testing.T) {
				result := ValidateLogLevel(tt.input)
				if result != tt.expected {
					t.Errorf("ValidateLogLevel(%q) = %v, want %v", tt.input, result, tt.expected)
				}
			},
		)
	}
}

func TestLogServiceDefaultLevel(t *testing.T) {
	var buf bytes.Buffer
	logger := GetLogger()
	logger.SetOutput(&buf)
	logger.SetLevel(LogLevelWarn) // Ensure we're at WARN level for this test

	// Test that default level is WARN (should block info messages)
	logger.Info(TestLoggerInfoMsg)
	if strings.TrimSpace(buf.String()) != "" {
		t.Error("Info message should be blocked at default WARN level")
	}

	// Test that warning messages are allowed
	buf.Reset()
	logger.Warn(TestLoggerWarnMsg)
	if !strings.Contains(buf.String(), TestLoggerWarnMsg) {
		t.Error("Warn message should be allowed at default WARN level")
	}
}

func TestLogServiceSetLevel(t *testing.T) {
	tests := []struct {
		name     string
		setLevel LogLevel
		logFunc  func(Logger)
		expected bool
	}{
		{
			name:     "set debug level allows debug",
			setLevel: LogLevelDebug,
			logFunc: func(l Logger) {
				l.Debug(TestLoggerDebugMsg)
			},
			expected: true,
		},
		{
			name:     "set info level blocks debug",
			setLevel: LogLevelInfo,
			logFunc: func(l Logger) {
				l.Debug(TestLoggerDebugMsg)
			},
			expected: false,
		},
		{
			name:     "set warn level blocks info",
			setLevel: LogLevelWarn,
			logFunc: func(l Logger) {
				l.Info(TestLoggerInfoMsg)
			},
			expected: false,
		},
		{
			name:     "set error level blocks warn",
			setLevel: LogLevelError,
			logFunc: func(l Logger) {
				l.Warn(TestLoggerWarnMsg)
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				var buf bytes.Buffer
				logger := GetLogger()
				logger.SetOutput(&buf)
				logger.SetLevel(tt.setLevel)

				tt.logFunc(logger)

				output := buf.String()
				hasOutput := len(strings.TrimSpace(output)) > 0
				if hasOutput != tt.expected {
					t.Errorf("Expected output: %v, got output: %v, level: %v", tt.expected, hasOutput, tt.setLevel)
				}
			},
		)
	}
}

// Benchmark tests.
func BenchmarkLoggerInfo(b *testing.B) {
	logger := GetLogger()
	logger.SetOutput(io.Discard)
	logger.SetLevel(LogLevelInfo)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("benchmark message")
	}
}

func BenchmarkLoggerWithFields(b *testing.B) {
	logger := GetLogger()
	logger.SetOutput(io.Discard)
	logger.SetLevel(LogLevelInfo)

	fields := map[string]any{
		"component": "benchmark",
		"iteration": 0,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fields["iteration"] = i
		logger.WithFields(fields).Info("benchmark message")
	}
}
