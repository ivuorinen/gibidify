package fileproc

import (
	"context"
	"testing"
	"time"

	"github.com/spf13/viper"

	"github.com/ivuorinen/gibidify/testutil"
)

func TestResourceMonitor_RateLimiting(t *testing.T) {
	testutil.ResetViperConfig(t, "")

	// Enable rate limiting with a low rate for testing
	viper.Set("resourceLimits.enabled", true)
	viper.Set("resourceLimits.rateLimitFilesPerSec", 5) // 5 files per second

	rm := NewResourceMonitor()
	defer rm.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	// First few requests should succeed quickly
	start := time.Now()
	for i := 0; i < 3; i++ {
		err := rm.WaitForRateLimit(ctx)
		if err != nil {
			t.Errorf("Expected no error for rate limit wait %d, got %v", i, err)
		}
	}

	// Should have taken some time due to rate limiting
	duration := time.Since(start)
	if duration < 200*time.Millisecond {
		t.Logf("Rate limiting may not be working as expected, took only %v", duration)
	}
}
