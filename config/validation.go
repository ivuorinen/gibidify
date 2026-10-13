// Package config handles application configuration management.
package config

import (
	"fmt"
	"math"
	"strings"

	"github.com/ivuorinen/gibidify/shared"
)

// Value kinds a configuration key can require.
const (
	kindInteger = "integer"
	kindBoolean = "boolean"
	kindList    = "list"
	kindMap     = "map"
)

// keyKinds maps each known key to the YAML shape it must have. The getters
// return a zero value for a mistyped key (`enabled: "true"` read as false), so
// a wrong type is rejected here instead of being silently coerced.
var keyKinds = map[string]string{
	shared.ConfigKeyFileSizeLimit:                     kindInteger,
	shared.ConfigKeyMaxConcurrency:                    kindInteger,
	shared.ConfigKeyIgnoreDirectories:                 kindList,
	shared.ConfigKeyFileTypesEnabled:                  kindBoolean,
	shared.ConfigKeyFileTypesCustomImageExtensions:    kindList,
	shared.ConfigKeyFileTypesCustomBinaryExtensions:   kindList,
	shared.ConfigKeyFileTypesCustomLanguages:          kindMap,
	shared.ConfigKeyFileTypesDisabledImageExtensions:  kindList,
	shared.ConfigKeyFileTypesDisabledBinaryExtensions: kindList,
	shared.ConfigKeyFileTypesDisabledLanguageExts:     kindList,
	shared.ConfigKeyResourceLimitsMaxFiles:            kindInteger,
	shared.ConfigKeyResourceLimitsMaxTotalSize:        kindInteger,
}

// ValidateConfig validates the loaded configuration.
func ValidateConfig() error {
	// Type errors first: the range checks below would otherwise report a
	// mistyped value as its zero value ("fileSizeLimit (0) is below minimum").
	validationErrors := validateTypes()

	// Validate basic settings
	validationErrors = append(validationErrors, validateBasicSettings()...)
	validationErrors = append(validationErrors, validateFileTypeSettings()...)
	validationErrors = append(validationErrors, validateResourceLimitSettings()...)

	if len(validationErrors) > 0 {
		return shared.NewStructuredError(
			shared.ErrorTypeConfiguration,
			shared.CodeConfigValidation,
			"configuration validation failed: "+strings.Join(validationErrors, "; "),
			"",
			map[string]any{"validation_errors": validationErrors},
		)
	}

	return nil
}

// validateTypes reports explicitly set keys whose value has the wrong YAML shape.
func validateTypes() []string {
	var validationErrors []string

	for _, key := range knownKeys {
		v, ok := getNested(cfg.values, key)
		if !ok {
			continue
		}
		if want := keyKinds[key]; !hasKind(v, want) {
			validationErrors = append(validationErrors, fmt.Sprintf("%s must be a %s (got %T)", key, want, v))
		}
	}

	return validationErrors
}

// hasKind reports whether v has the given kind. Whole floats count as integers
// because YAML decodes forms like 1e9 as float64.
func hasKind(v any, kind string) bool {
	switch kind {
	case kindInteger:
		return isInteger(v)
	case kindBoolean:
		_, ok := v.(bool)

		return ok
	case kindList:
		_, isAnyList := v.([]any)
		_, isStringList := v.([]string)

		return isAnyList || isStringList
	case kindMap:
		_, isAnyMap := v.(map[string]any)
		_, isStringMap := v.(map[string]string)

		return isAnyMap || isStringMap
	default:
		return false
	}
}

// isInteger reports whether v is an integer or a whole float.
func isInteger(v any) bool {
	switch n := v.(type) {
	case int, int64:
		return true
	case float64:
		return n == math.Trunc(n)
	default:
		return false
	}
}

// validateBasicSettings validates basic configuration settings.
func validateBasicSettings() []string {
	var validationErrors []string

	validationErrors = append(validationErrors, validateFileSizeLimit()...)
	validationErrors = append(validationErrors, validateIgnoreDirectories()...)
	validationErrors = append(validationErrors, validateConcurrencySettings()...)

	return validationErrors
}

