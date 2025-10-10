package fileproc

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestBackpressureManagerShouldApplyBackpressure(t *testing.T) {
	ctx := context.Background()

	t.Run("returns false when disabled", func(t *testing.T) {
		bm := NewBackpressureManager()
		bm.enabled = false

		shouldApply := bm.ShouldApplyBackpressure(ctx)
		assert.False(t, shouldApply)
	})

	t.Run("checks memory at intervals", func(_ *testing.T) {
		bm := NewBackpressureManager()
		bm.enabled = true
		bm.memoryCheckInterval = 10

		// Should not check memory on most calls
		for i := 1; i < 10; i++ {
			shouldApply := bm.ShouldApplyBackpressure(ctx)
			// Can't predict result, but shouldn't panic
			_ = shouldApply
		}

		// Should check memory on 10th call
		shouldApply := bm.ShouldApplyBackpressure(ctx)
		// Result depends on actual memory usage
		_ = shouldApply
	})

	t.Run("detects high memory usage", func(t *testing.T) {
		bm := NewBackpressureManager()
		bm.enabled = true
		bm.memoryCheckInterval = 1
		bm.maxMemoryUsage = 1 // Set very low limit to trigger

		shouldApply := bm.ShouldApplyBackpressure(ctx)
		// Should detect high memory usage
		assert.True(t, shouldApply)
	})
}

func TestBackpressureManagerApplyBackpressure(t *testing.T) {
	ctx := context.Background()

	t.Run("does nothing when disabled", func(t *testing.T) {
		bm := NewBackpressureManager()
		bm.enabled = false

		// Use a channel to verify the function returns quickly
		done := make(chan struct{})
		go func() {
			bm.ApplyBackpressure(ctx)
			close(done)
		}()

		// Should complete quickly when disabled
		select {
		case <-done:
			// Success - function returned
		case <-time.After(50 * time.Millisecond):
			t.Fatal("ApplyBackpressure did not return quickly when disabled")
		}
	})

	t.Run("applies delay when enabled", func(t *testing.T) {
		bm := NewBackpressureManager()
		bm.enabled = true

		// Use a channel to verify the function blocks for some time
		done := make(chan struct{})
		started := make(chan struct{})
		go func() {
			close(started)
			bm.ApplyBackpressure(ctx)
			close(done)
		}()

		// Wait for goroutine to start
		<-started

		// Should NOT complete immediately - verify it blocks for at least 5ms
		select {
		case <-done:
			t.Fatal("ApplyBackpressure returned too quickly when enabled")
		case <-time.After(5 * time.Millisecond):
			// Good - it's blocking as expected
		}

		// Now wait for it to complete (should finish within reasonable time)
		select {
		case <-done:
			// Success - function eventually returned
		case <-time.After(500 * time.Millisecond):
			t.Fatal("ApplyBackpressure did not complete within timeout")
		}
	})

	t.Run("respects context cancellation", func(t *testing.T) {
		bm := NewBackpressureManager()
		bm.enabled = true

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		start := time.Now()
		bm.ApplyBackpressure(ctx)
		duration := time.Since(start)

		// Should return quickly when context is cancelled
		assert.Less(t, duration, 5*time.Millisecond)
	})
}

func TestBackpressureManagerLogBackpressureInfo(t *testing.T) {
	bm := NewBackpressureManager()
	bm.enabled = true // Ensure enabled so filesProcessed is incremented

	// Apply some operations
	ctx := context.Background()
	bm.ShouldApplyBackpressure(ctx)
	bm.ApplyBackpressure(ctx)

	// This should not panic
	bm.LogBackpressureInfo()

	stats := bm.GetStats()
	assert.Greater(t, stats.FilesProcessed, int64(0))
}

func TestBackpressureManagerMemoryLimiting(t *testing.T) {
	t.Run("triggers on low memory limit", func(t *testing.T) {
		bm := NewBackpressureManager()
		bm.enabled = true
		bm.memoryCheckInterval = 1 // Check every file
		bm.maxMemoryUsage = 1      // Very low limit to guarantee trigger

		ctx := context.Background()

		// Should detect memory over limit
		shouldApply := bm.ShouldApplyBackpressure(ctx)
		assert.True(t, shouldApply)
		stats := bm.GetStats()
		assert.True(t, stats.MemoryWarningActive)
	})

	t.Run("resets warning when memory normalizes", func(t *testing.T) {
		bm := NewBackpressureManager()
		bm.enabled = true
		bm.memoryCheckInterval = 1
		// Simulate warning by first triggering high memory usage
		bm.maxMemoryUsage = 1 // Very low to trigger warning
		ctx := context.Background()
		_ = bm.ShouldApplyBackpressure(ctx)
		stats := bm.GetStats()
		assert.True(t, stats.MemoryWarningActive)

		// Now set high limit so we're under it
		bm.maxMemoryUsage = 1024 * 1024 * 1024 * 10 // 10GB

		shouldApply := bm.ShouldApplyBackpressure(ctx)
		assert.False(t, shouldApply)

		// Warning should be reset (via public API)
		stats = bm.GetStats()
		assert.False(t, stats.MemoryWarningActive)
	})
}
