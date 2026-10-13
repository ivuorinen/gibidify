package fileproc_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ivuorinen/gibidify/config"
	"github.com/ivuorinen/gibidify/fileproc"
	"github.com/ivuorinen/gibidify/shared"
	"github.com/ivuorinen/gibidify/testutil"
)

// writeTempConfig creates a temporary config file with the given YAML content
// and returns the directory path containing the config file.
func writeTempConfig(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(content), 0o600); err != nil {
		t.Fatalf("Failed to create temp config: %v", err)
	}
	return dir
}

const testContent = "package main\nfunc main() {}\n"

// processOne runs ProcessWithContext for one file and returns every request it
// emitted together with its error.
func processOne(
	ctx context.Context,
	processor *fileproc.FileProcessor,
	path string,
) ([]fileproc.WriteRequest, error) {
	ch := make(chan fileproc.WriteRequest, 1)
	var err error
	var wg sync.WaitGroup
	wg.Go(func() {
		defer close(ch)
		err = processor.ProcessWithContext(ctx, path, ch)
	})

	var reqs []fileproc.WriteRequest
	for req := range ch {
		reqs = append(reqs, req)
	}
	wg.Wait()

	return reqs, err
}

// TestProcessFileContentIsExact guards against the removed "\n---\n<path>\n"
// header: Content must be the file's bytes and Path the path relative to root.
func TestProcessFileContentIsExact(t *testing.T) {
	testutil.ResetConfig(t, "")
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "sub", "test.go")
	if err := os.MkdirAll(filepath.Dir(testFile), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(testFile, []byte(testContent), 0o600); err != nil {
		t.Fatalf(testutil.TestMsgFailedToCreateTestFile, err)
	}

	ch := make(chan fileproc.WriteRequest, 1)
	if err := fileproc.ProcessFileContext(context.Background(), testFile, ch, tmpDir); err != nil {
		t.Fatalf(testutil.TestMsgUnexpectedError, err)
	}
	close(ch)
	req := <-ch

	if req.Content != testContent {
		t.Errorf(testutil.TestFmtExpectedContent, testContent, req.Content)
	}
	if want := filepath.Join("sub", "test.go"); req.Path != want {
		t.Errorf(testutil.TestFmtExpectedFilePath, want, req.Path)
	}
}

// TestProcessorSkipsBinaryContent verifies a file with a NUL byte in its first
// bytes is skipped even though its extension is not a known binary extension.
func TestProcessorSkipsBinaryContent(t *testing.T) {
	testutil.ResetConfig(t, "")
	tmpDir := t.TempDir()
	blob := filepath.Join(tmpDir, "tool")
	if err := os.WriteFile(blob, []byte("ELF\x00\x01\x02binary"), 0o600); err != nil {
		t.Fatal(err)
	}

	reqs, err := processOne(context.Background(), fileproc.NewFileProcessor(tmpDir), blob)
	if !errors.Is(err, fileproc.ErrSkipped) {
		t.Fatalf("expected ErrSkipped, got %v", err)
	}
	if len(reqs) != 0 {
		t.Errorf("expected no output for binary content, got %d requests", len(reqs))
	}
}

