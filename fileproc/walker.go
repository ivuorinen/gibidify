// Package fileproc provides functions for file processing.
package fileproc

import (
	"github.com/boyter/gocodewalker"
	"github.com/sirupsen/logrus"
)

// Walker defines an interface for scanning directories.
type Walker interface {
	Walk(root string) ([]string, error)
}

// ProdWalker implements Walker using gocodewalker.
type ProdWalker struct{}

// Walk scans the given root directory using gocodewalker and returns a slice of file paths.
func (pw ProdWalker) Walk(root string) ([]string, error) {
	fileListQueue := make(chan *gocodewalker.File, 100)
	fileWalker := gocodewalker.NewFileWalker(root, fileListQueue)

	errorHandler := func(err error) bool {
		logrus.Errorf("error walking directory: %s", err.Error())
		return true
	}
	fileWalker.SetErrorHandler(errorHandler)
	go func() {
		err := fileWalker.Start()
		if err != nil {
			logrus.Errorf("error walking directory: %s", err.Error())
		}
	}()

	var files []string
	for f := range fileListQueue {
		files = append(files, f.Location)
	}

	return files, nil
}
