// Package fileproc provides functions for writing file contents concurrently.
package fileproc

import (
	"io"
	"os"

	"github.com/sirupsen/logrus"
)

// StartWriter listens on the write channel and writes content to outFile.
// When finished, it signals on the done channel.
func StartWriter(outFile *os.File, writeCh <-chan WriteRequest, done chan<- struct{}) {
	writer := io.Writer(outFile)
	for req := range writeCh {
		if _, err := writer.Write([]byte(req.Content)); err != nil {
			logrus.Errorf("Error writing to file: %v", err)
		}
	}
	done <- struct{}{}
}
