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
