package cli

import (
	"bytes"
	"os"
	"testing"

	"github.com/fatih/color"
	"github.com/stretchr/testify/assert"
)

func TestNewUIManager(t *testing.T) {
	tests := []struct {
		name             string
		env              terminalEnvSetup
		expectedColors   bool
		expectedProgress bool
	}{
		{
			name:             "default terminal",
			env:              envDefaultTerminal,
			expectedColors:   true,
			expectedProgress: false, // Not a tty in test environment
		},
		{
			name:             "dumb terminal",
			env:              envDumbTerminal,
			expectedColors:   false,
			expectedProgress: false,
		},
		{
			name:             "CI environment without GitHub Actions",
			env:              envCIWithoutGitHub,
			expectedColors:   false,
			expectedProgress: false,
		},
		{
			name:             "GitHub Actions CI",
			env:              envGitHubActions,
			expectedColors:   true,
			expectedProgress: false,
		},
		{
			name:             "NO_COLOR set",
			env:              envNoColor,
			expectedColors:   false,
			expectedProgress: false,
		},
		{
			name:             "FORCE_COLOR set",
			env:              envForceColor,
			expectedColors:   true,
			expectedProgress: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.env.apply(t)

			ui := NewUIManager()
			assert.NotNil(t, ui)
			assert.NotNil(t, ui.output)
			assert.Equal(t, tt.expectedColors, ui.enableColors, "color state mismatch")
			assert.Equal(t, tt.expectedProgress, ui.enableProgress, "progress state mismatch")
		})
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
