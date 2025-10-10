// Package main for the gibidify CLI application.
// Repository: github.com/ivuorinen/gibidify
package main

import (
	"context"
	"os"

	"github.com/sirupsen/logrus"

	"github.com/ivuorinen/gibidify/cli"
	"github.com/ivuorinen/gibidify/config"
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
		// System errors still go to logrus for debugging
		logrus.Errorf("System error: %v", err)
		ui.PrintError("An unexpected error occurred. Please check the logs.")
		os.Exit(2)
	}
}

// Run executes the main logic of the CLI application using the provided context.
func run(ctx context.Context) error {
	// Parse CLI flags
	flags, err := cli.ParseFlags()
	if err != nil {
		return err
	}

	// Load configuration
	config.LoadConfig()

	// Create and run processor
	processor := cli.NewProcessor(flags)
	return processor.Process(ctx)
}
