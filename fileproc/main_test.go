package fileproc_test

import (
	"testing"

	"github.com/ivuorinen/gibidify/testutil"
)

// TestMain keeps the developer's own gibidify config out of the walker and
// processor tests.
func TestMain(m *testing.M) {
	testutil.RunIsolated(m)
}