// validateFileSizeLimit validates the file size limit setting.
func validateFileSizeLimit() []string {
	var validationErrors []string

	fileSizeLimit := GetInt64(shared.ConfigKeyFileSizeLimit)
	if fileSizeLimit < shared.ConfigFileSizeLimitMin {
		validationErrors = append(
			validationErrors,
			fmt.Sprintf("fileSizeLimit (%d) is below minimum (%d)", fileSizeLimit, shared.ConfigFileSizeLimitMin),
		)
	}
	if fileSizeLimit > shared.ConfigFileSizeLimitMax {
		validationErrors = append(
			validationErrors,
			fmt.Sprintf("fileSizeLimit (%d) exceeds maximum (%d)", fileSizeLimit, shared.ConfigFileSizeLimitMax),
		)
	}

	return validationErrors
}

// validateIgnoreDirectories validates the ignore directories setting.
func validateIgnoreDirectories() []string {
	var validationErrors []string

	ignoreDirectories := GetStringSlice(shared.ConfigKeyIgnoreDirectories)
	for i, dir := range ignoreDirectories {
		if errMsg := validateEmptyElement(shared.ConfigKeyIgnoreDirectories, dir, i); errMsg != "" {
			validationErrors = append(validationErrors, errMsg)

			continue
		}
		dir = strings.TrimSpace(dir)
		if strings.Contains(dir, "/") {
			validationErrors = append(
				validationErrors,
				fmt.Sprintf(
					"ignoreDirectories[%d] (%s) contains path separator - only directory names are allowed", i, dir,
				),
			)
		}
	}

	return validationErrors
}

// validateConcurrencySettings validates the concurrency settings.
func validateConcurrencySettings() []string {
	var validationErrors []string

	if !IsSet(shared.ConfigKeyMaxConcurrency) {
		return validationErrors
	}

	maxConcurrency := GetInt(shared.ConfigKeyMaxConcurrency)
	if maxConcurrency < 1 {
		validationErrors = append(
			validationErrors, fmt.Sprintf("maxConcurrency (%d) must be at least 1", maxConcurrency),
		)
	}
	if maxConcurrency > 100 {
		validationErrors = append(
			validationErrors,
			fmt.Sprintf("maxConcurrency (%d) is unreasonably high (max 100)", maxConcurrency),
		)
	}

	return validationErrors
}

// validateFileTypeSettings validates file type configuration settings. The
// disabled lists are checked like the custom ones: a dot-less "svg" matched no
// registry key (keys are filepath.Ext results such as ".svg") and silently did
// nothing.
func validateFileTypeSettings() []string {
	var validationErrors []string

	for _, key := range []string{
		shared.ConfigKeyFileTypesCustomImageExtensions,
		shared.ConfigKeyFileTypesCustomBinaryExtensions,
		shared.ConfigKeyFileTypesDisabledImageExtensions,
		shared.ConfigKeyFileTypesDisabledBinaryExtensions,
		shared.ConfigKeyFileTypesDisabledLanguageExts,
	} {
		validationErrors = append(validationErrors, validateCustomExtensions(key)...)
	}
	validationErrors = append(validationErrors, validateCustomLanguages()...)

	return validationErrors
}

// validateCustomExtensions validates an extension list under the given config
// key (custom or disabled): each element must be non-empty and dot-prefixed.
func validateCustomExtensions(key string) []string {
	var validationErrors []string

	if !IsSet(key) {
		return validationErrors
	}

	for i, ext := range GetStringSlice(key) {
		if errMsg := validateEmptyElement(key, ext, i); errMsg != "" {
			validationErrors = append(validationErrors, errMsg)

			continue
		}
		ext = strings.TrimSpace(ext)
		if errMsg := validateDotPrefix(key, ext, i); errMsg != "" {
			validationErrors = append(validationErrors, errMsg)
		}
	}

	return validationErrors
}

