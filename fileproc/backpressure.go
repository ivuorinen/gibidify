// Package fileproc provides back-pressure management for memory optimization.
package fileproc

import (
	"context"
	"math"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/ivuorinen/gibidify/config"
	"github.com/ivuorinen/gibidify/gibidiutils"
)

// BackpressureManager manages memory usage and applies back-pressure when needed.
type BackpressureManager struct {
	enabled             bool
	maxMemoryUsage      int64
	memoryCheckInterval int
	maxPendingFiles     int
	maxPendingWrites    int
	filesProcessed      int64
	mu                  sync.RWMutex
	memoryWarningLogged bool
	lastMemoryCheck     time.Time
}

// NewBackpressureManager creates a new back-pressure manager with configuration.
func NewBackpressureManager() *BackpressureManager {
	return &BackpressureManager{
		enabled:             config.GetBackpressureEnabled(),
		maxMemoryUsage:      config.GetMaxMemoryUsage(),
		memoryCheckInterval: config.GetMemoryCheckInterval(),
		maxPendingFiles:     config.GetMaxPendingFiles(),
		maxPendingWrites:    config.GetMaxPendingWrites(),
		lastMemoryCheck:     time.Now(),
	}
}

// CreateChannels creates properly sized channels based on back-pressure configuration.
func (bp *BackpressureManager) CreateChannels() (chan string, chan WriteRequest) {
	var fileCh chan string
	var writeCh chan WriteRequest

	if bp.enabled {
		// Use buffered channels with configured limits
		fileCh = make(chan string, bp.maxPendingFiles)
		writeCh = make(chan WriteRequest, bp.maxPendingWrites)
		logrus.Debugf("Created buffered channels: files=%d, writes=%d", bp.maxPendingFiles, bp.maxPendingWrites)
	} else {
		// Use unbuffered channels (default behavior)
		fileCh = make(chan string)
		writeCh = make(chan WriteRequest)
		logrus.Debug("Created unbuffered channels (back-pressure disabled)")
	}

	return fileCh, writeCh
}

// ShouldApplyBackpressure checks if back-pressure should be applied.
func (bp *BackpressureManager) ShouldApplyBackpressure(_ context.Context) bool {
	if !bp.enabled {
		return false
	}

	// Check if we should evaluate memory usage
	filesProcessed := atomic.AddInt64(&bp.filesProcessed, 1)
	// Avoid divide by zero - if interval is 0, check every file
	if bp.memoryCheckInterval > 0 && int(filesProcessed)%bp.memoryCheckInterval != 0 {
		return false
	}

	// Get current memory usage
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	currentMemory := gibidiutils.SafeUint64ToInt64WithDefault(m.Alloc, math.MaxInt64)

	bp.mu.Lock()
	defer bp.mu.Unlock()

	bp.lastMemoryCheck = time.Now()

	// Check if we're over the memory limit
	if currentMemory > bp.maxMemoryUsage {
		if !bp.memoryWarningLogged {
			logrus.Warnf("Memory usage (%d bytes) exceeds limit (%d bytes), applying back-pressure",
				currentMemory, bp.maxMemoryUsage)
			bp.memoryWarningLogged = true
		}
		return true
	}

	// Reset warning flag if we're back under the limit
	if bp.memoryWarningLogged && currentMemory < bp.maxMemoryUsage*8/10 { // 80% of limit
		logrus.Infof("Memory usage normalized (%d bytes), removing back-pressure", currentMemory)
		bp.memoryWarningLogged = false
	}

	return false
}

// ApplyBackpressure applies back-pressure by triggering garbage collection and adding delay.
func (bp *BackpressureManager) ApplyBackpressure(ctx context.Context) {
	if !bp.enabled {
		return
	}

	// Force garbage collection to free up memory
	runtime.GC()

	// Add a small delay to allow memory to be freed
	select {
	case <-ctx.Done():
		return
	case <-time.After(10 * time.Millisecond):
		// Small delay to allow GC to complete
	}

	// Log memory usage after GC
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	logrus.Debugf("Applied back-pressure: memory after GC = %d bytes", m.Alloc)
}

// GetStats returns current back-pressure statistics.
func (bp *BackpressureManager) GetStats() BackpressureStats {
	bp.mu.RLock()
	defer bp.mu.RUnlock()

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return BackpressureStats{
		Enabled:             bp.enabled,
		FilesProcessed:      atomic.LoadInt64(&bp.filesProcessed),
		CurrentMemoryUsage:  gibidiutils.SafeUint64ToInt64WithDefault(m.Alloc, math.MaxInt64),
		MaxMemoryUsage:      bp.maxMemoryUsage,
		MemoryWarningActive: bp.memoryWarningLogged,
		LastMemoryCheck:     bp.lastMemoryCheck,
		MaxPendingFiles:     bp.maxPendingFiles,
		MaxPendingWrites:    bp.maxPendingWrites,
	}
}

// BackpressureStats represents back-pressure manager statistics.
type BackpressureStats struct {
	Enabled             bool      `json:"enabled"`
	FilesProcessed      int64     `json:"files_processed"`
	CurrentMemoryUsage  int64     `json:"current_memory_usage"`
	MaxMemoryUsage      int64     `json:"max_memory_usage"`
	MemoryWarningActive bool      `json:"memory_warning_active"`
	LastMemoryCheck     time.Time `json:"last_memory_check"`
	MaxPendingFiles     int       `json:"max_pending_files"`
	MaxPendingWrites    int       `json:"max_pending_writes"`
}

// WaitForChannelSpace waits for space in channels if they're getting full.
func (bp *BackpressureManager) WaitForChannelSpace(ctx context.Context, fileCh chan string, writeCh chan WriteRequest) {
	if !bp.enabled {
		return
	}

	// Check if file channel is getting full (>=90% capacity)
	if bp.maxPendingFiles > 0 && len(fileCh) >= bp.maxPendingFiles*9/10 {
		logrus.Debugf("File channel is %d%% full, waiting for space", len(fileCh)*100/bp.maxPendingFiles)

		// Wait a bit for the channel to drain
		select {
		case <-ctx.Done():
			return
		case <-time.After(5 * time.Millisecond):
		}
	}

	// Check if write channel is getting full (>=90% capacity)
	if bp.maxPendingWrites > 0 && len(writeCh) >= bp.maxPendingWrites*9/10 {
		logrus.Debugf("Write channel is %d%% full, waiting for space", len(writeCh)*100/bp.maxPendingWrites)

		// Wait a bit for the channel to drain
		select {
		case <-ctx.Done():
			return
		case <-time.After(5 * time.Millisecond):
		}
	}
}

// LogBackpressureInfo logs back-pressure configuration and status.
func (bp *BackpressureManager) LogBackpressureInfo() {
	if bp.enabled {
		logrus.Infof("Back-pressure enabled: maxMemory=%dMB, fileBuffer=%d, writeBuffer=%d, checkInterval=%d",
			bp.maxMemoryUsage/1024/1024, bp.maxPendingFiles, bp.maxPendingWrites, bp.memoryCheckInterval)
	} else {
		logrus.Info("Back-pressure disabled")
	}
}
