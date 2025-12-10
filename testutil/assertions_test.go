package testutil

import (
	"errors"
	"strings"
	"testing"

	"github.com/ivuorinen/gibidify/shared"
)

// TestAssertError tests the AssertError function.
func TestAssertError(t *testing.T) {
	t.Run(
		shared.TestCaseSuccessCases, func(t *testing.T) {
			// Test expect error and get error
			t.Run(
				"expect error and get error", func(t *testing.T) {
					AssertError(t, errors.New(shared.TestErrTestErrorMsg), true, shared.TestCaseTestOperation)
					// If we get here, the assertion passed
				},
			)

			// Test expect no error and get no error
			t.Run(
				"expect no error and get no error", func(t *testing.T) {
					AssertError(t, nil, false, "successful operation")
					// If we get here, the assertion passed
				},
			)

			// Test with empty operation name
			t.Run(
				shared.TestCaseEmptyOperationName, func(t *testing.T) {
					AssertError(t, nil, false, "")
					// Should still work with empty operation
				},
			)

			// Test with complex operation name
			t.Run(
				"complex operation name", func(t *testing.T) {
					AssertError(t, nil, false, "complex operation with special chars: !@#$%^&*()")
					// Should handle special characters
				},
			)
		},
	)

	// Test edge cases
	t.Run(
		"edge cases", func(t *testing.T) {
			// Test various error types
			t.Run(
				shared.TestCaseDifferentErrorTypes, func(t *testing.T) {
					AssertError(t, shared.ErrTestError, true, "using shared.ErrTestError")
					AssertError(t, errors.New("wrapped: original"), true, "wrapped error")
				},
			)
		},
	)
}

// TestAssertNoError tests the AssertNoError function.
func TestAssertNoError(t *testing.T) {
	t.Run(
		shared.TestCaseSuccessCases, func(t *testing.T) {
			// Test with nil error
			t.Run(
				"nil error", func(t *testing.T) {
					AssertNoError(t, nil, "successful operation")
				},
			)

			// Test with empty operation name
			t.Run(
				shared.TestCaseEmptyOperationName, func(t *testing.T) {
					AssertNoError(t, nil, "")
				},
			)

			// Test with complex operation name
			t.Run(
				"complex operation", func(t *testing.T) {
					AssertNoError(t, nil, "complex operation with special chars: !@#$%^&*()")
				},
			)
		},
	)

	// We can't easily test the failure case in a unit test since it would cause test failure
	// But we can verify the function signature and basic functionality
	t.Run(
		shared.TestCaseFunctionAvailability, func(t *testing.T) {
			// Verify the function doesn't panic with valid inputs
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("AssertNoError should not panic: %v", r)
				}
			}()

			// Call with success case to ensure function works
			AssertNoError(t, nil, shared.TestCaseTestOperation)
		},
	)
}

// TestAssertExpectedError tests the AssertExpectedError function.
func TestAssertExpectedError(t *testing.T) {
	t.Run(
		shared.TestCaseSuccessCases, func(t *testing.T) {
			// Test with error present
			t.Run(
				"error present as expected", func(t *testing.T) {
					AssertExpectedError(t, errors.New("expected error"), "operation that should fail")
				},
			)

			// Test with different error types
			t.Run(
				shared.TestCaseDifferentErrorTypes, func(t *testing.T) {
					AssertExpectedError(t, shared.ErrTestError, "test error operation")
					AssertExpectedError(t, errors.New("complex error with details"), "complex operation")
				},
			)

			// Test with empty operation name
			t.Run(
				shared.TestCaseEmptyOperationName, func(t *testing.T) {
					AssertExpectedError(t, errors.New("error"), "")
				},
			)
		},
	)

	// Verify function availability and basic properties
	t.Run(
		shared.TestCaseFunctionAvailability, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("AssertExpectedError should not panic: %v", r)
				}
			}()

			// Call with success case
			AssertExpectedError(t, errors.New("test"), shared.TestCaseTestOperation)
		},
	)
}

