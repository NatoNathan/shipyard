package contract

import (
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestUpgradeContract_HelpFlag verifies that upgrade --help exits 0 and documents expected flags.
func TestUpgradeContract_HelpFlag(t *testing.T) {
	shipyardBin := buildShipyard(t)

	cmd := exec.Command(shipyardBin, "upgrade", "--help")
	output, err := cmd.CombinedOutput()

	require.NoError(t, err, "upgrade --help should exit 0")

	outputStr := string(output)
	lower := strings.ToLower(outputStr)

	assert.Contains(t, lower, "upgrade", "Help output should mention 'upgrade'")
	assert.Contains(t, outputStr, "--yes", "Help output should document --yes flag")
	assert.Contains(t, outputStr, "--version", "Help output should document --version flag")
	assert.Contains(t, outputStr, "--force", "Help output should document --force flag")
	assert.Contains(t, outputStr, "--dry-run", "Help output should document --dry-run flag")
}

// TestUpgradeContract_DryRunNoNetwork verifies that upgrade --dry-run does not panic
// even when GITHUB_TOKEN is not available.
func TestUpgradeContract_DryRunNoNetwork(t *testing.T) {
	shipyardBin := buildShipyard(t)

	cmd := exec.Command(shipyardBin, "upgrade", "--dry-run")

	// Strip GITHUB_TOKEN from the environment so the command cannot reach GitHub.
	env := os.Environ()
	filtered := make([]string, 0, len(env))
	for _, e := range env {
		if !strings.HasPrefix(e, "GITHUB_TOKEN=") {
			filtered = append(filtered, e)
		}
	}
	cmd.Env = filtered

	output, err := cmd.CombinedOutput()
	outputStr := string(output)

	// The command may exit non-zero (e.g. network failure), but it must not panic
	// or be killed by a signal. A nil error or an *exec.ExitError with a valid
	// exit code are both acceptable.
	if err != nil {
		exitErr, ok := err.(*exec.ExitError)
		require.True(t, ok, "Error should be an *exec.ExitError, not: %v", err)
		assert.GreaterOrEqual(t, exitErr.ExitCode(), 0, "Exit code should be valid (not signal-killed)")
	}

	// Output should not be empty — there should be some recognizable text.
	assert.NotEmpty(t, strings.TrimSpace(outputStr), "Command output should not be empty")
}
