package cli

import (
	"strings"
	"testing"

	"github.com/ivuorinen/gibidify/shared"
	"github.com/ivuorinen/gibidify/testutil"
)

// suggestionCases lists, per error type and code, suggestions that must appear
// and one that belongs to another code and must not.
var suggestionCases = []struct {
	name    string
	errType shared.ErrorType
	code    string
	want    []string
	notWant string
}{
	{
		name:    "configuration",
		errType: shared.ErrorTypeConfiguration,
		code:    shared.CodeConfigValidation,
		want:    []string{testutil.TestSuggestionsPlain, "Fix the configuration file", "config.example.yaml"},
		notWant: "Check your command line arguments",
	},
	{
		name:    "validation concurrency",
		errType: shared.ErrorTypeValidation,
		code:    shared.CodeValidationConcurrency,
		want:    []string{"-concurrency value of at least 1", "Example: -concurrency 4"},
		notWant: "Use a supported format",
	},
	{
		name:    "validation size",
		errType: shared.ErrorTypeValidation,
		code:    shared.CodeValidationSize,
		want:    []string{"Increase file size limit", "exclude large files"},
		notWant: "Use a supported format",
	},
	{
		name:    "validation other code",
		errType: shared.ErrorTypeValidation,
		code:    shared.CodeValidationArgument,
		want:    []string{"Check your command line arguments", "Run with --help"},
		notWant: "Increase file size limit",
	},
	{
		name:    "processing file read",
		errType: shared.ErrorTypeProcessing,
		code:    shared.CodeProcessingFileRead,
		want:    []string{"Check file permissions", "not corrupted"},
		notWant: "Check if the source directory exists",
	},
	{
		name:    "processing other code",
		errType: shared.ErrorTypeProcessing,
		code:    shared.CodeProcessingEncode,
		want:    []string{"Try reducing concurrency: -concurrency 1", "system resources"},
		notWant: "not corrupted",
	},
	{
		name:    "io file create",
		errType: shared.ErrorTypeIO,
		code:    shared.CodeIOFileCreate,
		want:    []string{"destination directory exists", "Ensure sufficient disk space"},
		notWant: "Check file/directory permissions",
	},
	{
		name:    "io other code",
		errType: shared.ErrorTypeIO,
		code:    "OTHER",
		want:    []string{"Check file/directory permissions", "Verify available disk space"},
		notWant: "destination directory exists",
	},
}

// TestErrorFormatterSuggestionsByCode verifies each error type and code gets
// its own suggestions, and that an unlisted code falls back to the type's
// general advice instead of the advice for another code.
func TestErrorFormatterSuggestionsByCode(t *testing.T) {
	for _, tt := range suggestionCases {
		t.Run(tt.name, func(t *testing.T) {
			ui, output := createTestUI()
			NewErrorFormatter(ui).provideSuggestions(&shared.StructuredError{Type: tt.errType, Code: tt.code})

			got := output.String()
			testutil.VerifyContentContains(t, got, tt.want)
			if strings.Contains(got, tt.notWant) {
				t.Errorf("output contains %q, which belongs to another code:\n%s", tt.notWant, got)
			}
		})
	}
}

// TestErrorFormatterSuggestFileNotFoundListsAtMostThree verifies the similar
// file list stops after three matches.
func TestErrorFormatterSuggestFileNotFoundListsAtMostThree(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"match-a.go", "match-b.go", "match-c.go", "match-d.go", "other.go"} {
		testutil.CreateTestFile(t, dir, name, nil)
	}

	ui, output := createTestUI()
	NewErrorFormatter(ui).suggestFileNotFound(dir + "/match")

	got := output.String()
	if listed := strings.Count(got, "    - match-"); listed != 3 {
		t.Errorf("listed %d similar files, want 3:\n%s", listed, got)
	}
	if strings.Contains(got, "other.go") {
		t.Errorf("listed a file that does not match:\n%s", got)
	}
}
