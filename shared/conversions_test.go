package shared

import (
	"math"
	"testing"
)

func TestSafeUint64ToInt64(t *testing.T) {
	tests := []struct {
		name     string
		input    uint64
		expected int64
		wantOk   bool
	}{
		{
			name:     TestSafeConversion,
			input:    1000,
			expected: 1000,
			wantOk:   true,
		},
		{
			name:     "max safe value",
			input:    math.MaxInt64,
			expected: math.MaxInt64,
			wantOk:   true,
		},
		{
			name:     "overflow by one",
			input:    math.MaxInt64 + 1,
			expected: 0,
			wantOk:   false,
		},
		{
			name:     "max uint64 overflow",
			input:    math.MaxUint64,
			expected: 0,
			wantOk:   false,
		},
		{
			name:     "zero value",
			input:    0,
			expected: 0,
			wantOk:   true,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				got, ok := SafeUint64ToInt64(tt.input)
				if ok != tt.wantOk {
					t.Errorf("SafeUint64ToInt64() ok = %v, want %v", ok, tt.wantOk)
				}
				if got != tt.expected {
					t.Errorf("SafeUint64ToInt64() = %v, want %v", got, tt.expected)
				}
			},
		)
	}
}

func TestSafeIntToInt32(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		expected int32
		wantOk   bool
	}{
		{
			name:     TestSafeConversion,
			input:    1000,
			expected: 1000,
			wantOk:   true,
		},
		{
			name:     "max safe value",
			input:    math.MaxInt32,
			expected: math.MaxInt32,
			wantOk:   true,
		},
		{
			name:     "min safe value",
			input:    math.MinInt32,
			expected: math.MinInt32,
			wantOk:   true,
		},
		{
			name:     "overflow by one",
			input:    math.MaxInt32 + 1,
			expected: 0,
			wantOk:   false,
		},
		{
			name:     "underflow by one",
			input:    math.MinInt32 - 1,
			expected: 0,
			wantOk:   false,
		},
		{
			name:     "zero value",
			input:    0,
			expected: 0,
			wantOk:   true,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				got, ok := SafeIntToInt32(tt.input)
				if ok != tt.wantOk {
					t.Errorf("SafeIntToInt32() ok = %v, want %v", ok, tt.wantOk)
				}
				if got != tt.expected {
					t.Errorf("SafeIntToInt32() = %v, want %v", got, tt.expected)
				}
			},
		)
	}
}

func TestSafeUint64ToInt64WithDefault(t *testing.T) {
	tests := []struct {
		name         string
		input        uint64
		defaultValue int64
		expected     int64
	}{
		{
			name:         TestSafeConversion,
			input:        1000,
			defaultValue: -1,
			expected:     1000,
		},
		{
			name:         "overflow uses default",
			input:        math.MaxUint64,
			defaultValue: -1,
			expected:     -1,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				got := SafeUint64ToInt64WithDefault(tt.input, tt.defaultValue)
				if got != tt.expected {
					t.Errorf("SafeUint64ToInt64WithDefault() = %v, want %v", got, tt.expected)
				}
			},
		)
	}
}

func TestSafeIntToInt32WithDefault(t *testing.T) {
	tests := []struct {
		name         string
		input        int
		defaultValue int32
		expected     int32
	}{
		{
			name:         TestSafeConversion,
			input:        1000,
			defaultValue: -1,
			expected:     1000,
		},
		{
			name:         "overflow uses default",
			input:        math.MaxInt32 + 1,
			defaultValue: -1,
			expected:     -1,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				got := SafeIntToInt32WithDefault(tt.input, tt.defaultValue)
				if got != tt.expected {
					t.Errorf("SafeIntToInt32WithDefault() = %v, want %v", got, tt.expected)
				}
			},
		)
	}
}

func TestBytesToMB(t *testing.T) {
	tests := []struct {
		name     string
		input    uint64
		expected int64
	}{
		{
			name:     "zero bytes",
			input:    0,
			expected: 0,
		},
		{
			name:     "1MB",
			input:    1024 * 1024,
			expected: 1,
		},
		{
			name:     "1GB",
			input:    1024 * 1024 * 1024,
			expected: 1024,
		},
		{
			name:     "large value (no overflow)",
			input:    math.MaxUint64,
			expected: 17592186044415, // MaxUint64 / 1024 / 1024, which is still within int64 range
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				got := BytesToMB(tt.input)
				if got != tt.expected {
					t.Errorf("BytesToMB() = %v, want %v", got, tt.expected)
				}
			},
		)
	}
}

func TestBytesToMBFloat64(t *testing.T) {
	tests := []struct {
		name     string
		input    uint64
		expected float64
		delta    float64
	}{
		{
			name:     "zero bytes",
			input:    0,
			expected: 0,
			delta:    0.0001,
		},
		{
			name:     "1MB",
			input:    1024 * 1024,
			expected: 1.0,
			delta:    0.0001,
		},
		{
			name:     "1GB",
			input:    1024 * 1024 * 1024,
			expected: 1024.0,
			delta:    0.0001,
		},
		{
			name:     "large value near overflow",
			input:    math.MaxUint64 - 1,
			expected: float64((math.MaxUint64-1)/1024) / 1024.0,
			delta:    1.0, // Allow larger delta for very large numbers
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				got := BytesToMBFloat64(tt.input)
				if math.Abs(got-tt.expected) > tt.delta {
					t.Errorf("BytesToMBFloat64() = %v, want %v (±%v)", got, tt.expected, tt.delta)
				}
			},
		)
	}
}

func TestSafeMemoryDiffMB(t *testing.T) {
	tests := []struct {
		name     string
		after    uint64
		before   uint64
		expected float64
		delta    float64
	}{
		{
			name:     "normal increase",
			after:    2 * 1024 * 1024, // 2MB
			before:   1 * 1024 * 1024, // 1MB
			expected: 1.0,
			delta:    0.0001,
		},
		{
			name:     "no change",
			after:    1 * 1024 * 1024,
			before:   1 * 1024 * 1024,
			expected: 0.0,
			delta:    0.0001,
		},
		{
			name:     "underflow case",
			after:    1 * 1024 * 1024, // 1MB
			before:   2 * 1024 * 1024, // 2MB
			expected: 0.0,             // Should return 0 instead of negative
			delta:    0.0001,
		},
		{
			name:     "large difference",
			after:    2 * 1024 * 1024 * 1024, // 2GB
			before:   1 * 1024 * 1024 * 1024, // 1GB
			expected: 1024.0,                 // 1GB = 1024MB
			delta:    0.0001,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				got := SafeMemoryDiffMB(tt.after, tt.before)
				if math.Abs(got-tt.expected) > tt.delta {
					t.Errorf("SafeMemoryDiffMB() = %v, want %v (±%v)", got, tt.expected, tt.delta)
				}
			},
		)
	}
}
