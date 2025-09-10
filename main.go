// Package main for the gibidify CLI application.
// Repository: github.com/ivuorinen/gibidify
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/ivuorinen/gibidify/cli"
	"github.com/ivuorinen/gibidify/config"
	"github.com/ivuorinen/gibidify/shared"
)

func main() {
	// Initialize UI for error handling
	ui := cli.NewUIManager()
	errorFormatter := cli.NewErrorFormatter(ui)

	// In production, use a background context.
	if err := run(context.Background()); err != nil {
		// Handle errors with better formatting and suggestions
		if cli.IsUserError(err) {
			errorFormatter.FormatError(err)
			os.Exit(1)
		}
		// System errors still go to logger for debugging
		logger := shared.GetLogger()
		logger.Errorf("System error: %v", err)
		ui.PrintError("An unexpected error occurred. Please check the logs.")
		os.Exit(2)
	}
}

// Run executes the main logic of the CLI application using the provided context.
func run(ctx context.Context) error {
	// Parse CLI flags
	flags, err := cli.ParseFlags()
	if err != nil {
		return fmt.Errorf("parsing flags: %w", err)
	}

	// Initialize logger with provided log level
	logger := shared.GetLogger()
	logger.SetLevel(shared.ParseLogLevel(flags.LogLevel))

	// Load configuration
	config.LoadConfig()

	// Create and run processor
	processor := cli.NewProcessor(flags)

	if err := processor.Process(ctx); err != nil {
		return fmt.Errorf("processing: %w", err)
	}

	return nil
}
