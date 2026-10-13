// Package fileproc provides functions for file processing.
package fileproc

import (
	"os"
	"path/filepath"

	"github.com/ivuorinen/gibidify/shared"
)

// ProdWalker is a directory walker that respects .gitignore and .ignore files,
// configuration-defined ignore directories, and ignores binary, image and
// credential files by default.
type ProdWalker struct {
	filter *FileFilter
}

// NewProdWalker creates a new production walker with current configuration.
// opts.Exclude lists absolute paths never returned, such as the output file:
// when the destination lies inside the source tree, the previous run's output
// would otherwise be ingested as input.
func NewProdWalker(opts WalkOptions) *ProdWalker {
	return &ProdWalker{
		filter: NewFileFilter(opts),
	}
}

// Walk scans the given root directory recursively and returns a slice of file paths
// that are not ignored based on .gitignore/.ignore files, the configuration, or the default binary/image filter.
func (w *ProdWalker) Walk(root string) ([]string, error) {
	absRoot, err := shared.AbsolutePath(root)
	if err != nil {
		return nil, shared.WrapError(
			err,
			shared.ErrorTypeFileSystem,
			shared.CodeFSPathResolution,
			"failed to resolve root path",
		).WithFilePath(root)
	}

	return w.walkDir(absRoot, []ignoreRule{})
}

// Skipped returns how many files the walk left out for exceeding the size limit
// or as credential files. Each one was also logged at Warn.
func (w *ProdWalker) Skipped() int {
	return w.filter.skipped
}

// walkDir recursively walks the directory tree starting at currentDir.
// It loads any .gitignore and .ignore files found in each directory and
// appends the corresponding rules to the inherited list. Each file/directory is
// then checked against the accumulated ignore rules, the configuration's list of ignored directories,
// and a default filter that ignores binary and image files.
//
// Only an unreadable currentDir, or an unreadable ignore file in it, at the
// root fails the walk. Below the root the directory is logged and skipped: one
// root-owned cache directory must not make the whole tree unusable, and a
// directory whose ignore file cannot be applied must not be collected as if it
// had none.
func (w *ProdWalker) walkDir(currentDir string, parentRules []ignoreRule) ([]string, error) {
	var results []string

	entries, err := os.ReadDir(currentDir)
	if err != nil {
		return nil, shared.WrapError(
			err,
			shared.ErrorTypeFileSystem,
			shared.CodeFSAccess,
			"failed to read directory",
		).WithFilePath(currentDir)
	}

	rules, err := loadIgnoreRules(currentDir, parentRules)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		fullPath := filepath.Join(currentDir, entry.Name())

		if w.filter.shouldSkipEntry(entry, fullPath, rules) {
			continue
		}

		if !entry.IsDir() {
			results = append(results, fullPath)

			continue
		}

		subFiles, err := w.walkDir(fullPath, rules)
		if err != nil {
			shared.GetLogger().Warnf("Skipping unreadable directory %s: %v", fullPath, err)

			continue
		}
		results = append(results, subFiles...)
	}

	return results, nil
}
