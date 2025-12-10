// Package metrics provides performance monitoring and reporting capabilities.
package metrics

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/ivuorinen/gibidify/shared"
)

// reportBuilder wraps strings.Builder with error accumulation for robust error handling.
type reportBuilder struct {
	b   *strings.Builder
	err error
}

// newReportBuilder creates a new report builder.
func newReportBuilder() *reportBuilder {
	return &reportBuilder{b: &strings.Builder{}}
}

// writeString writes a string, accumulating any errors.
func (rb *reportBuilder) writeString(s string) {
	if rb.err != nil {
		return
	}
	_, rb.err = rb.b.WriteString(s)
}

// fprintf formats and writes, accumulating any errors.
func (rb *reportBuilder) fprintf(format string, args ...any) {
	if rb.err != nil {
		return
	}
	_, rb.err = fmt.Fprintf(rb.b, format, args...)
}

// String returns the accumulated string, or empty string if there was an error.
func (rb *reportBuilder) String() string {
	if rb.err != nil {
		return ""
	}
	return rb.b.String()
}

// Reporter handles metrics reporting and formatting.
type Reporter struct {
	collector *Collector
	verbose   bool
	colors    bool
}

// NewReporter creates a new metrics reporter.
func NewReporter(collector *Collector, verbose, colors bool) *Reporter {
	return &Reporter{
		collector: collector,
		verbose:   verbose,
		colors:    colors,
	}
}

// ReportProgress provides a real-time progress report suitable for CLI output.
func (r *Reporter) ReportProgress() string {
	if r == nil || r.collector == nil {
		return "no metrics available"
	}

	metrics := r.collector.CurrentMetrics()

	if r.verbose {
		return r.formatVerboseProgress(metrics)
	}

	return r.formatBasicProgress(metrics)
}

// ReportFinal provides a comprehensive final report.
func (r *Reporter) ReportFinal() string {
	if r == nil || r.collector == nil {
		return ""
	}

	report := r.collector.GenerateReport()

	if r.verbose {
		return r.formatVerboseReport(report)
	}

	return r.formatBasicReport(report.Summary)
}

// formatBasicProgress formats basic progress information.
func (r *Reporter) formatBasicProgress(metrics ProcessingMetrics) string {
	b := newReportBuilder()

	// Basic stats
	b.writeString(fmt.Sprintf("Processed: %d files", metrics.ProcessedFiles))

	if metrics.SkippedFiles > 0 {
		b.writeString(fmt.Sprintf(", Skipped: %d", metrics.SkippedFiles))
	}

	if metrics.ErrorFiles > 0 {
		if r.colors {
			b.writeString(fmt.Sprintf(", \033[31mErrors: %d\033[0m", metrics.ErrorFiles))
		} else {
			b.writeString(fmt.Sprintf(", Errors: %d", metrics.ErrorFiles))
		}
	}

	// Processing rate
	if metrics.FilesPerSecond > 0 {
		b.writeString(fmt.Sprintf(" (%.1f files/sec)", metrics.FilesPerSecond))
	}

	return b.String()
}

// formatVerboseProgress formats detailed progress information.
func (r *Reporter) formatVerboseProgress(metrics ProcessingMetrics) string {
	b := newReportBuilder()

	// Header
	b.writeString("=== Processing Statistics ===\n")

	// File counts
	b.writeString(
		fmt.Sprintf(
			"Files - Total: %d, Processed: %d, Skipped: %d, Errors: %d\n",
			metrics.TotalFiles, metrics.ProcessedFiles, metrics.SkippedFiles, metrics.ErrorFiles,
		),
	)

	// Size information
	b.writeString(
		fmt.Sprintf(
			"Size - Processed: %s, Average: %s\n",
			r.formatBytes(metrics.ProcessedSize),
			r.formatBytes(int64(metrics.AverageFileSize)),
		),
	)

	if metrics.LargestFile > 0 {
		b.writeString(
			fmt.Sprintf(
				"File Size Range: %s - %s\n",
				r.formatBytes(metrics.SmallestFile),
				r.formatBytes(metrics.LargestFile),
			),
		)
	}

	// Performance
	b.writeString(
		fmt.Sprintf(
			"Performance - Files/sec: %.1f, MB/sec: %.1f\n",
			metrics.FilesPerSecond,
			metrics.BytesPerSecond/float64(shared.BytesPerMB),
		),
	)

	// Memory usage
	b.writeString(
		fmt.Sprintf(
			"Memory - Current: %dMB, Peak: %dMB, Goroutines: %d\n",
			metrics.CurrentMemoryMB, metrics.PeakMemoryMB, metrics.GoroutineCount,
		),
	)

	// Concurrency
	b.writeString(
		fmt.Sprintf(
			"Concurrency - Current: %d, Max: %d\n",
			metrics.CurrentConcurrency, metrics.MaxConcurrency,
		),
	)

	// Format breakdown (if available)
	if len(metrics.FormatCounts) > 0 {
		b.writeString("Format Breakdown:\n")
		formats := r.sortedMapKeys(metrics.FormatCounts)
		for _, format := range formats {
			count := metrics.FormatCounts[format]
			b.writeString(fmt.Sprintf(shared.MetricsFmtFileCount, format, count))
		}
	}

	// Processing time
	b.writeString(fmt.Sprintf(shared.MetricsFmtProcessingTime, metrics.ProcessingTime.Truncate(time.Millisecond)))

	return b.String()
}

