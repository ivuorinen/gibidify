package fileproc_test

import (
	"encoding/json"
	"os"
	"strings"
	"sync"
	"testing"

	fileproc "github.com/ivuorinen/gibidify/fileproc"
	"gopkg.in/yaml.v3"
)

func TestStartWriter_Formats(t *testing.T) {
	// Define table-driven test cases
	tests := []struct {
		name        string
		format      string
		expectError bool
	}{
		{
			name:        "JSON format",
			format:      "json",
			expectError: false,
		},
		{
			name:        "YAML format",
			format:      "yaml",
			expectError: false,
		},
		{
			name:        "Markdown format",
			format:      "markdown",
			expectError: false,
		},
		{
			name:        "Invalid format",
			format:      "invalid",
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			outFile, err := os.CreateTemp("", "gibidify_test_output")
			if err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}
			defer func() {
				if err := outFile.Close(); err != nil {
					t.Errorf("close temp file: %v", err)
				}
				if err := os.Remove(outFile.Name()); err != nil {
					t.Errorf("remove temp file: %v", err)
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
				fileproc.StartWriter(outFile, writeCh, doneCh, tc.format, "PREFIX", "SUFFIX")
			}()

			// Wait until writer signals completion
			wg.Wait()
			<-doneCh // make sure all writes finished

			// Read output
			data, err := os.ReadFile(outFile.Name())
			if err != nil {
				t.Fatalf("Error reading output file: %v", err)
			}

			if tc.expectError {
				// For an invalid format, we expect StartWriter to log an error
				// and produce no content or minimal content. There's no official
				// error returned, so check if it's empty or obviously incorrect.
				if len(data) != 0 {
					t.Errorf("Expected no output for invalid format, got:\n%s", data)
				}
			} else {
				// Valid format: check basic properties in the output
				content := string(data)
				switch tc.format {
				case "json":
					// Quick parse check
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
					// Check presence of code fences or "## File: ..."
					if !strings.Contains(content, "```") {
						t.Error("Expected markdown code fences not found")
					}
				}

				// Prefix and suffix checks (common to JSON, YAML, markdown)
				if !strings.Contains(string(data), "PREFIX") {
					t.Errorf("Missing prefix in output: %s", data)
				}
				if !strings.Contains(string(data), "SUFFIX") {
					t.Errorf("Missing suffix in output: %s", data)
				}
			}
		})
	}
}
