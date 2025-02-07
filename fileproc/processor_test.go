package fileproc

import (
	"os"
	"strings"
	"sync"
	"testing"
)

func TestProcessFile(t *testing.T) {
	// Create a temporary file with known content.
	tmpFile, err := os.CreateTemp("", "testfile")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	content := "Test content"
	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()

	ch := make(chan WriteRequest, 1)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		ProcessFile(tmpFile.Name(), ch, nil)
	}()
	wg.Wait()
	close(ch)

	var result string
	for req := range ch {
		result = req.Content
	}

	if !strings.Contains(result, tmpFile.Name()) {
		t.Errorf("Output does not contain file path: %s", tmpFile.Name())
	}
	if !strings.Contains(result, content) {
		t.Errorf("Output does not contain file content: %s", content)
	}
}
