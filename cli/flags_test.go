package cli

import (
	"flag"
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/ivuorinen/gibidify/testutil"
)

const testDirPlaceholder = "testdir"

// setupTestArgs prepares test arguments by replacing testdir with actual temp directory.
func setupTestArgs(t *testing.T, args []string, want *Flags) ([]string, *Flags) {
	t.Helper()

	if !containsFlag(args, "-source") {
		return args, want
	}

	tempDir := t.TempDir()
	modifiedArgs := replaceTestDirInArgs(args, tempDir)

	// Handle nil want parameter (used for error test cases)
	if want == nil {
		return modifiedArgs, nil
	}

	modifiedWant := updateWantFlags(*want, tempDir)

	return modifiedArgs, &modifiedWant
}

// replaceTestDirInArgs replaces testdir placeholder with actual temp directory in args.
func replaceTestDirInArgs(args []string, tempDir string) []string {
	modifiedArgs := make([]string, len(args))
	copy(modifiedArgs, args)

	for i, arg := range modifiedArgs {
		if arg == testDirPlaceholder {
			modifiedArgs[i] = tempDir

			break
		}
	}

	return modifiedArgs
}

// updateWantFlags updates the want flags with temp directory replacements.
func updateWantFlags(want Flags, tempDir string) Flags {
	modifiedWant := want

	if want.SourceDir == testDirPlaceholder {
		modifiedWant.SourceDir = tempDir
		if strings.HasPrefix(want.Destination, testDirPlaceholder+".") {
			baseName := testutil.GetBaseName(tempDir)
			modifiedWant.Destination = baseName + "." + want.Format
		}
	}

	return modifiedWant
}

// runParseFlagsTest runs a single parse flags test.
func runParseFlagsTest(t *testing.T, args []string, want *Flags, wantErr bool, errContains string) {
	t.Helper()

	resetFlagsState()
	modifiedArgs, modifiedWant := setupTestArgs(t, args, want)
	setupCommandLineArgs(modifiedArgs)

	got, err := ParseFlags()

	if wantErr {
		if err == nil {
			t.Error("ParseFlags() expected error, got nil")

			return
		}
		if errContains != "" && !strings.Contains(err.Error(), errContains) {
			t.Errorf("ParseFlags() error = %v, want error containing %v", err, errContains)
		}

		return
	}

	if err != nil {
		t.Errorf("ParseFlags() unexpected error = %v", err)

		return
	}

	verifyFlags(t, got, modifiedWant)
}

func TestParseFlags(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		want        *Flags
		wantErr     bool
		errContains string
	}{
		{
			name: "valid basic flags",
			args: []string{"-source", "testdir", "-format", "markdown"},
			want: &Flags{
				SourceDir:   "testdir",
				Format:      "markdown",
				Concurrency: runtime.NumCPU(),
				Destination: "testdir.markdown",
			},
			wantErr: false,
		},
		{
			name: "valid with all flags",
			args: []string{
				"-source", "testdir",
				"-destination", "output.md",
				"-prefix", "# Header",
				"-suffix", "# Footer",
				"-format", "json",
				"-concurrency", "4",
				"-verbose",
				"-no-colors",
				"-no-progress",
			},
			want: &Flags{
				SourceDir:   "testdir",
				Destination: "output.md",
				Prefix:      "# Header",
				Suffix:      "# Footer",
				Format:      "json",
				Concurrency: 4,
				Verbose:     true,
				NoColors:    true,
				NoProgress:  true,
			},
			wantErr: false,
		},
		{
			name:        "missing source directory",
			args:        []string{"-format", "markdown"},
			wantErr:     true,
			errContains: "source directory is required",
		},
		{
			name:        "invalid format",
			args:        []string{"-source", "testdir", "-format", "invalid"},
			wantErr:     true,
			errContains: "validating output format",
		},
		{
			name:        "invalid concurrency zero",
			args:        []string{"-source", "testdir", "-concurrency", "0"},
			wantErr:     true,
			errContains: "validating concurrency",
		},
		{
			name:        "negative concurrency",
			args:        []string{"-source", "testdir", "-concurrency", "-1"},
			wantErr:     true,
			errContains: "validating concurrency",
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				runParseFlagsTest(t, tt.args, tt.want, tt.wantErr, tt.errContains)
			},
		)
	}
}

