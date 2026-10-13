package fileproc_test

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/ivuorinen/gibidify/fileproc"
	"github.com/ivuorinen/gibidify/testutil"
)

// writeAll runs StartWriter over reqs and returns the output and writer error.
func writeAll(t *testing.T, format, prefix, suffix string, reqs ...fileproc.WriteRequest) ([]byte, error) {
	t.Helper()
	outFile, err := os.CreateTemp(t.TempDir(), "gibidify_test_output")
	if err != nil {
		t.Fatalf(testutil.TestMsgFailedToCreateFile, err)
	}
	defer func() {
		if closeErr := outFile.Close(); closeErr != nil {
			t.Errorf("close temp file: %v", closeErr)
		}
	}()

	writeCh := make(chan fileproc.WriteRequest, len(reqs))
	for _, req := range reqs {
		writeCh <- req
	}
	close(writeCh)

	writeErr := fileproc.StartWriter(outFile, writeCh, format, prefix, suffix)

	data, err := os.ReadFile(outFile.Name())
	if err != nil {
		t.Fatalf("Error reading output file: %v", err)
	}

	return data, writeErr
}

func sampleRequests() []fileproc.WriteRequest {
	return []fileproc.WriteRequest{
		{Path: "sample.go", Content: testutil.LiteralPackageMain},
		{Path: "example.py", Content: "def foo(): pass"},
	}
}

func TestStartWriterFormats(t *testing.T) {
	for _, format := range []string{"json", "yaml", "markdown"} {
		t.Run(format, func(t *testing.T) {
			data, err := writeAll(t, format, "PREFIX", "SUFFIX", sampleRequests()...)
			if err != nil {
				t.Fatalf(testutil.TestMsgUnexpectedError, err)
			}
			verifyValidOutput(t, data, format)
			verifyPrefixSuffixWith(t, data, "PREFIX", "SUFFIX")
		})
	}
}

// TestStartWriterInvalidFormat verifies an unsupported format returns an error,
// writes nothing and still drains the channel.
func TestStartWriterInvalidFormat(t *testing.T) {
	data, err := writeAll(t, "invalid", "PREFIX", "SUFFIX", sampleRequests()...)
	if err == nil {
		t.Fatal(testutil.TestMsgExpectedError)
	}
	if len(data) != 0 {
		t.Errorf("Expected no output for invalid format, got:\n%s", data)
	}
}

// verifyValidOutput checks format-specific output validity.
func verifyValidOutput(t *testing.T, data []byte, format string) {
	t.Helper()
	switch format {
	case "json":
		var outStruct fileproc.OutputData
		if err := json.Unmarshal(data, &outStruct); err != nil {
			t.Errorf("JSON unmarshal failed: %v", err)
		}
	case "yaml":
		var outStruct fileproc.OutputData
		if err := yaml.Unmarshal(data, &outStruct); err != nil {
			t.Errorf("YAML unmarshal failed: %v", err)
		}
	case "markdown":
		if !strings.Contains(string(data), "```") {
			t.Error("Expected markdown code fences not found")
		}
	default:
		t.Fatalf("verifyValidOutput: unknown format %q", format)
	}
}

// verifyPrefixSuffixWith checks that output contains expected custom prefix and suffix.
func verifyPrefixSuffixWith(t *testing.T, data []byte, expectedPrefix, expectedSuffix string) {
	t.Helper()
	content := string(data)
	if !strings.Contains(content, expectedPrefix) {
		t.Errorf("Missing prefix '%s' in output: %s", expectedPrefix, data)
	}
	if !strings.Contains(content, expectedSuffix) {
		t.Errorf("Missing suffix '%s' in output: %s", expectedSuffix, data)
	}
}

// TestStructuredOutputRoundTrips verifies JSON and YAML output parse back to
// exactly the input: content bytes, paths, and a multi-line prefix. The cases
// cover YAML block-scalar edges (leading indentation, missing or repeated
// trailing newlines, empty content) and the characters YAML forbids or folds
// in a block scalar (ESC, form feed, bare CR, CRLF, DEL, NEL, BOM), each of
// which made the hand-written YAML output unparseable or altered it.
func TestStructuredOutputRoundTrips(t *testing.T) {
	reqs := []fileproc.WriteRequest{
		{Path: "plain.go", Content: "package main\n"},
		{Path: "indented.py", Content: "    x = 1\n  y = 2\n"},
		{Path: "no-newline.txt", Content: "last line"},
		{Path: "blank-tail.txt", Content: "a\n\n\n"},
		{Path: "empty.txt", Content: ""},
		{Path: "with space: and colon.md", Content: "# Title\n"},
		{Path: "ansi.txt", Content: "x\x1b[31mred\x1b[0m\n"},
		{Path: "formfeed.c", Content: "int a;\n\f\nint b;\n"},
		{Path: "bare-cr.txt", Content: "a\rb\n"},
		{Path: "crlf.txt", Content: "a\r\nb\r\n"},
		{Path: "del.txt", Content: "x\x7fy\n"},
		{Path: "nel.txt", Content: "x\u0085y\n"},
		{Path: "bom.txt", Content: "\xef\xbb\xbfhead\n"},
		{Path: "html.java", Content: "List<String> a && b > c\n"},
	}
	prefix, suffix := "line1\nline2 \"quoted\" \x1b", "tab\there\r"

	unmarshalers := map[string]func([]byte, any) error{"json": json.Unmarshal, "yaml": yaml.Unmarshal}
	for format, unmarshal := range unmarshalers {
		t.Run(format, func(t *testing.T) {
			data, err := writeAll(t, format, prefix, suffix, reqs...)
			if err != nil {
				t.Fatalf(testutil.TestMsgUnexpectedError, err)
			}

			var out fileproc.OutputData
			if err := unmarshal(data, &out); err != nil {
				t.Fatalf("unmarshal failed: %v\n%s", err, data)
			}
			if out.Prefix != prefix || out.Suffix != suffix {
				t.Errorf("prefix/suffix = %q/%q, want %q/%q", out.Prefix, out.Suffix, prefix, suffix)
			}
			assertFilesEqual(t, out.Files, reqs)
		})
	}
}

