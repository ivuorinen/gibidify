package testutil

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"

	"github.com/ivuorinen/gibidify/config"
)

// fakeTB records the Helper, Errorf and Fatalf calls a helper makes. Every
// other method goes to the embedded real testing.TB, so TempDir and Cleanup
// still work. Fatalf ends the calling goroutine like the real one does, so a
// helper never runs past it; runWithFakeTB runs the helper on its own goroutine.
type fakeTB struct {
	testing.TB

	mu      sync.Mutex
	helpers int
	errors  []string
	fatals  []string
}

func (f *fakeTB) Helper() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.helpers++
}

func (f *fakeTB) Errorf(format string, args ...any) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.errors = append(f.errors, fmt.Sprintf(format, args...))
}

func (f *fakeTB) Fatalf(format string, args ...any) {
	f.mu.Lock()
	f.fatals = append(f.fatals, fmt.Sprintf(format, args...))
	f.mu.Unlock()
	runtime.Goexit()
}

// runWithFakeTB runs fn with a fakeTB on a separate goroutine and waits for it
// to return or to stop at Fatalf.
func runWithFakeTB(t *testing.T, fn func(tb testing.TB)) *fakeTB {
	t.Helper()
	fake := &fakeTB{TB: t}
	done := make(chan struct{})
	go func() {
		defer close(done)
		fn(fake)
	}()
	<-done

	return fake
}

// helperFailureCase is one helper call that must report a failure.
type helperFailureCase struct {
	name      string
	call      func(tb testing.TB)
	wantFatal bool
	wantMsg   string
}

// missingPath returns a path under a fresh temp directory that does not exist.
func missingPath(tb testing.TB) string {
	return filepath.Join(tb.TempDir(), "missing", "entry")
}

var helperFailureCases = []helperFailureCase{
	{"CreateTestFile missing dir", func(tb testing.TB) {
		CreateTestFile(tb, missingPath(tb), TestFileTXT, nil)
	}, true, "Failed to write file"},
	{"ReadFile missing file", func(tb testing.TB) {
		ReadFile(tb, missingPath(tb))
	}, true, "Failed to read file"},
	{"CreateTempOutputFile bad pattern", func(tb testing.TB) {
		CreateTempOutputFile(tb, "bad/pattern-*")
	}, true, "Failed to create temp output file"},
	{"CreateTestDirectory missing parent", func(tb testing.TB) {
		CreateTestDirectory(tb, missingPath(tb), "child")
	}, true, "Failed to create directory"},
	{"CreateTestDirectoryStructure under a file", func(tb testing.TB) {
		file := CreateTestFile(tb, tb.TempDir(), TestFileTXT, nil)
		CreateTestDirectoryStructure(tb, file, []DirSpec{{Path: "sub"}})
	}, true, "Failed to create directory structure"},
	{"ResetConfig unparsable config", func(tb testing.TB) {
		tb.Cleanup(config.Reset)
		dir := tb.TempDir()
		CreateTestFile(tb, dir, "config.yaml", []byte("fileSizeLimit: [\n"))
		ResetConfig(tb, dir)
	}, true, "loading config"},
	{"MustSucceed error", func(tb testing.TB) {
		MustSucceed(tb, ErrTestError, TestCaseTestOperation)
	}, true, TestCaseTestOperation},
	{"VerifyContentContains missing substring", func(tb testing.TB) {
		VerifyContentContains(tb, "present", []string{"present", "absent"})
	}, false, "Content missing expected substring: absent"},
	{"CloseFile already closed", func(tb testing.TB) {
		file, _ := CreateTempOutputFile(tb, "closed-*")
		if err := file.Close(); err != nil {
			tb.Fatalf("first close: %v", err)
		}
		CloseFile(tb, file)
	}, false, "Failed to close file"},
	{"AssertNoError error", func(tb testing.TB) {
		AssertNoError(tb, ErrTestError, TestCaseTestOperation)
	}, false, TestErrTestErrorMsg},
	{"AssertExpectedError nil", func(tb testing.TB) {
		AssertExpectedError(tb, nil, TestCaseTestOperation)
	}, false, TestCaseTestOperation},
	{"AssertErrorContains nil", func(tb testing.TB) {
		AssertErrorContains(tb, nil, "needle", TestCaseTestOperation)
	}, false, `expected error containing "needle" but got none`},
	{"AssertErrorContains mismatch", func(tb testing.TB) {
		AssertErrorContains(tb, ErrTestError, "needle", TestCaseTestOperation)
	}, false, `should contain "needle"`},
}

// TestHelperFailureBranches verifies each helper reports its failure through
// the expected method, with the expected message, and marks itself a helper.
func TestHelperFailureBranches(t *testing.T) {
	for _, tc := range helperFailureCases {
		t.Run(tc.name, func(t *testing.T) {
			fake := runWithFakeTB(t, tc.call)

			got, other := fake.errors, fake.fatals
			if tc.wantFatal {
				got, other = fake.fatals, fake.errors
			}
			if len(got) != 1 || !strings.Contains(got[0], tc.wantMsg) {
				t.Errorf("reported %q, want one message containing %q", got, tc.wantMsg)
			}
			if len(other) != 0 {
				t.Errorf("unexpected messages through the other method: %q", other)
			}
			if fake.helpers == 0 {
				t.Error("helper did not call Helper()")
			}
		})
	}
}

// TestRunWithFakeTBStopsAtFatalf verifies the fake stops the helper at Fatalf,
// as the real testing.TB does, so code after it never runs.
func TestRunWithFakeTBStopsAtFatalf(t *testing.T) {
	ranPastFatal := false
	fake := runWithFakeTB(t, func(tb testing.TB) {
		tb.Fatalf("stop")
		ranPastFatal = true
	})

	if ranPastFatal {
		t.Error("code after Fatalf ran")
	}
	if len(fake.fatals) != 1 || fake.fatals[0] != "stop" {
		t.Errorf("fatals = %q, want [\"stop\"]", fake.fatals)
	}
}
