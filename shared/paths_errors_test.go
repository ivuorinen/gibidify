package shared

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// requireStructuredError asserts err is a *StructuredError with the given code
// and message.
func requireStructuredError(t *testing.T, err error, wantCode, wantMessage string) {
	t.Helper()
	structErr := &StructuredError{}
	if !errors.As(err, &structErr) {
		t.Fatalf("expected *StructuredError, got %T: %v", err, err)
	}
	if structErr.Code != wantCode {
		t.Errorf(TestFmtExpectedCode, wantCode, structErr.Code)
	}
	if structErr.Message != wantMessage {
		t.Errorf(TestFmtExpectedMessage, wantMessage, structErr.Message)
	}
}

// TestPathValidationStatFailures covers paths whose stat fails for a reason
// other than "does not exist", and a destination whose parent is a file.
func TestPathValidationStatFailures(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "regular.txt")
	if err := os.WriteFile(file, []byte(TestContentTest), 0o600); err != nil {
		t.Fatalf("writing fixture file: %v", err)
	}

	tests := []struct {
		name        string
		validate    func(string) error
		path        string
		wantCode    string
		wantMessage string
	}{
		{
			name:        "source below a regular file",
			validate:    ValidateSourcePath,
			path:        filepath.Join(file, "child"),
			wantCode:    CodeFSAccess,
			wantMessage: "cannot access source directory",
		},
		{
			name:        "destination parent below a regular file",
			validate:    ValidateDestinationPath,
			path:        filepath.Join(file, "child", TestFileOutputTXT),
			wantCode:    CodeFSAccess,
			wantMessage: "cannot access destination parent directory",
		},
		{
			name:        "destination parent is a regular file",
			validate:    ValidateDestinationPath,
			path:        filepath.Join(file, TestFileOutputTXT),
			wantCode:    CodeValidationPath,
			wantMessage: "destination parent is not a directory",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requireStructuredError(t, tt.validate(tt.path), tt.wantCode, tt.wantMessage)
		})
	}
}

// TestPathResolutionWithoutWorkingDirectory removes the working directory so
// filepath.Abs cannot resolve a relative path.
func TestPathResolutionWithoutWorkingDirectory(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("the working directory cannot be removed on Windows")
	}

	gone := filepath.Join(t.TempDir(), "gone")
	if err := os.Mkdir(gone, 0o750); err != nil {
		t.Fatalf(TestMsgFailedToCreateTestDir, err)
	}
	t.Chdir(gone)
	if err := os.Remove(gone); err != nil {
		t.Fatalf("removing working directory: %v", err)
	}

	if abs, err := AbsolutePath("relative"); err == nil {
		t.Errorf("AbsolutePath() = %q, want error without a working directory", abs)
	}

	err := ValidateDestinationPath(TestFileOutputTXT)
	requireStructuredError(t, err, CodeFSPathResolution, "cannot resolve destination path")
}
