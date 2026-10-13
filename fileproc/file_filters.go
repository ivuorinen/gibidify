// Package fileproc handles file processing, collection, and output formatting.
package fileproc

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/ivuorinen/gibidify/config"
	"github.com/ivuorinen/gibidify/shared"
)

// vcsDirs are version-control metadata directories, skipped regardless of
// ignoreDirectories. Setting ignoreDirectories replaces its default list, so
// keeping ".git" there re-included .git/ (whose config holds remote URLs that
// often carry tokens) as soon as a user added one directory of their own.
var vcsDirs = map[string]bool{".git": true, ".hg": true, ".svn": true}

// sensitiveNames are exact file names that hold credentials.
var sensitiveNames = map[string]bool{".env": true, ".npmrc": true, ".netrc": true, ".pgpass": true}

// sensitiveExtensions are key and certificate-bundle extensions.
var sensitiveExtensions = map[string]bool{".pem": true, ".key": true, ".p12": true, ".pfx": true}

// sensitivePrefixes are name prefixes of SSH private keys and dotenv variants
// (.env.local, .env.production).
var sensitivePrefixes = []string{".env.", "id_rsa", "id_dsa", "id_ecdsa", "id_ed25519"}

// WalkOptions configures a walk.
type WalkOptions struct {
	// Exclude lists absolute paths never returned, such as the output file.
	Exclude []string
	// IncludeSensitive keeps credential files (.env, *.pem, id_rsa, …) that
	// are skipped by default.
	IncludeSensitive bool
}

// FileFilter defines filtering criteria for files and directories.
type FileFilter struct {
	ignoredDirs      map[string]bool
	sizeLimit        int64
	includeSensitive bool
	excluded         map[string]struct{}
	// excludedFiles holds the excluded paths that exist, compared by identity:
	// a destination named through a symlinked root, a bind mount or a hard link
	// has a different path string than the walker produces for it.
	excludedFiles []os.FileInfo
	// skipped counts files left out for size or as sensitive. The walk is
	// single-goroutine, so it needs no synchronization.
	skipped int
}

// NewFileFilter creates a new file filter with current configuration.
// Excluded paths are compared after filepath.Clean, and by file identity when
// they exist.
func NewFileFilter(opts WalkOptions) *FileFilter {
	excluded := make(map[string]struct{}, len(opts.Exclude))
	var excludedFiles []os.FileInfo
	for _, p := range opts.Exclude {
		excluded[filepath.Clean(p)] = struct{}{}
		if info, err := os.Stat(p); err == nil {
			excludedFiles = append(excludedFiles, info)
		}
	}

	// Trimmed like validation trims them, so " .venv" matches ".venv".
	ignoredDirs := make(map[string]bool)
	for _, d := range config.IgnoredDirectories() {
		ignoredDirs[strings.TrimSpace(d)] = true
	}

	return &FileFilter{
		ignoredDirs:      ignoredDirs,
		sizeLimit:        config.FileSizeLimit(),
		includeSensitive: opts.IncludeSensitive,
		excluded:         excluded,
		excludedFiles:    excludedFiles,
	}
}

// shouldSkipEntry determines if an entry should be skipped based on ignore rules and filters.
//
// Symlinks are always skipped: following one would let a link planted in a
// scanned repository pull files from outside the tree (credentials, keys) into
// output that is pasted into an LLM. Other non-regular files (FIFOs, sockets,
// devices) are skipped because reading them can block forever.
func (f *FileFilter) shouldSkipEntry(entry os.DirEntry, fullPath string, rules []ignoreRule) bool {
	if entry.Type()&os.ModeSymlink != 0 {
		return true
	}

	if entry.IsDir() {
		return f.shouldSkipDirectory(entry) || matchesIgnoreRules(fullPath, true, rules)
	}

	if !entry.Type().IsRegular() || f.isExcluded(entry, fullPath) || f.shouldSkipFile(entry, fullPath) {
		return true
	}

	return matchesIgnoreRules(fullPath, false, rules)
}

// isExcluded reports whether the file is one of the excluded paths, by name or
// by identity.
func (f *FileFilter) isExcluded(entry os.DirEntry, fullPath string) bool {
	if _, ok := f.excluded[fullPath]; ok {
		return true
	}
	if len(f.excludedFiles) == 0 {
		return false
	}
	info, err := entry.Info()
	if err != nil {
		return false
	}
	for _, ex := range f.excludedFiles {
		if os.SameFile(info, ex) {
			return true
		}
	}

	return false
}

// shouldSkipDirectory reports whether a directory is VCS metadata or listed in
// ignoreDirectories.
func (f *FileFilter) shouldSkipDirectory(entry os.DirEntry) bool {
	return vcsDirs[entry.Name()] || f.ignoredDirs[entry.Name()]
}

// shouldSkipFile checks if a file should be skipped based on size limit, file
// type and the sensitive-file rules.
//
// Oversized and sensitive files are logged at Warn and counted: each is a file
// the user may expect in the output, and dropping it without a word left the
// LLM answering from absent context while the summary claimed nothing was
// skipped.
func (f *FileFilter) shouldSkipFile(entry os.DirEntry, fullPath string) bool {
	if info, err := entry.Info(); err == nil && info.Size() > f.sizeLimit {
		f.skipped++
		shared.GetLogger().Warnf(
			"Skipping %s: size %d bytes exceeds fileSizeLimit %d", fullPath, info.Size(), f.sizeLimit,
		)

		return true
	}

	if !f.includeSensitive && isSensitiveName(entry.Name()) {
		f.skipped++
		shared.GetLogger().Warnf("Skipping %s: it may hold credentials (use -include-sensitive to keep it)", fullPath)

		return true
	}

	// Apply the default filter to ignore binary and image files.
	reg := DefaultRegistry()

	return reg.IsBinary(fullPath) || reg.IsImage(fullPath)
}

// isSensitiveName reports whether a file name matches a credential-file rule.
func isSensitiveName(name string) bool {
	lower := strings.ToLower(name)
	if sensitiveNames[lower] || sensitiveExtensions[filepath.Ext(lower)] {
		return true
	}
	for _, prefix := range sensitivePrefixes {
		if strings.HasPrefix(lower, prefix) {
			return true
		}
	}

	return false
}
