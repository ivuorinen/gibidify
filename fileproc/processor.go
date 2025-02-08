// Package fileproc provides functions for processing files.
package fileproc

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

// WriteRequest represents the content to be written.
type WriteRequest struct {
	Path    string
	Content string
}

// ProcessFile reads the file at filePath and sends a formatted output to outCh.
func ProcessFile(filePath string, outCh chan<- WriteRequest, rootPath string) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		logrus.Errorf("Failed to read file %s: %v", filePath, err)
		return
	}

	// Compute path relative to rootPath, so /a/b/c/d.c becomes c/d.c
	relPath, err := filepath.Rel(rootPath, filePath)
	if err != nil {
		// Fallback if something unexpected happens
		relPath = filePath
	}

	// Format: separator, then relative path, then content
	formatted := fmt.Sprintf("\n---\n%s\n%s\n", relPath, string(content))
	outCh <- WriteRequest{Path: relPath, Content: formatted}
}
