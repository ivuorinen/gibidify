// Package fileproc handles file processing, collection, and output formatting.
package fileproc

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"

	ignore "github.com/sabhiram/go-gitignore"

	"github.com/ivuorinen/gibidify/shared"
)

// ignoreRule holds the matchers compiled from one ignore file and the
// directory it was loaded from.
type ignoreRule struct {
	// gi matches what the file ignores, with its own negations applied.
	gi *ignore.GitIgnore
	// neg matches what the file re-includes: its "!" patterns with the "!"
	// stripped. go-gitignore reports a lone negation the same as no match, so
	// without it a deeper "!keep.log" could not override a parent's "*.log".
	// Nil when the file has no negations.
	neg  *ignore.GitIgnore
	base string
}

// loadIgnoreRules loads ignore rules from the current directory and combines
// them with parent rules. An ignore file that exists but cannot be read is an
// error: skipping it silently put every file it excludes into the output.
func loadIgnoreRules(currentDir string, parentRules []ignoreRule) ([]ignoreRule, error) {
	// Pre-allocate for parent rules plus possible .gitignore and .ignore
	const expectedIgnoreFiles = 2
	rules := make([]ignoreRule, 0, len(parentRules)+expectedIgnoreFiles)
	rules = append(rules, parentRules...)

	// Check for .gitignore and .ignore files in the current directory.
	for _, fileName := range []string{".gitignore", ".ignore"} {
		rule, err := tryLoadIgnoreFile(currentDir, fileName)
		if err != nil {
			return nil, err
		}
		if rule != nil {
			rules = append(rules, *rule)
		}
	}

	return rules, nil
}

// tryLoadIgnoreFile loads an ignore file from the given directory. It returns
// nil and no error when the file does not exist.
func tryLoadIgnoreFile(dir, fileName string) (*ignoreRule, error) {
	ignorePath := filepath.Join(dir, fileName)
	info, err := os.Stat(ignorePath)
	if errors.Is(err, os.ErrNotExist) || (err == nil && info.IsDir()) {
		// Absent file: no rule, and nothing is wrong.
		return nil, nil
	}
	if err != nil {
		return nil, wrapIgnoreFileError(err, ignorePath)
	}

	data, err := os.ReadFile(filepath.Clean(ignorePath))
	if err != nil {
		return nil, wrapIgnoreFileError(err, ignorePath)
	}

	lines := strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n")
	var negations []string
	for _, line := range lines {
		if pattern, ok := strings.CutPrefix(line, "!"); ok && pattern != "" {
			negations = append(negations, pattern)
		}
	}

	rule := &ignoreRule{gi: ignore.CompileIgnoreLines(lines...), base: dir}
	if len(negations) > 0 {
		rule.neg = ignore.CompileIgnoreLines(negations...)
	}

	return rule, nil
}

// wrapIgnoreFileError reports an ignore file that exists but cannot be read.
func wrapIgnoreFileError(err error, path string) error {
	return shared.WrapError(
		err, shared.ErrorTypeFileSystem, shared.CodeFSAccess, "cannot read ignore file",
	).WithFilePath(path)
}

// matchesIgnoreRules reports whether a path is ignored. Rules are consulted
// deepest first and the first one with an opinion decides, as in git: a
// directory's "!pattern" re-includes what a parent directory's file ignores.
func matchesIgnoreRules(fullPath string, isDir bool, rules []ignoreRule) bool {
	for _, rule := range slices.Backward(rules) {
		rel, ok := relativeMatchPath(fullPath, isDir, rule.base)
		if !ok {
			continue
		}
		if rule.gi.MatchesPath(rel) {
			return true
		}
		if rule.neg != nil && rule.neg.MatchesPath(rel) {
			return false
		}
	}

	return false
}

// relativeMatchPath returns fullPath relative to base in the slash form
// go-gitignore matches against. Directories get a trailing slash: go-gitignore
// only applies a directory pattern such as "venv/" to "venv/", so without it
// ignored directories were descended into and every file inside was matched one
// by one.
func relativeMatchPath(fullPath string, isDir bool, base string) (string, bool) {
	rel, err := filepath.Rel(base, fullPath)
	if err != nil {
		return "", false
	}
	rel = filepath.ToSlash(rel)
	if isDir {
		rel += "/"
	}

	return rel, true
}