// TestProcessorReplacesInvalidUTF8 verifies a non-UTF-8 file is emitted with its
// invalid bytes replaced by U+FFFD and a warning naming it. Every output format
// is UTF-8 text: JSON used to replace the bytes silently and YAML produced an
// unparseable document.
func TestProcessorReplacesInvalidUTF8(t *testing.T) {
	testutil.ResetConfig(t, "")
	var logs strings.Builder
	logger := shared.GetLogger()
	logger.SetOutput(&logs)
	t.Cleanup(func() { logger.SetOutput(os.Stderr) })

	tmpDir := t.TempDir()
	latin1 := filepath.Join(tmpDir, "latin1.txt")
	if err := os.WriteFile(latin1, []byte("caf\xe9\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	reqs, err := processOne(context.Background(), fileproc.NewFileProcessor(tmpDir), latin1)
	if err != nil {
		t.Fatalf(shared.ErrorFmtWithCause, "processing", err)
	}
	if len(reqs) != 1 || reqs[0].Content != "caf\xef\xbf\xbd\n" {
		t.Fatalf("content = %q, want invalid byte replaced with U+FFFD", reqs)
	}
	if !strings.Contains(logs.String(), "latin1.txt is not valid UTF-8") {
		t.Errorf("expected a warning naming the file, got: %q", logs.String())
	}
}

// TestProcessorRefusesSymlink verifies a symlink presented at processing time
// (e.g. swapped in after collection) is skipped, not followed out of the tree.
func TestProcessorRefusesSymlink(t *testing.T) {
	testutil.ResetConfig(t, "")
	outside := filepath.Join(t.TempDir(), "secret.txt")
	if err := os.WriteFile(outside, []byte("secret"), 0o600); err != nil {
		t.Fatal(err)
	}
	tmpDir := t.TempDir()
	link := filepath.Join(tmpDir, "link.txt")
	if err := os.Symlink(outside, link); err != nil {
		t.Skipf("symlinks unsupported: %v", err)
	}

	reqs, err := processOne(context.Background(), fileproc.NewFileProcessor(tmpDir), link)
	if !errors.Is(err, fileproc.ErrSkipped) {
		t.Fatalf("expected ErrSkipped, got %v", err)
	}
	if len(reqs) != 0 {
		t.Errorf("symlink target leaked into output: %+v", reqs)
	}
}

// createLargeTestFile creates a large test file (> 1MB) for large-file tests.
func createLargeTestFile(t *testing.T) *os.File {
	t.Helper()

	tmpFile, err := os.CreateTemp(t.TempDir(), "large_file_*.go")
	if err != nil {
		t.Fatalf(testutil.TestMsgFailedToCreateFile, err)
	}

	lineContent := "// Repeated comment line to make a file larger than 1MB\n"
	repeatCount := (1048576 / len(lineContent)) + 1000
	largeContent := strings.Repeat(lineContent, repeatCount)

	if _, err := tmpFile.WriteString(largeContent); err != nil {
		t.Fatalf(testutil.TestMsgFailedToWriteContent, err)
	}
	if err := tmpFile.Close(); err != nil {
		t.Fatalf(testutil.TestMsgFailedToCloseFile, err)
	}

	t.Logf("Created test file size: %d bytes", len(largeContent))

	return tmpFile
}

// TestProcessorLargeFileInline verifies a large file (within the size cap) is
// buffered and emitted as a single inline request with its full content.
func TestProcessorLargeFileInline(t *testing.T) {
	configDir := writeTempConfig(t, `
fileSizeLimit: 5242880
`)
	testutil.ResetConfig(t, configDir)

	tmpFile := createLargeTestFile(t)
	want, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		t.Fatal(err)
	}

	processor := fileproc.NewFileProcessor(filepath.Dir(tmpFile.Name()))
	reqs, err := processOne(context.Background(), processor, tmpFile.Name())
	if err != nil {
		t.Fatalf(testutil.TestMsgUnexpectedError, err)
	}
	if len(reqs) != 1 {
		t.Fatalf("Expected 1 inline request, got %d", len(reqs))
	}
	if reqs[0].Content != string(want) {
		t.Errorf("large file content differs: got %d bytes, want %d", len(reqs[0].Content), len(want))
	}
}

// TestProcessorContextCancellation tests that an already-canceled context
// produces an error and no output.
func TestProcessorContextCancellation(t *testing.T) {
	testutil.ResetConfig(t, "")

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")
	if err := os.WriteFile(testFile, []byte(testContent), 0o600); err != nil {
		t.Fatalf(testutil.TestMsgFailedToCreateTestFile, err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	reqs, err := processOne(ctx, fileproc.NewFileProcessor(tmpDir), testFile)
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled, got %v", err)
	}
	if len(reqs) != 0 {
		t.Errorf("expected no output after cancellation, got %d", len(reqs))
	}
}

// TestProcessorValidationEdgeCases tests a missing file and a file over the size limit.
func TestProcessorValidationEdgeCases(t *testing.T) {
	configDir := writeTempConfig(t, `
fileSizeLimit: 1024
`)
	testutil.ResetConfig(t, configDir)

	tmpDir := t.TempDir()
	processor := fileproc.NewFileProcessor(tmpDir)

	reqs, err := processOne(context.Background(), processor, filepath.Join(tmpDir, "does-not-exist.go"))
	if err == nil || len(reqs) != 0 {
		t.Errorf("missing file: want error and no output, got %d requests, err %v", len(reqs), err)
	}

	largeFile := filepath.Join(tmpDir, "large.go")
	largeContent := strings.Repeat("// Large file content\n", 100) // > 1KB
	if err := os.WriteFile(largeFile, []byte(largeContent), 0o600); err != nil {
		t.Fatalf("Failed to create large file: %v", err)
	}

	reqs, err = processOne(context.Background(), processor, largeFile)
	structErr := &shared.StructuredError{}
	if !errors.As(err, &structErr) || structErr.Code != shared.CodeValidationSize {
		t.Errorf("oversized file: want %s error, got %v", shared.CodeValidationSize, err)
	}
	if len(reqs) != 0 {
		t.Errorf("oversized file: want no output, got %d requests", len(reqs))
	}
}

// TestProcessorContextCancellationDuringValidation tests an expired deadline before validation.
func TestProcessorContextCancellationDuringValidation(t *testing.T) {
	testutil.ResetConfig(t, "")

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")
	if err := os.WriteFile(testFile, []byte(testContent), 0o600); err != nil {
		t.Fatalf(testutil.TestMsgFailedToCreateTestFile, err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()
	<-ctx.Done()

	reqs, err := processOne(ctx, fileproc.NewFileProcessor(tmpDir), testFile)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("expected context.DeadlineExceeded, got %v", err)
	}
	if len(reqs) != 0 {
		t.Errorf("expected no output after deadline, got %d", len(reqs))
	}
}

// TestProcessorEmptyFile verifies an empty file is emitted with empty content.
func TestProcessorEmptyFile(t *testing.T) {
	testutil.ResetConfig(t, "")

	tmpDir := t.TempDir()
	emptyFile := filepath.Join(tmpDir, "empty.go")
	if err := os.WriteFile(emptyFile, []byte(""), 0o600); err != nil {
		t.Fatalf("Failed to create empty file: %v", err)
	}

	reqs, err := processOne(context.Background(), fileproc.NewFileProcessor(tmpDir), emptyFile)
	if err != nil {
		t.Fatalf(testutil.TestMsgUnexpectedError, err)
	}
	if len(reqs) != 1 {
		t.Fatalf("Expected 1 result for empty file, got %d", len(reqs))
	}
	if reqs[0].Path != "empty.go" || reqs[0].Content != "" {
		t.Errorf("unexpected result for empty file: %+v", reqs[0])
	}
}

// TestProcessorLargeFileCancellation tests context cancellation while processing a large file.
func TestProcessorLargeFileCancellation(t *testing.T) {
	testutil.ResetConfig(t, "")

	tmpDir := t.TempDir()

	largeFile := filepath.Join(tmpDir, "large_stream.go")
	largeContent := strings.Repeat("// Large file content line\n", 50000) // > 1MB
	if err := os.WriteFile(largeFile, []byte(largeContent), 0o600); err != nil {
		t.Fatalf("Failed to create large file: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(1 * time.Millisecond)
		cancel()
	}()

	reqs, err := processOne(ctx, fileproc.NewFileProcessor(tmpDir), largeFile)
	// Either the read finished first (one complete request) or cancellation won.
	switch {
	case err == nil && len(reqs) == 1 && reqs[0].Content == largeContent:
	case errors.Is(err, context.Canceled) && len(reqs) == 0:
	default:
		t.Errorf("unexpected outcome: %d requests, err %v", len(reqs), err)
	}
}

// Benchmarks for processor hot paths

// benchmarkProcessFile benchmarks ProcessWithContext on a file with the given content.
func benchmarkProcessFile(b *testing.B, content string) {
	b.Helper()
	config.Reset()
	if err := config.LoadConfig(); err != nil {
		b.Fatalf("loading config: %v", err)
	}

	tmpDir := b.TempDir()
	testFile := filepath.Join(tmpDir, "bench.go")
	if err := os.WriteFile(testFile, []byte(content), 0o600); err != nil {
		b.Fatalf(testutil.TestMsgFailedToCreateTestFile, err)
	}

	processor := fileproc.NewFileProcessor(tmpDir)
	ctx := context.Background()

	b.ResetTimer()
	for b.Loop() {
		ch := make(chan fileproc.WriteRequest, 1)
		_ = processor.ProcessWithContext(ctx, testFile, ch)
	}
}

// BenchmarkProcessFileInline benchmarks processing of a small file.
func BenchmarkProcessFileInline(b *testing.B) {
	benchmarkProcessFile(b, strings.Repeat("// Inline benchmark content\n", 100)) // ~2.6KB
}

// BenchmarkProcessFileLarge benchmarks buffered processing for large files (> 1MB).
func BenchmarkProcessFileLarge(b *testing.B) {
	lineContent := "// Large-file benchmark content line that will be repeated\n"
	benchmarkProcessFile(b, strings.Repeat(lineContent, (1048576/len(lineContent))+1000))
}

// BenchmarkProcessorWithContext benchmarks ProcessWithContext for a single file.
func BenchmarkProcessorWithContext(b *testing.B) {
	benchmarkProcessFile(b, strings.Repeat("// Benchmark file content\n", 50))
}

func BenchmarkProcessorConcurrent(b *testing.B) {
	tmpDir := b.TempDir()

	// Create multiple test files
	testFiles := make([]string, 10)
	for i := 0; i < 10; i++ {
		testFiles[i] = filepath.Join(tmpDir, fmt.Sprintf("bench_concurrent_%d.go", i))
		content := strings.Repeat(fmt.Sprintf("// Concurrent file %d content\n", i), 50)
		if err := os.WriteFile(testFiles[i], []byte(content), 0o600); err != nil {
			b.Fatalf(testutil.TestMsgFailedToCreateTestFile, err)
		}
	}

	processor := fileproc.NewFileProcessor(tmpDir)
	ctx := context.Background()
	fileCount := len(testFiles)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			ch := make(chan fileproc.WriteRequest, 1)
			_ = processor.ProcessWithContext(ctx, testFiles[i%fileCount], ch)
			i++
		}
	})
}
