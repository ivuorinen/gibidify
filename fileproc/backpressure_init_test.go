package fileproc

import (
	"context"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

// viperState captures the original state of a viper key.
type viperState struct {
	value  interface{}
	wasSet bool
}

// setupViperCleanup captures the current state of viper keys and registers a cleanup
// that restores previously-set keys to their original values and unsets keys that
// were not originally set, preventing pollution of global viper state between tests.
func setupViperCleanup(t *testing.T, keys []string) {
	t.Helper()
	original := make(map[string]viperState)
	for _, key := range keys {
		original[key] = viperState{
			value:  viper.Get(key),
			wasSet: viper.IsSet(key),
		}
	}

	t.Cleanup(func() {
		for key, state := range original {
			if state.wasSet {
				viper.Set(key, state.value)
			} else {
				// Key was not originally set, so unset it
				viper.Set(key, nil)
			}
		}
	})
}

func TestNewBackpressureManager(t *testing.T) {
	keys := []string{
		"backpressure.enabled",
		"backpressure.maxMemoryUsage",
		"backpressure.memoryCheckInterval",
		"backpressure.maxPendingFiles",
		"backpressure.maxPendingWrites",
	}
	setupViperCleanup(t, keys)

	viper.Set("backpressure.enabled", true)
	viper.Set("backpressure.maxMemoryUsage", 100)
	viper.Set("backpressure.memoryCheckInterval", 10)
	viper.Set("backpressure.maxPendingFiles", 10)
	viper.Set("backpressure.maxPendingWrites", 10)

	bm := NewBackpressureManager()
	assert.NotNil(t, bm)
	assert.True(t, bm.enabled)
	assert.Greater(t, bm.maxMemoryUsage, int64(0))
	assert.Greater(t, bm.memoryCheckInterval, 0)
	assert.Greater(t, bm.maxPendingFiles, 0)
	assert.Greater(t, bm.maxPendingWrites, 0)
	assert.Equal(t, int64(0), bm.filesProcessed)
}

func TestBackpressureStatsStructure(t *testing.T) {
	// Behavioral test that exercises BackpressureManager and validates stats
	keys := []string{
		"backpressure.enabled",
		"backpressure.maxMemoryUsage",
		"backpressure.memoryCheckInterval",
		"backpressure.maxPendingFiles",
		"backpressure.maxPendingWrites",
	}
	setupViperCleanup(t, keys)

	// Configure backpressure with realistic settings
	viper.Set("backpressure.enabled", true)
	viper.Set("backpressure.maxMemoryUsage", 100*1024*1024) // 100MB
	viper.Set("backpressure.memoryCheckInterval", 1)        // Check every file
	viper.Set("backpressure.maxPendingFiles", 1000)
	viper.Set("backpressure.maxPendingWrites", 500)

	bm := NewBackpressureManager()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Simulate processing files
	initialStats := bm.GetStats()
	assert.True(t, initialStats.Enabled, "backpressure should be enabled")
	assert.Equal(t, int64(0), initialStats.FilesProcessed, "initially no files processed")

	// Process some files to trigger memory checks
	for i := 0; i < 5; i++ {
		bm.ShouldApplyBackpressure(ctx)
	}

	// Verify stats reflect the operations
	stats := bm.GetStats()
	assert.True(t, stats.Enabled, "enabled flag should be set")
	assert.Equal(t, int64(5), stats.FilesProcessed, "should have processed 5 files")
	assert.Greater(t, stats.CurrentMemoryUsage, int64(0), "memory usage should be tracked")
	assert.Equal(t, int64(100*1024*1024), stats.MaxMemoryUsage, "max memory should match config")
	assert.Equal(t, 1000, stats.MaxPendingFiles, "maxPendingFiles should match config")
	assert.Equal(t, 500, stats.MaxPendingWrites, "maxPendingWrites should match config")
	assert.NotZero(t, stats.LastMemoryCheck, "lastMemoryCheck should be set after checks")
}

func TestBackpressureManagerGetStats(t *testing.T) {
	keys := []string{
		"backpressure.enabled",
		"backpressure.memoryCheckInterval",
	}
	setupViperCleanup(t, keys)

	// Ensure config enables backpressure and checks every call
	viper.Set("backpressure.enabled", true)
	viper.Set("backpressure.memoryCheckInterval", 1)

	bm := NewBackpressureManager()

	// Process some files to update stats
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for i := 0; i < 5; i++ {
		bm.ShouldApplyBackpressure(ctx)
	}

	stats := bm.GetStats()

	assert.True(t, stats.Enabled)
	assert.Equal(t, int64(5), stats.FilesProcessed)
	assert.Greater(t, stats.CurrentMemoryUsage, int64(0))
	assert.Equal(t, bm.maxMemoryUsage, stats.MaxMemoryUsage)
	assert.Equal(t, bm.maxPendingFiles, stats.MaxPendingFiles)
	assert.Equal(t, bm.maxPendingWrites, stats.MaxPendingWrites)

	// LastMemoryCheck should be set if we hit the interval
	if bm.memoryCheckInterval <= 5 {
		assert.NotZero(t, stats.LastMemoryCheck)
	}
}
