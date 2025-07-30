package fileproc

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
)

// WaitForRateLimit waits for rate limiting if enabled.
func (rm *ResourceMonitor) WaitForRateLimit(ctx context.Context) error {
	if !rm.enabled || rm.rateLimitFilesPerSec <= 0 {
		return nil
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-rm.rateLimitChan:
		return nil
	case <-time.After(time.Second): // Fallback timeout
		logrus.Warn("Rate limiting timeout exceeded, continuing without rate limit")
		return nil
	}
}

// rateLimiterRefill refills the rate limiting channel periodically.
func (rm *ResourceMonitor) rateLimiterRefill() {
	for range rm.rateLimiter.C {
		select {
		case rm.rateLimitChan <- struct{}{}:
		default:
			// Channel is full, skip
		}
	}
}