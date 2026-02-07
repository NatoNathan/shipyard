package contract

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/stretchr/testify/require"
)

// initGitRepo initializes a git repository for testing
func initGitRepo(t *testing.T, dir string) {
	t.Helper()
	_, err := gogit.PlainInit(dir, false)
	require.NoError(t, err, "Failed to initialize git repository")
}

// buildShipyard builds the shipyard binary and returns its path
func buildShipyard(t *testing.T) string {
	t.Helper()

	// Build the binary in a temp directory
	tempBin := filepath.Join(t.TempDir(), "shipyard")
	cmd := exec.Command("go", "build", "-o", tempBin, "../../cmd/shipyard")
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "Failed to build shipyard: %s", string(output))

	return tempBin
}

// initializeTestRepo creates a git repo, sets up a single "core" package, and runs shipyard init
func initializeTestRepo(t *testing.T, shipyardBin, dir string) {
	t.Helper()

	initGitRepo(t, dir)

	// Create a "core" package directory
	coreDir := filepath.Join(dir, "core")
	require.NoError(t, os.MkdirAll(coreDir, 0755))

	// Create go.mod for package detection
	goModContent := "module example.com/core\ngo 1.21\n"
	require.NoError(t, os.WriteFile(filepath.Join(coreDir, "go.mod"), []byte(goModContent), 0644))

	// Create version file
	versionContent := "package core\n\nconst Version = \"1.0.0\"\n"
	require.NoError(t, os.WriteFile(filepath.Join(coreDir, "version.go"), []byte(versionContent), 0644))

	// Run shipyard init
	cmd := exec.Command(shipyardBin, "init", "--yes")
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "failed to init shipyard: %s", string(output))
}

// initializeTestRepoMultiPackage creates a git repo with core/api/web packages and runs shipyard init
func initializeTestRepoMultiPackage(t *testing.T, shipyardBin, dir string) {
	t.Helper()

	initGitRepo(t, dir)

	packages := []struct {
		name string
		path string
	}{
		{"core", "core"},
		{"api", "api"},
		{"web", "web"},
	}

	for _, pkg := range packages {
		pkgDir := filepath.Join(dir, pkg.path)
		require.NoError(t, os.MkdirAll(pkgDir, 0755))

		versionContent := "package " + pkg.name + "\n\nconst Version = \"1.0.0\"\n"
		require.NoError(t, os.WriteFile(filepath.Join(pkgDir, "version.go"), []byte(versionContent), 0644))

		goModContent := "module example.com/" + pkg.name + "\ngo 1.21\n"
		require.NoError(t, os.WriteFile(filepath.Join(pkgDir, "go.mod"), []byte(goModContent), 0644))
	}

	// Run shipyard init
	cmd := exec.Command(shipyardBin, "init", "--yes")
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "failed to init shipyard: %s", string(output))
}

// createTestConsignment writes a consignment markdown file to the consignments directory
func createTestConsignment(t *testing.T, dir, id, packageName, changeType, summary string) {
	t.Helper()

	content := "---\nid: " + id + "\npackages:\n  - " + packageName + "\nchangeType: " + changeType + "\nsummary: " + summary + "\ntimestamp: 2026-01-30T00:00:00Z\n---\n# Change\n\n" + summary + "\n"

	consignmentPath := filepath.Join(dir, ".shipyard", "consignments", id+".md")
	require.NoError(t, os.WriteFile(consignmentPath, []byte(content), 0644))
}

// createInitialCommit stages all files in .shipyard/ and creates a git commit.
// This is required for commands that create tags or commits (version, prerelease, promote, snapshot).
func createInitialCommit(t *testing.T, dir string) {
	t.Helper()

	repo, err := gogit.PlainOpen(dir)
	require.NoError(t, err, "Failed to open git repository")

	wt, err := repo.Worktree()
	require.NoError(t, err, "Failed to get worktree")

	// Add all files
	_, err = wt.Add(".")
	require.NoError(t, err, "Failed to add files to staging")

	// Create commit
	_, err = wt.Commit("initial commit", &gogit.CommitOptions{
		Author: &object.Signature{
			Name:  "Test",
			Email: "test@example.com",
		},
	})
	require.NoError(t, err, "Failed to create initial commit")
}

// writeConfig overwrites .shipyard/shipyard.yaml with the given YAML content
func writeConfig(t *testing.T, dir, yamlContent string) {
	t.Helper()
	configPath := filepath.Join(dir, ".shipyard", "shipyard.yaml")
	require.NoError(t, os.WriteFile(configPath, []byte(yamlContent), 0644))
}

// writePrereleaseState writes .shipyard/prerelease.yml with the given YAML content
func writePrereleaseState(t *testing.T, dir, yamlContent string) {
	t.Helper()
	statePath := filepath.Join(dir, ".shipyard", "prerelease.yml")
	require.NoError(t, os.WriteFile(statePath, []byte(yamlContent), 0644))
}

// writeHistoryJSON overwrites .shipyard/history.json with the given JSON content
func writeHistoryJSON(t *testing.T, dir, jsonContent string) {
	t.Helper()
	historyPath := filepath.Join(dir, ".shipyard", "history.json")
	require.NoError(t, os.WriteFile(historyPath, []byte(jsonContent), 0644))
}

// readFileContent reads and returns the contents of a file as a string
func readFileContent(t *testing.T, path string) string {
	t.Helper()
	content, err := os.ReadFile(path)
	require.NoError(t, err, "Failed to read file: %s", path)
	return string(content)
}

// getExitCode extracts the exit code from an exec.ExitError. Returns -1 if not an ExitError.
func getExitCode(err error) int {
	if err == nil {
		return 0
	}
	if exitErr, ok := err.(*exec.ExitError); ok {
		return exitErr.ExitCode()
	}
	return -1
}
