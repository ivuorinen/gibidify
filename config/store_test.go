package config_test

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/ivuorinen/gibidify/config"
	"github.com/ivuorinen/gibidify/shared"
)

// isolateStore resets the config store and resets it again when t ends, so no
// explicit value set by the test leaks into later tests.
func isolateStore(t *testing.T) {
	t.Helper()
	config.Reset()
	t.Cleanup(config.Reset)
}

// numericGetterCases covers every value shape the numeric getters convert.
var numericGetterCases = []struct {
	name      string
	value     any
	wantInt   int
	wantInt64 int64
}{
	{"int", 42, 42, 42},
	{"int64", int64(43), 43, 43},
	{"whole float64 as decoded from 1e3", 1e3, 1000, 1000},
	{"fractional float64 truncates", 7.9, 7, 7},
	{"string is not converted", "12", 0, 0},
	{"bool is not converted", true, 0, 0},
}

// TestNumericGetterConversions verifies GetInt and GetInt64 convert the int,
// int64 and float64 shapes a YAML decoder or Set can produce, and return zero
// for any other type.
func TestNumericGetterConversions(t *testing.T) {
	const key = "numeric"
	for _, tt := range numericGetterCases {
		t.Run(tt.name, func(t *testing.T) {
			isolateStore(t)
			config.Set(key, tt.value)

			if got := config.GetInt(key); got != tt.wantInt {
				t.Errorf("GetInt(%v) = %d, want %d", tt.value, got, tt.wantInt)
			}
			if got := config.GetInt64(key); got != tt.wantInt64 {
				t.Errorf("GetInt64(%v) = %d, want %d", tt.value, got, tt.wantInt64)
			}
		})
	}
}

// TestGetStringSliceConversions verifies GetStringSlice accepts []string and
// []any (stringifying elements) and returns nil for any other type.
func TestGetStringSliceConversions(t *testing.T) {
	const key = "list"
	tests := []struct {
		name  string
		value any
		want  []string
	}{
		{"string slice", []string{"a", "b"}, []string{"a", "b"}},
		{"any slice stringified", []any{"a", 1, true}, []string{"a", "1", "true"}},
		{"scalar string", "a", nil},
		{"map", map[string]any{"a": "b"}, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isolateStore(t)
			config.Set(key, tt.value)

			if got := config.GetStringSlice(key); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetStringSlice(%#v) = %#v, want %#v", tt.value, got, tt.want)
			}
		})
	}
}

// TestGetStringMapStringConversions verifies GetStringMapString accepts
// map[string]string and map[string]any (stringifying values), and returns an
// empty, non-nil map for any other type so callers can range or index it.
func TestGetStringMapStringConversions(t *testing.T) {
	const key = "mapping"
	tests := []struct {
		name  string
		value any
		want  map[string]string
	}{
		{"string map", map[string]string{".x": "x"}, map[string]string{".x": "x"}},
		{"any map stringified", map[string]any{".x": "x", ".n": 5}, map[string]string{".x": "x", ".n": "5"}},
		{"scalar", "x", map[string]string{}},
		{"list", []any{"x"}, map[string]string{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isolateStore(t)
			config.Set(key, tt.value)

			got := config.GetStringMapString(key)
			if got == nil {
				t.Fatalf("GetStringMapString(%#v) returned nil", tt.value)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetStringMapString(%#v) = %#v, want %#v", tt.value, got, tt.want)
			}
		})
	}
}

// TestNestedLookupThroughScalar verifies that a scalar explicit value sitting
// where a section is expected does not satisfy a nested key: IsSet reports
// false and the lookup falls back to the registered default.
func TestNestedLookupThroughScalar(t *testing.T) {
	isolateStore(t)
	config.SetDefault("section.limit", 7)
	config.Set("section", "not-a-map")

	if config.IsSet("section.limit") {
		t.Error("IsSet(section.limit) = true, want false when section is a scalar")
	}
	if got := config.GetInt("section.limit"); got != 7 {
		t.Errorf("GetInt(section.limit) = %d, want default 7", got)
	}
}

// TestReadInConfigUnreadableFile verifies a config.yaml that exists but cannot
// be read is reported as a read error, not as a missing file, and that the
// search stops there instead of trying later paths.
func TestReadInConfigUnreadableFile(t *testing.T) {
	isolateStore(t)
	blocked := t.TempDir()
	// A directory named config.yaml exists but cannot be read as a file.
	if err := os.Mkdir(filepath.Join(blocked, "config."+shared.FormatYAML), 0o750); err != nil {
		t.Fatal(err)
	}
	fallback := t.TempDir()
	if err := os.WriteFile(filepath.Join(fallback, "config.yaml"), []byte("fileSizeLimit: 2048\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	config.AddConfigPath(blocked)
	config.AddConfigPath(fallback)

	err := config.ReadInConfig()
	if err == nil {
		t.Fatal("expected an error for an unreadable config file")
	}
	assertReadError(t, err)
	if config.FileUsed() != "" {
		t.Errorf("FileUsed() = %q, want empty after a read error", config.FileUsed())
	}
}

// assertReadError checks err describes a read failure rather than a missing or
// unparsable file.
func assertReadError(t *testing.T, err error) {
	t.Helper()
	msg := err.Error()
	if !strings.Contains(msg, "reading config file") || !strings.Contains(msg, "config.yaml") {
		t.Errorf("error %q does not describe the read failure", msg)
	}
	if strings.Contains(msg, "no config file found") {
		t.Errorf("read failure misreported as a missing file: %q", msg)
	}
	var pathErr *os.PathError
	if !errors.As(err, &pathErr) {
		t.Errorf("error %v does not wrap the underlying *os.PathError", err)
	}
}

// TestLoadConfigUnreadableFileFails verifies LoadConfig surfaces a config file
// that cannot be read instead of silently applying the defaults.
func TestLoadConfigUnreadableFileFails(t *testing.T) {
	xdg := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", xdg)
	t.Setenv("HOME", t.TempDir())
	if err := os.MkdirAll(filepath.Join(xdg, shared.AppName, "config.yaml"), 0o750); err != nil {
		t.Fatal(err)
	}
	isolateStore(t)

	err := config.LoadConfig()

	structErr := requireConfigError(t, err, "cannot load config file")
	if structErr.Code != shared.CodeConfigValidation {
		t.Errorf("error code = %q, want %q", structErr.Code, shared.CodeConfigValidation)
	}
	assertReadError(t, err)
}
