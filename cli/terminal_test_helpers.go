package cli

import "testing"

// terminalEnvSetup defines environment variables for terminal detection tests.
type terminalEnvSetup struct {
	Term          string
	CI            string
	GitHubActions string
	NoColor       string
	ForceColor    string
}

// apply sets up the environment variables using t.Setenv.
func (e terminalEnvSetup) apply(t *testing.T) {
	t.Helper()

	// Always set all environment variables to ensure isolation
	// Empty string explicitly unsets the variable in the test environment
	t.Setenv("TERM", e.Term)
	t.Setenv("CI", e.CI)
	t.Setenv("GITHUB_ACTIONS", e.GitHubActions)
	t.Setenv("NO_COLOR", e.NoColor)
	t.Setenv("FORCE_COLOR", e.ForceColor)
}

// Common terminal environment setups for reuse across tests.
var (
	envDefaultTerminal = terminalEnvSetup{
		Term:       "xterm-256color",
		CI:         "",
		NoColor:    "",
		ForceColor: "",
	}

	envDumbTerminal = terminalEnvSetup{
		Term: "dumb",
	}

	envCIWithoutGitHub = terminalEnvSetup{
		Term:          "xterm",
		CI:            "true",
		GitHubActions: "",
	}

	envGitHubActions = terminalEnvSetup{
		Term:          "xterm",
		CI:            "true",
		GitHubActions: "true",
		NoColor:       "",
	}

	envNoColor = terminalEnvSetup{
		Term:       "xterm-256color",
		CI:         "",
		NoColor:    "1",
		ForceColor: "",
	}

	envForceColor = terminalEnvSetup{
		Term:       "dumb",
		ForceColor: "1",
	}

	envEmptyTerm = terminalEnvSetup{
		Term: "",
	}
)
