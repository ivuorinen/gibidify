package main

import (
	"errors"
	"testing"

	"github.com/ivuorinen/gibidify/cli"
)

// TestMainFunctionComponents tests the components exitCode() builds to report
// errors. The exit statuses themselves are covered by TestExitCode.
func TestMainFunctionComponents(t *testing.T) {
	// Test UI manager creation (used in exitCode())
	ui := cli.NewUIManager()
	if ui == nil {
		t.Error("Expected NewUIManager to return non-nil UIManager")
	}

	// Test error formatter creation (used in exitCode())
	errorFormatter := cli.NewErrorFormatter(ui)
	if errorFormatter == nil {
		t.Error("Expected NewErrorFormatter to return non-nil ErrorFormatter")
	}
}

// TestUserErrorClassification tests the error classification used in exitCode().
func TestUserErrorClassification(t *testing.T) {
	// Test the cli.IsUserError function that exitCode() uses for error classification

	// Create a user error (MissingSourceError is a user error)
	userErr := &cli.MissingSourceError{}
	if !cli.IsUserError(userErr) {
		t.Error("Expected cli.IsUserError to return true for MissingSourceError")
	}

	// Test with a system error (generic error)
	systemErr := errors.New("test system error")
	if cli.IsUserError(systemErr) {
		t.Error("Expected cli.IsUserError to return false for generic error")
	}

	// Test with nil error
	if cli.IsUserError(nil) {
		t.Error("Expected cli.IsUserError to return false for nil error")
	}
}