// formatBasicReport formats a basic final report.
func (r *Reporter) formatBasicReport(metrics ProcessingMetrics) string {
	b := newReportBuilder()

	b.writeString("=== Processing Complete ===\n")
	b.writeString(
		fmt.Sprintf(
			"Total Files: %d (Processed: %d, Skipped: %d, Errors: %d)\n",
			metrics.TotalFiles, metrics.ProcessedFiles, metrics.SkippedFiles, metrics.ErrorFiles,
		),
	)

	b.writeString(
		fmt.Sprintf(
			"Total Size: %s, Average Rate: %.1f files/sec\n",
			r.formatBytes(metrics.ProcessedSize), metrics.FilesPerSecond,
		),
	)

	b.writeString(fmt.Sprintf(shared.MetricsFmtProcessingTime, metrics.ProcessingTime.Truncate(time.Millisecond)))

	return b.String()
}

// formatVerboseReport formats a comprehensive final report.
func (r *Reporter) formatVerboseReport(report ProfileReport) string {
	b := newReportBuilder()

	b.writeString("=== Comprehensive Processing Report ===\n\n")

	r.writeSummarySection(b, report)
	r.writeFormatBreakdown(b, report)
	r.writePhaseBreakdown(b, report)
	r.writeErrorBreakdown(b, report)
	r.writeResourceUsage(b, report)
	r.writeFileSizeStats(b, report)
	r.writeRecommendations(b, report)

	return b.String()
}

// writeSummarySection writes the summary section of the verbose report.
//
//goland:noinspection ALL
func (r *Reporter) writeSummarySection(b *reportBuilder, report ProfileReport) {
	metrics := report.Summary

	b.writeString("SUMMARY:\n")
	b.fprintf(
		"  Files: %d total (%d processed, %d skipped, %d errors)\n",
		metrics.TotalFiles, metrics.ProcessedFiles, metrics.SkippedFiles, metrics.ErrorFiles,
	)
	b.fprintf(
		"  Size: %s processed (avg: %s per file)\n",
		r.formatBytes(metrics.ProcessedSize), r.formatBytes(int64(metrics.AverageFileSize)),
	)
	b.fprintf(
		"  Time: %v (%.1f files/sec, %.1f MB/sec)\n",
		metrics.ProcessingTime.Truncate(time.Millisecond),
		metrics.FilesPerSecond, metrics.BytesPerSecond/float64(shared.BytesPerMB),
	)
	b.fprintf("  Performance Index: %.1f\n", report.PerformanceIndex)
}

// writeFormatBreakdown writes the format breakdown section.
func (r *Reporter) writeFormatBreakdown(b *reportBuilder, report ProfileReport) {
	if len(report.FormatBreakdown) == 0 {
		return
	}

	b.writeString("\nFORMAT BREAKDOWN:\n")
	formats := make([]string, 0, len(report.FormatBreakdown))
	for format := range report.FormatBreakdown {
		formats = append(formats, format)
	}
	sort.Strings(formats)

	for _, format := range formats {
		formatMetrics := report.FormatBreakdown[format]
		b.fprintf(shared.MetricsFmtFileCount, format, formatMetrics.Count)
	}
}

