// Package shared provides logging utilities for gibidify.
package shared

import (
	"io"
	"os"
	"sync"

	"github.com/sirupsen/logrus"
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

// logService implements the Logger interface using logrus.
type logService struct {
	logger *logrus.Logger
	entry  *logrus.Entry
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
			logger := logrus.New()
			logger.SetLevel(logrus.WarnLevel) // Default to WARNING level
			logger.SetOutput(os.Stderr)
			logger.SetFormatter(
				&logrus.TextFormatter{
					DisableColors: false,
					FullTimestamp: false,
				},
			)

			instance = &logService{
				logger: logger,
				entry:  logger.WithFields(logrus.Fields{}),
			}
		},
	)

	return instance
}

// Debug logs a debug message.
func (l *logService) Debug(args ...any) {
	l.entry.Debug(args...)
}

// Debugf logs a formatted debug message.
func (l *logService) Debugf(format string, args ...any) {
	l.entry.Debugf(format, args...)
}

// Info logs an info message.
func (l *logService) Info(args ...any) {
	l.entry.Info(args...)
}

// Infof logs a formatted info message.
func (l *logService) Infof(format string, args ...any) {
	l.entry.Infof(format, args...)
}

// Warn logs a warning message.
func (l *logService) Warn(args ...any) {
	l.entry.Warn(args...)
}

// Warnf logs a formatted warning message.
func (l *logService) Warnf(format string, args ...any) {
	l.entry.Warnf(format, args...)
}

// Error logs an error message.
func (l *logService) Error(args ...any) {
	l.entry.Error(args...)
}

// Errorf logs a formatted error message.
func (l *logService) Errorf(format string, args ...any) {
	l.entry.Errorf(format, args...)
}

// WithFields adds structured fields to log entries.
func (l *logService) WithFields(fields map[string]any) Logger {
	logrusFields := make(logrus.Fields)
	for k, v := range fields {
		logrusFields[k] = v
	}

	return &logService{
		logger: l.logger,
		entry:  l.entry.WithFields(logrusFields),
	}
}

// SetLevel sets the logging level.
func (l *logService) SetLevel(level LogLevel) {
	var logrusLevel logrus.Level
	switch level {
	case LogLevelDebug:
		logrusLevel = logrus.DebugLevel
	case LogLevelInfo:
		logrusLevel = logrus.InfoLevel
	case LogLevelError:
		logrusLevel = logrus.ErrorLevel
	default:
		// LogLevelWarn and unknown levels default to warn
		logrusLevel = logrus.WarnLevel
	}
	l.logger.SetLevel(logrusLevel)
}

// SetOutput sets the output destination for logs.
func (l *logService) SetOutput(output io.Writer) {
	l.logger.SetOutput(output)
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
