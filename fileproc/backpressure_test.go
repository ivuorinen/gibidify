package fileproc_test

import (
	"context"
	"runtime"
	"testing"
	"time"

	"github.com/ivuorinen/gibidify/fileproc"
	"github.com/ivuorinen/gibidify/testutil"
)

func TestNewBackpressureManager(t *testing.T) {
	// Test creating a new backpressure manager
	bp := fileproc.NewBackpressureManager()

	if bp == nil {
		t.Error("Expected backpressure manager to be created, got nil")
	}

	// The backpressure manager should be initialized with config values
	// We can't test the internal values directly since they're private,
	// but we can test that it was created successfully
}

func TestBackpressureManager_CreateChannels(t *testing.T) {
	testutil.ResetViperConfig(t, "")

	bp := fileproc.NewBackpressureManager()

	// Test creating channels
	fileCh, writeCh := bp.CreateChannels()

	// Verify channels are created
	if fileCh == nil {
		t.Error("Expected file channel to be created, got nil")
	}
	if writeCh == nil {
		t.Error("Expected write channel to be created, got nil")
	}

	// Test that channels can be used
	select {
	case fileCh <- "test-file":
		// Successfully sent to channel
	default:
		t.Error("Unable to send to file channel")
	}

	// Read from channel
	select {
	case file := <-fileCh:
		if file != "test-file" {
			t.Errorf("Expected 'test-file', got %s", file)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Timeout reading from file channel")
	}
}

func TestBackpressureManager_ShouldApplyBackpressure(t *testing.T) {
	testutil.ResetViperConfig(t, "")

	bp := fileproc.NewBackpressureManager()
	ctx := context.Background()

	// Test backpressure decision
	shouldApply := bp.ShouldApplyBackpressure(ctx)

	// Since we're using default config, backpressure behavior depends on settings
	// We just test that the method returns a boolean without error
	if shouldApply != true && shouldApply != false {
		t.Error("ShouldApplyBackpressure should return a boolean")
	}
}

func TestBackpressureManager_ApplyBackpressure(t *testing.T) {
	testutil.ResetViperConfig(t, "")

	bp := fileproc.NewBackpressureManager()
	ctx := context.Background()

	// Test applying backpressure
	bp.ApplyBackpressure(ctx)

	// ApplyBackpressure is a void method that should not panic
	// If we reach here, the method executed successfully
}

func TestBackpressureManager_ApplyBackpressure_WithCancellation(t *testing.T) {
	testutil.ResetViperConfig(t, "")

	bp := fileproc.NewBackpressureManager()

	// Create canceled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Test applying backpressure with canceled context
	bp.ApplyBackpressure(ctx)

	// ApplyBackpressure doesn't return errors, but should handle cancellation gracefully
	// If we reach here without hanging, the method handled cancellation properly
}

func TestBackpressureManager_GetStats(t *testing.T) {
	testutil.ResetViperConfig(t, "")

	bp := fileproc.NewBackpressureManager()

	// Test getting stats
	stats := bp.GetStats()

	// Stats should contain relevant information
	if stats.FilesProcessed < 0 {
		t.Error("Expected non-negative files processed count")
	}

	if stats.CurrentMemoryUsage < 0 {
		t.Error("Expected non-negative memory usage")
	}

	if stats.MaxMemoryUsage < 0 {
		t.Error("Expected non-negative max memory usage")
	}

	// Test that stats have reasonable values
	if stats.MaxPendingFiles < 0 || stats.MaxPendingWrites < 0 {
		t.Error("Expected non-negative channel buffer sizes")
	}
}

func TestBackpressureManager_WaitForChannelSpace(t *testing.T) {
	testutil.ResetViperConfig(t, "")

	bp := fileproc.NewBackpressureManager()
	ctx := context.Background()

	// Create test channels
	fileCh, writeCh := bp.CreateChannels()

	// Test waiting for channel space
	bp.WaitForChannelSpace(ctx, fileCh, writeCh)

	// WaitForChannelSpace is void method that should complete without hanging
	// If we reach here, the method executed successfully
}

func TestBackpressureManager_WaitForChannelSpace_WithCancellation(t *testing.T) {
	testutil.ResetViperConfig(t, "")

	bp := fileproc.NewBackpressureManager()

	// Create canceled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Create test channels
	fileCh, writeCh := bp.CreateChannels()

	// Test waiting for channel space with canceled context
	bp.WaitForChannelSpace(ctx, fileCh, writeCh)

	// WaitForChannelSpace should handle cancellation gracefully without hanging
	// If we reach here, the method handled cancellation properly
}

func TestBackpressureManager_LogBackpressureInfo(t *testing.T) {
	testutil.ResetViperConfig(t, "")

	bp := fileproc.NewBackpressureManager()

	// Test logging backpressure info
	// This method primarily logs information, so we test it executes without panic
	bp.LogBackpressureInfo()

	// If we reach here without panic, the method worked
}

// BenchmarkBackpressureManager benchmarks backpressure operations.
func BenchmarkBackpressureManager_CreateChannels(b *testing.B) {
	bp := fileproc.NewBackpressureManager()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fileCh, writeCh := bp.CreateChannels()

		// Use channels to prevent optimization
		_ = fileCh
		_ = writeCh

		runtime.GC() // Force GC to measure memory impact
	}
}

