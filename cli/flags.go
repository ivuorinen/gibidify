package cli

import (
	"flag"
	"runtime"

	"github.com/ivuorinen/gibidify/config"
	"github.com/ivuorinen/gibidify/gibidiutils"
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
}

var (
	flagsParsed bool
	globalFlags *Flags
)

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
	flag.StringVar(&flags.Format, "format", "markdown", "Output format (json, markdown, yaml)")
	flag.IntVar(
		&flags.Concurrency, "concurrency", runtime.NumCPU(),
		"Number of concurrent workers (default: number of CPU cores)",
	)
	flag.BoolVar(&flags.NoColors, "no-colors", false, "Disable colored output")
	flag.BoolVar(&flags.NoProgress, "no-progress", false, "Disable progress bars")
	flag.BoolVar(&flags.Verbose, "verbose", false, "Enable verbose output")

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
		return NewMissingSourceError()
	}

	// Validate source path for security
	if err := gibidiutils.ValidateSourcePath(f.SourceDir); err != nil {
		return err
	}

	// Validate output format
	if err := config.ValidateOutputFormat(f.Format); err != nil {
		return err
	}

	// Validate concurrency
	return config.ValidateConcurrency(f.Concurrency)
}

// setDefaultDestination sets the default destination if not provided.
func (f *Flags) setDefaultDestination() error {
	if f.Destination == "" {
		absRoot, err := gibidiutils.GetAbsolutePath(f.SourceDir)
		if err != nil {
			return err
		}
		baseName := gibidiutils.GetBaseName(absRoot)
		f.Destination = baseName + "." + f.Format
	}

	// Validate destination path for security
	return gibidiutils.ValidateDestinationPath(f.Destination)
}
