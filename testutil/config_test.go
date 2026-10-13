package testutil

import (
	"os"
	"strconv"
	"testing"

	"github.com/ivuorinen/gibidify/config"
	"github.com/ivuorinen/gibidify/shared"
)

func TestResetConfig(t *testing.T) {
	tests := []struct {
		name       string
		configPath string
		preSetup   func()
		verify     func(t *testing.T)
	}{
		{
			name:       "reset with empty config path",
			configPath: "",
			preSetup: func() {
				config.Set(TestKeyName, "value")
			},
			verify: func(t *testing.T) {
				t.Helper()
				if config.IsSet(TestKeyName) {
					t.Error("config store not reset properly")
				}
			},
		},
		{
			name:       "reset with config path",
			configPath: t.TempDir(),
			preSetup: func() {
				config.Set(TestKeyName, "value")
			},
			verify: func(t *testing.T) {
				t.Helper()
				if config.IsSet(TestKeyName) {
					t.Error("config store not reset properly")
				}
				// Verify config path was added
				paths := config.FileUsed()
				if paths == "" {
					// This is expected as no config file exists
					return
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				tt.preSetup()
				ResetConfig(t, tt.configPath)
				tt.verify(t)
			},
		)
	}
}

func TestSetupCLIArgs(t *testing.T) {
	// Save original args
	originalArgs := os.Args
	defer func() {
		os.Args = originalArgs
	}()

	tests := []struct {
		name        string
		srcDir      string
		outFile     string
		prefix      string
		suffix      string
		concurrency int
		wantLen     int
	}{
		{
			name:        "basic CLI args",
			srcDir:      "/src",
			outFile:     "/out.txt",
			prefix:      "PREFIX",
			suffix:      "SUFFIX",
			concurrency: 4,
			wantLen:     12,
		},
		{
			name:        "empty strings",
			srcDir:      "",
			outFile:     "",
			prefix:      "",
			suffix:      "",
			concurrency: 1,
			wantLen:     12,
		},
		{
			name:        "special characters in args",
			srcDir:      "/path with spaces/src",
			outFile:     "/path/to/output file.txt",
			prefix:      "Prefix with\nnewline",
			suffix:      "Suffix with\ttab",
			concurrency: 8,
			wantLen:     12,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				SetupCLIArgs(tt.srcDir, tt.outFile, tt.prefix, tt.suffix, tt.concurrency)
				verifySetupCLIArgs(t, tt.srcDir, tt.outFile, tt.prefix, tt.suffix, tt.concurrency, tt.wantLen)
			},
		)
	}
}

// verifySetupCLIArgs verifies that CLI arguments are set correctly.
func verifySetupCLIArgs(t *testing.T, srcDir, outFile, prefix, suffix string, concurrency, wantLen int) {
	t.Helper()

	if len(os.Args) != wantLen {
		t.Errorf("os.Args length = %d, want %d", len(os.Args), wantLen)
	}

	// Verify specific args
	if os.Args[0] != "gibidify" {
		t.Errorf("Program name = %s, want gibidify", os.Args[0])
	}
	if os.Args[2] != srcDir {
		t.Errorf("Source dir = %s, want %s", os.Args[2], srcDir)
	}
	if os.Args[4] != outFile {
		t.Errorf("Output file = %s, want %s", os.Args[4], outFile)
	}
	if os.Args[6] != prefix {
		t.Errorf("Prefix = %s, want %s", os.Args[6], prefix)
	}
	if os.Args[8] != suffix {
		t.Errorf("Suffix = %s, want %s", os.Args[8], suffix)
	}
	if os.Args[10] != strconv.Itoa(concurrency) {
		t.Errorf("Concurrency = %s, want %d", os.Args[10], concurrency)
	}

	// Verify the -no-ui flag is present
	if os.Args[11] != "-no-ui" {
		t.Errorf("NoUI flag = %s, want -no-ui", os.Args[11])
	}
}

// TestSetConfigKeys verifies SetConfigKeys drops earlier values, keeps the
// defaults and applies the explicit keys on top.
func TestSetConfigKeys(t *testing.T) {
	t.Cleanup(config.Reset)

	config.Reset()
	config.SetDefaultConfig()
	defaultMaxFiles := config.MaxFiles()

	config.Set(TestKeyName, "value")
	SetConfigKeys(t, map[string]any{
		shared.ConfigKeyFileSizeLimit:  int64(2048),
		shared.ConfigKeyMaxConcurrency: 3,
	})

	if config.IsSet(TestKeyName) {
		t.Errorf("%s survived SetConfigKeys", TestKeyName)
	}
	if got := config.FileSizeLimit(); got != 2048 {
		t.Errorf("FileSizeLimit() = %d, want 2048", got)
	}
	if got := config.MaxConcurrency(); got != 3 {
		t.Errorf("MaxConcurrency() = %d, want 3", got)
	}
	if got := config.MaxFiles(); got != defaultMaxFiles {
		t.Errorf("MaxFiles() = %d, want default %d", got, defaultMaxFiles)
	}
}
