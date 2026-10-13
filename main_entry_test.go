package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ivuorinen/gibidify/cli"
	"github.com/ivuorinen/gibidify/shared"
	"github.com/ivuorinen/gibidify/testutil"
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

// captureExitOutput runs fn with os.Stderr redirected to a file and the logger
// writing to a buffer, and returns what each received.
func captureExitOutput(t *testing.T, fn func()) (stderr, logged string) {
	t.Helper()

	origStderr := os.Stderr
	logger := shared.GetLogger()
	t.Cleanup(func() {
		os.Stderr = origStderr
		logger.SetOutput(os.Stderr)
	})

	stderrFile, err := os.Create(filepath.Join(t.TempDir(), "stderr"))
	if err != nil {
		t.Fatalf("creating stderr file: %v", err)
	}
	t.Cleanup(func() { _ = stderrFile.Close() })

	var logBuf bytes.Buffer
	logger.SetOutput(&logBuf)
	os.Stderr = stderrFile

	fn()

	os.Stderr = origStderr
	logger.SetOutput(os.Stderr)

	captured, err := os.ReadFile(stderrFile.Name())
	if err != nil {
		t.Fatalf("reading captured stderr: %v", err)
	}

	return string(captured), logBuf.String()
}

// TestExitCode covers every exit status main() can hand to os.Exit and what
// each one reports on stderr and in the log.
func TestExitCode(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantCode   int
		wantStderr []string
		wantLog    string
	}{
		{name: "success", err: nil, wantCode: 0},
		{
			name:       "interrupted",
			err:        fmt.Errorf("processing: %w", context.Canceled),
			wantCode:   exitInterrupted,
			wantStderr: []string{"Interrupted"},
		},
		{
			name:       "user error",
			err:        fmt.Errorf("parsing flags: %w", cli.NewCLIMissingSourceError()),
			wantCode:   1,
			wantStderr: []string{"source directory is required", "Suggestions"},
		},
		{
			name: "structured error is a user error",
			err: shared.NewStructuredError(
				shared.ErrorTypeFileSystem, shared.CodeFSNotFound, "source directory does not exist", "/missing", nil,
			),
			wantCode:   1,
			wantStderr: []string{"source directory does not exist", "/missing"},
		},
		{
			name:       "system error",
			err:        errors.New("disk exploded"),
			wantCode:   2,
			wantStderr: []string{"An unexpected error occurred. Please check the logs."},
			wantLog:    "System error: disk exploded",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var code int
			stderr, logged := captureExitOutput(t, func() { code = exitCode(tt.err) })

			if code != tt.wantCode {
				t.Errorf("exitCode(%v) = %d, want %d", tt.err, code, tt.wantCode)
			}
			if len(tt.wantStderr) == 0 && stderr != "" {
				t.Errorf("stderr = %q, want nothing", stderr)
			}
			for _, want := range tt.wantStderr {
				if !strings.Contains(stderr, want) {
					t.Errorf("stderr = %q, want it to contain %q", stderr, want)
				}
			}
			assertLogged(t, logged, tt.wantLog)
		})
	}
}

// assertLogged checks the log holds want, or nothing when want is empty.
func assertLogged(t *testing.T, logged, want string) {
	t.Helper()

	if want == "" {
		if logged != "" {
			t.Errorf("log = %q, want nothing", logged)
		}

		return
	}
	if !strings.Contains(logged, want) {
		t.Errorf("log = %q, want it to contain %q", logged, want)
	}
}

// TestExitCodeOfRun checks the status main() derives from errors run() really
// returns, so a change in how run() wraps them cannot silently move the code.
func TestExitCodeOfRun(t *testing.T) {
	canceled, cancel := context.WithCancel(context.Background())
	cancel()

	tests := []struct {
		name     string
		ctx      context.Context
		args     func(t *testing.T) []string
		wantCode int
	}{
		{
			name: "missing source",
			ctx:  context.Background(),
			args: func(_ *testing.T) []string {
				return []string{"gibidify", testutil.TestCLIFlagNoUI}
			},
			wantCode: 1,
		},
		{
			name: "canceled context",
			ctx:  canceled,
			args: func(t *testing.T) []string {
				srcDir := t.TempDir()
				testutil.CreateTestFile(t, srcDir, testutil.TestFileGo, []byte(testutil.LiteralPackageMain))

				return []string{
					"gibidify", testutil.TestCLIFlagSource, srcDir, testutil.TestCLIFlagDestination,
					filepath.Join(t.TempDir(), testutil.TestOutputMD), testutil.TestCLIFlagNoUI,
				}
			},
			wantCode: exitInterrupted,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			origArgs := os.Args
			t.Cleanup(func() {
				os.Args = origArgs
				resetFlagState()
			})
			resetFlagState()
			os.Args = tt.args(t)

			var code int
			_, _ = captureExitOutput(t, func() { code = exitCode(run(tt.ctx)) })

			if code != tt.wantCode {
				t.Errorf("exitCode(run()) = %d, want %d", code, tt.wantCode)
			}
		})
	}
}
