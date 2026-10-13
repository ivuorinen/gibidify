package fileproc_test

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/ivuorinen/gibidify/fileproc"
	"github.com/ivuorinen/gibidify/shared"
	"github.com/ivuorinen/gibidify/testutil"
)

// errInjectedWrite is the failure tokenFailWriter injects.
var errInjectedWrite = errors.New("injected write failure")

// tokenFailWriter fails any write whose payload contains token (or, when
// exact is set, equals it), so a test can fail one specific step of a writer.
type tokenFailWriter struct {
	token string
	exact bool
	buf   bytes.Buffer
}

func (w *tokenFailWriter) Write(p []byte) (int, error) {
	payload := string(p)
	if payload == w.token || (!w.exact && strings.Contains(payload, w.token)) {
		return 0, errInjectedWrite
	}

	return w.buf.Write(p)
}

var (
	firstReq  = fileproc.WriteRequest{Path: "first.go", Content: "package first"}
	secondReq = fileproc.WriteRequest{Path: "second.go", Content: "package second"}
)

// writeErrorCase fails the writes containing (or, with exact, equal to) token
// and expects run to return the injected error wrapped with wantMsg.
type writeErrorCase struct {
	name    string
	token   string
	exact   bool
	run     func(io.Writer) error
	wantMsg string
	// written, when set, must already be in the output: it shows the
	// earlier steps succeeded and the failure is in the step under test.
	written string
}

// assertWriteErrors runs each case against a tokenFailWriter.
func assertWriteErrors(t *testing.T, cases []writeErrorCase) {
	t.Helper()
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			w := &tokenFailWriter{token: tt.token, exact: tt.exact}
			err := tt.run(w)
			// yaml.v3 flattens the writer's error into its own message, so the
			// cause is checked by text rather than errors.Is.
			var se *shared.StructuredError
			if !errors.As(err, &se) || se.Code != shared.CodeIOWrite {
				t.Fatalf("error = %v, want code %s", err, shared.CodeIOWrite)
			}
			testutil.AssertErrorContains(t, err, errInjectedWrite.Error(), tt.name)
			testutil.AssertErrorContains(t, err, tt.wantMsg, tt.name)
			if !strings.Contains(w.buf.String(), tt.written) {
				t.Fatalf("output = %q, want it to contain %q", w.buf.String(), tt.written)
			}
		})
	}
}

// TestJSONWriterReturnsWriteErrors verifies every JSON write step returns the
// underlying write error wrapped with what was being written.
func TestJSONWriterReturnsWriteErrors(t *testing.T) {
	assertWriteErrors(t, []writeErrorCase{
		{"json start", `"files":[`, false, func(w io.Writer) error {
			return fileproc.NewJSONWriter(w).Start("P", "S")
		}, "failed to write JSON start", ""},
		{"json entry", "first.go", false, func(w io.Writer) error {
			return fileproc.NewJSONWriter(w).WriteFile(firstReq)
		}, "failed to write JSON file", ""},
		{"json separator", ",", true, func(w io.Writer) error {
			jw := fileproc.NewJSONWriter(w)
			if err := jw.WriteFile(firstReq); err != nil {
				return err
			}

			return jw.WriteFile(secondReq)
		}, "failed to write JSON separator", "first.go"},
		{"json end", "]}", false, func(w io.Writer) error {
			return fileproc.NewJSONWriter(w).Close()
		}, "failed to write JSON end", ""},
	})
}

// TestMarkdownWriterReturnsWriteErrors verifies every Markdown write step
// returns the underlying write error wrapped with what was being written.
func TestMarkdownWriterReturnsWriteErrors(t *testing.T) {
	assertWriteErrors(t, []writeErrorCase{
		{"markdown prefix", "PREFIX", false, func(w io.Writer) error {
			return fileproc.NewMarkdownWriter(w).Start("PREFIX", "SUFFIX")
		}, "failed to write prefix", ""},
		{"markdown entry", "first.go", false, func(w io.Writer) error {
			return fileproc.NewMarkdownWriter(w).WriteFile(firstReq)
		}, "failed to write inline content", ""},
		{"markdown suffix", "SUFFIX", false, func(w io.Writer) error {
			mw := fileproc.NewMarkdownWriter(w)
			if err := mw.Start("PREFIX", "SUFFIX"); err != nil {
				return err
			}

			return mw.Close()
		}, "failed to write suffix", "# PREFIX"},
	})
}

// TestYAMLWriterReturnsWriteErrors verifies every YAML write step returns the
// underlying write error wrapped with what was being written.
func TestYAMLWriterReturnsWriteErrors(t *testing.T) {
	assertWriteErrors(t, []writeErrorCase{
		{"yaml header", "PREFIX", false, func(w io.Writer) error {
			return fileproc.NewYAMLWriter(w).Start("PREFIX", "SUFFIX")
		}, "failed to write YAML header", ""},
		{"yaml files key", "files:", false, func(w io.Writer) error {
			return fileproc.NewYAMLWriter(w).Start("PREFIX", "SUFFIX")
		}, "failed to write YAML header", "prefix: PREFIX"},
		{"yaml entry", "first.go", false, func(w io.Writer) error {
			return fileproc.NewYAMLWriter(w).WriteFile(firstReq)
		}, "failed to write YAML entry", ""},
	})
}

// TestMarkdownWriterEmptyPrefixSuffixWritesNothing verifies an empty prefix
// and suffix produce no header or footer.
func TestMarkdownWriterEmptyPrefixSuffixWritesNothing(t *testing.T) {
	var buf bytes.Buffer
	mw := fileproc.NewMarkdownWriter(&buf)
	if err := mw.Start("", ""); err != nil {
		t.Fatalf(testutil.TestMsgUnexpectedError, err)
	}
	if err := mw.Close(); err != nil {
		t.Fatalf(testutil.TestMsgUnexpectedError, err)
	}
	if buf.Len() != 0 {
		t.Fatalf("output = %q, want empty", buf.String())
	}
}

// TestMarkdownWriterFenceOutlastsBacktickRuns verifies the code fence is longer
// than the longest backtick run in the content, so the content cannot close it.
func TestMarkdownWriterFenceOutlastsBacktickRuns(t *testing.T) {
	tests := []struct {
		name    string
		content string
		fence   string
	}{
		{"no backticks", "plain", "```"},
		{"short runs", "a `b` c ``d``", "```"},
		{"triple run", "```go\nx\n```", "````"},
		{"longest run wins", "``` and ````` then `", "``````"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			req := fileproc.WriteRequest{Path: "doc.txt", Content: tt.content}
			if err := fileproc.NewMarkdownWriter(&buf).WriteFile(req); err != nil {
				t.Fatalf(testutil.TestMsgUnexpectedError, err)
			}
			want := "## File: `doc.txt`\n" + tt.fence + "\n" + tt.content + "\n" + tt.fence + "\n\n"
			if got := buf.String(); got != want {
				t.Fatalf("output = %q, want %q", got, want)
			}
		})
	}
}