// TestAssertErrorContains tests the AssertErrorContains function.
func TestAssertErrorContains(t *testing.T) {
	t.Run(
		shared.TestCaseSuccessCases, func(t *testing.T) {
			// Test error contains substring
			t.Run(
				"error contains substring", func(t *testing.T) {
					err := errors.New("database connection failed")
					AssertErrorContains(t, err, "connection", "database operation")
				},
			)

			// Test exact match
			t.Run(
				"exact match", func(t *testing.T) {
					err := errors.New("exact error message")
					AssertErrorContains(t, err, "exact error message", "exact operation")
				},
			)

			// Test empty substring (should match any error)
			t.Run(
				"empty substring matches any error", func(t *testing.T) {
					err := errors.New("any error")
					AssertErrorContains(t, err, "", "any operation")
				},
			)

			// Test special characters
			t.Run(
				"special characters in substring", func(t *testing.T) {
					err := errors.New("error: failed with code 500")
					AssertErrorContains(t, err, "code 500", "special chars operation")
				},
			)

			// Test case sensitivity
			t.Run(
				"case sensitive operations", func(t *testing.T) {
					err := errors.New("error Message")
					AssertErrorContains(t, err, "error Message", "case operation")
				},
			)
		},
	)

	// Test with various error types
	t.Run(
		shared.TestCaseDifferentErrorTypes, func(t *testing.T) {
			t.Run(
				"standard error", func(t *testing.T) {
					AssertErrorContains(
						t, shared.ErrTestError, shared.TestErrTestErrorMsg, shared.TestCaseTestOperation,
					)
				},
			)

			t.Run(
				"wrapped error", func(t *testing.T) {
					wrappedErr := errors.New("wrapped: original error")
					AssertErrorContains(t, wrappedErr, "original", "wrapped operation")
				},
			)
		},
	)

	// Verify function availability
	t.Run(
		shared.TestCaseFunctionAvailability, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("AssertErrorContains should not panic: %v", r)
				}
			}()

			// Call with success case
			err := errors.New(shared.TestErrTestErrorMsg)
			AssertErrorContains(t, err, "test", "availability check")
		},
	)
}

// TestAssertionHelpers tests edge cases and combinations of assertion helpers.
func TestAssertionHelpers(t *testing.T) {
	t.Run(
		"error types compatibility", func(t *testing.T) {
			// Test that all assertion functions work with shared.ErrTestError
			AssertError(t, shared.ErrTestError, true, "shared.ErrTestError compatibility")
			AssertExpectedError(t, shared.ErrTestError, "shared.ErrTestError expected")
			AssertErrorContains(t, shared.ErrTestError, "test", "shared.ErrTestError contains")
		},
	)

	t.Run(
		"operation name handling", func(t *testing.T) {
			operations := []string{
				"",
				"simple operation",
				"operation with spaces",
				"operation-with-dashes",
				"operation_with_underscores",
				"operation.with.dots",
				"operation/with/slashes",
				"operation\\with\\backslashes",
				"operation with special chars: !@#$%^&*()",
				"operation with unicode: αβγ",
				strings.Repeat("very long operation name ", 10),
			}

			for i, op := range operations {
				t.Run(
					"operation_"+string(rune(i+'a')), func(t *testing.T) {
						// Test each assertion function with this operation name
						AssertError(t, nil, false, op)
						AssertNoError(t, nil, op)
						AssertExpectedError(t, errors.New("test"), op)
						AssertErrorContains(t, errors.New(shared.TestErrTestErrorMsg), "test", op)
					},
				)
			}
		},
	)

	t.Run(
		"error message variations", func(t *testing.T) {
			errorMessages := []string{
				"",
				"simple error",
				"error with spaces",
				"error\nwith\nnewlines",
				"error\twith\ttabs",
				"error with unicode: αβγδε",
				"error: with: colons: everywhere",
				strings.Repeat("very long error message ", 20),
				"error with special chars: !@#$%^&*()",
			}

			for i, msg := range errorMessages {
				t.Run(
					"error_message_"+string(rune(i+'a')), func(t *testing.T) {
						err := errors.New(msg)
						AssertError(t, err, true, shared.TestCaseMessageTest)
						AssertExpectedError(t, err, shared.TestCaseMessageTest)
						if msg != "" {
							// Only test contains if message is not empty
							AssertErrorContains(t, err, msg, shared.TestCaseMessageTest)
						}
					},
				)
			}
		},
	)
}

// BenchmarkStringOperations benchmarks string operations used by assertion functions.
func BenchmarkStringOperations(b *testing.B) {
	testErr := errors.New("this is a long error message with many words for testing performance of substring matching")
	errorMessage := testErr.Error()
	substring := "error message"

	b.Run(
		"contains_operation", func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = strings.Contains(errorMessage, substring)
			}
		},
	)

	b.Run(
		"error_to_string", func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = testErr.Error()
			}
		},
	)
}

// BenchmarkAssertionLogic benchmarks the core logic of assertion functions.
func BenchmarkAssertionLogic(b *testing.B) {
	testErr := errors.New("benchmark error")

	b.Run(
		"error_nil_check", func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = testErr != nil
			}
		},
	)

	b.Run(
		"boolean_comparison", func(b *testing.B) {
			hasErr := testErr != nil

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = !hasErr
			}
		},
	)
}
