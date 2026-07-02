// Package shared provides logging utilities for gibidify.
package shared

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"sync"
)

// Logger interface defines the logging contract for gibidify.
type Logger interface {
	Debug(args ...any)
	Debugf(format string, args ...any)
	Info(args ...any)
	Infof(format string, args ...any)
	Warn(args ...any)
	Warnf(format string, args ...any)
	Error(args ...any)
	Errorf(format string, args ...any)
	WithFields(fields map[string]any) Logger
	SetLevel(level LogLevel)
	SetOutput(output io.Writer)
}

// LogLevel represents available log levels.
type LogLevel string

// swappableWriter is an io.Writer whose destination can be changed at runtime.
// The default slog handler binds to its writer at construction, so all derived
// loggers share this indirection to make SetOutput affect them uniformly.
type swappableWriter struct {
	mu sync.Mutex
	w  io.Writer
}

func (s *swappableWriter) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.w.Write(p)
}

func (s *swappableWriter) set(w io.Writer) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.w = w
}

// logService implements the Logger interface on top of the standard log/slog.
// out and level are shared across WithFields-derived instances so SetOutput and
// SetLevel affect the whole logger tree, matching the previous behavior.
type logService struct {
	logger *slog.Logger
	out    *swappableWriter
	level  *slog.LevelVar
}

var (
	instance Logger
	once     sync.Once
)

// GetLogger returns the singleton logger instance.
// Default level is WARNING to reduce noise in CLI output.
func GetLogger() Logger {
	once.Do(
		func() {
			out := &swappableWriter{w: os.Stderr}
			level := &slog.LevelVar{}
			level.Set(slog.LevelWarn) // Default to WARNING level

			handler := slog.NewTextHandler(out, &slog.HandlerOptions{
				Level: level,
				ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
					switch a.Key {
					case slog.TimeKey:
						// Drop the timestamp for concise CLI output.
						return slog.Attr{}
					case slog.LevelKey:
						// Lowercase the level (level=error) to match the prior format.
						return slog.String(slog.LevelKey, strings.ToLower(a.Value.String()))
					default:
						return a
					}
				},
			})

			instance = &logService{
				logger: slog.New(handler),
				out:    out,
				level:  level,
			}
		},
	)

	return instance
}

// slogLevel maps a gibidify LogLevel to its slog equivalent.
func slogLevel(level LogLevel) slog.Level {
	switch level {
	case LogLevelDebug:
		return slog.LevelDebug
	case LogLevelInfo:
		return slog.LevelInfo
	case LogLevelError:
		return slog.LevelError
	default:
		// LogLevelWarn and unknown levels default to warn
		return slog.LevelWarn
	}
}

// log emits args joined with fmt.Sprint semantics at the given level.
func (l *logService) log(level slog.Level, args ...any) {
	if !l.logger.Enabled(context.Background(), level) {
		return
	}
	l.logger.Log(context.Background(), level, fmt.Sprint(args...))
}

// logf emits a printf-formatted message at the given level.
func (l *logService) logf(level slog.Level, format string, args ...any) {
	if !l.logger.Enabled(context.Background(), level) {
		return
	}
	l.logger.Log(context.Background(), level, fmt.Sprintf(format, args...))
}

// Debug logs a debug message.
func (l *logService) Debug(args ...any) { l.log(slog.LevelDebug, args...) }

// Debugf logs a formatted debug message.
func (l *logService) Debugf(format string, args ...any) { l.logf(slog.LevelDebug, format, args...) }

// Info logs an info message.
func (l *logService) Info(args ...any) { l.log(slog.LevelInfo, args...) }

// Infof logs a formatted info message.
func (l *logService) Infof(format string, args ...any) { l.logf(slog.LevelInfo, format, args...) }

// Warn logs a warning message.
func (l *logService) Warn(args ...any) { l.log(slog.LevelWarn, args...) }

// Warnf logs a formatted warning message.
func (l *logService) Warnf(format string, args ...any) { l.logf(slog.LevelWarn, format, args...) }

// Error logs an error message.
func (l *logService) Error(args ...any) { l.log(slog.LevelError, args...) }

// Errorf logs a formatted error message.
func (l *logService) Errorf(format string, args ...any) { l.logf(slog.LevelError, format, args...) }

// WithFields adds structured fields to log entries.
func (l *logService) WithFields(fields map[string]any) Logger {
	attrs := make([]any, 0, len(fields)*2)
	for k, v := range fields {
		attrs = append(attrs, k, v)
	}

	return &logService{
		logger: l.logger.With(attrs...),
		out:    l.out,
		level:  l.level,
	}
}

// SetLevel sets the logging level for the whole logger tree.
func (l *logService) SetLevel(level LogLevel) {
	l.level.Set(slogLevel(level))
}

// SetOutput sets the output destination for the whole logger tree.
func (l *logService) SetOutput(output io.Writer) {
	l.out.set(output)
}

// ParseLogLevel parses string log level to LogLevel.
func ParseLogLevel(level string) LogLevel {
	switch level {
	case string(LogLevelDebug):
		return LogLevelDebug
	case string(LogLevelInfo):
		return LogLevelInfo
	case string(LogLevelError):
		return LogLevelError
	default:
		// "warn", "warning", and unknown levels default to warn
		return LogLevelWarn
	}
}

// ValidateLogLevel validates if the provided log level is valid.
func ValidateLogLevel(level string) bool {
	switch level {
	case string(LogLevelDebug), string(LogLevelInfo), string(LogLevelWarn), LogLevelWarningAlias, string(LogLevelError):
		return true
	default:
		return false
	}
}
