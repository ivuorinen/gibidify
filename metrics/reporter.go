// Package metrics provides performance monitoring and reporting capabilities.
package metrics

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

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
	metrics := r.collector.GetCurrentMetrics()

	if r.verbose {
		return r.formatVerboseProgress(metrics)
	}

	return r.formatBasicProgress(metrics)
}

// ReportFinal provides a comprehensive final report.
func (r *Reporter) ReportFinal() string {
	report := r.collector.GenerateReport()

	if r.verbose {
		return r.formatVerboseReport(report)
	}

	return r.formatBasicReport(report.Summary)
}

// formatBasicProgress formats basic progress information.
func (r *Reporter) formatBasicProgress(metrics ProcessingMetrics) string {
	var b strings.Builder

	// Basic stats
	_, _ = b.WriteString(fmt.Sprintf("Processed: %d files", metrics.ProcessedFiles))

	if metrics.SkippedFiles > 0 {
		_, _ = b.WriteString(fmt.Sprintf(", Skipped: %d", metrics.SkippedFiles))
	}

	if metrics.ErrorFiles > 0 {
		if r.colors {
			_, _ = b.WriteString(fmt.Sprintf(", \033[31mErrors: %d\033[0m", metrics.ErrorFiles))
		} else {
			_, _ = b.WriteString(fmt.Sprintf(", Errors: %d", metrics.ErrorFiles))
		}
	}

	// Processing rate
	if metrics.FilesPerSecond > 0 {
		_, _ = b.WriteString(fmt.Sprintf(" (%.1f files/sec)", metrics.FilesPerSecond))
	}

	return b.String()
}

// formatVerboseProgress formats detailed progress information.
func (r *Reporter) formatVerboseProgress(metrics ProcessingMetrics) string {
	var b strings.Builder

	// Header
	_, _ = b.WriteString("=== Processing Statistics ===\n")

	// File counts
	_, _ = b.WriteString(
		fmt.Sprintf(
			"Files - Total: %d, Processed: %d, Skipped: %d, Errors: %d\n",
			metrics.TotalFiles, metrics.ProcessedFiles, metrics.SkippedFiles, metrics.ErrorFiles,
		),
	)

	// Size information
	_, _ = b.WriteString(
		fmt.Sprintf(
			"Size - Processed: %s, Average: %s\n",
			r.formatBytes(metrics.ProcessedSize),
			r.formatBytes(int64(metrics.AverageFileSize)),
		),
	)

	if metrics.LargestFile > 0 {
		_, _ = b.WriteString(
			fmt.Sprintf(
				"File Size Range: %s - %s\n",
				r.formatBytes(metrics.SmallestFile),
				r.formatBytes(metrics.LargestFile),
			),
		)
	}

	// Performance
	_, _ = b.WriteString(
		fmt.Sprintf(
			"Performance - Files/sec: %.1f, MB/sec: %.1f\n",
			metrics.FilesPerSecond,
			metrics.BytesPerSecond/1024/1024,
		),
	)

	// Memory usage
	_, _ = b.WriteString(
		fmt.Sprintf(
			"Memory - Current: %dMB, Peak: %dMB, Goroutines: %d\n",
			metrics.CurrentMemoryMB, metrics.PeakMemoryMB, metrics.GoroutineCount,
		),
	)

	// Concurrency
	_, _ = b.WriteString(
		fmt.Sprintf(
			"Concurrency - Current: %d, Max: %d\n",
			metrics.CurrentConcurrency, metrics.MaxConcurrency,
		),
	)

	// Format breakdown (if available)
	if len(metrics.FormatCounts) > 0 {
		_, _ = b.WriteString("Format Breakdown:\n")
		formats := r.sortedMapKeys(metrics.FormatCounts)
		for _, format := range formats {
			count := metrics.FormatCounts[format]
			_, _ = b.WriteString(fmt.Sprintf("  %s: %d files\n", format, count))
		}
	}

	// Processing time
	_, _ = b.WriteString(fmt.Sprintf("Processing Time: %v\n", metrics.ProcessingTime.Truncate(time.Millisecond)))

	return b.String()
}

// formatBasicReport formats a basic final report.
func (r *Reporter) formatBasicReport(metrics ProcessingMetrics) string {
	var b strings.Builder

	_, _ = b.WriteString("=== Processing Complete ===\n")
	_, _ = b.WriteString(
		fmt.Sprintf(
			"Total Files: %d (Processed: %d, Skipped: %d, Errors: %d)\n",
			metrics.TotalFiles, metrics.ProcessedFiles, metrics.SkippedFiles, metrics.ErrorFiles,
		),
	)

	_, _ = b.WriteString(
		fmt.Sprintf(
			"Total Size: %s, Average Rate: %.1f files/sec\n",
			r.formatBytes(metrics.ProcessedSize), metrics.FilesPerSecond,
		),
	)

	_, _ = b.WriteString(fmt.Sprintf("Processing Time: %v\n", metrics.ProcessingTime.Truncate(time.Millisecond)))

	return b.String()
}

// formatVerboseReport formats a comprehensive final report.
func (r *Reporter) formatVerboseReport(report ProfileReport) string {
	var b strings.Builder

	_, _ = b.WriteString("=== Comprehensive Processing Report ===\n\n")

	r.writeSummarySection(&b, report)
	r.writeFormatBreakdown(&b, report)
	r.writePhaseBreakdown(&b, report)
	r.writeErrorBreakdown(&b, report)
	r.writeResourceUsage(&b, report)
	r.writeFileSizeStats(&b, report)
	r.writeRecommendations(&b, report)

	return b.String()
}

