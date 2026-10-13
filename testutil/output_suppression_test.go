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

// redirectStderr points os.Stderr at a temporary file and returns it with its
// path. os.Stderr, os.Stdout and the logger output are restored on cleanup.
func redirectStderr(t *testing.T) (file *os.File, path string) {
	t.Helper()
	origStdout, origStderr := os.Stdout, os.Stderr
	t.Cleanup(func() {
		os.Stdout, os.Stderr = origStdout, origStderr
		shared.GetLogger().SetOutput(origStderr)
	})

	file, path = CreateTempOutputFile(t, "stderr-*.log")
	os.Stderr = file

	return file, path
}

// verifyStderrLog closes the redirected stderr file and checks that it holds
// the restored message and not the suppressed one.
func verifyStderrLog(t *testing.T, file *os.File, path string) {
	t.Helper()
	CloseFile(t, file)
	content := string(ReadFile(t, path))
	if !strings.Contains(content, restoredLogMsg) {
		t.Errorf("stderr after restore = %q, want it to contain %q", content, restoredLogMsg)
	}
	if strings.Contains(content, suppressedLogMsg) {
		t.Errorf("stderr = %q, suppressed message leaked", content)
	}
}

// TestSuppressLogs verifies the logger is silenced and that restore sends it
// to os.Stderr.
func TestSuppressLogs(t *testing.T) {
	file, path := redirectStderr(t)
	logger := shared.GetLogger()
	var buf bytes.Buffer
	logger.SetOutput(&buf)

	restore := SuppressLogs(t)
	logger.Errorf(suppressedLogMsg)
	if buf.Len() != 0 {
		t.Errorf("logger wrote %q while suppressed", buf.String())
	}

	restore()
	logger.Errorf(restoredLogMsg)
	verifyStderrLog(t, file, path)
}

// TestSuppressAllOutput verifies stdout, stderr and the logger go to the null
// device until restore puts the originals back and closes the null device.
func TestSuppressAllOutput(t *testing.T) {
	file, path := redirectStderr(t)
	origStdout := os.Stdout
	logger := shared.GetLogger()
	var buf bytes.Buffer
	logger.SetOutput(&buf)

	restore := SuppressAllOutput(t)
	devNull := os.Stdout
	if os.Stdout.Name() != os.DevNull || os.Stderr.Name() != os.DevNull {
		t.Errorf("stdout = %q, stderr = %q, want both %q", os.Stdout.Name(), os.Stderr.Name(), os.DevNull)
	}
	if _, err := fmt.Fprintln(os.Stdout, "discarded"); err != nil {
		t.Errorf("writing to suppressed stdout: %v", err)
	}
	logger.Errorf(suppressedLogMsg)
	if buf.Len() != 0 {
		t.Errorf("logger wrote %q while suppressed", buf.String())
	}

	restore()
	if os.Stdout != origStdout || os.Stderr != file {
		t.Error("restore did not put the original stdout and stderr back")
	}
	if err := devNull.Close(); !errors.Is(err, os.ErrClosed) {
		t.Errorf("closing null device after restore: got %v, want %v", err, os.ErrClosed)
	}
	logger.Errorf(restoredLogMsg)
	verifyStderrLog(t, file, path)
}
