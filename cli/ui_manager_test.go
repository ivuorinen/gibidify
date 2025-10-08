package cli

import (
	"bytes"
	"os"
	"testing"

	"github.com/fatih/color"
	"github.com/stretchr/testify/assert"
)

func TestNewUIManager(t *testing.T) {
	// Use t.Setenv for environment isolation

	tests := []struct {
		name             string
		setupEnv         func(t *testing.T)
		expectedColors   bool
		expectedProgress bool
	}{
		{
			name: "default terminal",
			setupEnv: func(t *testing.T) {
				t.Setenv("TERM", "xterm-256color")
				t.Setenv("CI", "")
				t.Setenv("NO_COLOR", "")
				t.Setenv("FORCE_COLOR", "")
			},
			expectedColors:   true,
			expectedProgress: false, // Not a tty in test environment
		},
		{
			name: "dumb terminal",
			setupEnv: func(t *testing.T) {
				t.Setenv("TERM", "dumb")
			},
			expectedColors:   false,
			expectedProgress: false,
		},
		{
			name: "CI environment without GitHub Actions",
			setupEnv: func(t *testing.T) {
				t.Setenv("CI", "true")
				t.Setenv("GITHUB_ACTIONS", "")
			},
			expectedColors:   false,
			expectedProgress: false,
		},
		{
			name: "GitHub Actions CI",
			setupEnv: func(t *testing.T) {
				t.Setenv("TERM", "xterm")
				t.Setenv("CI", "true")
				t.Setenv("GITHUB_ACTIONS", "true")
				t.Setenv("NO_COLOR", "")
			},
			expectedColors:   true,
			expectedProgress: false,
		},
		{
			name: "NO_COLOR set",
			setupEnv: func(t *testing.T) {
				t.Setenv("TERM", "xterm-256color")
				t.Setenv("CI", "")
				t.Setenv("NO_COLOR", "1")
				t.Setenv("FORCE_COLOR", "")
			},
			expectedColors:   false,
			expectedProgress: false,
		},
		{
			name: "FORCE_COLOR set",
			setupEnv: func(t *testing.T) {
				t.Setenv("TERM", "dumb")
				t.Setenv("FORCE_COLOR", "1")
			},
			expectedColors:   true,
			expectedProgress: false,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				tt.setupEnv(t)

				ui := NewUIManager()
				assert.NotNil(t, ui)
				assert.NotNil(t, ui.output)
				assert.Equal(t, tt.expectedColors, ui.enableColors, "color state mismatch")
				assert.Equal(t, tt.expectedProgress, ui.enableProgress, "progress state mismatch")
			},
		)
	}
}

func TestSetColorOutput(t *testing.T) {
	// Capture original color.NoColor state and restore after test
	orig := color.NoColor
	defer func() { color.NoColor = orig }()

	ui := &UIManager{output: os.Stderr}

	// Test enabling colors
	ui.SetColorOutput(true)
	assert.False(t, color.NoColor)
	assert.True(t, ui.enableColors)

	// Test disabling colors
	ui.SetColorOutput(false)
	assert.True(t, color.NoColor)
	assert.False(t, ui.enableColors)
}

func TestSetProgressOutput(t *testing.T) {
	ui := &UIManager{output: os.Stderr}

	// Test enabling progress
	ui.SetProgressOutput(true)
	assert.True(t, ui.enableProgress)

	// Test disabling progress
	ui.SetProgressOutput(false)
	assert.False(t, ui.enableProgress)
}

func TestPrintf(t *testing.T) {
	buf := &bytes.Buffer{}
	ui := &UIManager{
		output: buf,
	}

	ui.printf("Test %s %d", "output", 123)

	assert.Equal(t, "Test output 123", buf.String())
}
