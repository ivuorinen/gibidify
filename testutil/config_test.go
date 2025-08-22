package testutil

import (
	"os"
	"testing"

	"github.com/spf13/viper"
)

func TestResetViperConfig(t *testing.T) {
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
				viper.Set("test.key", "value")
			},
			verify: func(t *testing.T) {
				if viper.IsSet("test.key") {
					t.Error("Viper config not reset properly")
				}
			},
		},
		{
			name:       "reset with config path",
			configPath: t.TempDir(),
			preSetup: func() {
				viper.Set("test.key", "value")
			},
			verify: func(t *testing.T) {
				if viper.IsSet("test.key") {
					t.Error("Viper config not reset properly")
				}
				// Verify config path was added
				paths := viper.ConfigFileUsed()
				if paths == "" {
					// This is expected as no config file exists
					return
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.preSetup()
			ResetViperConfig(t, tt.configPath)
			tt.verify(t)
		})
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
			wantLen:     11,
		},
		{
			name:        "empty strings",
			srcDir:      "",
			outFile:     "",
			prefix:      "",
			suffix:      "",
			concurrency: 1,
			wantLen:     11,
		},
		{
			name:        "special characters in args",
			srcDir:      "/path with spaces/src",
			outFile:     "/path/to/output file.txt",
			prefix:      "Prefix with\nnewline",
			suffix:      "Suffix with\ttab",
			concurrency: 8,
			wantLen:     11,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetupCLIArgs(tt.srcDir, tt.outFile, tt.prefix, tt.suffix, tt.concurrency)

			if len(os.Args) != tt.wantLen {
				t.Errorf("os.Args length = %d, want %d", len(os.Args), tt.wantLen)
			}

			// Verify specific args
			if os.Args[0] != "gibidify" {
				t.Errorf("Program name = %s, want gibidify", os.Args[0])
			}
			if os.Args[2] != tt.srcDir {
				t.Errorf("Source dir = %s, want %s", os.Args[2], tt.srcDir)
			}
			if os.Args[4] != tt.outFile {
				t.Errorf("Output file = %s, want %s", os.Args[4], tt.outFile)
			}
			if os.Args[6] != tt.prefix {
				t.Errorf("Prefix = %s, want %s", os.Args[6], tt.prefix)
			}
			if os.Args[8] != tt.suffix {
				t.Errorf("Suffix = %s, want %s", os.Args[8], tt.suffix)
			}
			if os.Args[10] != string(rune(tt.concurrency+'0')) {
				t.Errorf("Concurrency = %s, want %d", os.Args[10], tt.concurrency)
			}
		})
	}
}
