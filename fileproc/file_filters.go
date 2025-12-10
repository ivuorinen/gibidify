// Package fileproc handles file processing, collection, and output formatting.
package fileproc

import (
	"os"

	"github.com/ivuorinen/gibidify/config"
)

// FileFilter defines filtering criteria for files and directories.
type FileFilter struct {
	ignoredDirs []string
	sizeLimit   int64
}

// NewFileFilter creates a new file filter with current configuration.
func NewFileFilter() *FileFilter {
	return &FileFilter{
		ignoredDirs: config.IgnoredDirectories(),
		sizeLimit:   config.FileSizeLimit(),
	}
}

// shouldSkipEntry determines if an entry should be skipped based on ignore rules and filters.
func (f *FileFilter) shouldSkipEntry(entry os.DirEntry, fullPath string, rules []ignoreRule) bool {
	if entry.IsDir() {
		return f.shouldSkipDirectory(entry)
	}

	if f.shouldSkipFile(entry, fullPath) {
		return true
	}

	return matchesIgnoreRules(fullPath, rules)
}

// shouldSkipDirectory checks if a directory should be skipped based on the ignored directories list.
func (f *FileFilter) shouldSkipDirectory(entry os.DirEntry) bool {
	for _, d := range f.ignoredDirs {
		if entry.Name() == d {
			return true
		}
	}

	return false
}

// shouldSkipFile checks if a file should be skipped based on size limit and file type.
func (f *FileFilter) shouldSkipFile(entry os.DirEntry, fullPath string) bool {
	// Check if file exceeds the configured size limit.
	if info, err := entry.Info(); err == nil && info.Size() > f.sizeLimit {
		return true
	}

	// Apply the default filter to ignore binary and image files.
	return IsBinary(fullPath) || IsImage(fullPath)
}
