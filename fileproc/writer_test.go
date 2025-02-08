package fileproc

import (
	"encoding/json"
	"os"
	"testing"
)

func TestStartWriter_JSONOutput(t *testing.T) {
	outFile, err := os.CreateTemp("", "output.json")
	if err != nil {
		t.Fatal(err)
	}
	defer func(name string) {
		err := os.Remove(name)
		if err != nil {
			t.Fatal(err)
		}
	}(outFile.Name())

	writeCh := make(chan WriteRequest)
	done := make(chan struct{})

	go StartWriter(outFile, writeCh, done, "json", "Prefix", "Suffix")

	writeCh <- WriteRequest{Path: "file1.go", Content: "package main"}
	writeCh <- WriteRequest{Path: "file2.py", Content: "def hello(): print('Hello')"}

	close(writeCh)
	<-done

	data, err := os.ReadFile(outFile.Name())
	if err != nil {
		t.Fatal(err)
	}

	var output OutputData
	if err := json.Unmarshal(data, &output); err != nil {
		t.Fatalf("JSON output is invalid: %v", err)
	}

	if len(output.Files) != 2 {
		t.Errorf("Expected 2 files, got %d", len(output.Files))
	}
}
