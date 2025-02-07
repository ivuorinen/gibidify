// Package fileproc provides functions for processing files.
package fileproc

import (
	"fmt"
	"io/ioutil"

	"github.com/sirupsen/logrus"
)

// WriteRequest represents the content to be written.
type WriteRequest struct {
	Content string
}

// ProcessFile reads the file at filePath and sends a formatted output to outCh.
// The optional wg parameter is used when the caller wants to wait on file-level processing.
func ProcessFile(filePath string, outCh chan<- WriteRequest, wg *interface{}) {
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		logrus.Errorf("Failed to read file %s: %v", filePath, err)
		return
	}
	// Format: separator, file path, then content.
	formatted := fmt.Sprintf("\n---\n%s\n%s\n", filePath, string(content))
	outCh <- WriteRequest{Content: formatted}
}
