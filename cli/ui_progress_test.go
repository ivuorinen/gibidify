package cli

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStartProgress(t *testing.T) {
	tests := []struct {
		name        string
		total       int
		description string
		enabled     bool
		expectBar   bool
	}{
		{
			name:        "progress enabled with valid total",
			total:       100,
			description: "Processing files",
			enabled:     true,
			expectBar:   true,
		},
		{
			name:        "progress disabled",
			total:       100,
			description: "Processing files",
			enabled:     false,
			expectBar:   false,
		},
		{
			name:        "zero total",
			total:       0,
			description: "Processing files",
			enabled:     true,
			expectBar:   false,
		},
		{
			name:        "negative total",
			total:       -5,
			description: "Processing files",
			enabled:     true,
			expectBar:   false,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				ui := &UIManager{
					enableProgress: tt.enabled,
					output:         &bytes.Buffer{},
				}

				ui.StartProgress(tt.total, tt.description)

				if tt.expectBar {
					assert.NotNil(t, ui.progressBar)
				} else {
					assert.Nil(t, ui.progressBar)
				}
			},
		)
	}
}

func TestUpdateProgress(t *testing.T) {
	tests := []struct {
		name         string
		setupBar     bool
		enabledProg  bool
		expectUpdate bool
	}{
		{
			name:         "with progress bar",
			setupBar:     true,
			enabledProg:  true,
			expectUpdate: true,
		},
		{
			name:         "without progress bar",
			setupBar:     false,
			enabledProg:  false,
			expectUpdate: false,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(_ *testing.T) {
				ui := &UIManager{
					enableProgress: tt.enabledProg,
					output:         &bytes.Buffer{},
				}

				if tt.setupBar {
					ui.StartProgress(10, "Test")
				}

				// Should not panic
				ui.UpdateProgress(1)

				// Multiple updates should not panic
				ui.UpdateProgress(2)
				ui.UpdateProgress(3)
			},
		)
	}
}

func TestFinishProgress(t *testing.T) {
	tests := []struct {
		name     string
		setupBar bool
	}{
		{
			name:     "with progress bar",
			setupBar: true,
		},
		{
			name:     "without progress bar",
			setupBar: false,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				ui := &UIManager{
					enableProgress: true,
					output:         &bytes.Buffer{},
				}

				if tt.setupBar {
					ui.StartProgress(10, "Test")
				}

				// Should not panic
				ui.FinishProgress()

				// Bar should be cleared
				assert.Nil(t, ui.progressBar)
			},
		)
	}
}
