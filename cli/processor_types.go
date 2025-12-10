// Package cli provides command-line interface functionality for gibidify.
package cli

import (
	"github.com/ivuorinen/gibidify/config"
	"github.com/ivuorinen/gibidify/fileproc"
	"github.com/ivuorinen/gibidify/metrics"
)

// Processor handles the main file processing logic.
type Processor struct {
	flags            *Flags
	backpressure     *fileproc.BackpressureManager
	resourceMonitor  *fileproc.ResourceMonitor
	ui               *UIManager
	metricsCollector *metrics.Collector
	metricsReporter  *metrics.Reporter
}

// NewProcessor creates a new processor with the given flags.
func NewProcessor(flags *Flags) *Processor {
	ui := NewUIManager()

	// Configure UI based on flags
	ui.SetColorOutput(!flags.NoColors && !flags.NoUI)
	ui.SetProgressOutput(!flags.NoProgress && !flags.NoUI)
	ui.SetSilentMode(flags.NoUI)

	// Initialize metrics system
	metricsCollector := metrics.NewCollector()
	metricsReporter := metrics.NewReporter(
		metricsCollector,
		flags.Verbose && !flags.NoUI,
		!flags.NoColors && !flags.NoUI,
	)

	return &Processor{
		flags:            flags,
		backpressure:     fileproc.NewBackpressureManager(),
		resourceMonitor:  fileproc.NewResourceMonitor(),
		ui:               ui,
		metricsCollector: metricsCollector,
		metricsReporter:  metricsReporter,
	}
}

// configureFileTypes configures the file type registry.
func (p *Processor) configureFileTypes() {
	if config.FileTypesEnabled() {
		fileproc.ConfigureFromSettings(
			config.CustomImageExtensions(),
			config.CustomBinaryExtensions(),
			config.CustomLanguages(),
			config.DisabledImageExtensions(),
			config.DisabledBinaryExtensions(),
			config.DisabledLanguageExtensions(),
		)
	}
}
