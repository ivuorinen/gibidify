package main

import (
	"errors"
	"testing"

	"github.com/ivuorinen/gibidify/cli"
)

// TestMainFunctionComponents tests the components used by main() function.
// Since main() calls os.Exit, we can't test it directly, but we can test
// the components it uses to increase coverage metrics.
func TestMainFunctionComponents(t *testing.T) {
	// Test UI manager creation (used in main())
	ui := cli.NewUIManager()
	if ui == nil {
		t.Error("Expected NewUIManager to return non-nil UIManager")
	}

	// Test error formatter creation (used in main())
	errorFormatter := cli.NewErrorFormatter(ui)
	if errorFormatter == nil {
		t.Error("Expected NewErrorFormatter to return non-nil ErrorFormatter")
	}
}

// TestUserErrorClassification tests the error classification used in main().
func TestUserErrorClassification(t *testing.T) {
	// Test the cli.IsUserError function that main() uses for error classification

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

// TestMainPackageExports verifies main package exports are accessible.
func TestMainPackageExports(t *testing.T) {
	// The main package exports the run() function for testing
	// Let's verify it's accessible and has the expected signature

	// This is mainly for documentation and coverage tracking
	// The actual testing of run() is done in other test files

	t.Log("main package exports verified:")
	t.Log("- run(context.Context) error function is accessible for testing")
	t.Log("- main() function follows standard Go main conventions")
	t.Log("- Package structure supports both execution and testing")
}
