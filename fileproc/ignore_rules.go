// Package fileproc handles file processing, collection, and output formatting.
package fileproc

import (
	"os"
	"path/filepath"

	ignore "github.com/sabhiram/go-gitignore"
)

// ignoreRule holds an ignore matcher along with the base directory where it was loaded.
type ignoreRule struct {
	gi   *ignore.GitIgnore
	base string
}

// loadIgnoreRules loads ignore rules from the current directory and combines them with parent rules.
func loadIgnoreRules(currentDir string, parentRules []ignoreRule) []ignoreRule {
	// Pre-allocate for parent rules plus possible .gitignore and .ignore
	const expectedIgnoreFiles = 2
	rules := make([]ignoreRule, 0, len(parentRules)+expectedIgnoreFiles)
	rules = append(rules, parentRules...)

	// Check for .gitignore and .ignore files in the current directory.
	for _, fileName := range []string{".gitignore", ".ignore"} {
		if rule := tryLoadIgnoreFile(currentDir, fileName); rule != nil {
			rules = append(rules, *rule)
		}
	}

	return rules
}

// tryLoadIgnoreFile attempts to load an ignore file from the given directory.
func tryLoadIgnoreFile(dir, fileName string) *ignoreRule {
	ignorePath := filepath.Join(dir, fileName)
	if info, err := os.Stat(ignorePath); err == nil && !info.IsDir() {
		//nolint:errcheck // Regex compile error handled by validation, safe to ignore here
		if gi, err := ignore.CompileIgnoreFile(ignorePath); err == nil {
			return &ignoreRule{
				base: dir,
				gi:   gi,
			}
		}
	}

	return nil
}

// matchesIgnoreRules checks if a path matches any of the ignore rules.
func matchesIgnoreRules(fullPath string, rules []ignoreRule) bool {
	for _, rule := range rules {
		if matchesRule(fullPath, rule) {
			return true
		}
	}

	return false
}

// matchesRule checks if a path matches a specific ignore rule.
func matchesRule(fullPath string, rule ignoreRule) bool {
	// Compute the path relative to the base where the ignore rule was defined.
	rel, err := filepath.Rel(rule.base, fullPath)
	if err != nil {
		return false
	}
	// If the rule matches, skip this entry.
	return rule.gi.MatchesPath(rel)
}
