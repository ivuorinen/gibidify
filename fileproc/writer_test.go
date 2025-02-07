package fileproc

import (
	"bytes"
	"sync"
	"testing"
)

func TestStartWriter(t *testing.T) {
	var buf bytes.Buffer

	writeCh := make(chan WriteRequest)
	done := make(chan struct{})

	go StartWriter(&buf, writeCh, done)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		writeCh <- WriteRequest{Content: "Hello"}
		writeCh <- WriteRequest{Content: " World"}
	}()
	wg.Wait()
	close(writeCh)
	<-done

	if buf.String() != "Hello World" {
		t.Errorf("Expected 'Hello World', got '%s'", buf.String())
	}
}
