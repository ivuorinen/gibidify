package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsColorTerminal(t *testing.T) {
	tests := []struct {
		name     string
		env      terminalEnvSetup
		expected bool
	}{
		{
			name:     "dumb terminal",
			env:      envDumbTerminal,
			expected: false,
		},
		{
			name:     "empty TERM",
			env:      envEmptyTerm,
			expected: false,
		},
		{
			name:     "CI without GitHub Actions",
			env:      envCIWithoutGitHub,
			expected: false,
		},
		{
			name:     "GitHub Actions",
			env:      envGitHubActions,
			expected: true,
		},
		{
			name:     "NO_COLOR set",
			env:      envNoColor,
			expected: false,
		},
		{
			name:     "FORCE_COLOR set",
			env:      envForceColor,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.env.apply(t)

			result := isColorTerminal()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsInteractiveTerminal(t *testing.T) {
	// This function checks if stderr is a terminal
	// In test environment, it will typically return false
	result := isInteractiveTerminal()
	assert.False(t, result)
}
