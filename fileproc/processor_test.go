package fileproc_test

import (
	"os"
	"strings"
	"sync"
	"testing"

	fileproc "github.com/ivuorinen/gibidify/fileproc"
)

func TestProcessFile(t *testing.T) {
	// Create a temporary file with known content.
	tmpFile, err := os.CreateTemp("", "testfile")
	if err != nil {
		t.Fatal(err)
	}
	defer func(name string) {
		err := os.Remove(name)
		if err != nil {
			t.Fatal(err)
		}
	}(tmpFile.Name())

	content := "Test content"
	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatal(err)
	}
	errTmpFile := tmpFile.Close()
	if errTmpFile != nil {
		t.Fatal(errTmpFile)
		return
	}

	ch := make(chan fileproc.WriteRequest, 1)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		fileproc.ProcessFile(tmpFile.Name(), ch, "")
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