// validateFlagsValidationResult validates flag validation test results.
func validateFlagsValidationResult(t *testing.T, err error, wantErr bool, errContains string) {
	t.Helper()

	if wantErr {
		if err == nil {
			t.Error("Flags.validate() expected error, got nil")

			return
		}
		if errContains != "" && !strings.Contains(err.Error(), errContains) {
			t.Errorf("Flags.validate() error = %v, want error containing %v", err, errContains)
		}

		return
	}

	if err != nil {
		t.Errorf("Flags.validate() unexpected error = %v", err)
	}
}

func TestFlags_validate(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name        string
		flags       *Flags
		wantErr     bool
		errContains string
	}{
		{
			name: "valid flags",
			flags: &Flags{
				SourceDir:   tempDir,
				Format:      "markdown",
				Concurrency: 4,
				LogLevel:    "warn",
			},
			wantErr: false,
		},
		{
			name: "empty source directory",
			flags: &Flags{
				Format:      "markdown",
				Concurrency: 4,
				LogLevel:    "warn",
			},
			wantErr:     true,
			errContains: "source directory is required",
		},
		{
			name: "invalid format",
			flags: &Flags{
				SourceDir:   tempDir,
				Format:      "invalid",
				Concurrency: 4,
				LogLevel:    "warn",
			},
			wantErr:     true,
			errContains: "validating output format",
		},
		{
			name: "zero concurrency",
			flags: &Flags{
				SourceDir:   tempDir,
				Format:      "markdown",
				Concurrency: 0,
				LogLevel:    "warn",
			},
			wantErr:     true,
			errContains: "validating concurrency",
		},
		{
			name: "negative concurrency",
			flags: &Flags{
				SourceDir:   tempDir,
				Format:      "json",
				Concurrency: -1,
				LogLevel:    "warn",
			},
			wantErr:     true,
			errContains: "validating concurrency",
		},
		{
			name: "invalid log level",
			flags: &Flags{
				SourceDir:   tempDir,
				Format:      "json",
				Concurrency: 4,
				LogLevel:    "invalid",
			},
			wantErr:     true,
			errContains: "invalid log level",
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				err := tt.flags.validate()
				validateFlagsValidationResult(t, err, tt.wantErr, tt.errContains)
			},
		)
	}
}

// validateDefaultDestinationResult validates default destination test results.
func validateDefaultDestinationResult(
	t *testing.T,
	flags *Flags,
	err error,
	wantDestination string,
	wantErr bool,
	errContains string,
) {
	t.Helper()

	if wantErr {
		if err == nil {
			t.Error("Flags.setDefaultDestination() expected error, got nil")

			return
		}
		if errContains != "" && !strings.Contains(err.Error(), errContains) {
			t.Errorf("Flags.setDefaultDestination() error = %v, want error containing %v", err, errContains)
		}

		return
	}

	if err != nil {
		t.Errorf("Flags.setDefaultDestination() unexpected error = %v", err)

		return
	}

	if flags.Destination != wantDestination {
		t.Errorf("Flags.Destination = %v, want %v", flags.Destination, wantDestination)
	}
}

