package fileproc_test

import (
	"context"
	"runtime"
	"testing"
	"time"

	"github.com/ivuorinen/gibidify/fileproc"
	"github.com/ivuorinen/gibidify/shared"
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

func TestBackpressureManagerCreateChannels(t *testing.T) {
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

func TestBackpressureManagerShouldApplyBackpressure(t *testing.T) {
	testutil.ResetViperConfig(t, "")

	bp := fileproc.NewBackpressureManager()
	ctx := context.Background()

	// Test backpressure decision
	shouldApply := bp.ShouldApplyBackpressure(ctx)

	// Since we're using default config, backpressure behavior depends on settings
	// We just test that the method returns without error
	// shouldApply is a valid boolean value
	_ = shouldApply
}

func TestBackpressureManagerApplyBackpressure(t *testing.T) {
	testutil.ResetViperConfig(t, "")

	bp := fileproc.NewBackpressureManager()
	ctx := context.Background()

	// Test applying backpressure
	bp.ApplyBackpressure(ctx)

	// ApplyBackpressure is a void method that should not panic
	// If we reach here, the method executed successfully
}

func TestBackpressureManagerApplyBackpressureWithCancellation(t *testing.T) {
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

func TestBackpressureManagerGetStats(t *testing.T) {
	testutil.ResetViperConfig(t, "")

	bp := fileproc.NewBackpressureManager()

	// Test getting stats
	stats := bp.Stats()

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

func TestBackpressureManagerWaitForChannelSpace(t *testing.T) {
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

func TestBackpressureManagerWaitForChannelSpaceWithCancellation(t *testing.T) {
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

func TestBackpressureManagerLogBackpressureInfo(t *testing.T) {
	testutil.ResetViperConfig(t, "")

	bp := fileproc.NewBackpressureManager()

	// Test logging backpressure info
	// This method primarily logs information, so we test it executes without panic
	bp.LogBackpressureInfo()

	// If we reach here without panic, the method worked
}

// BenchmarkBackpressureManager benchmarks backpressure operations.
func BenchmarkBackpressureManagerCreateChannels(b *testing.B) {
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

func BenchmarkBackpressureManagerShouldApplyBackpressure(b *testing.B) {
	bp := fileproc.NewBackpressureManager()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		shouldApply := bp.ShouldApplyBackpressure(ctx)
		_ = shouldApply // Prevent optimization
	}
}

func BenchmarkBackpressureManagerApplyBackpressure(b *testing.B) {
	bp := fileproc.NewBackpressureManager()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bp.ApplyBackpressure(ctx)
	}
}

func BenchmarkBackpressureManagerGetStats(b *testing.B) {
	bp := fileproc.NewBackpressureManager()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		stats := bp.Stats()
		_ = stats // Prevent optimization
	}
}

// TestBackpressureManager_ShouldApplyBackpressure_EdgeCases tests various edge cases for backpressure decision.
func TestBackpressureManagerShouldApplyBackpressureEdgeCases(t *testing.T) {
	testutil.ApplyBackpressureOverrides(t, map[string]any{
		shared.ConfigKeyBackpressureEnabled:  true,
		"backpressure.memory_check_interval": 2,
		"backpressure.memory_limit_mb":       1,
	})

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
func TestBackpressureManagerCreateChannelsEdgeCases(t *testing.T) {
	// Test with custom configuration that might trigger different buffer sizes
	testutil.ApplyBackpressureOverrides(t, map[string]any{
		"backpressure.file_buffer_size":  50,
		"backpressure.write_buffer_size": 25,
	})

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
func TestBackpressureManagerWaitForChannelSpaceEdgeCases(t *testing.T) {
	testutil.ApplyBackpressureOverrides(t, map[string]any{
		shared.ConfigKeyBackpressureEnabled: true,
		"backpressure.wait_timeout_ms":      10,
	})

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
func TestBackpressureManagerMemoryPressure(t *testing.T) {
	// Test with very low memory limit to trigger backpressure
	testutil.ApplyBackpressureOverrides(t, map[string]any{
		shared.ConfigKeyBackpressureEnabled:  true,
		"backpressure.memory_limit_mb":       0.001,
		"backpressure.memory_check_interval": 1,
	})

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
