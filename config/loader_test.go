package config_test

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ivuorinen/gibidify/config"
	"github.com/ivuorinen/gibidify/shared"
	"github.com/ivuorinen/gibidify/testutil"
)

const (
	defaultFileSizeLimit = 5242880
	testFileSizeLimit    = 123456
)

// TestDefaultConfig verifies that if no config file is found,
// the default configuration values are correctly set.
func TestDefaultConfig(t *testing.T) {
	// Create a temporary directory to ensure no config file is present, and
	// keep the user's real config out of the search path.
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmpDir)
	t.Setenv("HOME", tmpDir)

	testutil.ResetConfig(t, tmpDir)

	// Check defaults
	defaultSizeLimit := config.FileSizeLimit()
	if defaultSizeLimit != defaultFileSizeLimit {
		t.Errorf("Expected default file size limit of 5242880, got %d", defaultSizeLimit)
	}

	ignoredDirs := config.IgnoredDirectories()
	if len(ignoredDirs) == 0 {
		t.Error("Expected some default ignored directories, got none")
	}
}

// TestLoadConfigFile verifies that when a valid config file is present,
// the store loads the specified values correctly.
func TestLoadConfigFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Prepare a minimal config file
	configContent := []byte(`---
fileSizeLimit: 123456
ignoreDirectories:
- "testdir1"
- "testdir2"
`)

	testutil.CreateTestFile(t, tmpDir, "config.yaml", configContent)

	// Reset the store and point it at the new config path
	isolateStore(t)
	config.AddConfigPath(tmpDir)

	testutil.MustSucceed(t, config.ReadInConfig(), "reading config file")

	// Validate loaded data
	if got := config.GetInt64("fileSizeLimit"); got != testFileSizeLimit {
		t.Errorf("Expected fileSizeLimit=123456, got %d", got)
	}

	ignored := config.GetStringSlice("ignoreDirectories")
	if len(ignored) != 2 || ignored[0] != "testdir1" || ignored[1] != "testdir2" {
		t.Errorf("Expected [\"testdir1\", \"testdir2\"], got %v", ignored)
	}
}

// loadWithConfig writes content as $XDG_CONFIG_HOME/gibidify/config.yaml, runs
// LoadConfig, and returns its error plus everything logged at warn level or
// above. An empty content writes no file.
func loadWithConfig(t *testing.T, content string) (string, error) {
	t.Helper()
	xdg := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", xdg)
	t.Setenv("HOME", t.TempDir())
	if content != "" {
		dir := filepath.Join(xdg, shared.AppName)
		if err := os.MkdirAll(dir, 0o750); err != nil {
			t.Fatal(err)
		}
		testutil.CreateTestFile(t, dir, "config.yaml", []byte(content))
	}

	t.Cleanup(config.Reset)

	var err error
	out := captureWarnings(t, func() {
		config.Reset()
		err = config.LoadConfig()
	})

	return out, err
}

// captureWarnings runs fn with the logger at warn level writing to a buffer.
func captureWarnings(t *testing.T, fn func()) string {
	t.Helper()
	var buf bytes.Buffer
	logger := shared.GetLogger()
	logger.SetLevel(shared.LogLevelWarn)
	origOutput := logger.Output()
	logger.SetOutput(&buf)
	t.Cleanup(func() { logger.SetOutput(origOutput) })
	fn()

	return buf.String()
}

// requireConfigError fails the test unless err is a configuration
// StructuredError whose message contains want.
func requireConfigError(t *testing.T, err error, want string) *shared.StructuredError {
	t.Helper()
	var structErr *shared.StructuredError
	if !errors.As(err, &structErr) {
		t.Fatalf("expected a StructuredError, got %v", err)
	}
	if structErr.Type != shared.ErrorTypeConfiguration {
		t.Errorf("error type = %v, want Configuration", structErr.Type)
	}
	if !strings.Contains(err.Error(), want) {
		t.Errorf("error %q does not contain %q", err, want)
	}

	return structErr
}

