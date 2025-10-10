// Package fileproc provides functions for file processing.
package fileproc

import (
	"os"
	"path/filepath"

	"github.com/ivuorinen/gibidify/gibidiutils"
)

// Walker defines an interface for scanning directories.
type Walker interface {
	Walk(root string) ([]string, error)
}

// ProdWalker implements Walker using a custom directory walker that
// respects .gitignore and .ignore files, configuration-defined ignore directories,
// and ignores binary and image files by default.
type ProdWalker struct {
	filter *FileFilter
}

// NewProdWalker creates a new production walker with current configuration.
func NewProdWalker() *ProdWalker {
	return &ProdWalker{
		filter: NewFileFilter(),
	}
}

// Walk scans the given root directory recursively and returns a slice of file paths
// that are not ignored based on .gitignore/.ignore files, the configuration, or the default binary/image filter.
func (w *ProdWalker) Walk(root string) ([]string, error) {
	absRoot, err := gibidiutils.GetAbsolutePath(root)
	if err != nil {
		return nil, gibidiutils.WrapError(
			err, gibidiutils.ErrorTypeFileSystem, gibidiutils.CodeFSPathResolution,
			"failed to resolve root path",
		).WithFilePath(root)
	}
	return w.walkDir(absRoot, []ignoreRule{})
}

// walkDir recursively walks the directory tree starting at currentDir.
// It loads any .gitignore and .ignore files found in each directory and
// appends the corresponding rules to the inherited list. Each file/directory is
// then checked against the accumulated ignore rules, the configuration's list of ignored directories,
// and a default filter that ignores binary and image files.
func (w *ProdWalker) walkDir(currentDir string, parentRules []ignoreRule) ([]string, error) {
	var results []string

	entries, err := os.ReadDir(currentDir)
	if err != nil {
		return nil, gibidiutils.WrapError(
			err, gibidiutils.ErrorTypeFileSystem, gibidiutils.CodeFSAccess,
			"failed to read directory",
		).WithFilePath(currentDir)
	}

	rules := loadIgnoreRules(currentDir, parentRules)

	for _, entry := range entries {
		fullPath := filepath.Join(currentDir, entry.Name())

		if w.filter.shouldSkipEntry(entry, fullPath, rules) {
			continue
		}

		// Process entry
		if entry.IsDir() {
			subFiles, err := w.walkDir(fullPath, rules)
			if err != nil {
				return nil, gibidiutils.WrapError(
					err, gibidiutils.ErrorTypeProcessing, gibidiutils.CodeProcessingTraversal,
					"failed to traverse subdirectory",
				).WithFilePath(fullPath)
			}
			results = append(results, subFiles...)
		} else {
			results = append(results, fullPath)
		}
	}

	return results, nil
}
