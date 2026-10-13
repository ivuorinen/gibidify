// Package cli provides command-line interface functionality for gibidify.
package cli

import (
	"github.com/ivuorinen/gibidify/config"
	"github.com/ivuorinen/gibidify/fileproc"
)

// Processor handles the main file processing logic.
type Processor struct {
	flags *Flags
	ui    *UIManager
}

// NewProcessor creates a new processor with the given flags.
func NewProcessor(flags *Flags) *Processor {
	ui := NewUIManager()

	// Flags can only turn features off. Setting them from the flags alone
	// overrode NewUIManager's terminal detection, so a redirected stderr got
	// ANSI codes and progress-bar redraws.
	ui.SetColorOutput(ui.enableColors && !flags.NoColors && !flags.NoUI)
	ui.SetProgressOutput(ui.enableProgress && !flags.NoProgress && !flags.NoUI)
	ui.SetSilentMode(flags.NoUI)

	return &Processor{
		flags: flags,
		ui:    ui,
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
