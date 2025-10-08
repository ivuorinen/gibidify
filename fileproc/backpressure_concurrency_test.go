package fileproc

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBackpressureManagerConcurrency(t *testing.T) {
	// Configure via viper instead of direct field access
	origEnabled := viper.Get("backpressure.enabled")
	t.Cleanup(func() {
		if origEnabled != nil {
			viper.Set("backpressure.enabled", origEnabled)
		}
	})
	viper.Set("backpressure.enabled", true)

	bm := NewBackpressureManager()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup

	// Multiple goroutines checking backpressure
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			bm.ShouldApplyBackpressure(ctx)
		}()
	}

	// Multiple goroutines applying backpressure
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			bm.ApplyBackpressure(ctx)
		}()
	}

	// Multiple goroutines getting stats
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			bm.GetStats()
		}()
	}

	// Multiple goroutines creating channels
	// Note: CreateChannels returns new channels each time, caller owns them
	type channelResult struct {
		fileCh  chan string
		writeCh chan WriteRequest
	}
	results := make(chan channelResult, 3)
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			fileCh, writeCh := bm.CreateChannels()
			results <- channelResult{fileCh, writeCh}
		}()
	}

	wg.Wait()
	close(results)

	// Verify channels are created and have expected properties
	for result := range results {
		assert.NotNil(t, result.fileCh)
		assert.NotNil(t, result.writeCh)
		// Close channels to prevent resource leak (caller owns them)
		close(result.fileCh)
		close(result.writeCh)
	}

	// Verify stats are consistent
	stats := bm.GetStats()
	assert.GreaterOrEqual(t, stats.FilesProcessed, int64(10))
}

func TestBackpressureManagerIntegration(t *testing.T) {
	// Configure via viper instead of direct field access
	origEnabled := viper.Get("backpressure.enabled")
	origMaxFiles := viper.Get("backpressure.maxPendingFiles")
	origMaxWrites := viper.Get("backpressure.maxPendingWrites")
	origCheckInterval := viper.Get("backpressure.memoryCheckInterval")
	origMaxMemory := viper.Get("backpressure.maxMemoryUsage")
	t.Cleanup(func() {
		if origEnabled != nil {
			viper.Set("backpressure.enabled", origEnabled)
		}
		if origMaxFiles != nil {
			viper.Set("backpressure.maxPendingFiles", origMaxFiles)
		}
		if origMaxWrites != nil {
			viper.Set("backpressure.maxPendingWrites", origMaxWrites)
		}
		if origCheckInterval != nil {
			viper.Set("backpressure.memoryCheckInterval", origCheckInterval)
		}
		if origMaxMemory != nil {
			viper.Set("backpressure.maxMemoryUsage", origMaxMemory)
		}
	})

	viper.Set("backpressure.enabled", true)
	viper.Set("backpressure.maxPendingFiles", 10)
	viper.Set("backpressure.maxPendingWrites", 10)
	viper.Set("backpressure.memoryCheckInterval", 10)
	viper.Set("backpressure.maxMemoryUsage", 100*1024*1024) // 100MB

	bm := NewBackpressureManager()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create channels - caller owns these channels and is responsible for closing them
	fileCh, writeCh := bm.CreateChannels()
	require.NotNil(t, fileCh)
	require.NotNil(t, writeCh)
	require.Greater(t, cap(fileCh), 0, "fileCh should be buffered")
	require.Greater(t, cap(writeCh), 0, "writeCh should be buffered")

	// Simulate file processing
	var wg sync.WaitGroup

	// Producer
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			// Check for backpressure
			if bm.ShouldApplyBackpressure(ctx) {
				bm.ApplyBackpressure(ctx)
			}

			// Wait for channel space if needed
			bm.WaitForChannelSpace(ctx, fileCh, writeCh)

			select {
			case fileCh <- "file.txt":
				// File sent
			case <-ctx.Done():
				return
			}
		}
	}()

	// Consumer
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			select {
			case <-fileCh:
				// Process file (do not manually increment filesProcessed)
			case <-ctx.Done():
				return
			}
		}
	}()

	// Wait for completion
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Success
	case <-time.After(5 * time.Second):
		t.Fatal("Integration test timeout")
	}

	// Log final info
	bm.LogBackpressureInfo()

	// Check final stats
	stats := bm.GetStats()
	assert.GreaterOrEqual(t, stats.FilesProcessed, int64(100))

	// Clean up - caller owns the channels, safe to close now that goroutines have finished
	close(fileCh)
	close(writeCh)
}
