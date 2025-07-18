package fileproc_test

import (
	"encoding/json"
	"os"
	"strings"
	"sync"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/ivuorinen/gibidify/fileproc"
)

func TestStartWriter_Formats(t *testing.T) {
	// Define table-driven test cases
	tests := []struct {
		name        string
		format      string
		expectError bool
	}{
		{"JSON format", "json", false},
		{"YAML format", "yaml", false},
		{"Markdown format", "markdown", false},
		{"Invalid format", "invalid", true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			data := runWriterTest(t, tc.format)
			if tc.expectError {
				verifyErrorOutput(t, data)
			} else {
				verifyValidOutput(t, data, tc.format)
				verifyPrefixSuffix(t, data)
			}
		})
	}
}

// runWriterTest executes the writer with the given format and returns the output data.
func runWriterTest(t *testing.T, format string) []byte {
	t.Helper()
	outFile, err := os.CreateTemp(t.TempDir(), "gibidify_test_output")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer func() {
		if closeErr := outFile.Close(); closeErr != nil {
			t.Errorf("close temp file: %v", closeErr)
		}
		if removeErr := os.Remove(outFile.Name()); removeErr != nil {
			t.Errorf("remove temp file: %v", removeErr)
		}
	}()

	// Prepare channels
	writeCh := make(chan fileproc.WriteRequest, 2)
	doneCh := make(chan struct{})

	// Write a couple of sample requests
	writeCh <- fileproc.WriteRequest{Path: "sample.go", Content: "package main"}
	writeCh <- fileproc.WriteRequest{Path: "example.py", Content: "def foo(): pass"}
	close(writeCh)

	// Start the writer
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		fileproc.StartWriter(outFile, writeCh, doneCh, format, "PREFIX", "SUFFIX")
	}()

	// Wait until writer signals completion
	wg.Wait()
	<-doneCh // make sure all writes finished

	// Read output
	data, err := os.ReadFile(outFile.Name())
	if err != nil {
		t.Fatalf("Error reading output file: %v", err)
	}

	return data
}

// verifyErrorOutput checks that error cases produce no output.
func verifyErrorOutput(t *testing.T, data []byte) {
	t.Helper()
	if len(data) != 0 {
		t.Errorf("Expected no output for invalid format, got:\n%s", data)
	}
}

// verifyValidOutput checks format-specific output validity.
func verifyValidOutput(t *testing.T, data []byte, format string) {
	t.Helper()
	content := string(data)
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
		if !strings.Contains(content, "```") {
			t.Error("Expected markdown code fences not found")
		}
	}
}

// verifyPrefixSuffix checks that output contains expected prefix and suffix.
func verifyPrefixSuffix(t *testing.T, data []byte) {
	t.Helper()
	content := string(data)
	if !strings.Contains(content, "PREFIX") {
		t.Errorf("Missing prefix in output: %s", data)
	}
	if !strings.Contains(content, "SUFFIX") {
		t.Errorf("Missing suffix in output: %s", data)
	}
}
