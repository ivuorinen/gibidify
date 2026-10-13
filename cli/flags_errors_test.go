package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ivuorinen/gibidify/shared"
	"github.com/ivuorinen/gibidify/testutil"
)

// parseArgs runs ParseFlags on args with fresh parse state, restoring os.Args
// and the parse state when the test ends.
func parseArgs(t *testing.T, args ...string) (*Flags, error) {
	t.Helper()
	origArgs := os.Args
	t.Cleanup(func() {
		os.Args = origArgs
		resetFlagsState()
	})
	resetFlagsState()
	setupCommandLineArgs(args)

	return ParseFlags()
}

// removedWorkingDir makes the test's working directory a directory that no
// longer exists, so resolving a relative path fails.
func removedWorkingDir(t *testing.T) {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "gone")
	testutil.MustSucceed(t, os.Mkdir(dir, 0o750), "creating working directory")
	t.Chdir(dir)
	testutil.MustSucceed(t, os.Remove(dir), "removing working directory")
}

// TestParseFlagsRejectsInvalidInput verifies ParseFlags fails, and stays
// unparsed, for input the flag package, the source check or the destination
// check rejects.
func TestParseFlagsRejectsInvalidInput(t *testing.T) {
	src := t.TempDir()
	file := testutil.CreateTestFile(t, src, "file.txt", []byte("x"))

	tests := []struct {
		name        string
		args        []string
		errContains string
	}{
		{"unknown flag", []string{"-bogus"}, "flag provided but not defined"},
		{"non-numeric concurrency", []string{testutil.TestCLIFlagConcurrency, "many"}, "invalid value"},
		{
			"missing source",
			[]string{testutil.TestCLIFlagSource, filepath.Join(src, "missing")},
			"validating source path",
		},
		{"source is a file", []string{testutil.TestCLIFlagSource, file}, "validating source path"},
		{
			"destination is a directory",
			[]string{testutil.TestCLIFlagSource, src, testutil.TestCLIFlagDestination, t.TempDir()},
			"validating destination path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			redirectStderrToPipe(t)
			got, err := parseArgs(t, tt.args...)
			testutil.AssertErrorContains(t, err, tt.errContains, "ParseFlags")
			if got != nil || flagsParsed {
				t.Errorf("ParseFlags() = %+v, parsed = %v; want nil and unparsed after an error", got, flagsParsed)
			}
		})
	}
}

// TestParseFlagsVersionSkipsValidation verifies -version parses without a
// source and leaves the destination unset, and that the result is cached.
func TestParseFlagsVersionSkipsValidation(t *testing.T) {
	got, err := parseArgs(t, "-version")
	testutil.AssertNoError(t, err, "ParseFlags")
	if got == nil || !got.ShowVersion {
		t.Fatalf("ParseFlags() = %+v, want ShowVersion", got)
	}
	if got.SourceDir != "" || got.Destination != "" {
		t.Errorf("SourceDir = %q, Destination = %q; want both empty", got.SourceDir, got.Destination)
	}

	again, err := ParseFlags()
	testutil.AssertNoError(t, err, "second ParseFlags")
	if again != got {
		t.Error("second ParseFlags() did not return the cached flags")
	}
}

// TestFlagsValidateRejectsBadSource verifies validate reports the source path
// check before checking any other flag.
func TestFlagsValidateRejectsBadSource(t *testing.T) {
	file := testutil.CreateTestFile(t, t.TempDir(), "file.txt", []byte("x"))

	for name, src := range map[string]string{"missing": filepath.Join(t.TempDir(), "missing"), "file": file} {
		t.Run(name, func(t *testing.T) {
			flags := &Flags{SourceDir: src, Format: "invalid", Concurrency: 0, LogLevel: "invalid"}
			testutil.AssertErrorContains(t, flags.validate(), "validating source path", "validate")
		})
	}
}

// TestFlagsSetDefaultDestinationErrors verifies setDefaultDestination fails
// when the source cannot be resolved or the destination is a directory.
func TestFlagsSetDefaultDestinationErrors(t *testing.T) {
	t.Run("relative source without working directory", func(t *testing.T) {
		removedWorkingDir(t)
		flags := &Flags{SourceDir: "src", Format: shared.FormatJSON}
		testutil.AssertErrorContains(t, flags.setDefaultDestination(), "getting absolute path", "setDefaultDestination")
		if flags.Destination != "" {
			t.Errorf("Destination = %q, want it left unset", flags.Destination)
		}
	})

	t.Run("destination is a directory", func(t *testing.T) {
		flags := &Flags{SourceDir: t.TempDir(), Format: shared.FormatJSON, Destination: t.TempDir()}
		err := flags.setDefaultDestination()
		testutil.AssertErrorContains(t, err, "validating destination path", "setDefaultDestination")
	})
}
