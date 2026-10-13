package config_test

import (
	"testing"

	"github.com/ivuorinen/gibidify/config"
	"github.com/ivuorinen/gibidify/shared"
	"github.com/ivuorinen/gibidify/testutil"
)

// validationBoundaryCases pairs explicit settings with the validation error
// they must produce; an empty errContains means the settings must validate.
var validationBoundaryCases = []struct {
	name        string
	settings    map[string]any
	errContains string
}{
	{"maxConcurrency at ceiling", map[string]any{shared.ConfigKeyMaxConcurrency: 100}, ""},
	{
		"maxConcurrency above ceiling", map[string]any{shared.ConfigKeyMaxConcurrency: 101},
		"maxConcurrency (101) is unreasonably high (max 100)",
	},
	{"maxFiles at minimum", map[string]any{shared.ConfigKeyResourceLimitsMaxFiles: shared.ConfigMaxFilesMin}, ""},
	{"maxFiles at maximum", map[string]any{shared.ConfigKeyResourceLimitsMaxFiles: shared.ConfigMaxFilesMax}, ""},
	{
		"maxFiles below minimum", map[string]any{shared.ConfigKeyResourceLimitsMaxFiles: shared.ConfigMaxFilesMin - 1},
		"resourceLimits.maxFiles (0) must be at least 1",
	},
	{
		"maxFiles above maximum", map[string]any{shared.ConfigKeyResourceLimitsMaxFiles: shared.ConfigMaxFilesMax + 1},
		"resourceLimits.maxFiles (1000001) exceeds maximum (1000000)",
	},
	{
		"maxTotalSize at minimum",
		map[string]any{shared.ConfigKeyResourceLimitsMaxTotalSize: int64(shared.ConfigMaxTotalSizeMin)}, "",
	},
	{
		"maxTotalSize at maximum",
		map[string]any{shared.ConfigKeyResourceLimitsMaxTotalSize: int64(shared.ConfigMaxTotalSizeMax)}, "",
	},
	{
		"maxTotalSize below minimum",
		map[string]any{shared.ConfigKeyResourceLimitsMaxTotalSize: int64(shared.ConfigMaxTotalSizeMin - 1)},
		"resourceLimits.maxTotalSize (1048575) must be at least 1048576",
	},
	{
		"maxTotalSize above maximum",
		map[string]any{shared.ConfigKeyResourceLimitsMaxTotalSize: int64(shared.ConfigMaxTotalSizeMax + 1)},
		"resourceLimits.maxTotalSize (107374182401) exceeds maximum (107374182400)",
	},
	{
		"custom language with blank extension",
		map[string]any{shared.ConfigKeyFileTypesCustomLanguages: map[string]any{"  ": "text"}},
		shared.ConfigKeyFileTypesCustomLanguages + " contains empty extension key",
	},
	{"whole float accepted as integer", map[string]any{shared.ConfigKeyResourceLimitsMaxFiles: 500.0}, ""},
	{
		"fractional float rejected as integer", map[string]any{shared.ConfigKeyResourceLimitsMaxFiles: 500.5},
		shared.ConfigKeyResourceLimitsMaxFiles + " must be an integer (got float64)",
	},
	{
		"string rejected as integer", map[string]any{shared.ConfigKeyFileSizeLimit: "5MB"},
		shared.ConfigKeyFileSizeLimit + " must be an integer (got string)",
	},
}

// TestValidateConfigBoundaries verifies the concurrency ceiling, the resource
// limit ranges (inclusive at both ends), blank custom-language keys and the
// integer type check.
func TestValidateConfigBoundaries(t *testing.T) {
	for _, tt := range validationBoundaryCases {
		t.Run(tt.name, func(t *testing.T) {
			testutil.SetConfigKeys(t, tt.settings)

			err := config.ValidateConfig()

			if tt.errContains == "" {
				if err != nil {
					t.Errorf("expected valid config, got: %v", err)
				}

				return
			}
			validateExpectedError(t, err, tt.errContains)
		})
	}
}

// TestLoadConfigAcceptsExponentIntegers verifies a whole number written in
// exponent form, which YAML decodes as a float, is accepted and read back as
// the integer it denotes.
func TestLoadConfigAcceptsExponentIntegers(t *testing.T) {
	_, err := loadWithConfig(t, "maxConcurrency: 4e1\nresourceLimits:\n  maxTotalSize: 2e9\n")
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}

	if got := config.GetInt(shared.ConfigKeyMaxConcurrency); got != 40 {
		t.Errorf("maxConcurrency = %d, want 40", got)
	}
	if got := config.GetInt64(shared.ConfigKeyResourceLimitsMaxTotalSize); got != 2_000_000_000 {
		t.Errorf("maxTotalSize = %d, want 2000000000", got)
	}
}
