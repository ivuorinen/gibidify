package config

import (
	"testing"

	"github.com/spf13/viper"
)

func TestGetSupportedFormats(t *testing.T) {
	// Save and restore original value
	originalValue := viper.Get("supportedFormats")
	defer viper.Set("supportedFormats", originalValue)

	// Set test value
	testFormats := []string{"json", "yaml", "markdown"}
	viper.Set("supportedFormats", testFormats)

	formats := GetSupportedFormats()
	if len(formats) != len(testFormats) {
		t.Errorf("Expected %d formats, got %d", len(testFormats), len(formats))
	}

	for i, format := range formats {
		if format != testFormats[i] {
			t.Errorf("Expected format %s at index %d, got %s", testFormats[i], i, format)
		}
	}
}

func TestGetFilePatterns(t *testing.T) {
	// Save and restore original value
	originalValue := viper.Get("filePatterns")
	defer viper.Set("filePatterns", originalValue)

	// Set test value
	testPatterns := []string{"*.go", "*.txt", "*.md"}
	viper.Set("filePatterns", testPatterns)

	patterns := GetFilePatterns()
	if len(patterns) != len(testPatterns) {
		t.Errorf("Expected %d patterns, got %d", len(testPatterns), len(patterns))
	}

	for i, pattern := range patterns {
		if pattern != testPatterns[i] {
			t.Errorf("Expected pattern %s at index %d, got %s", testPatterns[i], i, pattern)
		}
	}
}

