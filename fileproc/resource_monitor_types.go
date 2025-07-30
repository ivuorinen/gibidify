package fileproc

import (
	"sync"
	"time"

	"github.com/ivuorinen/gibidify/config"
)

// ResourceMonitor monitors resource usage and enforces limits to prevent DoS attacks.
type ResourceMonitor struct {
	enabled               bool
	maxFiles              int
	maxTotalSize          int64
	fileProcessingTimeout time.Duration
	overallTimeout        time.Duration
	maxConcurrentReads    int
	rateLimitFilesPerSec  int
	hardMemoryLimitMB     int
	enableGracefulDegr    bool
	enableResourceMon     bool

	// Current state tracking
	filesProcessed       int64
	totalSizeProcessed   int64
	concurrentReads      int64
	startTime            time.Time
	lastRateLimitCheck   time.Time
	hardMemoryLimitBytes int64

	// Rate limiting
	rateLimiter   *time.Ticker
	rateLimitChan chan struct{}

	// Synchronization
	mu                     sync.RWMutex
	violationLogged        map[string]bool
	degradationActive      bool
	emergencyStopRequested bool
}

// ResourceMetrics holds comprehensive resource usage metrics.
type ResourceMetrics struct {
	FilesProcessed      int64         `json:"files_processed"`
	TotalSizeProcessed  int64         `json:"total_size_processed"`
	ConcurrentReads     int64         `json:"concurrent_reads"`
	ProcessingDuration  time.Duration `json:"processing_duration"`
	AverageFileSize     float64       `json:"average_file_size"`
	ProcessingRate      float64       `json:"processing_rate_files_per_sec"`
	MemoryUsageMB       int64         `json:"memory_usage_mb"`
	MaxMemoryUsageMB    int64         `json:"max_memory_usage_mb"`
	ViolationsDetected  []string      `json:"violations_detected"`
	DegradationActive   bool          `json:"degradation_active"`
	EmergencyStopActive bool          `json:"emergency_stop_active"`
	LastUpdated         time.Time     `json:"last_updated"`
}

// ResourceViolation represents a detected resource limit violation.
type ResourceViolation struct {
	Type      string                 `json:"type"`
	Message   string                 `json:"message"`
	Current   interface{}            `json:"current"`
	Limit     interface{}            `json:"limit"`
	Timestamp time.Time              `json:"timestamp"`
	Context   map[string]interface{} `json:"context"`
}

// NewResourceMonitor creates a new resource monitor with configuration.
func NewResourceMonitor() *ResourceMonitor {
	rm := &ResourceMonitor{
		enabled:               config.GetResourceLimitsEnabled(),
		maxFiles:              config.GetMaxFiles(),
		maxTotalSize:          config.GetMaxTotalSize(),
		fileProcessingTimeout: time.Duration(config.GetFileProcessingTimeoutSec()) * time.Second,
		overallTimeout:        time.Duration(config.GetOverallTimeoutSec()) * time.Second,
		maxConcurrentReads:    config.GetMaxConcurrentReads(),
		rateLimitFilesPerSec:  config.GetRateLimitFilesPerSec(),
		hardMemoryLimitMB:     config.GetHardMemoryLimitMB(),
		enableGracefulDegr:    config.GetEnableGracefulDegradation(),
		enableResourceMon:     config.GetEnableResourceMonitoring(),
		startTime:             time.Now(),
		lastRateLimitCheck:    time.Now(),
		violationLogged:       make(map[string]bool),
		hardMemoryLimitBytes:  int64(config.GetHardMemoryLimitMB()) * 1024 * 1024,
	}

	// Initialize rate limiter if rate limiting is enabled
	if rm.enabled && rm.rateLimitFilesPerSec > 0 {
		interval := time.Second / time.Duration(rm.rateLimitFilesPerSec)
		rm.rateLimiter = time.NewTicker(interval)
		rm.rateLimitChan = make(chan struct{}, rm.rateLimitFilesPerSec)

		// Pre-fill the rate limit channel
		for i := 0; i < rm.rateLimitFilesPerSec; i++ {
			select {
			case rm.rateLimitChan <- struct{}{}:
			default:
				goto rateLimitFull
			}
		}
	rateLimitFull:

		// Start rate limiter refill goroutine  
		go rm.rateLimiterRefill()
	}

	return rm
}