package config_test

import (
	"testing"

	"github.com/ivuorinen/gibidify/testutil"
)

// TestMain keeps the developer's own gibidify config out of the loader tests.
// It lives in the external test package: the internal one cannot import
// testutil, which imports config.
func TestMain(m *testing.M) {
	testutil.RunIsolated(m)
}