// writePhaseBreakdown writes the phase timing breakdown section.
func (r *Reporter) writePhaseBreakdown(b *reportBuilder, report ProfileReport) {
	if len(report.PhaseBreakdown) == 0 {
		return
	}

	b.writeString("\nPHASE BREAKDOWN:\n")
	phases := []string{
		shared.MetricsPhaseCollection,
		shared.MetricsPhaseProcessing,
		shared.MetricsPhaseWriting,
		shared.MetricsPhaseFinalize,
	}
	for _, phase := range phases {
		if phaseMetrics, exists := report.PhaseBreakdown[phase]; exists {
			b.fprintf(
				"  %s: %v (%.1f%%)\n",
				cases.Title(language.English).String(phase),
				phaseMetrics.TotalTime.Truncate(time.Millisecond),
				phaseMetrics.Percentage,
			)
		}
	}
}

// writeErrorBreakdown writes the error breakdown section.
func (r *Reporter) writeErrorBreakdown(b *reportBuilder, report ProfileReport) {
	if len(report.ErrorBreakdown) == 0 {
		return
	}

	b.writeString("\nERROR BREAKDOWN:\n")
	errors := r.sortedMapKeys(report.ErrorBreakdown)
	for _, errorType := range errors {
		count := report.ErrorBreakdown[errorType]
		b.fprintf("  %s: %d occurrences\n", errorType, count)
	}
}

// writeResourceUsage writes the resource usage section.
func (r *Reporter) writeResourceUsage(b *reportBuilder, report ProfileReport) {
	metrics := report.Summary
	b.writeString("\nRESOURCE USAGE:\n")
	b.fprintf(
		"  Memory: %dMB current, %dMB peak\n",
		metrics.CurrentMemoryMB, metrics.PeakMemoryMB,
	)
	b.fprintf(
		"  Concurrency: %d current, %d max, %d goroutines\n",
		metrics.CurrentConcurrency, metrics.MaxConcurrency, metrics.GoroutineCount,
	)
}

// writeFileSizeStats writes the file size statistics section.
func (r *Reporter) writeFileSizeStats(b *reportBuilder, report ProfileReport) {
	metrics := report.Summary
	if metrics.ProcessedFiles == 0 {
		return
	}

	b.writeString("\nFILE SIZE STATISTICS:\n")
	b.fprintf(
		"  Range: %s - %s\n",
		r.formatBytes(metrics.SmallestFile), r.formatBytes(metrics.LargestFile),
	)
	b.fprintf("  Average: %s\n", r.formatBytes(int64(metrics.AverageFileSize)))
}

// writeRecommendations writes the recommendations section.
func (r *Reporter) writeRecommendations(b *reportBuilder, report ProfileReport) {
	if len(report.Recommendations) == 0 {
		return
	}

	b.writeString("\nRECOMMENDATIONS:\n")
	for i, rec := range report.Recommendations {
		b.fprintf("  %d. %s\n", i+1, rec)
	}
}

// formatBytes formats byte counts in human-readable format.
func (r *Reporter) formatBytes(bytes int64) string {
	if bytes == 0 {
		return "0B"
	}

	if bytes < shared.BytesPerKB {
		return fmt.Sprintf(shared.MetricsFmtBytesShort, bytes)
	}

	exp := 0
	for n := bytes / shared.BytesPerKB; n >= shared.BytesPerKB; n /= shared.BytesPerKB {
		exp++
	}

	divisor := int64(1)
	for i := 0; i < exp+1; i++ {
		divisor *= shared.BytesPerKB
	}

	return fmt.Sprintf(shared.MetricsFmtBytesHuman, float64(bytes)/float64(divisor), "KMGTPE"[exp])
}

// sortedMapKeys returns sorted keys from a map for consistent output.
func (r *Reporter) sortedMapKeys(m map[string]int64) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	return keys
}

// QuickStats returns a quick one-line status suitable for progress bars.
func (r *Reporter) QuickStats() string {
	if r == nil || r.collector == nil {
		return "0/0 files"
	}

	metrics := r.collector.CurrentMetrics()

	status := fmt.Sprintf("%d/%d files", metrics.ProcessedFiles, metrics.TotalFiles)
	if metrics.FilesPerSecond > 0 {
		status += fmt.Sprintf(" (%.1f/s)", metrics.FilesPerSecond)
	}

	if metrics.ErrorFiles > 0 {
		if r.colors {
			status += fmt.Sprintf(" \033[31m%d errors\033[0m", metrics.ErrorFiles)
		} else {
			status += fmt.Sprintf(" %d errors", metrics.ErrorFiles)
		}
	}

	return status
}
