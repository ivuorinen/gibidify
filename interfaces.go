// Package main provides core interfaces for the gibidify application.
package main

import (
	"context"
	"io"

	"github.com/ivuorinen/gibidify/shared"
)

// Processor defines the interface for file processors.
// This interface allows for easier testing and mocking of the main processing logic.
type Processor interface {
	// Process starts the file processing workflow with the given context.
	// It returns an error if processing fails at any stage.
	Process(ctx context.Context) error
}

// FileProcessorInterface defines the interface for individual file processing.
// This abstracts the file processing logic for better testability.
type FileProcessorInterface interface {
	// ProcessFile processes a single file and sends the result to the output channel.
	ProcessFile(ctx context.Context, filePath string, outCh chan<- WriteRequest)

	// ProcessWithContext processes a file and returns the content directly.
	ProcessWithContext(ctx context.Context, filePath string) (string, error)
}

// ResourceMonitorInterface defines the interface for resource monitoring.
// This allows for mocking and testing of resource management functionality.
type ResourceMonitorInterface interface {
	// Start begins resource monitoring.
	Start() error

	// Stop stops resource monitoring and cleanup.
	Stop() error

	// CheckResourceLimits validates current resource usage against limits.
	CheckResourceLimits() error

	// Metrics returns current resource usage metrics.
	Metrics() ResourceMetrics
}

// MetricsCollectorInterface defines the interface for metrics collection.
// This enables easier testing and different metrics backend implementations.
type MetricsCollectorInterface interface {
	// RecordFileProcessed records the processing of a single file.
	RecordFileProcessed(result FileProcessingResult)

	// IncrementConcurrency increments the current concurrency counter.
	IncrementConcurrency()

	// DecrementConcurrency decrements the current concurrency counter.
	DecrementConcurrency()

	// CurrentMetrics returns the current processing metrics.
	CurrentMetrics() ProcessingMetrics

	// GenerateReport generates a comprehensive processing report.
	GenerateReport() ProfileReport

	// Reset resets all metrics to initial state.
	Reset()
}

// UIManagerInterface defines the interface for user interface management.
// This abstracts UI operations for better testing and different UI implementations.
type UIManagerInterface interface {
	// PrintInfo prints an informational message.
	PrintInfo(message string)

	// PrintWarning prints a warning message.
	PrintWarning(message string)

	// PrintError prints an error message.
	PrintError(message string)

	// PrintSuccess prints a success message.
	PrintSuccess(message string)

	// SetColorOutput enables or disables colored output.
	SetColorOutput(enabled bool)

	// SetProgressOutput enables or disables progress indicators.
	SetProgressOutput(enabled bool)
}

// WriterInterface defines the interface for output writers.
// This allows for different output formats and destinations.
type WriterInterface interface {
	// Write writes the processed content to the destination.
	Write(req WriteRequest) error

	// Close finalizes the output and closes any resources.
	Close() error

	// GetFormat returns the output format supported by this writer.
	GetFormat() string
}

// BackpressureManagerInterface defines the interface for backpressure management.
// This abstracts memory and flow control for better testing.
type BackpressureManagerInterface interface {
	// CheckBackpressure returns true if backpressure should be applied.
	CheckBackpressure() bool

	// UpdateMemoryUsage updates the current memory usage tracking.
	UpdateMemoryUsage(bytes int64)

	// GetMemoryUsage returns current memory usage statistics.
	GetMemoryUsage() int64

	// Reset resets backpressure state to initial values.
	Reset()
}

// TemplateEngineInterface defines the interface for template processing.
// This allows for different templating systems and easier testing.
type TemplateEngineInterface interface {
	// RenderHeader renders the document header using the configured template.
	RenderHeader(ctx TemplateContext) (string, error)

	// RenderFooter renders the document footer using the configured template.
	RenderFooter(ctx TemplateContext) (string, error)

	// RenderFileContent renders individual file content with formatting.
	RenderFileContent(ctx FileContext) (string, error)

	// RenderMetadata renders metadata section if enabled.
	RenderMetadata(ctx TemplateContext) (string, error)
}

// ConfigLoaderInterface defines the interface for configuration management.
// This enables different configuration sources and easier testing.
type ConfigLoaderInterface interface {
	// LoadConfig loads configuration from the appropriate source.
	LoadConfig() error

	// GetString returns a string configuration value.
	GetString(key string) string

	// GetInt returns an integer configuration value.
	GetInt(key string) int

	// GetBool returns a boolean configuration value.
	GetBool(key string) bool

	// GetStringSlice returns a string slice configuration value.
	GetStringSlice(key string) []string
}

// LoggerInterface defines the interface for logging operations.
// This abstracts logging for better testing and different log backends.
type LoggerInterface = shared.Logger

// These types are referenced by the interfaces but need to be defined
// elsewhere in the codebase. They are included here for documentation.

type WriteRequest struct {
	Path     string
	Content  string
	IsStream bool
	Reader   io.Reader
	Size     int64
}

type ResourceMetrics struct {
	FilesProcessed     int64
	TotalSizeProcessed int64
	ConcurrentReads    int64
	MaxConcurrentReads int64
}

type FileProcessingResult struct {
	FilePath   string
	FileSize   int64
	Format     string
	Success    bool
	Error      error
	Skipped    bool
	SkipReason string
}

type ProcessingMetrics struct {
	TotalFiles     int64
	ProcessedFiles int64
	ErrorFiles     int64
	SkippedFiles   int64
	TotalSize      int64
	ProcessedSize  int64
}

type ProfileReport struct {
	Summary ProcessingMetrics
	// Additional report fields would be defined in the metrics package
}

type TemplateContext struct {
	Files []FileContext
	// Additional context fields would be defined in the templates package
}

type FileContext struct {
	Path    string
	Content string
	// Additional file context fields would be defined in the templates package
}

type LogLevel int

const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelWarn
	LogLevelError
)