// validateCustomLanguages validates custom language mappings.
func validateCustomLanguages() []string {
	var validationErrors []string

	if !IsSet(shared.ConfigKeyFileTypesCustomLanguages) {
		return validationErrors
	}

	customLangs := GetStringMapString(shared.ConfigKeyFileTypesCustomLanguages)
	for ext, lang := range customLangs {
		ext = strings.TrimSpace(ext)
		if ext == "" {
			validationErrors = append(
				validationErrors,
				shared.ConfigKeyFileTypesCustomLanguages+" contains empty extension key",
			)

			continue
		}
		if errMsg := validateDotPrefixMap(shared.ConfigKeyFileTypesCustomLanguages, ext); errMsg != "" {
			validationErrors = append(validationErrors, errMsg)
		}
		if errMsg := validateEmptyMapValue(shared.ConfigKeyFileTypesCustomLanguages, ext, lang); errMsg != "" {
			validationErrors = append(validationErrors, errMsg)
		}
	}

	return validationErrors
}

// validateResourceLimitSettings validates the file-count and total-size caps.
func validateResourceLimitSettings() []string {
	var validationErrors []string

	validationErrors = append(validationErrors, validateMaxFilesLimit()...)
	validationErrors = append(validationErrors, validateMaxTotalSizeLimit()...)

	return validationErrors
}

// validateMaxFilesLimit validates resourceLimits.maxFiles setting.
func validateMaxFilesLimit() []string {
	var validationErrors []string

	if !IsSet(shared.ConfigKeyResourceLimitsMaxFiles) {
		return validationErrors
	}

	maxFiles := GetInt(shared.ConfigKeyResourceLimitsMaxFiles)
	if maxFiles < shared.ConfigMaxFilesMin {
		validationErrors = append(
			validationErrors,
			fmt.Sprintf("resourceLimits.maxFiles (%d) must be at least %d", maxFiles, shared.ConfigMaxFilesMin),
		)
	}
	if maxFiles > shared.ConfigMaxFilesMax {
		validationErrors = append(
			validationErrors,
			fmt.Sprintf("resourceLimits.maxFiles (%d) exceeds maximum (%d)", maxFiles, shared.ConfigMaxFilesMax),
		)
	}

	return validationErrors
}

// validateMaxTotalSizeLimit validates resourceLimits.maxTotalSize setting.
func validateMaxTotalSizeLimit() []string {
	var validationErrors []string

	if !IsSet(shared.ConfigKeyResourceLimitsMaxTotalSize) {
		return validationErrors
	}

	maxTotalSize := GetInt64(shared.ConfigKeyResourceLimitsMaxTotalSize)
	minTotalSize := int64(shared.ConfigMaxTotalSizeMin)
	maxTotalSizeLimit := int64(shared.ConfigMaxTotalSizeMax)
	if maxTotalSize < minTotalSize {
		validationErrors = append(
			validationErrors,
			fmt.Sprintf("resourceLimits.maxTotalSize (%d) must be at least %d", maxTotalSize, minTotalSize),
		)
	}
	if maxTotalSize > maxTotalSizeLimit {
		validationErrors = append(
			validationErrors,
			fmt.Sprintf("resourceLimits.maxTotalSize (%d) exceeds maximum (%d)", maxTotalSize, maxTotalSizeLimit),
		)
	}

	return validationErrors
}

// ValidateOutputFormat checks if an output format is valid.
func ValidateOutputFormat(format string) error {
	if !IsValidFormat(format) {
		return shared.NewStructuredError(
			shared.ErrorTypeValidation,
			shared.CodeValidationFormat,
			fmt.Sprintf("unsupported output format: %s (supported: json, yaml, markdown)", format),
			"",
			map[string]any{"format": format},
		)
	}

	return nil
}

// ValidateConcurrency checks if a concurrency level is valid.
func ValidateConcurrency(concurrency int) error {
	if concurrency < 1 {
		return shared.NewStructuredError(
			shared.ErrorTypeValidation,
			shared.CodeValidationConcurrency,
			fmt.Sprintf("concurrency (%d) must be at least 1", concurrency),
			"",
			map[string]any{"concurrency": concurrency},
		)
	}

	if IsSet(shared.ConfigKeyMaxConcurrency) {
		maxConcurrency := MaxConcurrency()
		if concurrency > maxConcurrency {
			return shared.NewStructuredError(
				shared.ErrorTypeValidation,
				shared.CodeValidationConcurrency,
				fmt.Sprintf("concurrency (%d) exceeds maximum (%d)", concurrency, maxConcurrency),
				"",
				map[string]any{"concurrency": concurrency, "max_concurrency": maxConcurrency},
			)
		}
	}

	return nil
}
