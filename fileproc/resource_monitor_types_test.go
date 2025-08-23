package fileproc

import (
	"context"
	"testing"
	"time"

	"github.com/spf13/viper"

	"github.com/ivuorinen/gibidify/config"
	"github.com/ivuorinen/gibidify/testutil"
)

func TestResourceMonitor_NewResourceMonitor(t *testing.T) {
	// Reset viper for clean test state
	testutil.ResetViperConfig(t, "")

	rm := NewResourceMonitor()
	if rm == nil {
		t.Fatal("NewResourceMonitor() returned nil")
	}

	// Test default values are set correctly
	if !rm.enabled {
		t.Error("Expected resource monitor to be enabled by default")
	}

	if rm.maxFiles != config.DefaultMaxFiles {
		t.Errorf("Expected maxFiles to be %d, got %d", config.DefaultMaxFiles, rm.maxFiles)
	}

	if rm.maxTotalSize != config.DefaultMaxTotalSize {
		t.Errorf("Expected maxTotalSize to be %d, got %d", config.DefaultMaxTotalSize, rm.maxTotalSize)
	}

	if rm.fileProcessingTimeout != time.Duration(config.DefaultFileProcessingTimeoutSec)*time.Second {
		t.Errorf("Expected fileProcessingTimeout to be %v, got %v",
			time.Duration(config.DefaultFileProcessingTimeoutSec)*time.Second, rm.fileProcessingTimeout)
	}

	// Clean up
	rm.Close()
}

func TestResourceMonitor_DisabledResourceLimits(t *testing.T) {
	// Reset viper for clean test state
	testutil.ResetViperConfig(t, "")

	// Set resource limits disabled
	viper.Set("resourceLimits.enabled", false)

	rm := NewResourceMonitor()
	defer rm.Close()

	// Test that validation passes when disabled
	err := rm.ValidateFileProcessing("/tmp/test.txt", 1000)
	if err != nil {
		t.Errorf("Expected no error when resource limits disabled, got %v", err)
	}

	// Test that read slot acquisition works when disabled
	ctx := context.Background()
	err = rm.AcquireReadSlot(ctx)
	if err != nil {
		t.Errorf("Expected no error when acquiring read slot with disabled limits, got %v", err)
	}
	rm.ReleaseReadSlot()

	// Test that rate limiting is bypassed when disabled
	err = rm.WaitForRateLimit(ctx)
	if err != nil {
		t.Errorf("Expected no error when rate limiting disabled, got %v", err)
	}
}

// TestResourceMonitor_StateQueries tests state query functions.
func TestResourceMonitor_StateQueries(t *testing.T) {
	testutil.ResetViperConfig(t, "")

	rm := NewResourceMonitor()
	defer rm.Close()

	// Test IsEmergencyStopActive - should be false initially
	if rm.IsEmergencyStopActive() {
		t.Error("Expected emergency stop to be inactive initially")
	}

	// Test IsDegradationActive - should be false initially
	if rm.IsDegradationActive() {
		t.Error("Expected degradation mode to be inactive initially")
	}
}

// TestResourceMonitor_IsEmergencyStopActive tests the IsEmergencyStopActive method.
func TestResourceMonitor_IsEmergencyStopActive(t *testing.T) {
	testutil.ResetViperConfig(t, "")

	rm := NewResourceMonitor()
	defer rm.Close()

	// Test initial state
	active := rm.IsEmergencyStopActive()
	if active {
		t.Error("Expected emergency stop to be inactive initially")
	}

	// The method should return a consistent value on multiple calls
	for i := 0; i < 5; i++ {
		if rm.IsEmergencyStopActive() != active {
			t.Error("IsEmergencyStopActive should return consistent values")
		}
	}
}

// TestResourceMonitor_IsDegradationActive tests the IsDegradationActive method.
func TestResourceMonitor_IsDegradationActive(t *testing.T) {
	testutil.ResetViperConfig(t, "")

	rm := NewResourceMonitor()
	defer rm.Close()

	// Test initial state
	active := rm.IsDegradationActive()
	if active {
		t.Error("Expected degradation mode to be inactive initially")
	}

	// The method should return a consistent value on multiple calls
	for i := 0; i < 5; i++ {
		if rm.IsDegradationActive() != active {
			t.Error("IsDegradationActive should return consistent values")
		}
	}
}

// TestResourceMonitor_Close tests the Close method.
func TestResourceMonitor_Close(t *testing.T) {
	testutil.ResetViperConfig(t, "")

	rm := NewResourceMonitor()

	// Close should not panic
	rm.Close()

	// Multiple closes should be safe
	rm.Close()
	rm.Close()
}
