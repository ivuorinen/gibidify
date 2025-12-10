package fileproc

import (
	"errors"
	"strings"
	"testing"

	"github.com/spf13/viper"

	"github.com/ivuorinen/gibidify/shared"
	"github.com/ivuorinen/gibidify/testutil"
)

// assertStructuredError verifies that an error is a StructuredError with the expected code.
func assertStructuredError(t *testing.T, err error, expectedCode string) {
	t.Helper()
	structErr := &shared.StructuredError{}
	ok := errors.As(err, &structErr)
	if !ok {
		t.Errorf("Expected StructuredError, got %T", err)
	} else if structErr.Code != expectedCode {
		t.Errorf("Expected error code %s, got %s", expectedCode, structErr.Code)
	}
}

// validateMemoryLimitError validates that an error is a proper memory limit StructuredError.
func validateMemoryLimitError(t *testing.T, err error) {
	t.Helper()

	structErr := &shared.StructuredError{}
	if errors.As(err, &structErr) {
		if structErr.Code != shared.CodeResourceLimitMemory {
			t.Errorf("Expected memory limit error code, got %s", structErr.Code)
		}
	} else {
		t.Errorf("Expected StructuredError, got %T", err)
	}
}

func TestResourceMonitorFileCountLimit(t *testing.T) {
	testutil.ResetViperConfig(t, "")

	// Set a very low file count limit for testing
	viper.Set(shared.TestCfgResourceLimitsEnabled, true)
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
	assertStructuredError(t, err, shared.CodeResourceLimitFiles)
}

func TestResourceMonitorTotalSizeLimit(t *testing.T) {
	testutil.ResetViperConfig(t, "")

	// Set a low total size limit for testing (1KB)
	viper.Set(shared.TestCfgResourceLimitsEnabled, true)
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
	assertStructuredError(t, err, shared.CodeResourceLimitTotalSize)
}

// TestResourceMonitor_MemoryLimitExceeded tests memory limit violation scenarios.
func TestResourceMonitorMemoryLimitExceeded(t *testing.T) {
	testutil.ResetViperConfig(t, "")

	// Set very low memory limit to try to force violations
	viper.Set(shared.TestCfgResourceLimitsEnabled, true)
	viper.Set("resourceLimits.hardMemoryLimitMB", 0.001) // 1KB - extremely low

	rm := NewResourceMonitor()
	defer rm.Close()

	// Allocate large buffer to increase memory usage before check
	largeBuffer := make([]byte, 10*1024*1024) // 10MB allocation
	_ = largeBuffer[0]                        // Use the buffer to prevent optimization

	// Check hard memory limit - might trigger if actual memory is high enough
	err := rm.CheckHardMemoryLimit()

	// Note: This test might not always fail since it depends on actual runtime memory
	// But if it does fail, verify it's the correct error type
	if err != nil {
		validateMemoryLimitError(t, err)
		t.Log("Successfully triggered memory limit violation")
	} else {
		t.Log("Memory limit check passed - actual memory usage may be within limits")
	}
}

// TestResourceMonitor_MemoryLimitHandling tests the memory violation detection.
func TestResourceMonitorMemoryLimitHandling(t *testing.T) {
	testutil.ResetViperConfig(t, "")

	// Enable resource limits with very small hard limit
	viper.Set(shared.TestCfgResourceLimitsEnabled, true)
	viper.Set("resourceLimits.hardMemoryLimitMB", 0.0001) // Very tiny limit
	viper.Set("resourceLimits.enableGracefulDegradation", true)

	rm := NewResourceMonitor()
	defer rm.Close()

	// Allocate more memory to increase chances of triggering limit
	buffers := make([][]byte, 0, 100) // Pre-allocate capacity
	for i := 0; i < 100; i++ {
		buffer := make([]byte, 1024*1024) // 1MB each
		buffers = append(buffers, buffer)
		_ = buffer[0] // Use buffer
		_ = buffers   // Use the slice to prevent unused variable warning

		// Check periodically
		if i%10 == 0 {
			err := rm.CheckHardMemoryLimit()
			if err != nil {
				// Successfully triggered memory limit
				if !strings.Contains(err.Error(), "memory limit") {
					t.Errorf("Expected error message to mention memory limit, got: %v", err)
				}
				t.Log("Successfully triggered memory limit handling")

				return
			}
		}
	}

	t.Log("Could not trigger memory limit - actual memory usage may be lower than limit")
}

// TestResourceMonitorGracefulRecovery tests graceful recovery attempts.
func TestResourceMonitorGracefulRecovery(t *testing.T) {
	testutil.ResetViperConfig(t, "")

	// Set memory limits that will trigger recovery
	viper.Set(shared.TestCfgResourceLimitsEnabled, true)

	rm := NewResourceMonitor()
	defer rm.Close()

	// Force a deterministic 1-byte hard memory limit to trigger recovery
	rm.hardMemoryLimitBytes = 1

	// Process multiple files to accumulate memory usage
	for i := 0; i < 3; i++ {
		filePath := "/tmp/test" + string(rune('1'+i)) + ".txt"
		fileSize := int64(400) // Each file is 400 bytes

		// First few might pass, but eventually should trigger recovery mechanisms
		err := rm.ValidateFileProcessing(filePath, fileSize)
		if err != nil {
			// Once we hit the limit, test that the error is appropriate
			if !strings.Contains(err.Error(), "resource") && !strings.Contains(err.Error(), "limit") {
				t.Errorf("Expected resource limit error, got: %v", err)
			}

			break
		}
		rm.RecordFileProcessed(fileSize)
	}
}
