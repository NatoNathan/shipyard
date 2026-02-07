package contract

import (
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCompletionContract_Bash tests that bash completion script is generated
func TestCompletionContract_Bash(t *testing.T) {
	shipyardBin := buildShipyard(t)

	cmd := exec.Command(shipyardBin, "completion", "bash")
	output, err := cmd.CombinedOutput()

	require.NoError(t, err, "completion bash should exit 0")
	outputStr := strings.ToLower(string(output))
	assert.True(t,
		strings.Contains(outputStr, "compreply") ||
			strings.Contains(outputStr, "_shipyard") ||
			strings.Contains(outputStr, "bash"),
		"Output should contain bash completion markers (COMPREPLY, _shipyard, or bash), got: %s", string(output))
}

// TestCompletionContract_Zsh tests that zsh completion script is generated
func TestCompletionContract_Zsh(t *testing.T) {
	shipyardBin := buildShipyard(t)

	cmd := exec.Command(shipyardBin, "completion", "zsh")
	output, err := cmd.CombinedOutput()

	require.NoError(t, err, "completion zsh should exit 0")
	outputStr := string(output)
	assert.True(t,
		strings.Contains(outputStr, "compdef") ||
			strings.Contains(outputStr, "_shipyard") ||
			strings.Contains(outputStr, "zsh"),
		"Output should contain zsh completion markers (compdef, _shipyard, or zsh), got: %s", outputStr)
}

// TestCompletionContract_Fish tests that fish completion script is generated
func TestCompletionContract_Fish(t *testing.T) {
	shipyardBin := buildShipyard(t)

	cmd := exec.Command(shipyardBin, "completion", "fish")
	output, err := cmd.CombinedOutput()

	require.NoError(t, err, "completion fish should exit 0")
	outputStr := string(output)
	assert.Contains(t, outputStr, "complete", "Output should contain 'complete' for fish")
	assert.Contains(t, outputStr, "shipyard", "Output should contain 'shipyard' for fish")
}

// TestCompletionContract_Powershell tests that powershell completion script is generated
func TestCompletionContract_Powershell(t *testing.T) {
	shipyardBin := buildShipyard(t)

	cmd := exec.Command(shipyardBin, "completion", "powershell")
	output, err := cmd.CombinedOutput()

	require.NoError(t, err, "completion powershell should exit 0")
	outputStr := string(output)
	assert.True(t,
		strings.Contains(outputStr, "Register-ArgumentCompleter") ||
			strings.Contains(outputStr, "shipyard"),
		"Output should contain powershell completion markers (Register-ArgumentCompleter or shipyard), got: %s", outputStr)
}

// TestCompletionContract_InvalidArg tests that an invalid shell argument causes a non-zero exit
func TestCompletionContract_InvalidArg(t *testing.T) {
	shipyardBin := buildShipyard(t)

	cmd := exec.Command(shipyardBin, "completion", "invalid")
	_, err := cmd.CombinedOutput()

	assert.Error(t, err, "completion with invalid shell should exit non-zero")
}

// TestCompletionContract_NoArg tests that missing shell argument causes a non-zero exit
func TestCompletionContract_NoArg(t *testing.T) {
	shipyardBin := buildShipyard(t)

	cmd := exec.Command(shipyardBin, "completion")
	_, err := cmd.CombinedOutput()

	assert.Error(t, err, "completion with no argument should exit non-zero")
}

// TestCompletionContract_HelpFlag tests that --help flag shows completion help text
func TestCompletionContract_HelpFlag(t *testing.T) {
	shipyardBin := buildShipyard(t)

	cmd := exec.Command(shipyardBin, "completion", "--help")
	output, err := cmd.CombinedOutput()

	require.NoError(t, err, "completion --help should exit 0")
	outputStr := string(output)
	assert.True(t,
		strings.Contains(outputStr, "Teach your shell to speak") ||
			strings.Contains(outputStr, "completion"),
		"Output should contain help text about completion, got: %s", outputStr)
}
