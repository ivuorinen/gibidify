package fileproc

import (
	"testing"
	"time"

	"github.com/spf13/viper"

	"github.com/ivuorinen/gibidify/testutil"
)

func TestResourceMonitorMetrics(t *testing.T) {
	testutil.ResetViperConfig(t, "")

	viper.Set("resourceLimits.enabled", true)
	viper.Set("resourceLimits.enableResourceMonitoring", true)

	rm := NewResourceMonitor()
	defer rm.Close()

	// Process some files to generate metrics
	rm.RecordFileProcessed(1000)
	rm.RecordFileProcessed(2000)
	rm.RecordFileProcessed(500)

	metrics := rm.Metrics()

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
