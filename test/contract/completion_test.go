package contract

import (
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
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

func TestDocumentationLocalLinks(t *testing.T) {
	repoRoot, err := filepath.Abs(filepath.Join("..", ".."))
	require.NoError(t, err)

	linkPattern := regexp.MustCompile(`\[[^\]]+\]\(([^)]+)\)`)
	roots := []string{"README.md", "CONTRIBUTING.md", "docs", filepath.Join("skills", "shipyard-cli")}

	for _, root := range roots {
		rootPath := filepath.Join(repoRoot, root)
		err := filepath.WalkDir(rootPath, func(path string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if entry.IsDir() || filepath.Ext(path) != ".md" {
				return nil
			}

			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			for _, match := range linkPattern.FindAllStringSubmatch(string(content), -1) {
				target := strings.Trim(match[1], "<>")
				if strings.HasPrefix(target, "http://") || strings.HasPrefix(target, "https://") || strings.HasPrefix(target, "mailto:") || strings.HasPrefix(target, "#") {
					continue
				}
				target = strings.SplitN(target, "#", 2)[0]
				if target == "" {
					continue
				}
				resolved := filepath.Clean(filepath.Join(filepath.Dir(path), filepath.FromSlash(target)))
				if _, err := os.Stat(resolved); err != nil {
					relPath, _ := filepath.Rel(repoRoot, path)
					t.Errorf("broken local link in %s: %s", relPath, match[1])
				}
			}
			return nil
		})
		require.NoError(t, err)
	}
}

func TestDocumentedCommandInventoryMatchesCLI(t *testing.T) {
	skillPath := filepath.Join("..", "..", "skills", "shipyard-cli", "SKILL.md")
	content, err := os.ReadFile(skillPath)
	require.NoError(t, err)

	rowPattern := regexp.MustCompile(`(?m)^\| ` + "`" + `([^` + "`" + `]+)` + "`" + ` \|`)
	var documented []string
	for _, match := range rowPattern.FindAllStringSubmatch(string(content), -1) {
		if !strings.HasPrefix(match[1], "--") {
			documented = append(documented, match[1])
		}
	}

	shipyardBin := buildShipyard(t)
	actual := helpCommandNames(t, shipyardBin)
	for _, parent := range []string{"version", "config"} {
		for _, child := range helpCommandNames(t, shipyardBin, parent) {
			actual = append(actual, parent+" "+child)
		}
	}

	assert.ElementsMatch(t, actual, documented)
}

func helpCommandNames(t *testing.T, shipyardBin string, args ...string) []string {
	t.Helper()
	cmdArgs := append(append([]string{}, args...), "--help")
	output, err := exec.Command(shipyardBin, cmdArgs...).CombinedOutput()
	require.NoError(t, err, "help command failed: %s", output)

	commandPattern := regexp.MustCompile(`(?m)^  ([a-z][a-z-]*)\s{2,}`)
	var commands []string
	for _, match := range commandPattern.FindAllStringSubmatch(string(output), -1) {
		if match[1] != "help" && !strings.Contains(match[1], "shipyard") {
			commands = append(commands, match[1])
		}
	}
	return commands
}
