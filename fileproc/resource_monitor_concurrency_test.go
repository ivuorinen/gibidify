package fileproc

import (
	"context"
	"testing"
	"time"

	"github.com/spf13/viper"

	"github.com/ivuorinen/gibidify/testutil"
)

func TestResourceMonitor_ConcurrentReadsLimit(t *testing.T) {
	testutil.ResetViperConfig(t, "")

	// Set a low concurrent reads limit for testing
	viper.Set("resourceLimits.enabled", true)
	viper.Set("resourceLimits.maxConcurrentReads", 2)

	rm := NewResourceMonitor()
	defer rm.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// First read slot should succeed
	err := rm.AcquireReadSlot(ctx)
	if err != nil {
		t.Errorf("Expected no error for first read slot, got %v", err)
	}

	// Second read slot should succeed
	err = rm.AcquireReadSlot(ctx)
	if err != nil {
		t.Errorf("Expected no error for second read slot, got %v", err)
	}

	// Third read slot should timeout (context deadline exceeded)
	err = rm.AcquireReadSlot(ctx)
	if err == nil {
		t.Error("Expected timeout error for third read slot, got nil")
	}

	// Release one slot and try again
	rm.ReleaseReadSlot()
	
	// Create new context for the next attempt
	ctx2, cancel2 := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel2()
	
	err = rm.AcquireReadSlot(ctx2)
	if err != nil {
		t.Errorf("Expected no error after releasing a slot, got %v", err)
	}

	// Clean up remaining slots
	rm.ReleaseReadSlot()
	rm.ReleaseReadSlot()
}

func TestResourceMonitor_TimeoutContexts(t *testing.T) {
	testutil.ResetViperConfig(t, "")

	// Set short timeouts for testing
	viper.Set("resourceLimits.enabled", true)
	viper.Set("resourceLimits.fileProcessingTimeoutSec", 1) // 1 second
	viper.Set("resourceLimits.overallTimeoutSec", 2)        // 2 seconds

	rm := NewResourceMonitor()
	defer rm.Close()

	parentCtx := context.Background()

	// Test file processing context
	fileCtx, fileCancel := rm.CreateFileProcessingContext(parentCtx)
	defer fileCancel()

	deadline, ok := fileCtx.Deadline()
	if !ok {
		t.Error("Expected file processing context to have a deadline")
	} else if time.Until(deadline) > time.Second+100*time.Millisecond {
		t.Error("File processing timeout appears to be too long")
	}

	// Test overall processing context
	overallCtx, overallCancel := rm.CreateOverallProcessingContext(parentCtx)
	defer overallCancel()

	deadline, ok = overallCtx.Deadline()
	if !ok {
		t.Error("Expected overall processing context to have a deadline")
	} else if time.Until(deadline) > 2*time.Second+100*time.Millisecond {
		t.Error("Overall processing timeout appears to be too long")
	}
}