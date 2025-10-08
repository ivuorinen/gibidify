package fileproc

import (
	"context"
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

const (
	// CI-safe timeout constants
	fastOpTimeout = 100 * time.Millisecond // Operations that should complete quickly
	slowOpMinTime = 10 * time.Millisecond  // Minimum time for blocking operations
)

// cleanupViperConfig is a test helper that captures and restores viper configuration.
// It takes a testing.T and a list of config keys to save/restore.
// Returns a cleanup function that should be called via t.Cleanup.
func cleanupViperConfig(t *testing.T, keys ...string) {
	t.Helper()
	// Capture original values
	origValues := make(map[string]interface{})
	for _, key := range keys {
		origValues[key] = viper.Get(key)
	}
	// Register cleanup to restore values
	t.Cleanup(func() {
		for key, val := range origValues {
			if val != nil {
				viper.Set(key, val)
			}
		}
	})
}

func TestBackpressureManagerCreateChannels(t *testing.T) {
	t.Run("creates buffered channels when enabled", func(t *testing.T) {
		// Capture and restore viper config
		cleanupViperConfig(t, "backpressure.enabled", "backpressure.maxPendingFiles", "backpressure.maxPendingWrites")

		viper.Set("backpressure.enabled", true)
		viper.Set("backpressure.maxPendingFiles", 10)
		viper.Set("backpressure.maxPendingWrites", 10)
		bm := NewBackpressureManager()

		fileCh, writeCh := bm.CreateChannels()
		assert.NotNil(t, fileCh)
		assert.NotNil(t, writeCh)

		// Test that channels have buffer capacity
		assert.Greater(t, cap(fileCh), 0)
		assert.Greater(t, cap(writeCh), 0)

		// Test sending and receiving
		go func() {
			fileCh <- "test.go"
		}()

		val := <-fileCh
		assert.Equal(t, "test.go", val)

		go func() {
			writeCh <- WriteRequest{Content: "test content"}
		}()

		writeReq := <-writeCh
		assert.Equal(t, "test content", writeReq.Content)

		close(fileCh)
		close(writeCh)
	})

	t.Run("creates unbuffered channels when disabled", func(t *testing.T) {
		// Use viper to configure instead of direct field access
		cleanupViperConfig(t, "backpressure.enabled")

		viper.Set("backpressure.enabled", false)
		bm := NewBackpressureManager()

		fileCh, writeCh := bm.CreateChannels()
		assert.NotNil(t, fileCh)
		assert.NotNil(t, writeCh)

		// Unbuffered channels have capacity 0
		assert.Equal(t, 0, cap(fileCh))
		assert.Equal(t, 0, cap(writeCh))

		close(fileCh)
		close(writeCh)
	})
}

func TestBackpressureManagerWaitForChannelSpace(t *testing.T) {
	t.Run("does nothing when disabled", func(t *testing.T) {
		// Use viper to configure instead of direct field access
		cleanupViperConfig(t, "backpressure.enabled")

		viper.Set("backpressure.enabled", false)
		bm := NewBackpressureManager()

		fileCh := make(chan string, 1)
		writeCh := make(chan WriteRequest, 1)

		// Use context with timeout instead of measuring elapsed time
		ctx, cancel := context.WithTimeout(context.Background(), fastOpTimeout)
		defer cancel()

		done := make(chan struct{})
		go func() {
			bm.WaitForChannelSpace(ctx, fileCh, writeCh)
			close(done)
		}()

		// Should return immediately (before timeout)
		select {
		case <-done:
			// Success - operation completed quickly
		case <-ctx.Done():
			t.Fatal("WaitForChannelSpace should return immediately when disabled")
		}

		close(fileCh)
		close(writeCh)
	})

	t.Run("waits when file channel is nearly full", func(t *testing.T) {
		// Use viper to configure instead of direct field access
		cleanupViperConfig(t, "backpressure.enabled", "backpressure.maxPendingFiles")

		viper.Set("backpressure.enabled", true)
		viper.Set("backpressure.maxPendingFiles", 10)
		bm := NewBackpressureManager()

		// Create channel with exact capacity
		fileCh := make(chan string, 10)
		writeCh := make(chan WriteRequest, 10)

		// Fill file channel to >90% (with minimum of 1)
		target := max(1, int(float64(cap(fileCh))*0.9))
		for i := 0; i < target; i++ {
			fileCh <- "file.txt"
		}

		// Test that it blocks by verifying it doesn't complete immediately
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		done := make(chan struct{})
		start := time.Now()
		go func() {
			bm.WaitForChannelSpace(ctx, fileCh, writeCh)
			close(done)
		}()

		// Verify it doesn't complete immediately (within first millisecond)
		select {
		case <-done:
			t.Fatal("WaitForChannelSpace should block when channel is nearly full")
		case <-time.After(1 * time.Millisecond):
			// Good - it's blocking as expected
		}

		// Wait for it to complete
		<-done
		duration := time.Since(start)
		// Just verify it took some measurable time (very lenient for CI)
		assert.GreaterOrEqual(t, duration, 1*time.Millisecond)

		// Clean up
		for i := 0; i < target; i++ {
			<-fileCh
		}
		close(fileCh)
		close(writeCh)
	})

	t.Run("waits when write channel is nearly full", func(t *testing.T) {
		// Use viper to configure instead of direct field access
		cleanupViperConfig(t, "backpressure.enabled", "backpressure.maxPendingWrites")

		viper.Set("backpressure.enabled", true)
		viper.Set("backpressure.maxPendingWrites", 10)
		bm := NewBackpressureManager()

		fileCh := make(chan string, 10)
		writeCh := make(chan WriteRequest, 10)

		// Fill write channel to >90% (with minimum of 1)
		target := max(1, int(float64(cap(writeCh))*0.9))
		for i := 0; i < target; i++ {
			writeCh <- WriteRequest{}
		}

		// Test that it blocks by verifying it doesn't complete immediately
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		done := make(chan struct{})
		start := time.Now()
		go func() {
			bm.WaitForChannelSpace(ctx, fileCh, writeCh)
			close(done)
		}()

		// Verify it doesn't complete immediately (within first millisecond)
		select {
		case <-done:
			t.Fatal("WaitForChannelSpace should block when channel is nearly full")
		case <-time.After(1 * time.Millisecond):
			// Good - it's blocking as expected
		}

		// Wait for it to complete
		<-done
		duration := time.Since(start)
		// Just verify it took some measurable time (very lenient for CI)
		assert.GreaterOrEqual(t, duration, 1*time.Millisecond)

		// Clean up
		for i := 0; i < target; i++ {
			<-writeCh
		}
		close(fileCh)
		close(writeCh)
	})

	t.Run("respects context cancellation", func(t *testing.T) {
		// Use viper to configure instead of direct field access
		cleanupViperConfig(t, "backpressure.enabled", "backpressure.maxPendingFiles")

		viper.Set("backpressure.enabled", true)
		viper.Set("backpressure.maxPendingFiles", 10)
		bm := NewBackpressureManager()

		fileCh := make(chan string, 10)
		writeCh := make(chan WriteRequest, 10)

		// Fill channel
		for i := 0; i < 10; i++ {
			fileCh <- "file.txt"
		}

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		// Use timeout to verify it returns quickly
		done := make(chan struct{})
		go func() {
			bm.WaitForChannelSpace(ctx, fileCh, writeCh)
			close(done)
		}()

		// Should return quickly when context is cancelled
		select {
		case <-done:
			// Success - returned due to cancellation
		case <-time.After(fastOpTimeout):
			t.Fatal("WaitForChannelSpace should return immediately when context is cancelled")
		}

		// Clean up
		for i := 0; i < 10; i++ {
			<-fileCh
		}
		close(fileCh)
		close(writeCh)
	})
}
