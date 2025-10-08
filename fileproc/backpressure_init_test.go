package fileproc

import (
	"context"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

// setupViperCleanup is a test helper that captures and restores viper configuration.
// It takes a testing.T and a list of config keys to save/restore.
func setupViperCleanup(t *testing.T, keys []string) {
	t.Helper()
	// Capture original values and track which keys existed
	origValues := make(map[string]interface{})
	keysExisted := make(map[string]bool)
	for _, key := range keys {
		val := viper.Get(key)
		origValues[key] = val
		keysExisted[key] = viper.IsSet(key)
	}
	// Register cleanup to restore values
	t.Cleanup(func() {
		for _, key := range keys {
			if keysExisted[key] {
				viper.Set(key, origValues[key])
			} else {
				// Key didn't exist originally, so remove it
				allSettings := viper.AllSettings()
				delete(allSettings, key)
				viper.Reset()
				for k, v := range allSettings {
					viper.Set(k, v)
				}
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

	// Capture initial timestamp to verify it gets updated
	initialLastCheck := initialStats.LastMemoryCheck

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
	assert.True(t, stats.LastMemoryCheck.After(initialLastCheck) || stats.LastMemoryCheck.Equal(initialLastCheck),
		"lastMemoryCheck should be updated or remain initialized")
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

	// Capture initial timestamp to verify it gets updated
	initialStats := bm.GetStats()
	initialLastCheck := initialStats.LastMemoryCheck

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

	// LastMemoryCheck should be updated after processing files (memoryCheckInterval=1)
	assert.True(t, stats.LastMemoryCheck.After(initialLastCheck),
		"lastMemoryCheck should be updated after memory checks")
}
