package config

const (
	// DefaultFileSizeLimit is the default maximum file size (5MB).
	DefaultFileSizeLimit = 5242880
	// MinFileSizeLimit is the minimum allowed file size limit (1KB).
	MinFileSizeLimit = 1024
	// MaxFileSizeLimit is the maximum allowed file size limit (100MB).
	MaxFileSizeLimit = 104857600

	// Resource Limit Constants

	// DefaultMaxFiles is the default maximum number of files to process.
	DefaultMaxFiles = 10000
	// MinMaxFiles is the minimum allowed file count limit.
	MinMaxFiles = 1
	// MaxMaxFiles is the maximum allowed file count limit.
	MaxMaxFiles = 1000000

	// DefaultMaxTotalSize is the default maximum total size of files (1GB).
	DefaultMaxTotalSize = 1073741824
	// MinMaxTotalSize is the minimum allowed total size limit (1MB).
	MinMaxTotalSize = 1048576
	// MaxMaxTotalSize is the maximum allowed total size limit (100GB).
	MaxMaxTotalSize = 107374182400

	// DefaultFileProcessingTimeoutSec is the default timeout for individual file processing (30 seconds).
	DefaultFileProcessingTimeoutSec = 30
	// MinFileProcessingTimeoutSec is the minimum allowed file processing timeout (1 second).
	MinFileProcessingTimeoutSec = 1
	// MaxFileProcessingTimeoutSec is the maximum allowed file processing timeout (300 seconds).
	MaxFileProcessingTimeoutSec = 300

	// DefaultOverallTimeoutSec is the default timeout for overall processing (3600 seconds = 1 hour).
	DefaultOverallTimeoutSec = 3600
	// MinOverallTimeoutSec is the minimum allowed overall timeout (10 seconds).
	MinOverallTimeoutSec = 10
	// MaxOverallTimeoutSec is the maximum allowed overall timeout (86400 seconds = 24 hours).
	MaxOverallTimeoutSec = 86400

	// DefaultMaxConcurrentReads is the default maximum concurrent file reading operations.
	DefaultMaxConcurrentReads = 10
	// MinMaxConcurrentReads is the minimum allowed concurrent reads.
	MinMaxConcurrentReads = 1
	// MaxMaxConcurrentReads is the maximum allowed concurrent reads.
	MaxMaxConcurrentReads = 100

	// DefaultRateLimitFilesPerSec is the default rate limit for file processing (0 = disabled).
	DefaultRateLimitFilesPerSec = 0
	// MinRateLimitFilesPerSec is the minimum rate limit.
	MinRateLimitFilesPerSec = 0
	// MaxRateLimitFilesPerSec is the maximum rate limit.
	MaxRateLimitFilesPerSec = 10000

	// DefaultHardMemoryLimitMB is the default hard memory limit (512MB).
	DefaultHardMemoryLimitMB = 512
	// MinHardMemoryLimitMB is the minimum hard memory limit (64MB).
	MinHardMemoryLimitMB = 64
	// MaxHardMemoryLimitMB is the maximum hard memory limit (8192MB = 8GB).
	MaxHardMemoryLimitMB = 8192
)