// Package gibidiutils provides common utility functions for gibidify.
package gibidiutils

import (
	"math"
	"testing"
)

func TestSafeUint64ToInt64WithDefault(t *testing.T) {
	tests := []struct {
		name         string
		value        uint64
		defaultValue int64
		want         int64
	}{
		{
			name:         "normal value within range",
			value:        1000,
			defaultValue: 0,
			want:         1000,
		},
		{
			name:         "zero value",
			value:        0,
			defaultValue: 0,
			want:         0,
		},
		{
			name:         "max int64 exactly",
			value:        math.MaxInt64,
			defaultValue: 0,
			want:         math.MaxInt64,
		},
		{
			name:         "overflow with zero default clamps to max",
			value:        math.MaxInt64 + 1,
			defaultValue: 0,
			want:         math.MaxInt64,
		},
		{
			name:         "large overflow with zero default clamps to max",
			value:        math.MaxUint64,
			defaultValue: 0,
			want:         math.MaxInt64,
		},
		{
			name:         "overflow with custom default returns custom",
			value:        math.MaxInt64 + 1,
			defaultValue: -1,
			want:         -1,
		},
		{
			name:         "overflow with custom positive default",
			value:        math.MaxUint64,
			defaultValue: 12345,
			want:         12345,
		},
		{
			name:         "large value within range",
			value:        uint64(math.MaxInt64 - 1000),
			defaultValue: 0,
			want:         math.MaxInt64 - 1000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SafeUint64ToInt64WithDefault(tt.value, tt.defaultValue)
			if got != tt.want {
				t.Errorf("SafeUint64ToInt64WithDefault(%d, %d) = %d, want %d",
					tt.value, tt.defaultValue, got, tt.want)
			}
		})
	}
}

func TestSafeUint64ToInt64WithDefault_GuardrailsBehavior(t *testing.T) {
	// Test that overflow with default=0 returns MaxInt64, not 0
	// This is critical for back-pressure and resource monitors
	result := SafeUint64ToInt64WithDefault(math.MaxUint64, 0)
	if result == 0 {
		t.Error("Overflow with default=0 returned 0, which would disable guardrails")
	}
	if result != math.MaxInt64 {
		t.Errorf("Overflow with default=0 should clamp to MaxInt64, got %d", result)
	}
}

// BenchmarkSafeUint64ToInt64WithDefault benchmarks the conversion function
func BenchmarkSafeUint64ToInt64WithDefault(b *testing.B) {
	b.Run("normal_value", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = SafeUint64ToInt64WithDefault(1000, 0)
		}
	})

	b.Run("overflow_zero_default", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = SafeUint64ToInt64WithDefault(math.MaxUint64, 0)
		}
	})

	b.Run("overflow_custom_default", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = SafeUint64ToInt64WithDefault(math.MaxUint64, -1)
		}
	})
}
