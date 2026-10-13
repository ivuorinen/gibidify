// Package shared provides common utility functions.
package shared

import (
	"context"
	"fmt"
)

// CheckContextCancellation is a helper function that checks if context is canceled and returns appropriate error.
func CheckContextCancellation(ctx context.Context, operation string) error {
	select {
	case <-ctx.Done():
		return fmt.Errorf("%s canceled: %w", operation, ctx.Err())
	default:
		return nil
	}
}
