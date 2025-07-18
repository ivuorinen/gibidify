// Package fileproc provides tests for resource monitoring functionality.
package fileproc

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/spf13/viper"

	"github.com/ivuorinen/gibidify/config"
	"github.com/ivuorinen/gibidify/testutil"
	"github.com/ivuorinen/gibidify/utils"
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

func TestResourceMonitor_FileCountLimit(t *testing.T) {
	testutil.ResetViperConfig(t, "")

	// Set a very low file count limit for testing
	viper.Set("resourceLimits.enabled", true)
	viper.Set("resourceLimits.maxFiles", 2)

	rm := NewResourceMonitor()
	defer rm.Close()

	// First file should pass
	err := rm.ValidateFileProcessing("/tmp/file1.txt", 100)
	if err != nil {
		t.Errorf("Expected no error for first file, got %v", err)
	}
	rm.RecordFileProcessed(100)

	// Second file should pass
	err = rm.ValidateFileProcessing("/tmp/file2.txt", 100)
	if err != nil {
		t.Errorf("Expected no error for second file, got %v", err)
	}
	rm.RecordFileProcessed(100)

	// Third file should fail
	err = rm.ValidateFileProcessing("/tmp/file3.txt", 100)
	if err == nil {
		t.Error("Expected error for third file (exceeds limit), got nil")
	}

	// Verify it's the correct error type
	structErr, ok := err.(*utils.StructuredError)
	if !ok {
		t.Errorf("Expected StructuredError, got %T", err)
	} else if structErr.Code != utils.CodeResourceLimitFiles {
		t.Errorf("Expected error code %s, got %s", utils.CodeResourceLimitFiles, structErr.Code)
	}
}

func TestResourceMonitor_TotalSizeLimit(t *testing.T) {
	testutil.ResetViperConfig(t, "")

	// Set a low total size limit for testing (1KB)
	viper.Set("resourceLimits.enabled", true)
	viper.Set("resourceLimits.maxTotalSize", 1024)

	rm := NewResourceMonitor()
	defer rm.Close()

	// First small file should pass
	err := rm.ValidateFileProcessing("/tmp/small.txt", 500)
	if err != nil {
		t.Errorf("Expected no error for small file, got %v", err)
	}
	rm.RecordFileProcessed(500)

	// Second small file should pass
	err = rm.ValidateFileProcessing("/tmp/small2.txt", 400)
	if err != nil {
		t.Errorf("Expected no error for second small file, got %v", err)
	}
	rm.RecordFileProcessed(400)

	// Large file that would exceed limit should fail
	err = rm.ValidateFileProcessing("/tmp/large.txt", 200)
	if err == nil {
		t.Error("Expected error for file that would exceed size limit, got nil")
	}

	// Verify it's the correct error type
	structErr, ok := err.(*utils.StructuredError)
	if !ok {
		t.Errorf("Expected StructuredError, got %T", err)
	} else if structErr.Code != utils.CodeResourceLimitTotalSize {
		t.Errorf("Expected error code %s, got %s", utils.CodeResourceLimitTotalSize, structErr.Code)
	}
}

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

func TestResourceMonitor_RateLimiting(t *testing.T) {
	testutil.ResetViperConfig(t, "")

	// Enable rate limiting with a low rate for testing
	viper.Set("resourceLimits.enabled", true)
	viper.Set("resourceLimits.rateLimitFilesPerSec", 5) // 5 files per second

	rm := NewResourceMonitor()
	defer rm.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	// First few requests should succeed quickly
	start := time.Now()
	for i := 0; i < 3; i++ {
		err := rm.WaitForRateLimit(ctx)
		if err != nil {
			t.Errorf("Expected no error for rate limit wait %d, got %v", i, err)
		}
	}

	// Should have taken some time due to rate limiting
	duration := time.Since(start)
	if duration < 200*time.Millisecond {
		t.Logf("Rate limiting may not be working as expected, took only %v", duration)
	}
}

func TestResourceMonitor_Metrics(t *testing.T) {
	testutil.ResetViperConfig(t, "")

	viper.Set("resourceLimits.enabled", true)
	viper.Set("resourceLimits.enableResourceMonitoring", true)

	rm := NewResourceMonitor()
	defer rm.Close()

	// Process some files to generate metrics
	rm.RecordFileProcessed(1000)
	rm.RecordFileProcessed(2000)
	rm.RecordFileProcessed(500)

	metrics := rm.GetMetrics()

	// Verify metrics
	if metrics.FilesProcessed != 3 {
		t.Errorf("Expected 3 files processed, got %d", metrics.FilesProcessed)
	}

	if metrics.TotalSizeProcessed != 3500 {
		t.Errorf("Expected total size 3500, got %d", metrics.TotalSizeProcessed)
	}

	expectedAvgSize := float64(3500) / float64(3)
	if metrics.AverageFileSize != expectedAvgSize {
		t.Errorf("Expected average file size %.2f, got %.2f", expectedAvgSize, metrics.AverageFileSize)
	}

	if metrics.ProcessingRate <= 0 {
		t.Error("Expected positive processing rate")
	}

	if !metrics.LastUpdated.After(time.Now().Add(-time.Second)) {
		t.Error("Expected recent LastUpdated timestamp")
	}
}

func TestResourceMonitor_Integration(t *testing.T) {
	// Create temporary test directory
	tempDir := t.TempDir()

	// Create test files
	testFiles := []string{"test1.txt", "test2.txt", "test3.txt"}
	for _, filename := range testFiles {
		testutil.CreateTestFile(t, tempDir, filename, []byte("test content"))
	}

	testutil.ResetViperConfig(t, "")

	// Configure resource limits
	viper.Set("resourceLimits.enabled", true)
	viper.Set("resourceLimits.maxFiles", 5)
	viper.Set("resourceLimits.maxTotalSize", 1024*1024) // 1MB
	viper.Set("resourceLimits.fileProcessingTimeoutSec", 10)
	viper.Set("resourceLimits.maxConcurrentReads", 3)

	rm := NewResourceMonitor()
	defer rm.Close()

	ctx := context.Background()

	// Test file processing workflow
	for _, filename := range testFiles {
		filePath := filepath.Join(tempDir, filename)
		fileInfo, err := os.Stat(filePath)
		if err != nil {
			t.Fatalf("Failed to stat test file %s: %v", filePath, err)
		}

		// Validate file can be processed
		err = rm.ValidateFileProcessing(filePath, fileInfo.Size())
		if err != nil {
			t.Errorf("Failed to validate file %s: %v", filePath, err)
			continue
		}

		// Acquire read slot
		err = rm.AcquireReadSlot(ctx)
		if err != nil {
			t.Errorf("Failed to acquire read slot for %s: %v", filePath, err)
			continue
		}

		// Check memory limits
		err = rm.CheckHardMemoryLimit()
		if err != nil {
			t.Errorf("Memory limit check failed for %s: %v", filePath, err)
		}

		// Record processing
		rm.RecordFileProcessed(fileInfo.Size())

		// Release read slot
		rm.ReleaseReadSlot()
	}

	// Verify final metrics
	metrics := rm.GetMetrics()
	if metrics.FilesProcessed != int64(len(testFiles)) {
		t.Errorf("Expected %d files processed, got %d", len(testFiles), metrics.FilesProcessed)
	}

	// Test resource limit logging
	rm.LogResourceInfo()
}