package main

import (
	"io"
	"os"
	"strings"
	"testing"
)

// TestMainReturnsOnSuccess runs main() itself with --version: a run that
// succeeds returns from main without exiting and prints the build metadata.
func TestMainReturnsOnSuccess(t *testing.T) {
	origVersion, origArgs, origStdout := version, os.Args, os.Stdout
	t.Cleanup(func() {
		version, os.Args, os.Stdout = origVersion, origArgs, origStdout
		resetFlagState()
	})

	version = "7.7.7-main"
	resetFlagState()
	os.Args = []string{"gibidify", "--version"}

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("creating pipe: %v", err)
	}
	os.Stdout = w

	main()

	if closeErr := w.Close(); closeErr != nil {
		t.Fatalf("closing pipe writer: %v", closeErr)
	}
	captured, readErr := io.ReadAll(r)
	if readErr != nil {
		t.Fatalf("reading pipe: %v", readErr)
	}
	if out := string(captured); !strings.Contains(out, "gibidify 7.7.7-main") {
		t.Errorf("main() --version stdout = %q, want it to contain the version", out)
	}
}
