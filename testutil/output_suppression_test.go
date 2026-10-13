package testutil

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/ivuorinen/gibidify/shared"
)

const (
	suppressedLogMsg = "message logged while suppressed"
	restoredLogMsg   = "message logged after restore"
)

// redirectOutput points os.Stderr at a temporary file and the logger at a
// buffer, and returns both. os.Stdout, os.Stderr and the logger output are
// restored on cleanup.
func redirectOutput(t *testing.T) (stderr *os.File, logs *bytes.Buffer) {
	t.Helper()
	origStdout, origStderr := os.Stdout, os.Stderr
	logger := shared.GetLogger()
	origLogOutput := logger.Output()
	t.Cleanup(func() {
		os.Stdout, os.Stderr = origStdout, origStderr
		logger.SetOutput(origLogOutput)
	})

	stderr, _ = CreateTempOutputFile(t, "stderr-*.log")
	os.Stderr = stderr
	logs = &bytes.Buffer{}
	logger.SetOutput(logs)

	return stderr, logs
}

// verifyRestoredLog checks that the logger writes to logs again, that the
// suppressed message never reached it, and that nothing reached stderr.
func verifyRestoredLog(t *testing.T, stderr *os.File, logs *bytes.Buffer) {
	t.Helper()
	shared.GetLogger().Errorf(restoredLogMsg)
	if got := logs.String(); !strings.Contains(got, restoredLogMsg) || strings.Contains(got, suppressedLogMsg) {
		t.Errorf("logger output after restore = %q, want only %q", got, restoredLogMsg)
	}
	path := stderr.Name()
	CloseFile(t, stderr)
	if content := ReadFile(t, path); len(content) != 0 {
		t.Errorf("restore sent the logger to os.Stderr, which got %q", content)
	}
}

// TestSuppressLogs verifies the logger is silenced and that restore puts back
// the writer it had before, not os.Stderr.
func TestSuppressLogs(t *testing.T) {
	stderr, logs := redirectOutput(t)
	logger := shared.GetLogger()

	restore := SuppressLogs(t)
	logger.Errorf(suppressedLogMsg)
	if logs.Len() != 0 {
		t.Errorf("logger wrote %q while suppressed", logs.String())
	}

	restore()
	verifyRestoredLog(t, stderr, logs)
}

// TestSuppressAllOutput verifies stdout, stderr and the logger go to the null
// device until restore puts the originals back and closes the null device.
func TestSuppressAllOutput(t *testing.T) {
	stderr, logs := redirectOutput(t)
	origStdout := os.Stdout
	logger := shared.GetLogger()

	restore := SuppressAllOutput(t)
	devNull := os.Stdout
	if os.Stdout.Name() != os.DevNull || os.Stderr.Name() != os.DevNull {
		t.Errorf("stdout = %q, stderr = %q, want both %q", os.Stdout.Name(), os.Stderr.Name(), os.DevNull)
	}
	if _, err := fmt.Fprintln(os.Stdout, "discarded"); err != nil {
		t.Errorf("writing to suppressed stdout: %v", err)
	}
	logger.Errorf(suppressedLogMsg)
	if logs.Len() != 0 {
		t.Errorf("logger wrote %q while suppressed", logs.String())
	}

	restore()
	if os.Stdout != origStdout || os.Stderr != stderr {
		t.Error("restore did not put the original stdout and stderr back")
	}
	if err := devNull.Close(); !errors.Is(err, os.ErrClosed) {
		t.Errorf("closing null device after restore: got %v, want %v", err, os.ErrClosed)
	}
	verifyRestoredLog(t, stderr, logs)
}
