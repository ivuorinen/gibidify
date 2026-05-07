package main

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

// TestPrintVersion verifies the version helper writes all build metadata fields.
func TestPrintVersion(t *testing.T) {
	origVersion, origCommit, origDate, origBuiltBy := version, commit, date, builtBy
	t.Cleanup(func() {
		version, commit, date, builtBy = origVersion, origCommit, origDate, origBuiltBy
	})

	version = "1.2.3"
	commit = "deadbeef"
	date = "2026-05-07"
	builtBy = "unit-test"

	var buf bytes.Buffer
	printVersion(&buf)

	got := buf.String()
	for _, want := range []string{"gibidify 1.2.3", "commit: deadbeef", "built: 2026-05-07", "by: unit-test"} {
		if !strings.Contains(got, want) {
			t.Errorf("printVersion output missing %q\nfull output:\n%s", want, got)
		}
	}
}

// TestRunVersionFlag verifies --version short-circuits run() with metadata on stdout.
func TestRunVersionFlag(t *testing.T) {
	origVersion, origCommit, origDate, origBuiltBy := version, commit, date, builtBy
	origArgs := os.Args
	origStdout := os.Stdout
	t.Cleanup(func() {
		version, commit, date, builtBy = origVersion, origCommit, origDate, origBuiltBy
		os.Args = origArgs
		os.Stdout = origStdout
	})

	version = "9.9.9"
	commit = "abc123"
	date = "2026-05-07"
	builtBy = "release-test"

	resetFlagState()
	os.Args = []string{"gibidify", "--version"}

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("creating pipe: %v", err)
	}
	os.Stdout = w

	runErr := run(t.Context())

	if closeErr := w.Close(); closeErr != nil {
		t.Fatalf("closing pipe writer: %v", closeErr)
	}
	captured, readErr := io.ReadAll(r)
	if readErr != nil {
		t.Fatalf("reading pipe: %v", readErr)
	}

	if runErr != nil {
		t.Fatalf("run with --version returned error: %v", runErr)
	}

	out := string(captured)
	for _, want := range []string{"gibidify 9.9.9", "commit: abc123", "built: 2026-05-07", "by: release-test"} {
		if !strings.Contains(out, want) {
			t.Errorf("run --version stdout missing %q\nfull output:\n%s", want, out)
		}
	}
}
