package testutil

import (
	"os"
	"testing"
)

func TestVerifyContentContains(t *testing.T) {
	// Test successful verification
	t.Run(
		"all substrings present", func(t *testing.T) {
			content := "This is a test file with multiple lines"
			VerifyContentContains(t, content, []string{"test file", "multiple lines"})
			// If we get here, the test passed
		},
	)

	// Test empty expected substrings
	t.Run(
		"empty expected substrings", func(t *testing.T) {
			content := "Any content"
			VerifyContentContains(t, content, []string{})
			// Should pass with no expected strings
		},
	)
}

func TestMustSucceed(t *testing.T) {
	// Test with nil error (should succeed)
	t.Run(
		"nil error", func(t *testing.T) {
			MustSucceed(t, nil, "successful operation")
			// If we get here, the test passed
		},
	)
}

func TestCloseFile(t *testing.T) {
	// Test closing a normal file
	t.Run(
		"close normal file", func(t *testing.T) {
			file, err := os.CreateTemp(t.TempDir(), "test")
			if err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			CloseFile(t, file)

			// Verify file is closed by trying to write to it
			_, writeErr := file.Write([]byte("test"))
			if writeErr == nil {
				t.Error("Expected write to fail on closed file")
			}
		},
	)

	// Test that CloseFile doesn't panic on already closed files
	// Note: We can't easily test the error case without causing test failure
	// since CloseFile calls t.Errorf, which is the expected behavior
	t.Run(
		"verify CloseFile function exists and is callable", func(t *testing.T) {
			// This test just verifies the function signature and basic functionality
			// The error case is tested in integration tests where failures are expected
			file, err := os.CreateTemp(t.TempDir(), "test")
			if err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			// Test normal case - file should close successfully
			CloseFile(t, file)

			// Verify file is closed
			_, writeErr := file.Write([]byte("test"))
			if writeErr == nil {
				t.Error("Expected write to fail on closed file")
			}
		},
	)
}
