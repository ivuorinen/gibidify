package cli

import (
	"os"
	"testing"

	"github.com/ivuorinen/gibidify/testutil"
)

// TestIsInteractiveTerminalFalseWhenStderrUnusable verifies a stderr that
// cannot be inspected is treated as non-interactive.
func TestIsInteractiveTerminalFalseWhenStderrUnusable(t *testing.T) {
	f, _ := testutil.CreateTempOutputFile(t, "stderr-*")
	testutil.CloseFile(t, f)

	orig := os.Stderr
	os.Stderr = f
	t.Cleanup(func() { os.Stderr = orig })

	if isInteractiveTerminal() {
		t.Error("isInteractiveTerminal() = true for a closed stderr, want false")
	}
}
