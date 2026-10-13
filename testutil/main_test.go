package testutil

import (
	"testing"

	"github.com/ivuorinen/gibidify/config"
)

// TestMain keeps the developer's own gibidify config out of the helper tests.
func TestMain(m *testing.M) {
	RunIsolated(m)
}

// TestRunIsolatedHidesUserConfig verifies that, under RunIsolated, loading the
// configuration finds no config file at all, whatever the person running the
// tests has installed.
func TestRunIsolatedHidesUserConfig(t *testing.T) {
	config.Reset()
	t.Cleanup(config.Reset)

	if err := config.LoadConfig(); err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	if used := config.FileUsed(); used != "" {
		t.Errorf("a config file was loaded under RunIsolated: %s", used)
	}
}
