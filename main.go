// Package main for the gibidify CLI application.
// Repository: github.com/ivuorinen/gibidify
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/ivuorinen/gibidify/cli"
	"github.com/ivuorinen/gibidify/config"
	"github.com/ivuorinen/gibidify/shared"
)

// exitInterrupted is the exit status of a run stopped by a signal (128+SIGINT).
const exitInterrupted = 130

// Build information populated at link time via -ldflags by goreleaser.
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
	builtBy = "source"
)

func main() {
	// Ctrl-C and SIGTERM cancel the run; processing then stops the workers,
	// waits for the writer and discards its temporary output, leaving any
	// previous output in place.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	err := run(ctx)
	stop()

	// A successful run returns from main instead of calling os.Exit(0).
	if code := exitCode(err); code != 0 {
		os.Exit(code)
	}
}

// exitCode reports err to the user and returns the process exit status for it:
// 0 for nil, 130 for an interrupt, 1 for a user error and 2 for anything else.
func exitCode(err error) int {
	if err == nil {
		return 0
	}

	ui := cli.NewUIManager()

	// Ctrl-C or SIGTERM is not a failure of gibidify: report it plainly and use
	// the conventional 128+SIGINT status instead of the "unexpected error" path.
	if errors.Is(err, context.Canceled) {
		ui.PrintError("Interrupted")

		return exitInterrupted
	}

	// Handle errors with better formatting and suggestions
	if cli.IsUserError(err) {
		cli.NewErrorFormatter(ui).FormatError(err)

		return 1
	}

	// System errors still go to logger for debugging
	shared.GetLogger().Errorf("System error: %v", err)
	ui.PrintError("An unexpected error occurred. Please check the logs.")

	return 2
}

// printVersion writes build metadata to w. Extracted for testability.
func printVersion(w io.Writer) {
	_, _ = fmt.Fprintf(w, "gibidify %s\ncommit: %s\nbuilt: %s\nby: %s\n", version, commit, date, builtBy)
}

// Run executes the main logic of the CLI application using the provided context.
func run(ctx context.Context) error {
	// Parse CLI flags
	flags, err := cli.ParseFlags()
	if errors.Is(err, flag.ErrHelp) {
		// Usage has already been printed; asking for help is not a failure.
		return nil
	}
	if err != nil {
		return fmt.Errorf("parsing flags: %w", err)
	}

	// Honour --version regardless of argument position.
	if flags.ShowVersion {
		printVersion(os.Stdout)
		return nil
	}

	// Initialize logger with provided log level
	logger := shared.GetLogger()
	logger.SetLevel(shared.ParseLogLevel(flags.LogLevel))

	// A broken config aborts the run: continuing on defaults would drop the
	// user's exclusions and send the files they excluded to the output.
	if err := config.LoadConfig(); err != nil {
		return fmt.Errorf("loading configuration: %w", err)
	}

	// ParseFlags runs before the config file is loaded, so it cannot apply the
	// configured maxConcurrency ceiling; check it now that the config is in.
	if err := config.ValidateConcurrency(flags.Concurrency); err != nil {
		return fmt.Errorf("validating concurrency: %w", err)
	}

	// Create and run processor
	processor := cli.NewProcessor(flags)

	if err := processor.Process(ctx); err != nil {
		return fmt.Errorf("processing: %w", err)
	}

	return nil
}
