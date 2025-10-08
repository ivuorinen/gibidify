package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsColorTerminal(t *testing.T) {
	// Use t.Setenv for environment isolation

	tests := []struct {
		name     string
		setupEnv func(t *testing.T)
		expected bool
	}{
		{
			name: "dumb terminal",
			setupEnv: func(t *testing.T) {
				t.Setenv("TERM", "dumb")
				t.Setenv("CI", "")
				t.Setenv("NO_COLOR", "")
				t.Setenv("FORCE_COLOR", "")
			},
			expected: false,
		},
		{
			name: "empty TERM",
			setupEnv: func(t *testing.T) {
				t.Setenv("TERM", "")
			},
			expected: false,
		},
		{
			name: "CI without GitHub Actions",
			setupEnv: func(t *testing.T) {
				t.Setenv("TERM", "xterm")
				t.Setenv("CI", "true")
				t.Setenv("GITHUB_ACTIONS", "")
			},
			expected: false,
		},
		{
			name: "GitHub Actions",
			setupEnv: func(t *testing.T) {
				t.Setenv("TERM", "xterm")
				t.Setenv("CI", "true")
				t.Setenv("GITHUB_ACTIONS", "true")
			},
			expected: true,
		},
		{
			name: "NO_COLOR set",
			setupEnv: func(t *testing.T) {
				t.Setenv("TERM", "xterm")
				t.Setenv("NO_COLOR", "1")
				t.Setenv("CI", "")
			},
			expected: false,
		},
		{
			name: "FORCE_COLOR set",
			setupEnv: func(t *testing.T) {
				t.Setenv("TERM", "dumb")
				t.Setenv("FORCE_COLOR", "1")
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				tt.setupEnv(t)

				result := isColorTerminal()
				assert.Equal(t, tt.expected, result)
			},
		)
	}
}

func TestIsInteractiveTerminal(t *testing.T) {
	// This function checks if stderr is a terminal
	// In test environment, it will typically return false
	result := isInteractiveTerminal()
	assert.False(t, result)
}