// writeSummarySection writes the summary section of the verbose report.
//
//goland:noinspection ALL
func (r *Reporter) writeSummarySection(b *strings.Builder, report ProfileReport) {
	metrics := report.Summary

	_, _ = b.WriteString("SUMMARY:\n")
	_, _ = fmt.Fprintf(
		b, "  Files: %d total (%d processed, %d skipped, %d errors)\n",
		metrics.TotalFiles, metrics.ProcessedFiles, metrics.SkippedFiles, metrics.ErrorFiles,
	)
	_, _ = fmt.Fprintf(
		b, "  Size: %s processed (avg: %s per file)\n",
		r.formatBytes(metrics.ProcessedSize), r.formatBytes(int64(metrics.AverageFileSize)),
	)
	_, _ = fmt.Fprintf(
		b, "  Time: %v (%.1f files/sec, %.1f MB/sec)\n",
		metrics.ProcessingTime.Truncate(time.Millisecond),
		metrics.FilesPerSecond, metrics.BytesPerSecond/1024/1024,
	)
	_, _ = fmt.Fprintf(b, "  Performance Index: %.1f\n", report.PerformanceIndex)
}

// writeFormatBreakdown writes the format breakdown section.
func (r *Reporter) writeFormatBreakdown(b *strings.Builder, report ProfileReport) {
	if len(report.FormatBreakdown) == 0 {
		return
	}

	_, _ = b.WriteString("\nFORMAT BREAKDOWN:\n")
	formats := make([]string, 0, len(report.FormatBreakdown))
	for format := range report.FormatBreakdown {
		formats = append(formats, format)
	}
	sort.Strings(formats)

	for _, format := range formats {
		formatMetrics := report.FormatBreakdown[format]
		_, _ = fmt.Fprintf(b, "  %s: %d files\n", format, formatMetrics.Count)
	}
}

// writePhaseBreakdown writes the phase timing breakdown section.
func (r *Reporter) writePhaseBreakdown(b *strings.Builder, report ProfileReport) {
	if len(report.PhaseBreakdown) == 0 {
		return
	}

	_, _ = b.WriteString("\nPHASE BREAKDOWN:\n")
	phases := []string{PhaseCollection, PhaseProcessing, PhaseWriting, PhaseFinalize}
	for _, phase := range phases {
		if phaseMetrics, exists := report.PhaseBreakdown[phase]; exists {
			_, _ = fmt.Fprintf(
				b, "  %s: %v (%.1f%%)\n",
				cases.Title(language.English).String(phase),
				phaseMetrics.TotalTime.Truncate(time.Millisecond),
				phaseMetrics.Percentage,
			)
		}
	}
}

// writeErrorBreakdown writes the error breakdown section.
func (r *Reporter) writeErrorBreakdown(b *strings.Builder, report ProfileReport) {
	if len(report.ErrorBreakdown) == 0 {
		return
	}

	_, _ = b.WriteString("\nERROR BREAKDOWN:\n")
	errors := r.sortedMapKeys(report.ErrorBreakdown)
	for _, errorType := range errors {
		count := report.ErrorBreakdown[errorType]
		_, _ = fmt.Fprintf(b, "  %s: %d occurrences\n", errorType, count)
	}
}

// writeResourceUsage writes the resource usage section.
func (r *Reporter) writeResourceUsage(b *strings.Builder, report ProfileReport) {
	metrics := report.Summary
	_, _ = b.WriteString("\nRESOURCE USAGE:\n")
	_, _ = fmt.Fprintf(
		b, "  Memory: %dMB current, %dMB peak\n",
		metrics.CurrentMemoryMB, metrics.PeakMemoryMB,
	)
	_, _ = fmt.Fprintf(
		b, "  Concurrency: %d current, %d max, %d goroutines\n",
		metrics.CurrentConcurrency, metrics.MaxConcurrency, metrics.GoroutineCount,
	)
}

// writeFileSizeStats writes the file size statistics section.
func (r *Reporter) writeFileSizeStats(b *strings.Builder, report ProfileReport) {
	metrics := report.Summary
	if metrics.ProcessedFiles == 0 {
		return
	}

	_, _ = b.WriteString("\nFILE SIZE STATISTICS:\n")
	_, _ = fmt.Fprintf(
		b, "  Range: %s - %s\n",
		r.formatBytes(metrics.SmallestFile), r.formatBytes(metrics.LargestFile),
	)
	_, _ = fmt.Fprintf(b, "  Average: %s\n", r.formatBytes(int64(metrics.AverageFileSize)))
}

// writeRecommendations writes the recommendations section.
func (r *Reporter) writeRecommendations(b *strings.Builder, report ProfileReport) {
	if len(report.Recommendations) == 0 {
		return
	}

	_, _ = b.WriteString("\nRECOMMENDATIONS:\n")
	for i, rec := range report.Recommendations {
		_, _ = fmt.Fprintf(b, "  %d. %s\n", i+1, rec)
	}
}

// formatBytes formats byte counts in human-readable format.
func (r *Reporter) formatBytes(bytes int64) string {
	if bytes == 0 {
		return "0B"
	}

	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%dB", bytes)
	}

	exp := 0
	for n := bytes / unit; n >= unit; n /= unit {
		exp++
	}

	divisor := int64(1)
	for i := 0; i < exp+1; i++ {
		divisor *= 1024
	}

	return fmt.Sprintf("%.1f%cB", float64(bytes)/float64(divisor), "KMGTPE"[exp])
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

// GetQuickStats returns a quick one-line status suitable for progress bars.
func (r *Reporter) GetQuickStats() string {
	metrics := r.collector.GetCurrentMetrics()

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
