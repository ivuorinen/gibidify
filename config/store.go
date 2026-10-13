// Package config handles application configuration management.
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/ivuorinen/gibidify/shared"
)

// store is a minimal configuration store replacing the small slice of
// spf13/viper the application used. It holds two nested trees keyed by
// dot-separated paths: explicit values (from a config file or Set) shadow
// registered defaults. IsSet consults only the explicit tree, preserving the
// "validate a setting only when the user provided it" semantics the validators
// depend on. Not safe for concurrent writes — config is loaded once at startup.
type store struct {
	defaults    map[string]any
	values      map[string]any
	searchPaths []string
	file        string // path of the config file that was actually loaded
}

func newStore() *store {
	return &store{defaults: map[string]any{}, values: map[string]any{}}
}

var cfg = newStore()

// setNested assigns val at the dot-separated key, creating intermediate maps.
func setNested(root map[string]any, key string, val any) {
	parts := strings.Split(key, ".")
	m := root
	for _, p := range parts[:len(parts)-1] {
		next, ok := m[p].(map[string]any)
		if !ok {
			next = map[string]any{}
			m[p] = next
		}
		m = next
	}
	m[parts[len(parts)-1]] = val
}

// getNested returns the value at the dot-separated key, if present.
func getNested(root map[string]any, key string) (any, bool) {
	var cur any = root
	for _, p := range strings.Split(key, ".") {
		m, ok := cur.(map[string]any)
		if !ok {
			return nil, false
		}
		cur, ok = m[p]
		if !ok {
			return nil, false
		}
	}

	return cur, true
}

// lookup resolves a key against explicit values first, then defaults.
func (s *store) lookup(key string) any {
	if v, ok := getNested(s.values, key); ok {
		return v
	}
	v, _ := getNested(s.defaults, key)

	return v
}

// Set assigns an explicit value for key (making IsSet(key) report true).
func Set(key string, val any) { setNested(cfg.values, key, val) }

// SetDefault registers a fallback value for key.
func SetDefault(key string, val any) { setNested(cfg.defaults, key, val) }

// IsSet reports whether key was explicitly provided (defaults do not count).
func IsSet(key string) bool {
	_, ok := getNested(cfg.values, key)

	return ok
}

// Reset clears all values, defaults, search paths, and loaded-file state.
func Reset() { cfg = newStore() }

// AddConfigPath registers a directory to search for the config file.
func AddConfigPath(dir string) { cfg.searchPaths = append(cfg.searchPaths, dir) }

// FileUsed returns the path of the config file that was loaded, if any.
func FileUsed() string { return cfg.file }

// ReadInConfig loads the first config file found (each search path's
// config.yaml, in order) into the explicit values tree.
func ReadInConfig() error {
	candidates := make([]string, 0, len(cfg.searchPaths))
	for _, dir := range cfg.searchPaths {
		candidates = append(candidates, filepath.Join(dir, "config."+shared.FormatYAML))
	}

	for _, path := range candidates {
		data, err := os.ReadFile(filepath.Clean(path))
		if err != nil {
			if os.IsNotExist(err) {
				continue // try the next candidate
			}

			return fmt.Errorf("reading config file %s: %w", path, err)
		}
		parsed := map[string]any{}
		if err := yaml.Unmarshal(data, &parsed); err != nil {
			return fmt.Errorf("parsing config file %s: %w", path, err)
		}
		cfg.values = parsed
		cfg.file = path

		return nil
	}

	return fmt.Errorf("%w in %v", errNoConfigFile, candidates)
}

// errNoConfigFile reports that no candidate config file exists, as opposed to
// one that exists but cannot be read or parsed.
var errNoConfigFile = errors.New("no config file found")

// knownKeys lists every configuration key the application reads. The store
// keeps no schema, so without this list a typo or a removed setting in a config
// file was dropped without any message.
var knownKeys = []string{
	shared.ConfigKeyFileSizeLimit,
	shared.ConfigKeyMaxConcurrency,
	shared.ConfigKeyIgnoreDirectories,
	shared.ConfigKeyFileTypesEnabled,
	shared.ConfigKeyFileTypesCustomImageExtensions,
	shared.ConfigKeyFileTypesCustomBinaryExtensions,
	shared.ConfigKeyFileTypesCustomLanguages,
	shared.ConfigKeyFileTypesDisabledImageExtensions,
	shared.ConfigKeyFileTypesDisabledBinaryExtensions,
	shared.ConfigKeyFileTypesDisabledLanguageExts,
	shared.ConfigKeyResourceLimitsMaxFiles,
	shared.ConfigKeyResourceLimitsMaxTotalSize,
}

// unknownKeys returns the sorted dot-separated paths of explicitly set values
// that are neither known keys nor parents of known keys.
func unknownKeys() []string {
	var unknown []string

	var walk func(prefix string, m map[string]any)
	walk = func(prefix string, m map[string]any) {
		for k, v := range m {
			path := k
			if prefix != "" {
				path = prefix + "." + k
			}
			if slices.Contains(knownKeys, path) {
				continue
			}
			if child, ok := v.(map[string]any); ok && isKnownParent(path) {
				walk(path, child)

				continue
			}
			unknown = append(unknown, path)
		}
	}
	walk("", cfg.values)
	slices.Sort(unknown)

	return unknown
}

// isKnownParent reports whether path is a section containing known keys.
func isKnownParent(path string) bool {
	for _, k := range knownKeys {
		if strings.HasPrefix(k, path+".") {
			return true
		}
	}

	return false
}

// GetInt returns key as an int.
func GetInt(key string) int { return toInt(cfg.lookup(key)) }

// GetInt64 returns key as an int64.
func GetInt64(key string) int64 { return toInt64(cfg.lookup(key)) }

// GetBool returns key as a bool.
func GetBool(key string) bool {
	b, ok := cfg.lookup(key).(bool)

	return ok && b
}

// GetStringSlice returns key as a []string.
func GetStringSlice(key string) []string { return toStringSlice(cfg.lookup(key)) }

// GetStringMapString returns key as a map[string]string.
func GetStringMapString(key string) map[string]string {
	return toStringMapString(cfg.lookup(key))
}

func toInt(v any) int {
	switch n := v.(type) {
	case int:
		return n
	case int64:
		return int(n)
	case float64:
		return int(n)
	default:
		return 0
	}
}

func toInt64(v any) int64 {
	switch n := v.(type) {
	case int64:
		return n
	case int:
		return int64(n)
	case float64:
		return int64(n)
	default:
		return 0
	}
}

func toStringSlice(v any) []string {
	switch x := v.(type) {
	case []string:
		return x
	case []any:
		out := make([]string, 0, len(x))
		for _, e := range x {
			out = append(out, fmt.Sprint(e))
		}

		return out
	default:
		return nil
	}
}

func toStringMapString(v any) map[string]string {
	switch x := v.(type) {
	case map[string]string:
		return x
	case map[string]any:
		out := make(map[string]string, len(x))
		for k, e := range x {
			out[k] = fmt.Sprint(e)
		}

		return out
	default:
		return map[string]string{}
	}
}
