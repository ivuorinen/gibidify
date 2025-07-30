package fileproc

import (
	"testing"

	"github.com/spf13/viper"

	"github.com/ivuorinen/gibidify/testutil"
	"github.com/ivuorinen/gibidify/utils"
)

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