// TestJSONOutputKeepsHTMLCharacters verifies <, > and & are written literally.
// json.Marshal's HTML escaping turned each into a six-byte \u escape, which
// inflated LLM-bound output for generics, JSX and shell redirects.
func TestJSONOutputKeepsHTMLCharacters(t *testing.T) {
	content := "List<String> a && b > c\n"
	data, err := writeAll(t, "json", "<p>", "&",
		fileproc.WriteRequest{Path: "a.java", Content: content})
	if err != nil {
		t.Fatalf(testutil.TestMsgUnexpectedError, err)
	}

	// Built from parts so the source holds no \u sequence an editor could fold.
	for _, code := range []string{"003c", "003e", "0026"} {
		escaped := `\` + "u" + code
		if strings.Contains(string(data), escaped) {
			t.Errorf("output contains %s: %s", escaped, data)
		}
	}
	if !strings.Contains(string(data), `List<String> a && b > c`) {
		t.Errorf("content not written literally: %s", data)
	}
	if !json.Valid(data) {
		t.Errorf("output is not valid JSON: %s", data)
	}
}

// assertFilesEqual checks parsed output files match the requests exactly, in order.
func assertFilesEqual(t *testing.T, got []fileproc.FileData, want []fileproc.WriteRequest) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("got %d files, want %d", len(got), len(want))
	}
	for i, req := range want {
		if got[i].Path != req.Path {
			t.Errorf(testutil.TestFmtExpectedFilePath, req.Path, got[i].Path)
		}
		if got[i].Content != req.Content {
			t.Errorf("%s: "+testutil.TestFmtExpectedContent, req.Path, req.Content, got[i].Content)
		}
	}
}

// TestStartWriterReadOnlyFileReturnsError verifies a write failure is returned,
// not just logged.
func TestStartWriterReadOnlyFileReturnsError(t *testing.T) {
	outPath := filepath.Join(t.TempDir(), "readonly_out")
	if err := os.WriteFile(outPath, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	outFile, err := os.Open(filepath.Clean(outPath))
	if err != nil {
		t.Fatalf("Failed to reopen as read-only: %v", err)
	}
	defer func() { _ = outFile.Close() }()

	writeCh := make(chan fileproc.WriteRequest, 1)
	writeCh <- fileproc.WriteRequest{Path: testutil.TestFileGo, Content: testutil.LiteralPackageMain}
	close(writeCh)

	if err := fileproc.StartWriter(outFile, writeCh, "json", "PREFIX", "SUFFIX"); err == nil {
		t.Fatal(testutil.TestMsgExpectedError)
	}
}

// TestStartWriterDrainsAfterStartFailure guards against the deadlock where a
// failed Start returned without reading writeCh, blocking every producer.
func TestStartWriterDrainsAfterStartFailure(t *testing.T) {
	outPath := filepath.Join(t.TempDir(), "readonly_out")
	if err := os.WriteFile(outPath, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	outFile, err := os.Open(filepath.Clean(outPath))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = outFile.Close() }()

	writeCh := make(chan fileproc.WriteRequest) // unbuffered: every send needs the writer
	producerDone := make(chan struct{})
	go func() {
		defer close(producerDone)
		defer close(writeCh)
		for i := range 200 {
			writeCh <- fileproc.WriteRequest{Path: fmt.Sprintf("f%d.go", i), Content: "x"}
		}
	}()

	writerErr := make(chan error, 1)
	go func() { writerErr <- fileproc.StartWriter(outFile, writeCh, "json", "PREFIX", "SUFFIX") }()

	select {
	case <-producerDone:
	case <-time.After(5 * time.Second):
		t.Fatal("producer blocked: writer stopped draining after Start failed")
	}
	if err := <-writerErr; err == nil {
		t.Fatal(testutil.TestMsgExpectedError)
	}
}

// TestMarkdownWriterSpecialCharacters verifies unusual paths and content are written.
func TestMarkdownWriterSpecialCharacters(t *testing.T) {
	data, err := writeAll(t, "markdown", "", "",
		fileproc.WriteRequest{Path: "special\ncharacters.md", Content: "Content with\ttabs"},
		fileproc.WriteRequest{Path: "empty.md", Content: ""},
	)
	if err != nil {
		t.Fatalf(testutil.TestMsgUnexpectedError, err)
	}
	if !strings.Contains(string(data), "Content with\ttabs") {
		t.Errorf("content missing from output:\n%s", data)
	}
}

// Benchmarks for writer performance

// BenchmarkStartWriter benchmarks basic writer operations across formats.
func BenchmarkStartWriter(b *testing.B) {
	formats := []string{"json", "yaml", "markdown"}

	for _, format := range formats {
		b.Run(format, func(b *testing.B) {
			for b.Loop() {
				outFile, err := os.CreateTemp(b.TempDir(), "bench_output_*")
				if err != nil {
					b.Fatalf("Failed to create temp file: %v", err)
				}

				writeCh := make(chan fileproc.WriteRequest, 2)
				writeCh <- fileproc.WriteRequest{Path: "sample.go", Content: testutil.LiteralPackageMain}
				writeCh <- fileproc.WriteRequest{Path: "example.py", Content: "def foo(): pass"}
				close(writeCh)

				_ = fileproc.StartWriter(outFile, writeCh, format, "PREFIX", "SUFFIX")
				_ = outFile.Close()
			}
		})
	}
}
