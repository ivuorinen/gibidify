package shared

import (
	"context"
	"strings"
	"testing"
	"time"
)

// TestCheckContextCancellation tests the CheckContextCancellation function.
func TestCheckContextCancellation(t *testing.T) {
	tests := []struct {
		name          string
		setupContext  func() context.Context
		operation     string
		expectError   bool
		errorContains string
	}{
		{
			name:         "active context",
			setupContext: context.Background,
			operation:    "test operation",
			expectError:  false,
		},
		{
			name: "canceled context",
			setupContext: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel() // Cancel immediately
				return ctx
			},
			operation:     "test operation",
			expectError:   true,
			errorContains: "test operation canceled",
		},
		{
			name: "timeout context",
			setupContext: func() context.Context {
				ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
				defer cancel()
				// Wait for timeout
				time.Sleep(1 * time.Millisecond)
				return ctx
			},
			operation:     "timeout operation",
			expectError:   true,
			errorContains: "timeout operation canceled",
		},
		{
			name: "context with deadline exceeded",
			setupContext: func() context.Context {
				ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-1*time.Hour))
				defer cancel()
				return ctx
			},
			operation:     "deadline operation",
			expectError:   true,
			errorContains: "deadline operation canceled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.setupContext()
			err := CheckContextCancellation(ctx, tt.operation)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for %s, but got none", tt.name)
				} else if !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("Error %q should contain %q", err.Error(), tt.errorContains)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for %s: %v", tt.name, err)
				}
			}
		})
	}
}

// assertNoError is a helper that fails the test if err is not nil.
func assertNoError(t *testing.T, err error, msg string) {
	t.Helper()
	if err != nil {
		t.Errorf("%s: %v", msg, err)
	}
}

// assertError is a helper that fails the test if err is nil.
func assertError(t *testing.T, err error, msg string) {
	t.Helper()
	if err == nil {
		t.Error(msg)
	}
}

// TestContextCancellationIntegration tests integration scenarios.
func TestContextCancellationIntegration(t *testing.T) {
	t.Run("chained context checks", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())

		// First check should pass
		err := CheckContextCancellation(ctx, "first check")
		assertNoError(t, err, "First check should pass")

		// Cancel context
		cancel()

		// Second check should fail
		err = CheckContextCancellation(ctx, "second check")
		assertError(t, err, "Second check should fail after cancellation")
	})

	t.Run("context cancellation propagation", func(t *testing.T) {
		// Test that context cancellation properly propagates through nested calls
		parentCtx, parentCancel := context.WithCancel(context.Background())
		childCtx, childCancel := context.WithCancel(parentCtx)

		defer parentCancel()
		defer childCancel()

		// Both contexts should be active initially
		err := CheckContextCancellation(parentCtx, "parent")
		assertNoError(t, err, "Parent context should be active")

		err = CheckContextCancellation(childCtx, "child")
		assertNoError(t, err, "Child context should be active")

		// Cancel parent - child should also be canceled
		parentCancel()

		err = CheckContextCancellation(childCtx, "child after parent cancel")
		assertError(t, err, "Child context should be canceled when parent is canceled")
	})
}
