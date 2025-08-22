package cli

import (
	"github.com/ivuorinen/gibidify/config"
	"github.com/ivuorinen/gibidify/fileproc"
)

// Processor handles the main file processing logic.
type Processor struct {
	flags           *Flags
	backpressure    *fileproc.BackpressureManager
	resourceMonitor *fileproc.ResourceMonitor
	ui              *UIManager
}

// NewProcessor creates a new processor with the given flags.
func NewProcessor(flags *Flags) *Processor {
	ui := NewUIManager()

	// Configure UI based on flags
	ui.SetColorOutput(!flags.NoColors)
	ui.SetProgressOutput(!flags.NoProgress)

	return &Processor{
		flags:           flags,
		backpressure:    fileproc.NewBackpressureManager(),
		resourceMonitor: fileproc.NewResourceMonitor(),
		ui:              ui,
	}
}

// configureFileTypes configures the file type registry.
func (p *Processor) configureFileTypes() {
	if config.GetFileTypesEnabled() {
		fileproc.ConfigureFromSettings(
			config.GetCustomImageExtensions(),
			config.GetCustomBinaryExtensions(),
			config.GetCustomLanguages(),
			config.GetDisabledImageExtensions(),
			config.GetDisabledBinaryExtensions(),
			config.GetDisabledLanguageExtensions(),
		)
	}
}
