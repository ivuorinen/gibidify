package fileproc

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"

	ignore "github.com/sabhiram/go-gitignore"

	"github.com/ivuorinen/gibidify/shared"
	"github.com/ivuorinen/gibidify/testutil"
)

// openForRead opens path with flag and closes it when the test ends.
func openForRead(t *testing.T, path string, flag int) *os.File {
	t.Helper()
	f, err := os.OpenFile(filepath.Clean(path), flag, 0)
	if err != nil {
		t.Fatalf("open %s: %v", path, err)
	}
	t.Cleanup(func() { _ = f.Close() })

	return f
}

// TestProcessInMemoryReadFailures covers a file that outgrew the limit after
// validation and a descriptor that cannot be read: both are returned as
// structured errors and nothing is sent.
func TestProcessInMemoryReadFailures(t *testing.T) {
	path := testutil.CreateTestFile(t, t.TempDir(), "grown.txt", []byte("0123456789"))

	tests := []struct {
		name     string
		flag     int
		limit    int64
		wantCode string
	}{
		{"over limit", os.O_RDONLY, 4, shared.CodeValidationSize},
		{"read error", os.O_WRONLY, 100, shared.CodeProcessingFileRead},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &FileProcessor{sizeLimit: tt.limit}
			outCh := make(chan WriteRequest, 1)
			f := openForRead(t, path, tt.flag)
			err := p.processInMemoryWithContext(context.Background(), f, path, "grown.txt", outCh)
			if len(outCh) != 0 {
				t.Fatalf("sent %d requests, want 0", len(outCh))
			}
			var se *shared.StructuredError
			if !errors.As(err, &se) || se.Code != tt.wantCode {
				t.Fatalf("error = %v, want code %s", err, tt.wantCode)
			}
			if se.FilePath != path {
				t.Fatalf("FilePath = %q, want %q", se.FilePath, path)
			}
		})
	}
}

// TestProcessInMemoryCanceledBeforeRead verifies a cancellation that lands
// after validation stops the read and sends nothing.
func TestProcessInMemoryCanceledBeforeRead(t *testing.T) {
	path := testutil.CreateTestFile(t, t.TempDir(), "a.go", []byte("package a"))
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	outCh := make(chan WriteRequest, 1)
	p := &FileProcessor{sizeLimit: 100}
	err := p.processInMemoryWithContext(ctx, openForRead(t, path, os.O_RDONLY), path, "a.go", outCh)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("error = %v, want context.Canceled", err)
	}
	if len(outCh) != 0 {
		t.Fatalf("sent %d requests, want 0", len(outCh))
	}
}

// cancelAfterFirstCheck is a context that reports nothing on its first Done
// call and is canceled on every later one, so the pre-read check passes and
// the cancellation lands while the result is waiting to be sent.
type cancelAfterFirstCheck struct {
	context.Context
	calls atomic.Int32
	done  chan struct{}
}

func (c *cancelAfterFirstCheck) Done() <-chan struct{} {
	if c.calls.Add(1) == 1 {
		return nil
	}

	return c.done
}

func (*cancelAfterFirstCheck) Err() error { return context.Canceled }

// TestProcessInMemoryCanceledWhileSending verifies a cancellation while the
// output channel is full returns the cancellation instead of blocking.
func TestProcessInMemoryCanceledWhileSending(t *testing.T) {
	path := testutil.CreateTestFile(t, t.TempDir(), "a.go", []byte("package a"))
	ctx := &cancelAfterFirstCheck{Context: context.Background(), done: make(chan struct{})}
	close(ctx.done)

	outCh := make(chan WriteRequest) // never read: the send cannot proceed
	p := &FileProcessor{sizeLimit: 100}
	err := p.processInMemoryWithContext(ctx, openForRead(t, path, os.O_RDONLY), path, "a.go", outCh)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("error = %v, want context.Canceled", err)
	}
	testutil.AssertErrorContains(t, err, "canceled before output", "send")
}

// infoErrEntry is a directory entry whose file vanished before Info was called.
type infoErrEntry struct{ name string }

func (e infoErrEntry) Name() string             { return e.name }
func (infoErrEntry) IsDir() bool                { return false }
func (infoErrEntry) Type() fs.FileMode          { return 0 }
func (infoErrEntry) Info() (fs.FileInfo, error) { return nil, fs.ErrNotExist }
func (e infoErrEntry) String() string           { return e.name }

// TestIsExcludedVanishedEntry verifies an entry that cannot be stat'ed is not
// treated as excluded by identity.
func TestIsExcludedVanishedEntry(t *testing.T) {
	outPath := testutil.CreateTestFile(t, t.TempDir(), "out.md", nil)
	info, err := os.Stat(outPath)
	if err != nil {
		t.Fatal(err)
	}
	f := &FileFilter{excludedFiles: []os.FileInfo{info}}

	if f.isExcluded(infoErrEntry{name: "gone.go"}, filepath.Join(filepath.Dir(outPath), "gone.go")) {
		t.Fatal("isExcluded = true for an entry whose Info failed, want false")
	}
}

// TestTryLoadIgnoreFileStatError verifies a stat failure other than "does not
// exist" is reported instead of being treated as an absent ignore file.
func TestTryLoadIgnoreFileStatError(t *testing.T) {
	// A regular file used as the directory makes stat fail with ENOTDIR.
	notDir := testutil.CreateTestFile(t, t.TempDir(), "file", []byte("x"))

	rule, err := tryLoadIgnoreFile(notDir, ".gitignore")
	if rule != nil {
		t.Fatalf("rule = %+v, want nil", rule)
	}
	testutil.AssertErrorContains(t, err, "cannot read ignore file", "tryLoadIgnoreFile")
}

// TestMatchesIgnoreRulesSkipsUnrelatableBase verifies a rule whose base cannot
// be related to the path has no opinion, so a matching pattern does not ignore it.
func TestMatchesIgnoreRulesSkipsUnrelatableBase(t *testing.T) {
	// filepath.Rel fails between a relative base and an absolute path.
	gi := ignore.CompileIgnoreLines("*.log")
	dir := t.TempDir()
	path := filepath.Join(dir, "debug.log")

	// Control: the same pattern under a relatable base does ignore the path.
	if !matchesIgnoreRules(path, false, []ignoreRule{{gi: gi, base: dir}}) {
		t.Fatal("matchesIgnoreRules = false under the file's own directory, want true")
	}
	if matchesIgnoreRules(path, false, []ignoreRule{{gi: gi, base: "relative"}}) {
		t.Fatal("matchesIgnoreRules = true, want false for an unrelatable base")
	}
	if rel, ok := relativeMatchPath(path, false, "relative"); ok || rel != "" {
		t.Fatalf("relativeMatchPath = (%q, %v), want (\"\", false)", rel, ok)
	}
}
