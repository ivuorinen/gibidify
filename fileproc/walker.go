// Package fileproc provides functions for file processing.
package fileproc

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/ivuorinen/gibidify/config"
	ignore "github.com/sabhiram/go-gitignore"
)

// Walker defines an interface for scanning directories.
type Walker interface {
	Walk(root string) ([]string, error)
}

// ProdWalker implements Walker using a custom directory walker that
// respects .gitignore and .ignore files, configuration-defined ignore directories,
// and ignores binary and image files by default.
type ProdWalker struct{}

// ignoreRule holds an ignore matcher along with the base directory where it was loaded.
type ignoreRule struct {
	base string
	gi   *ignore.GitIgnore
}

// Walk scans the given root directory recursively and returns a slice of file paths
// that are not ignored based on .gitignore/.ignore files, the configuration, or the default binary/image filter.
func (pw ProdWalker) Walk(root string) ([]string, error) {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	return walkDir(absRoot, absRoot, []ignoreRule{})
}

// walkDir recursively walks the directory tree starting at currentDir.
// It loads any .gitignore and .ignore files found in each directory and
// appends the corresponding rules to the inherited list. Each file/directory is
// then checked against the accumulated ignore rules, the configuration's list of ignored directories,
// and a default filter that ignores binary and image files.
func walkDir(root string, currentDir string, parentRules []ignoreRule) ([]string, error) {
	var results []string

	entries, err := os.ReadDir(currentDir)
	if err != nil {
		return nil, err
	}

	// Start with the parent's ignore rules.
	rules := make([]ignoreRule, len(parentRules))
	copy(rules, parentRules)

	// Check for .gitignore and .ignore files in the current directory.
	for _, fileName := range []string{".gitignore", ".ignore"} {
		ignorePath := filepath.Join(currentDir, fileName)
		if info, err := os.Stat(ignorePath); err == nil && !info.IsDir() {
			gi, err := ignore.CompileIgnoreFile(ignorePath)
			if err == nil {
				rules = append(rules, ignoreRule{
					base: currentDir,
					gi:   gi,
				})
			}
		}
	}

	// Get the list of directories to ignore from configuration.
	ignoredDirs := config.GetIgnoredDirectories()
	sizeLimit := config.GetFileSizeLimit() // e.g., 5242880 for 5 MB

	for _, entry := range entries {
		fullPath := filepath.Join(currentDir, entry.Name())

		// For directories, check if its name is in the config ignore list.
		if entry.IsDir() {
			for _, d := range ignoredDirs {
				if entry.Name() == d {
					// Skip this directory entirely.
					goto SkipEntry
				}
			}
		} else {
			// Check if file exceeds the configured size limit.
			info, err := entry.Info()
			if err == nil && info.Size() > sizeLimit {
				goto SkipEntry
			}

			// For files, apply the default filter to ignore binary and image files.
			if isBinaryOrImage(fullPath) {
				goto SkipEntry
			}
		}

		// Check accumulated ignore rules.
		for _, rule := range rules {
			// Compute the path relative to the base where the ignore rule was defined.
			rel, err := filepath.Rel(rule.base, fullPath)
			if err != nil {
				continue
			}
			// If the rule matches, skip this entry.
			if rule.gi.MatchesPath(rel) {
				goto SkipEntry
			}
		}

		// If not ignored, then process the entry.
		if entry.IsDir() {
			subFiles, err := walkDir(root, fullPath, rules)
			if err != nil {
				return nil, err
			}
			results = append(results, subFiles...)
		} else {
			results = append(results, fullPath)
		}
	SkipEntry:
		continue
	}

	return results, nil
}

// isBinaryOrImage checks if a file should be considered binary or an image based on its extension.
// The check is case-insensitive.
func isBinaryOrImage(filePath string) bool {
	ext := strings.ToLower(filepath.Ext(filePath))
	// Common image file extensions.
	imageExtensions := map[string]bool{
		".png":  true,
		".jpg":  true,
		".jpeg": true,
		".gif":  true,
		".bmp":  true,
		".tiff": true,
		".ico":  true,
		".svg":  true,
		".webp": true,
	}
	// Common binary file extensions.
	binaryExtensions := map[string]bool{
		".exe":      true,
		".dll":      true,
		".so":       true,
		".bin":      true,
		".dat":      true,
		".zip":      true,
		".tar":      true,
		".gz":       true,
		".7z":       true,
		".rar":      true,
		".DS_Store": true,
	}
	if imageExtensions[ext] || binaryExtensions[ext] {
		return true
	}
	return false
}
