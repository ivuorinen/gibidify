// Package fileproc handles file processing, collection, and output formatting.
package fileproc

import (
	"context"
	"fmt"
	"time"

	"github.com/ivuorinen/gibidify/shared"
)

// WaitForRateLimit waits for rate limiting if enabled.
func (rm *ResourceMonitor) WaitForRateLimit(ctx context.Context) error {
	if !rm.enabled || rm.rateLimitFilesPerSec <= 0 {
		return nil
	}

	select {
	case <-ctx.Done():
		return fmt.Errorf("context canceled while waiting for rate limit: %w", ctx.Err())
	case <-rm.rateLimitChan:
		return nil
	case <-time.After(time.Second): // Fallback timeout
		logger := shared.GetLogger()
		logger.Warn("Rate limiting timeout exceeded, continuing without rate limit")

		return nil
	}
}

// rateLimiterRefill refills the rate limiting channel periodically.
func (rm *ResourceMonitor) rateLimiterRefill() {
	for {
		select {
		case <-rm.done:
			return
		case <-rm.rateLimiter.C:
			select {
			case rm.rateLimitChan <- struct{}{}:
			default:
				// Channel is full, skip
			}
		}
	}
}