func TestFlags_setDefaultDestination(t *testing.T) {
	tempDir := t.TempDir()
	baseName := testutil.GetBaseName(tempDir)

	tests := []struct {
		name            string
		flags           *Flags
		wantDestination string
		wantErr         bool
		errContains     string
	}{
		{
			name: "set default destination markdown",
			flags: &Flags{
				SourceDir: tempDir,
				Format:    "markdown",
				LogLevel:  "warn",
			},
			wantDestination: baseName + ".markdown",
			wantErr:         false,
		},
		{
			name: "set default destination json",
			flags: &Flags{
				SourceDir: tempDir,
				Format:    "json",
				LogLevel:  "warn",
			},
			wantDestination: baseName + ".json",
			wantErr:         false,
		},
		{
			name: "preserve existing destination",
			flags: &Flags{
				SourceDir:   tempDir,
				Format:      "yaml",
				Destination: "custom-output.yaml",
				LogLevel:    "warn",
			},
			wantDestination: "custom-output.yaml",
			wantErr:         false,
		},
		{
			name: "nonexistent source path still generates destination",
			flags: &Flags{
				SourceDir: "/nonexistent/path/that/should/not/exist",
				Format:    "markdown",
				LogLevel:  "warn",
			},
			wantDestination: "exist.markdown", // Based on filepath.Base of the path
			wantErr:         false,            // GetAbsolutePath doesn't validate existence, only converts to absolute
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				err := tt.flags.setDefaultDestination()
				validateDefaultDestinationResult(t, tt.flags, err, tt.wantDestination, tt.wantErr, tt.errContains)
			},
		)
	}
}

func TestParseFlagsSingleton(t *testing.T) {
	resetFlagsState()
	tempDir := t.TempDir()

	// First call
	setupCommandLineArgs([]string{"-source", tempDir, "-format", "markdown"})
	flags1, err := ParseFlags()
	if err != nil {
		t.Fatalf("First ParseFlags() failed: %v", err)
	}

	// Second call should return the same instance
	flags2, err := ParseFlags()
	if err != nil {
		t.Fatalf("Second ParseFlags() failed: %v", err)
	}

	if flags1 != flags2 {
		t.Error("ParseFlags() should return singleton instance, got different pointers")
	}
}

// Helper functions

// resetFlagsState resets the global flags state for testing.
func resetFlagsState() {
	flagsParsed = false
	globalFlags = nil
	// Reset the flag.CommandLine for clean testing
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
}

// setupCommandLineArgs sets up command line arguments for testing.
func setupCommandLineArgs(args []string) {
	os.Args = append([]string{"gibidify"}, args...)
}

// containsFlag checks if a flag is present in the arguments.
func containsFlag(args []string, flagName string) bool {
	for _, arg := range args {
		if arg == flagName {
			return true
		}
	}

	return false
}

// verifyFlags compares two Flags structs for testing.
func verifyFlags(t *testing.T, got, want *Flags) {
	t.Helper()

	if got.SourceDir != want.SourceDir {
		t.Errorf("SourceDir = %v, want %v", got.SourceDir, want.SourceDir)
	}
	if got.Destination != want.Destination {
		t.Errorf("Destination = %v, want %v", got.Destination, want.Destination)
	}
	if got.Prefix != want.Prefix {
		t.Errorf("Prefix = %v, want %v", got.Prefix, want.Prefix)
	}
	if got.Suffix != want.Suffix {
		t.Errorf("Suffix = %v, want %v", got.Suffix, want.Suffix)
	}
	if got.Format != want.Format {
		t.Errorf("Format = %v, want %v", got.Format, want.Format)
	}
	if got.Concurrency != want.Concurrency {
		t.Errorf("Concurrency = %v, want %v", got.Concurrency, want.Concurrency)
	}
	if got.NoColors != want.NoColors {
		t.Errorf("NoColors = %v, want %v", got.NoColors, want.NoColors)
	}
	if got.NoProgress != want.NoProgress {
		t.Errorf("NoProgress = %v, want %v", got.NoProgress, want.NoProgress)
	}
	if got.Verbose != want.Verbose {
		t.Errorf("Verbose = %v, want %v", got.Verbose, want.Verbose)
	}
}