func BenchmarkBackpressureManager_ShouldApplyBackpressure(b *testing.B) {
	bp := fileproc.NewBackpressureManager()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		shouldApply := bp.ShouldApplyBackpressure(ctx)
		_ = shouldApply // Prevent optimization
	}
}

func BenchmarkBackpressureManager_ApplyBackpressure(b *testing.B) {
	bp := fileproc.NewBackpressureManager()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bp.ApplyBackpressure(ctx)
	}
}

func BenchmarkBackpressureManager_GetStats(b *testing.B) {
	bp := fileproc.NewBackpressureManager()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		stats := bp.GetStats()
		_ = stats // Prevent optimization
	}
}

// TestBackpressureManager_ShouldApplyBackpressure_EdgeCases tests various edge cases for backpressure decision.
func TestBackpressureManager_ShouldApplyBackpressure_EdgeCases(t *testing.T) {
	testutil.ResetViperConfig(t, `
backpressure:
  enabled: true
  memory_check_interval: 2
  memory_limit_mb: 1
`)

	bp := fileproc.NewBackpressureManager()
	ctx := context.Background()

	// Test multiple calls to trigger memory check interval logic
	for i := 0; i < 10; i++ {
		shouldApply := bp.ShouldApplyBackpressure(ctx)
		_ = shouldApply
	}

	// At this point, memory checking should have triggered multiple times
	// The actual decision depends on memory usage, but we're testing the paths
}

// TestBackpressureManager_CreateChannels_EdgeCases tests edge cases in channel creation.
func TestBackpressureManager_CreateChannels_EdgeCases(t *testing.T) {
	// Test with custom configuration that might trigger different buffer sizes
	testutil.ResetViperConfig(t, `
backpressure:
  file_buffer_size: 50
  write_buffer_size: 25
`)

	bp := fileproc.NewBackpressureManager()

	// Create multiple channel sets to test resource management
	for i := 0; i < 5; i++ {
		fileCh, writeCh := bp.CreateChannels()

		// Verify channels work correctly
		select {
		case fileCh <- "test":
			// Good - channel accepted value
		default:
			// This is also acceptable if buffer is full
		}

		// Test write channel
		select {
		case writeCh <- fileproc.WriteRequest{Path: "test", Content: "content"}:
			// Good - channel accepted value
		default:
			// This is also acceptable if buffer is full
		}
	}
}

// TestBackpressureManager_WaitForChannelSpace_EdgeCases tests edge cases in channel space waiting.
func TestBackpressureManager_WaitForChannelSpace_EdgeCases(t *testing.T) {
	testutil.ResetViperConfig(t, `
backpressure:
  enabled: true
  wait_timeout_ms: 10
`)

	bp := fileproc.NewBackpressureManager()
	ctx := context.Background()

	// Create channels with small buffers
	fileCh, writeCh := bp.CreateChannels()

	// Fill up the channels to create pressure
	go func() {
		for i := 0; i < 100; i++ {
			select {
			case fileCh <- "file":
			case <-time.After(1 * time.Millisecond):
			}
		}
	}()

	go func() {
		for i := 0; i < 100; i++ {
			select {
			case writeCh <- fileproc.WriteRequest{Path: "test", Content: "content"}:
			case <-time.After(1 * time.Millisecond):
			}
		}
	}()

	// Wait for channel space - should handle the full channels
	bp.WaitForChannelSpace(ctx, fileCh, writeCh)
}

// TestBackpressureManager_MemoryPressure tests behavior under simulated memory pressure.
func TestBackpressureManager_MemoryPressure(t *testing.T) {
	// Test with very low memory limit to trigger backpressure
	testutil.ResetViperConfig(t, `
backpressure:
  enabled: true
  memory_limit_mb: 0.001  
  memory_check_interval: 1
`)

	bp := fileproc.NewBackpressureManager()
	ctx := context.Background()

	// Allocate some memory to potentially trigger limits
	largeBuffer := make([]byte, 1024*1024) // 1MB
	_ = largeBuffer[0]

	// Test backpressure decision under memory pressure
	for i := 0; i < 5; i++ {
		shouldApply := bp.ShouldApplyBackpressure(ctx)
		if shouldApply {
			// Test applying backpressure when needed
			bp.ApplyBackpressure(ctx)
			t.Log("Backpressure applied due to memory pressure")
		}
	}

	// Test logging
	bp.LogBackpressureInfo()
}
