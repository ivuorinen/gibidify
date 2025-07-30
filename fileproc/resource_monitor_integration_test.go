package fileproc

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"

	"github.com/ivuorinen/gibidify/testutil"
)

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