// TestLoadConfigInvalidValueFails verifies a config file that fails validation
// is returned as an error naming the file. Falling back to the defaults used to
// drop the user's ignoreDirectories, so excluded files reached the output.
func TestLoadConfigInvalidValueFails(t *testing.T) {
	_, err := loadWithConfig(t, "fileSizeLimit: 100\nignoreDirectories:\n- node_modules\n- \"\"\n")

	structErr := requireConfigError(t, err, "fileSizeLimit (100) is below minimum")
	if !strings.HasSuffix(structErr.FilePath, filepath.Join(shared.AppName, "config.yaml")) {
		t.Errorf("FilePath = %q, want the loaded config file", structErr.FilePath)
	}
}

// TestLoadConfigMalformedFileFails verifies a config file that exists but does
// not parse aborts loading instead of being replaced by defaults.
func TestLoadConfigMalformedFileFails(t *testing.T) {
	_, err := loadWithConfig(t, "ignoreDirectories: [secrets\n")

	requireConfigError(t, err, "parsing config file")
}

// TestLoadConfigMissingFileIsQuiet verifies the no-config case loads the
// defaults and logs nothing at warn level.
func TestLoadConfigMissingFileIsQuiet(t *testing.T) {
	out, err := loadWithConfig(t, "")
	if err != nil {
		t.Fatalf("missing config must not be an error: %v", err)
	}
	if out != "" {
		t.Errorf("expected no warnings without a config file, got: %q", out)
	}
}

// TestLoadConfigAcceptsDotDirectories verifies dot-prefixed directory names are
// valid ignoreDirectories entries and the rest of the list is kept.
func TestLoadConfigAcceptsDotDirectories(t *testing.T) {
	_, err := loadWithConfig(t, "ignoreDirectories:\n  - secrets\n  - .venv\n")
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}

	got := config.IgnoredDirectories()
	if len(got) != 2 || got[0] != "secrets" || got[1] != ".venv" {
		t.Errorf("IgnoredDirectories() = %v, want [secrets .venv]", got)
	}
}

// TestLoadConfigUnknownKeysWarn verifies typos and removed settings are reported.
func TestLoadConfigUnknownKeysWarn(t *testing.T) {
	out, err := loadWithConfig(t, "fileSizeLimit: 2048\nbackpressure:\n  enabled: true\nfileTypes:\n  enabld: true\n")
	if err != nil {
		t.Fatalf("unknown keys must not fail loading: %v", err)
	}

	for _, key := range []string{"backpressure", "fileTypes.enabld"} {
		if !strings.Contains(out, "Unknown configuration key "+key+" ") {
			t.Errorf("expected warning for %s, got: %q", key, out)
		}
	}
	if got := config.FileSizeLimit(); got != 2048 {
		t.Errorf("known keys must still apply: fileSizeLimit = %d, want 2048", got)
	}
}

// TestLoadConfigRejectsWrongTypes verifies a quoted boolean is rejected rather
// than silently read as false.
func TestLoadConfigRejectsWrongTypes(t *testing.T) {
	_, err := loadWithConfig(t, "fileTypes:\n  enabled: \"false\"\n")

	requireConfigError(t, err, "fileTypes.enabled must be a boolean")
}

// TestLoadConfigIgnoresRelativeXDG verifies a relative XDG_CONFIG_HOME is
// ignored (per the XDG spec) and $HOME/.config is used instead.
func TestLoadConfigIgnoresRelativeXDG(t *testing.T) {
	home := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", "relative/dir")
	t.Setenv("HOME", home)
	dir := filepath.Join(home, ".config", shared.AppName)
	if err := os.MkdirAll(dir, 0o750); err != nil {
		t.Fatal(err)
	}
	testutil.CreateTestFile(t, dir, "config.yaml", []byte("fileSizeLimit: 4096\n"))
	t.Cleanup(config.Reset)

	var err error
	out := captureWarnings(t, func() {
		config.Reset()
		err = config.LoadConfig()
	})
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}

	if !strings.Contains(out, "Ignoring relative XDG_CONFIG_HOME") {
		t.Errorf("expected a relative XDG warning, got: %q", out)
	}
	if got := config.FileSizeLimit(); got != 4096 {
		t.Errorf("fileSizeLimit = %d, want 4096 from $HOME/.config", got)
	}
}
