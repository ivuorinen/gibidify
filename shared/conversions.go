// Package shared provides common utility functions for gibidify.
package shared

import (
	"math"
)

// SafeUint64ToInt64 safely converts uint64 to int64, checking for overflow.
// Returns the converted value and true if conversion is safe, or 0 and false if overflow would occur.
func SafeUint64ToInt64(value uint64) (int64, bool) {
	if value > math.MaxInt64 {
		return 0, false
	}

	return int64(value), true
}

// SafeIntToInt32 safely converts int to int32, checking for overflow.
// Returns the converted value and true if conversion is safe, or 0 and false if overflow would occur.
func SafeIntToInt32(value int) (int32, bool) {
	if value > math.MaxInt32 || value < math.MinInt32 {
		return 0, false
	}

	return int32(value), true
}

// SafeUint64ToInt64WithDefault safely converts uint64 to int64 with a default value on overflow.
func SafeUint64ToInt64WithDefault(value uint64, defaultValue int64) int64 {
	if converted, ok := SafeUint64ToInt64(value); ok {
		return converted
	}

	return defaultValue
}

// SafeIntToInt32WithDefault safely converts int to int32 with a default value on overflow.
func SafeIntToInt32WithDefault(value int, defaultValue int32) int32 {
	if converted, ok := SafeIntToInt32(value); ok {
		return converted
	}

	return defaultValue
}

// BytesToMB safely converts bytes (uint64) to megabytes (int64), handling overflow.
func BytesToMB(bytes uint64) int64 {
	mb := bytes / uint64(BytesPerMB)

	return SafeUint64ToInt64WithDefault(mb, math.MaxInt64)
}

// BytesToMBFloat64 safely converts bytes (uint64) to megabytes (float64), handling overflow.
func BytesToMBFloat64(bytes uint64) float64 {
	const bytesPerMB = float64(BytesPerMB)
	if bytes > math.MaxUint64/2 {
		// Prevent overflow in arithmetic by dividing step by step
		return float64(bytes/uint64(BytesPerKB)) / float64(BytesPerKB)
	}

	return float64(bytes) / bytesPerMB
}

// SafeMemoryDiffMB safely calculates the difference between two uint64 memory values
// and converts to MB as float64, handling potential underflow.
func SafeMemoryDiffMB(after, before uint64) float64 {
	if after >= before {
		diff := after - before

		return BytesToMBFloat64(diff)
	}
	// Handle underflow case - return 0 instead of negative
	return 0.0
}
