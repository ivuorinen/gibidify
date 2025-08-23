// Package cli provides command-line interface functionality for gibidify.
package cli

import (
	"flag"
	"fmt"
	"runtime"

	"github.com/ivuorinen/gibidify/config"
	"github.com/ivuorinen/gibidify/utils"
)

// Flags holds CLI flags values.
type Flags struct {
	SourceDir   string
	Destination string
	Prefix      string
	Suffix      string
	Concurrency int
	Format      string
	NoColors    bool
	NoProgress  bool
	Verbose     bool
	LogLevel    string
}

var (
	flagsParsed bool
	globalFlags *Flags
)

// ResetFlags resets the global flag parsing state for testing.
// This function should only be used in tests to ensure proper isolation.
func ResetFlags() {
	flagsParsed = false
	globalFlags = nil
}

// ParseFlags parses and validates CLI flags.
func ParseFlags() (*Flags, error) {
	if flagsParsed {
		return globalFlags, nil
	}

	flags := &Flags{}

	flag.StringVar(&flags.SourceDir, "source", "", "Source directory to scan recursively")
	flag.StringVar(&flags.Destination, "destination", "", "Output file to write aggregated code")
	flag.StringVar(&flags.Prefix, "prefix", "", "Text to add at the beginning of the output file")
	flag.StringVar(&flags.Suffix, "suffix", "", "Text to add at the end of the output file")
	flag.StringVar(&flags.Format, "format", "json", "Output format (json, markdown, yaml)")
	flag.IntVar(&flags.Concurrency, "concurrency", runtime.NumCPU(),
		"Number of concurrent workers (default: number of CPU cores)")
	flag.BoolVar(&flags.NoColors, "no-colors", false, "Disable colored output")
	flag.BoolVar(&flags.NoProgress, "no-progress", false, "Disable progress bars")
	flag.BoolVar(&flags.Verbose, "verbose", false, "Enable verbose output")
	flag.StringVar(&flags.LogLevel, "log-level", "warn", "Set log level (debug, info, warn, error)")

	flag.Parse()

	if err := flags.validate(); err != nil {
		return nil, err
	}

	if err := flags.setDefaultDestination(); err != nil {
		return nil, err
	}

	flagsParsed = true
	globalFlags = flags

	return flags, nil
}

// validate validates the CLI flags.
func (f *Flags) validate() error {
	if f.SourceDir == "" {
		return NewCLIMissingSourceError()
	}

	// Validate source path for security
	if err := utils.ValidateSourcePath(f.SourceDir); err != nil {
		return fmt.Errorf("validating source path: %w", err)
	}

	// Validate output format
	if err := config.ValidateOutputFormat(f.Format); err != nil {
		return fmt.Errorf("validating output format: %w", err)
	}

	// Validate concurrency
	if err := config.ValidateConcurrency(f.Concurrency); err != nil {
		return fmt.Errorf("validating concurrency: %w", err)
	}

	// Validate log level
	if !utils.ValidateLogLevel(f.LogLevel) {
		return fmt.Errorf("invalid log level: %s (must be: debug, info, warn, error)", f.LogLevel)
	}

	return nil
}

// setDefaultDestination sets the default destination if not provided.
func (f *Flags) setDefaultDestination() error {
	if f.Destination == "" {
		absRoot, err := utils.GetAbsolutePath(f.SourceDir)
		if err != nil {
			return fmt.Errorf("getting absolute path: %w", err)
		}
		baseName := utils.GetBaseName(absRoot)
		f.Destination = baseName + "." + f.Format
	}

	// Validate destination path for security
	if err := utils.ValidateDestinationPath(f.Destination); err != nil {
		return fmt.Errorf("validating destination path: %w", err)
	}

	return nil
}
