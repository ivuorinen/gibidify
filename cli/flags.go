// Package cli provides command-line interface functionality for gibidify.
package cli

import (
	"flag"
	"fmt"
	"os"
	"runtime"

	"github.com/ivuorinen/gibidify/config"
	"github.com/ivuorinen/gibidify/shared"
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
	NoUI        bool
	Verbose     bool
	ShowVersion bool
	LogLevel    string
	// IncludeSensitive keeps credential files (.env, keys) skipped by default.
	IncludeSensitive bool
}

var (
	flagsParsed bool
	globalFlags *Flags
)

// ResetFlags resets the global flag parsing state for testing.
// This function should only be used in tests to ensure proper isolation.
// It deliberately leaves flag.CommandLine alone: ParseFlags uses its own
// FlagSet, and replacing CommandLine drops the testing package's flags, which
// panics under `go test -count=N`.
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

	fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	fs.StringVar(&flags.SourceDir, shared.CLIArgSource, "", "Source directory to scan recursively")
	fs.StringVar(&flags.Destination, "destination", "", "Output file to write aggregated code")
	fs.StringVar(&flags.Prefix, "prefix", "", "Text to add at the beginning of the output file")
	fs.StringVar(&flags.Suffix, "suffix", "", "Text to add at the end of the output file")
	fs.StringVar(&flags.Format, shared.CLIArgFormat, shared.FormatJSON, "Output format (json, markdown, yaml)")
	fs.IntVar(&flags.Concurrency, shared.CLIArgConcurrency, runtime.NumCPU(),
		"Number of concurrent workers (default: number of CPU cores)")
	fs.BoolVar(&flags.NoColors, "no-colors", false, "Disable colored output")
	fs.BoolVar(&flags.NoProgress, "no-progress", false, "Disable progress bars")
	fs.BoolVar(&flags.NoUI, "no-ui", false, "Disable all UI output (implies no-colors and no-progress)")
	fs.BoolVar(&flags.Verbose, "verbose", false, "Shorthand for --log-level info")
	fs.BoolVar(&flags.ShowVersion, "version", false, "Print version information and exit")
	fs.BoolVar(&flags.IncludeSensitive, "include-sensitive", false,
		"Include credential files (.env, *.pem, *.key, id_rsa, .npmrc, .netrc) that are skipped by default")
	fs.StringVar(
		&flags.LogLevel, "log-level", string(shared.LogLevelWarn), "Set log level (debug, info, warn, error)",
	)

	if err := fs.Parse(os.Args[1:]); err != nil {
		return nil, err
	}

	// flag stops at the first non-flag argument, so everything after a stray
	// word (a glob expansion, or a source given without -source) was dropped
	// and the run went on with default format and destination.
	if fs.NArg() > 0 {
		return nil, shared.NewStructuredError(
			shared.ErrorTypeValidation,
			shared.CodeValidationArgument,
			fmt.Sprintf("unexpected argument %q: flags must come before arguments; use -source <dir>", fs.Arg(0)),
			"",
			map[string]any{"arguments": fs.Args()},
		)
	}

	// --verbose only raises the default level; an explicit --log-level wins.
	if flags.Verbose && !isFlagSet(fs, "log-level") {
		flags.LogLevel = string(shared.LogLevelInfo)
	}

	// --version is a terminal action that does not require source/destination validation.
	if flags.ShowVersion {
		flagsParsed = true
		globalFlags = flags
		return flags, nil
	}

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

// isFlagSet reports whether the named flag was given on the command line.
func isFlagSet(fs *flag.FlagSet, name string) bool {
	set := false
	fs.Visit(func(f *flag.Flag) {
		if f.Name == name {
			set = true
		}
	})

	return set
}

// validate validates the CLI flags.
func (f *Flags) validate() error {
	if f.SourceDir == "" {
		return NewCLIMissingSourceError()
	}

	// Validate source path for security
	if err := shared.ValidateSourcePath(f.SourceDir); err != nil {
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
	if !shared.ValidateLogLevel(f.LogLevel) {
		return fmt.Errorf("invalid log level: %s (must be: debug, info, warn, error)", f.LogLevel)
	}

	return nil
}

// setDefaultDestination sets the default destination if not provided.
func (f *Flags) setDefaultDestination() error {
	if f.Destination == "" {
		absRoot, err := shared.AbsolutePath(f.SourceDir)
		if err != nil {
			return fmt.Errorf("getting absolute path: %w", err)
		}
		baseName := shared.BaseName(absRoot)
		f.Destination = baseName + "." + f.Format
	}

	// Validate destination path for security
	if err := shared.ValidateDestinationPath(f.Destination); err != nil {
		return fmt.Errorf("validating destination path: %w", err)
	}

	return nil
}
