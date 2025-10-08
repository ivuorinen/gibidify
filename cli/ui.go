package cli

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/fatih/color"
	"github.com/schollz/progressbar/v3"

	"github.com/ivuorinen/gibidify/gibidiutils"
)

// UIManager handles CLI user interface elements.
type UIManager struct {
	enableColors   bool
	enableProgress bool
	progressBar    *progressbar.ProgressBar
	output         io.Writer
}

// NewUIManager creates a new UI manager.
func NewUIManager() *UIManager {
	return &UIManager{
		enableColors:   isColorTerminal(),
		enableProgress: isInteractiveTerminal(),
		output:         os.Stderr, // Progress and colors go to stderr
	}
}

// SetColorOutput enables or disables colored output.
func (ui *UIManager) SetColorOutput(enabled bool) {
	ui.enableColors = enabled
	color.NoColor = !enabled
}

// SetProgressOutput enables or disables progress bars.
func (ui *UIManager) SetProgressOutput(enabled bool) {
	ui.enableProgress = enabled
}

// StartProgress initializes a progress bar for file processing.
func (ui *UIManager) StartProgress(total int, description string) {
	if !ui.enableProgress || total <= 0 {
		return
	}

	// Set progress bar theme based on color support
	var theme progressbar.Theme
	if ui.enableColors {
		theme = progressbar.Theme{
			Saucer:        color.GreenString("█"),
			SaucerHead:    color.GreenString("█"),
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}
	} else {
		theme = progressbar.Theme{
			Saucer:        "█",
			SaucerHead:    "█",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}
	}

	ui.progressBar = progressbar.NewOptions(
		total,
		progressbar.OptionSetWriter(ui.output),
		progressbar.OptionSetDescription(description),
		progressbar.OptionSetTheme(theme),
		progressbar.OptionShowCount(),
		progressbar.OptionShowIts(),
		progressbar.OptionSetWidth(40),
		progressbar.OptionThrottle(100*time.Millisecond),
		progressbar.OptionOnCompletion(
			func() {
				_, _ = fmt.Fprint(ui.output, "\n")
			},
		),
		progressbar.OptionSetRenderBlankState(true),
	)
}

// UpdateProgress increments the progress bar.
func (ui *UIManager) UpdateProgress(increment int) {
	if ui.progressBar != nil {
		_ = ui.progressBar.Add(increment)
	}
}

// FinishProgress completes the progress bar.
func (ui *UIManager) FinishProgress() {
	if ui.progressBar != nil {
		_ = ui.progressBar.Finish()
		ui.progressBar = nil
	}
}

// writeMessage writes a formatted message with optional colorization.
// It handles color enablement, formatting, writing to output, and error logging.
func (ui *UIManager) writeMessage(
	icon, methodName, format string,
	colorFunc func(string, ...interface{}) string,
	args ...interface{},
) {
	msg := icon + " " + format
	var output string
	if ui.enableColors && colorFunc != nil {
		output = colorFunc(msg, args...)
	} else {
		output = fmt.Sprintf(msg, args...)
	}

	if _, err := fmt.Fprintf(ui.output, "%s\n", output); err != nil {
		gibidiutils.LogError(fmt.Sprintf("UIManager.%s: failed to write to output", methodName), err)
	}
}

// PrintSuccess prints a success message in green (to ui.output if set).
func (ui *UIManager) PrintSuccess(format string, args ...interface{}) {
	ui.writeMessage(gibidiutils.IconSuccess, "PrintSuccess", format, color.GreenString, args...)
}

// PrintError prints an error message in red (to ui.output if set).
func (ui *UIManager) PrintError(format string, args ...interface{}) {
	ui.writeMessage(gibidiutils.IconError, "PrintError", format, color.RedString, args...)
}

// PrintWarning prints a warning message in yellow (to ui.output if set).
func (ui *UIManager) PrintWarning(format string, args ...interface{}) {
	ui.writeMessage(gibidiutils.IconWarning, "PrintWarning", format, color.YellowString, args...)
}

// PrintInfo prints an info message in blue (to ui.output if set).
func (ui *UIManager) PrintInfo(format string, args ...interface{}) {
	ui.writeMessage(gibidiutils.IconInfo, "PrintInfo", format, color.BlueString, args...)
}

// PrintHeader prints a header message in bold.
func (ui *UIManager) PrintHeader(format string, args ...interface{}) {
	if ui.enableColors {
		_, _ = color.New(color.Bold).Fprintf(ui.output, format+"\n", args...)
	} else {
		ui.printf(format+"\n", args...)
	}
}

// isColorTerminal checks if the terminal supports colors.
func isColorTerminal() bool {
	// Check if FORCE_COLOR is set
	if os.Getenv("FORCE_COLOR") != "" {
		return true
	}

	// Check common environment variables
	term := os.Getenv("TERM")
	if term == "" || term == "dumb" {
		return false
	}

	// Check for CI environments that typically don't support colors
	if os.Getenv("CI") != "" {
		// GitHub Actions supports colors
		if os.Getenv("GITHUB_ACTIONS") == "true" {
			return true
		}
		// Most other CI systems don't
		return false
	}

	// Check if NO_COLOR is set (https://no-color.org/)
	if os.Getenv("NO_COLOR") != "" {
		return false
	}

	return true
}

// isInteractiveTerminal checks if we're running in an interactive terminal.
func isInteractiveTerminal() bool {
	// Check if stderr is a terminal (where we output progress/colors)
	fileInfo, err := os.Stderr.Stat()
	if err != nil {
		return false
	}
	return (fileInfo.Mode() & os.ModeCharDevice) != 0
}

// printf is a helper that ignores printf errors (for UI output).
func (ui *UIManager) printf(format string, args ...interface{}) {
	_, _ = fmt.Fprintf(ui.output, format, args...)
}
