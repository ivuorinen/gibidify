package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/fatih/color"
	"github.com/stretchr/testify/assert"

	"github.com/ivuorinen/gibidify/gibidiutils"
)

func TestPrintSuccess(t *testing.T) {
	tests := []struct {
		name         string
		enableColors bool
		format       string
		args         []interface{}
		expectSymbol string
	}{
		{
			name:         "with colors",
			enableColors: true,
			format:       "Operation %s",
			args:         []interface{}{"completed"},
			expectSymbol: gibidiutils.IconSuccess,
		},
		{
			name:         "without colors",
			enableColors: false,
			format:       "Operation %s",
			args:         []interface{}{"completed"},
			expectSymbol: gibidiutils.IconSuccess,
		},
		{
			name:         "no arguments",
			enableColors: true,
			format:       "Success",
			args:         nil,
			expectSymbol: gibidiutils.IconSuccess,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				buf := &bytes.Buffer{}
				ui := &UIManager{
					enableColors: tt.enableColors,
					output:       buf,
				}
				prev := color.NoColor
				color.NoColor = !tt.enableColors
				defer func() { color.NoColor = prev }()

				ui.PrintSuccess(tt.format, tt.args...)

				output := buf.String()
				assert.Contains(t, output, tt.expectSymbol)
				if len(tt.args) > 0 {
					assert.Contains(t, output, "completed")
				}
			},
		)
	}
}

func TestPrintError(t *testing.T) {
	tests := []struct {
		name         string
		enableColors bool
		format       string
		args         []interface{}
		expectSymbol string
	}{
		{
			name:         "with colors",
			enableColors: true,
			format:       "Failed to %s",
			args:         []interface{}{"process"},
			expectSymbol: gibidiutils.IconError,
		},
		{
			name:         "without colors",
			enableColors: false,
			format:       "Failed to %s",
			args:         []interface{}{"process"},
			expectSymbol: gibidiutils.IconError,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				buf := &bytes.Buffer{}
				ui := &UIManager{
					enableColors: tt.enableColors,
					output:       buf,
				}
				prev := color.NoColor
				color.NoColor = !tt.enableColors
				defer func() { color.NoColor = prev }()

				ui.PrintError(tt.format, tt.args...)

				output := buf.String()
				assert.Contains(t, output, tt.expectSymbol)
				if len(tt.args) > 0 {
					assert.Contains(t, output, "process")
				}
			},
		)
	}
}

func TestPrintWarning(t *testing.T) {
	buf := &bytes.Buffer{}
	ui := &UIManager{
		enableColors: true,
		output:       buf,
	}

	ui.PrintWarning("This is a %s", "warning")

	output := buf.String()
	assert.Contains(t, output, gibidiutils.IconWarning)
}

func TestPrintInfo(t *testing.T) {
	// Capture original color.NoColor state and restore after test
	orig := color.NoColor
	defer func() { color.NoColor = orig }()

	buf := &bytes.Buffer{}
	ui := &UIManager{
		enableColors: true,
		output:       buf,
	}

	color.NoColor = false

	ui.PrintInfo("Information: %d items", 42)

	output := buf.String()
	assert.Contains(t, output, gibidiutils.IconInfo)
	assert.Contains(t, output, "42")
}

func TestPrintHeader(t *testing.T) {
	tests := []struct {
		name         string
		enableColors bool
		format       string
		args         []interface{}
	}{
		{
			name:         "with colors",
			enableColors: true,
			format:       "Header %s",
			args:         []interface{}{"Title"},
		},
		{
			name:         "without colors",
			enableColors: false,
			format:       "Header %s",
			args:         []interface{}{"Title"},
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				// Capture original color.NoColor state and restore after test
				orig := color.NoColor
				defer func() { color.NoColor = orig }()

				buf := &bytes.Buffer{}
				ui := &UIManager{
					enableColors: tt.enableColors,
					output:       buf,
				}
				color.NoColor = !tt.enableColors

				ui.PrintHeader(tt.format, tt.args...)

				output := buf.String()
				assert.Contains(t, output, "Title")
			},
		)
	}
}

// Test that all print methods handle newlines correctly
func TestPrintMethodsNewlines(t *testing.T) {
	tests := []struct {
		name   string
		method func(*UIManager, string, ...interface{})
		symbol string
	}{
		{
			name:   "PrintSuccess",
			method: (*UIManager).PrintSuccess,
			symbol: gibidiutils.IconSuccess,
		},
		{
			name:   "PrintError",
			method: (*UIManager).PrintError,
			symbol: gibidiutils.IconError,
		},
		{
			name:   "PrintWarning",
			method: (*UIManager).PrintWarning,
			symbol: gibidiutils.IconWarning,
		},
		{
			name:   "PrintInfo",
			method: (*UIManager).PrintInfo,
			symbol: gibidiutils.IconInfo,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				// Disable colors for consistent testing
				oldNoColor := color.NoColor
				color.NoColor = true
				defer func() { color.NoColor = oldNoColor }()

				buf := &bytes.Buffer{}
				ui := &UIManager{
					enableColors: false,
					output:       buf,
				}

				tt.method(ui, "Test message")

				output := buf.String()
				assert.True(t, strings.HasSuffix(output, "\n"))
				assert.Contains(t, output, tt.symbol)
			},
		)
	}
}
