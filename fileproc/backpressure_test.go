package fileproc

import (
	"context"
	"testing"
	"time"

	"github.com/spf13/viper"
)

func TestNewBackpressureManager(t *testing.T) {
	// Save original values
	origEnabled := viper.Get("backpressure.enabled")
	origMaxMemory := viper.Get("backpressure.maxMemoryUsage")
	origCheckInterval := viper.Get("backpressure.memoryCheckInterval")
	origMaxPendingFiles := viper.Get("backpressure.maxPendingFiles")
	origMaxPendingWrites := viper.Get("backpressure.maxPendingWrites")

	defer func() {
		viper.Set("backpressure.enabled", origEnabled)
		viper.Set("backpressure.maxMemoryUsage", origMaxMemory)
		viper.Set("backpressure.memoryCheckInterval", origCheckInterval)
		viper.Set("backpressure.maxPendingFiles", origMaxPendingFiles)
		viper.Set("backpressure.maxPendingWrites", origMaxPendingWrites)
	}()

	// Set test values
	viper.Set("backpressure.enabled", true)
	viper.Set("backpressure.maxMemoryUsage", int64(1024*1024*100))
	viper.Set("backpressure.memoryCheckInterval", 5)
	viper.Set("backpressure.maxPendingFiles", 50)
	viper.Set("backpressure.maxPendingWrites", 25)

	bp := NewBackpressureManager()

	if !bp.enabled {
		t.Error("Expected backpressure to be enabled")
	}

	if bp.maxMemoryUsage != int64(1024*1024*100) {
		t.Errorf("Expected maxMemoryUsage to be %d, got %d", int64(1024*1024*100), bp.maxMemoryUsage)
	}

	if bp.memoryCheckInterval != 5 {
		t.Errorf("Expected memoryCheckInterval to be 5, got %d", bp.memoryCheckInterval)
	}

	if bp.maxPendingFiles != 50 {
		t.Errorf("Expected maxPendingFiles to be 50, got %d", bp.maxPendingFiles)
	}

	if bp.maxPendingWrites != 25 {
		t.Errorf("Expected maxPendingWrites to be 25, got %d", bp.maxPendingWrites)
	}
}

func TestCreateChannels(t *testing.T) {
	tests := []struct {
		name                string
		enabled             bool
		maxPendingFiles     int
		maxPendingWrites    int
		wantFileChBuffered  bool
		wantWriteChBuffered bool
	}{
		{
			name:                "buffered channels when enabled",
			enabled:             true,
			maxPendingFiles:     10,
			maxPendingWrites:    5,
			wantFileChBuffered:  true,
			wantWriteChBuffered: true,
		},
		{
			name:                "unbuffered channels when disabled",
			enabled:             false,
			maxPendingFiles:     10,
			maxPendingWrites:    5,
			wantFileChBuffered:  false,
			wantWriteChBuffered: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bp := &BackpressureManager{
				enabled:          tt.enabled,
				maxPendingFiles:  tt.maxPendingFiles,
				maxPendingWrites: tt.maxPendingWrites,
			}

			fileCh, writeCh := bp.CreateChannels()

			fileChCap := cap(fileCh)
			writeChCap := cap(writeCh)

			if tt.wantFileChBuffered && fileChCap == 0 {
				t.Error("Expected file channel to be buffered, but it's unbuffered")
			}

			if !tt.wantFileChBuffered && fileChCap != 0 {
				t.Errorf("Expected file channel to be unbuffered, but it has capacity %d", fileChCap)
			}

			if tt.wantWriteChBuffered && writeChCap == 0 {
				t.Error("Expected write channel to be buffered, but it's unbuffered")
			}

			if !tt.wantWriteChBuffered && writeChCap != 0 {
				t.Errorf("Expected write channel to be unbuffered, but it has capacity %d", writeChCap)
			}
		})
	}
}

func TestShouldApplyBackpressure(t *testing.T) {
	bp := &BackpressureManager{
		enabled:             true,
		maxMemoryUsage:      1024 * 1024 * 10, // 10MB
		memoryCheckInterval: 1,
		lastMemoryCheck:     time.Now().Add(-2 * time.Second),
	}

	// Test with context
	ctx := context.Background()

	// Should check memory since interval passed
	bp.ShouldApplyBackpressure(ctx)

	// Test disabled backpressure
	bpDisabled := &BackpressureManager{
		enabled: false,
	}

	bpDisabled.ShouldApplyBackpressure(ctx)
}

func TestApplyBackpressure(t *testing.T) {
	bp := &BackpressureManager{
		enabled:          true,
		maxPendingFiles:  10,
		maxPendingWrites: 5,
	}

	ctx := context.Background()

	// Test applying backpressure
	bp.ApplyBackpressure(ctx)

	// Test when disabled
	bpDisabled := &BackpressureManager{
		enabled: false,
	}

	bpDisabled.ApplyBackpressure(ctx)
}

func TestGetStats(t *testing.T) {
	bp := &BackpressureManager{
		filesProcessed: 100,
	}

	stats := bp.GetStats()

	if stats.FilesProcessed != int64(100) {
		t.Errorf("Expected files_processed to be 100, got %d", stats.FilesProcessed)
	}

	if stats.CurrentMemoryUsage == 0 {
		t.Error("Expected current memory usage to be non-zero")
	}
}

func TestWaitForChannelSpace(t *testing.T) {
	bp := &BackpressureManager{
		enabled:          true,
		maxPendingFiles:  2,
		maxPendingWrites: 2,
	}

	fileCh, writeCh := bp.CreateChannels()

	// Fill the channels
	fileCh <- "file1"
	fileCh <- "file2"
	writeCh <- WriteRequest{Path: "test1"}
	writeCh <- WriteRequest{Path: "test2"}

	// Test waiting for space (should not block since we'll drain immediately)
	go func() {
		time.Sleep(10 * time.Millisecond)
		<-fileCh
		<-writeCh
	}()

	ctx := context.Background()
	bp.WaitForChannelSpace(ctx, fileCh, writeCh)

	// Clean up channels
	close(fileCh)
	close(writeCh)
}

func TestLogBackpressureInfo(t *testing.T) {
	bp := &BackpressureManager{
		enabled:          true,
		filesProcessed:   50,
		maxPendingFiles:  10,
		maxPendingWrites: 5,
		maxMemoryUsage:   1024 * 1024 * 100,
	}

	// Test logging (just ensure it doesn't panic)
	bp.LogBackpressureInfo()

	// Test with disabled
	bpDisabled := &BackpressureManager{
		enabled: false,
	}
	bpDisabled.LogBackpressureInfo()
}