func TestGetBackpressureEnabled(t *testing.T) {
	// Save and restore original value
	originalValue := viper.Get("backpressure.enabled")
	defer viper.Set("backpressure.enabled", originalValue)

	tests := []struct {
		name     string
		setValue bool
		want     bool
	}{
		{"enabled", true, true},
		{"disabled", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viper.Set("backpressure.enabled", tt.setValue)
			got := GetBackpressureEnabled()
			if got != tt.want {
				t.Errorf("GetBackpressureEnabled() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetMaxPendingFiles(t *testing.T) {
	// Save and restore original value
	originalValue := viper.Get("backpressure.maxPendingFiles")
	defer viper.Set("backpressure.maxPendingFiles", originalValue)

	testValue := 100
	viper.Set("backpressure.maxPendingFiles", testValue)

	got := GetMaxPendingFiles()
	if got != testValue {
		t.Errorf("GetMaxPendingFiles() = %d, want %d", got, testValue)
	}
}

func TestGetMaxPendingWrites(t *testing.T) {
	// Save and restore original value
	originalValue := viper.Get("backpressure.maxPendingWrites")
	defer viper.Set("backpressure.maxPendingWrites", originalValue)

	testValue := 50
	viper.Set("backpressure.maxPendingWrites", testValue)

	got := GetMaxPendingWrites()
	if got != testValue {
		t.Errorf("GetMaxPendingWrites() = %d, want %d", got, testValue)
	}
}

func TestGetMaxMemoryUsage(t *testing.T) {
	// Save and restore original value
	originalValue := viper.Get("backpressure.maxMemoryUsage")
	defer viper.Set("backpressure.maxMemoryUsage", originalValue)

	testValue := int64(1024 * 1024 * 100) // 100MB
	viper.Set("backpressure.maxMemoryUsage", testValue)

	got := GetMaxMemoryUsage()
	if got != testValue {
		t.Errorf("GetMaxMemoryUsage() = %d, want %d", got, testValue)
	}
}

func TestGetMemoryCheckInterval(t *testing.T) {
	// Save and restore original value
	originalValue := viper.Get("backpressure.memoryCheckInterval")
	defer viper.Set("backpressure.memoryCheckInterval", originalValue)

	testValue := 5
	viper.Set("backpressure.memoryCheckInterval", testValue)

	got := GetMemoryCheckInterval()
	if got != testValue {
		t.Errorf("GetMemoryCheckInterval() = %d, want %d", got, testValue)
	}
}

func TestGetResourceLimitsEnabled(t *testing.T) {
	// Save and restore original value
	originalValue := viper.Get("resourceLimits.enabled")
	defer viper.Set("resourceLimits.enabled", originalValue)

	tests := []struct {
		name     string
		setValue bool
		want     bool
	}{
		{"enabled", true, true},
		{"disabled", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viper.Set("resourceLimits.enabled", tt.setValue)
			got := GetResourceLimitsEnabled()
			if got != tt.want {
				t.Errorf("GetResourceLimitsEnabled() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetMaxFiles(t *testing.T) {
	// Save and restore original value
	originalValue := viper.Get("resourceLimits.maxFiles")
	defer viper.Set("resourceLimits.maxFiles", originalValue)

	testValue := 10000
	viper.Set("resourceLimits.maxFiles", testValue)

	got := GetMaxFiles()
	if got != testValue {
		t.Errorf("GetMaxFiles() = %d, want %d", got, testValue)
	}
}

func TestGetMaxTotalSize(t *testing.T) {
	// Save and restore original value
	originalValue := viper.Get("resourceLimits.maxTotalSize")
	defer viper.Set("resourceLimits.maxTotalSize", originalValue)

	testValue := int64(1024 * 1024 * 1024) // 1GB
	viper.Set("resourceLimits.maxTotalSize", testValue)

	got := GetMaxTotalSize()
	if got != testValue {
		t.Errorf("GetMaxTotalSize() = %d, want %d", got, testValue)
	}
}

func TestGetFileProcessingTimeoutSec(t *testing.T) {
	// Save and restore original value
	originalValue := viper.Get("resourceLimits.fileProcessingTimeoutSec")
	defer viper.Set("resourceLimits.fileProcessingTimeoutSec", originalValue)

	testValue := 30
	viper.Set("resourceLimits.fileProcessingTimeoutSec", testValue)

	got := GetFileProcessingTimeoutSec()
	if got != testValue {
		t.Errorf("GetFileProcessingTimeoutSec() = %d, want %d", got, testValue)
	}
}

func TestGetOverallTimeoutSec(t *testing.T) {
	// Save and restore original value
	originalValue := viper.Get("resourceLimits.overallTimeoutSec")
	defer viper.Set("resourceLimits.overallTimeoutSec", originalValue)

	testValue := 3600
	viper.Set("resourceLimits.overallTimeoutSec", testValue)

	got := GetOverallTimeoutSec()
	if got != testValue {
		t.Errorf("GetOverallTimeoutSec() = %d, want %d", got, testValue)
	}
}

func TestGetMaxConcurrentReads(t *testing.T) {
	// Save and restore original value
	originalValue := viper.Get("resourceLimits.maxConcurrentReads")
	defer viper.Set("resourceLimits.maxConcurrentReads", originalValue)

	testValue := 10
	viper.Set("resourceLimits.maxConcurrentReads", testValue)

	got := GetMaxConcurrentReads()
	if got != testValue {
		t.Errorf("GetMaxConcurrentReads() = %d, want %d", got, testValue)
	}
}

func TestGetRateLimitFilesPerSec(t *testing.T) {
	// Save and restore original value
	originalValue := viper.Get("resourceLimits.rateLimitFilesPerSec")
	defer viper.Set("resourceLimits.rateLimitFilesPerSec", originalValue)

	testValue := 100
	viper.Set("resourceLimits.rateLimitFilesPerSec", testValue)

	got := GetRateLimitFilesPerSec()
	if got != testValue {
		t.Errorf("GetRateLimitFilesPerSec() = %d, want %d", got, testValue)
	}
}

func TestGetHardMemoryLimitMB(t *testing.T) {
	// Save and restore original value
	originalValue := viper.Get("resourceLimits.hardMemoryLimitMB")
	defer viper.Set("resourceLimits.hardMemoryLimitMB", originalValue)

	testValue := 512
	viper.Set("resourceLimits.hardMemoryLimitMB", testValue)

	got := GetHardMemoryLimitMB()
	if got != testValue {
		t.Errorf("GetHardMemoryLimitMB() = %d, want %d", got, testValue)
	}
}

func TestGetEnableGracefulDegradation(t *testing.T) {
	// Save and restore original value
	originalValue := viper.Get("resourceLimits.enableGracefulDegradation")
	defer viper.Set("resourceLimits.enableGracefulDegradation", originalValue)

	tests := []struct {
		name     string
		setValue bool
		want     bool
	}{
		{"enabled", true, true},
		{"disabled", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viper.Set("resourceLimits.enableGracefulDegradation", tt.setValue)
			got := GetEnableGracefulDegradation()
			if got != tt.want {
				t.Errorf("GetEnableGracefulDegradation() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetEnableResourceMonitoring(t *testing.T) {
	// Save and restore original value
	originalValue := viper.Get("resourceLimits.enableResourceMonitoring")
	defer viper.Set("resourceLimits.enableResourceMonitoring", originalValue)

	tests := []struct {
		name     string
		setValue bool
		want     bool
	}{
		{"enabled", true, true},
		{"disabled", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viper.Set("resourceLimits.enableResourceMonitoring", tt.setValue)
			got := GetEnableResourceMonitoring()
			if got != tt.want {
				t.Errorf("GetEnableResourceMonitoring() = %v, want %v", got, tt.want)
			}
		})
	}
}
