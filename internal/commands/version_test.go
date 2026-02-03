package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestVersionCommand_BasicExecution tests the basic version command execution
func TestVersionCommand_BasicExecution(t *testing.T) {
	t.Run("no consignments no-op", func(t *testing.T) {
		// Setup: Create a git repo with shipyard initialized
		tempDir := t.TempDir()
		initGitRepo(t, tempDir)

		// Initialize shipyard
		err := runInit(tempDir, InitOptions{Yes: true})
		require.NoError(t, err)

		// Run version command with no consignments
		opts := &VersionCommandOptions{
			Preview:  false,
			NoCommit: true, // Don't create commits in tests
			NoTag:    true, // Don't create tags in tests
		}

		err = runVersionInDir(tempDir, opts)

		// Should succeed with no consignments (no-op)
		require.NoError(t, err)
	})

	t.Run("single patch consignment", func(t *testing.T) {
		// Setup: Create a git repo with shipyard initialized
		tempDir := t.TempDir()
		initGitRepo(t, tempDir)

		// Create a Go version file BEFORE initializing shipyard
		versionFile := filepath.Join(tempDir, "version.go")
		initialVersion := `package main

const Version = "1.0.0"
`
		require.NoError(t, os.WriteFile(versionFile, []byte(initialVersion), 0644))

		// Create go.mod so shipyard auto-detects this as a Go package
		goModFile := filepath.Join(tempDir, "go.mod")
		goModContent := `module example.com/test

go 1.21
`
		require.NoError(t, os.WriteFile(goModFile, []byte(goModContent), 0644))

		// Initialize shipyard with auto-detection
		err := runInit(tempDir, InitOptions{Yes: true})
		require.NoError(t, err)

		// Create a single patch consignment (use "test" as package name - matches module name)
		consignmentDir := filepath.Join(tempDir, ".shipyard", "consignments")
		consignmentContent := `---
id: test-001
packages:
  - test
changeType: patch
summary: Fix a bug
timestamp: 2026-01-30T00:00:00Z
---
# Bug Fix

Fixed an important bug.
`
		consignmentPath := filepath.Join(consignmentDir, "test-001.md")
		require.NoError(t, os.WriteFile(consignmentPath, []byte(consignmentContent), 0644))

		// Run version command
		opts := &VersionCommandOptions{
			Preview:  false,
			NoCommit: true,
			NoTag:    true,
		}

		err = runVersionInDir(tempDir, opts)

		// Should succeed
		require.NoError(t, err)

		// Verify version was bumped to 1.0.1
		content, err := os.ReadFile(versionFile)
		require.NoError(t, err)
		assert.Contains(t, string(content), `const Version = "1.0.1"`, "version should be bumped from 1.0.0 to 1.0.1")

		// Verify consignment was deleted
		_, err = os.Stat(consignmentPath)
		assert.True(t, os.IsNotExist(err), "consignment should be deleted after processing")

		// Verify consignment was archived to history
		historyPath := filepath.Join(tempDir, ".shipyard", "history.json")
		historyContent, err := os.ReadFile(historyPath)
		require.NoError(t, err)
		assert.Contains(t, string(historyContent), "test-001", "consignment should be archived in history")
	})
}

// TestVersionCommand_PreviewMode tests the --preview flag behavior
func TestVersionCommand_PreviewMode(t *testing.T) {
	t.Run("preview shows changes without applying", func(t *testing.T) {
		// Setup: Create initialized repo with consignment
		tempDir := setupVersionTestRepo(t)

		// Create a consignment
		consignmentsDir := filepath.Join(tempDir, ".shipyard", "consignments")
		createTestConsignmentForVersion(t, consignmentsDir, "c1", []string{"test-package"}, "minor", "Add feature")

		// Run version command with --preview
		opts := &VersionCommandOptions{
			Preview: true,
		}

		// Capture output
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := runVersionWithDir(tempDir, opts)

		w.Close()
		os.Stdout = oldStdout

		// Read captured output
		var buf [4096]byte
		n, _ := r.Read(buf[:])
		output := string(buf[:n])

		// Verify: Command succeeds and shows preview
		require.NoError(t, err, "Preview should not return error")
		assert.Contains(t, output, "Preview Mode", "Should show preview mode message")
		assert.Contains(t, output, "test-package", "Should show package name")
	})

	t.Run("preview does not modify files", func(t *testing.T) {
		// Setup: Create initialized repo with consignment
		tempDir := setupVersionTestRepo(t)

		// Record version file timestamp before
		versionFile := filepath.Join(tempDir, "test-package", "version.go")
		beforeStat, err := os.Stat(versionFile)
		require.NoError(t, err)
		beforeModTime := beforeStat.ModTime()

		// Create a consignment
		consignmentsDir := filepath.Join(tempDir, ".shipyard", "consignments")
		createTestConsignmentForVersion(t, consignmentsDir, "c1", []string{"test-package"}, "patch", "Fix bug")

		// Wait a moment to ensure timestamp would change if file was modified
		time.Sleep(10 * time.Millisecond)

		// Run version command with --preview
		opts := &VersionCommandOptions{
			Preview: true,
		}

		err = runVersionWithDir(tempDir, opts)
		require.NoError(t, err)

		// Verify: Version file was NOT modified
		afterStat, err := os.Stat(versionFile)
		require.NoError(t, err)
		afterModTime := afterStat.ModTime()

		assert.Equal(t, beforeModTime, afterModTime, "Preview mode should not modify version files")

		// Verify: Consignment files still exist (not archived)
		consignmentFile := filepath.Join(consignmentsDir, "c1.md")
		assert.FileExists(t, consignmentFile, "Preview mode should not archive consignments")
	})
}

// TestVersionCommand_VersionBumpPropagation tests dependency version propagation
func TestVersionCommand_VersionBumpPropagation(t *testing.T) {
	tests := []struct {
		name              string
		packages          []string
		dependencies      map[string][]string
		directChanges     map[string]string // package -> changeType
		expectedVersions  map[string]string // package -> version
		description       string
	}{
		{
			name:     "linked dependency propagation",
			packages: []string{"core", "api-client"},
			dependencies: map[string][]string{
				"api-client": {"core"},
			},
			directChanges: map[string]string{
				"core": "minor",
			},
			expectedVersions: map[string]string{
				"core":       "1.1.0", // minor bump
				"api-client": "2.1.0", // propagated minor
			},
			description: "linked dependencies should propagate version bumps",
		},
		{
			name:     "fixed dependency no propagation",
			packages: []string{"core", "tool"},
			dependencies: map[string][]string{
				"tool": {"core"}, // but with fixed strategy
			},
			directChanges: map[string]string{
				"core": "major",
			},
			expectedVersions: map[string]string{
				"core": "2.0.0", // major bump
				"tool": "0.5.0", // no change (fixed)
			},
			description: "fixed dependencies should not propagate",
		},
		{
			name:     "cycle bump resolution",
			packages: []string{"service-a", "service-b", "service-c"},
			dependencies: map[string][]string{
				"service-a": {"service-b"},
				"service-b": {"service-c"},
				"service-c": {"service-a"}, // creates cycle
			},
			directChanges: map[string]string{
				"service-a": "minor",
				"service-b": "patch",
			},
			expectedVersions: map[string]string{
				"service-a": "1.1.0", // all get max bump (minor)
				"service-b": "1.1.0",
				"service-c": "1.1.0",
			},
			description: "cycles should apply max bump to all members",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Skip("Implementation pending")
		})
	}
}

// TestVersionCommand_FileUpdates tests version file updates for different ecosystems
func TestVersionCommand_FileUpdates(t *testing.T) {
	tests := []struct {
		name         string
		ecosystem    string
		versionFile  string
		oldVersion   string
		newVersion   string
		expectedContent string
		description  string
	}{
		{
			name:        "go version.go update",
			ecosystem:   "go",
			versionFile: "version.go",
			oldVersion:  "1.0.0",
			newVersion:  "1.1.0",
			expectedContent: `const Version = "1.1.0"`,
			description: "should update Go version constant",
		},
		{
			name:        "npm package.json update",
			ecosystem:   "npm",
			versionFile: "package.json",
			oldVersion:  "2.0.0",
			newVersion:  "2.1.0",
			expectedContent: `"version": "2.1.0"`,
			description: "should update package.json version field",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Skip("Implementation pending")
		})
	}
}

// TestVersionCommand_ChangelogGeneration tests changelog generation
func TestVersionCommand_ChangelogGeneration(t *testing.T) {
	tests := []struct {
		name            string
		consignments    []string
		template        string
		expectedContent []string // content that should be present
		description     string
	}{
		{
			name:         "default template",
			consignments: []string{"fix: bug", "feat: feature"},
			template:     "builtin:default",
			expectedContent: []string{
				"## ",      // section header
				"- ",       // list item
				"fix: bug",
				"feat: feature",
			},
			description: "default template should format changes as list",
		},
		{
			name:         "custom template with metadata",
			consignments: []string{"fix: bug"},
			template:     "custom",
			expectedContent: []string{
				"Author:",
				"Issue:",
			},
			description: "custom template should include metadata",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Skip("Implementation pending")
		})
	}
}

// TestVersionCommand_GitOperations tests git tag and commit creation
func TestVersionCommand_GitOperations(t *testing.T) {
	tests := []struct {
		name        string
		packages    []string
		versions    map[string]string
		expectedTag string
		noCommit    bool
		noTag       bool
		description string
	}{
		{
			name:     "create annotated tag",
			packages: []string{"core"},
			versions: map[string]string{
				"core": "1.2.0",
			},
			expectedTag: "v1.2.0",
			description: "should create annotated git tag",
		},
		{
			name:     "skip commit with --no-commit",
			packages: []string{"core"},
			versions: map[string]string{
				"core": "1.2.0",
			},
			noCommit:    true,
			description: "should not create commit when --no-commit flag set",
		},
		{
			name:     "skip tags with --no-tag",
			packages: []string{"core"},
			versions: map[string]string{
				"core": "1.2.0",
			},
			noTag:       true,
			description: "should not create tags when --no-tag flag set",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Skip("Implementation pending")
		})
	}
}

// TestVersionCommand_HistoryArchival tests consignment archival to history.json
func TestVersionCommand_HistoryArchival(t *testing.T) {
	tests := []struct {
		name              string
		consignmentCount  int
		existingHistory   int
		expectedTotal     int
		description       string
	}{
		{
			name:             "append to empty history",
			consignmentCount: 1,
			existingHistory:  0,
			expectedTotal:    1,
			description:      "should create history file with first entry",
		},
		{
			name:             "append to existing history",
			consignmentCount: 2,
			existingHistory:  5,
			expectedTotal:    7,
			description:      "should append to existing history file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Skip("Implementation pending")
		})
	}
}

// TestVersionCommand_ConsignmentDeletion tests deletion of processed consignments
func TestVersionCommand_ConsignmentDeletion(t *testing.T) {
	t.Run("delete after successful processing", func(t *testing.T) {
		t.Skip("Implementation pending")
	})

	t.Run("preserve on error", func(t *testing.T) {
		t.Skip("Implementation pending")
	})
}

// TestVersionCommand_PackageFilter tests --package flag filtering
func TestVersionCommand_PackageFilter(t *testing.T) {
	tests := []struct {
		name            string
		allPackages     []string
		filterPackages  []string
		expectedUpdated []string
		description     string
	}{
		{
			name:            "single package filter",
			allPackages:     []string{"core", "api", "web"},
			filterPackages:  []string{"core"},
			expectedUpdated: []string{"core"},
			description:     "should only update specified package",
		},
		{
			name:            "multiple package filter",
			allPackages:     []string{"core", "api", "web"},
			filterPackages:  []string{"core", "api"},
			expectedUpdated: []string{"core", "api"},
			description:     "should update multiple specified packages",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Skip("Implementation pending")
		})
	}
}

// TestVersionCommand_ErrorHandling tests error scenarios
func TestVersionCommand_ErrorHandling(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(t *testing.T, dir string) error
		expectedErr string
		description string
	}{
		{
			name:        "not initialized",
			setup:       func(t *testing.T, dir string) error { return nil },
			expectedErr: "not initialized",
			description: "should error if shipyard not initialized",
		},
		{
			name: "invalid consignment",
			setup: func(t *testing.T, dir string) error {
				// Create malformed consignment
				return nil
			},
			expectedErr: "invalid consignment",
			description: "should error on malformed consignment",
		},
		{
			name: "dirty working tree",
			setup: func(t *testing.T, dir string) error {
				// Create uncommitted changes
				return nil
			},
			expectedErr: "uncommitted changes",
			description: "should error if git working tree is dirty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Skip("Implementation pending")
		})
	}
}

// TestVersionCommand_ConcurrentSafety tests concurrent execution safety
func TestVersionCommand_ConcurrentSafety(t *testing.T) {
	t.Run("file locking prevents corruption", func(t *testing.T) {
		t.Skip("Implementation pending - requires concurrent process simulation")
	})
}

// Acceptance test covering all scenarios from spec.md
func TestVersionCommand_AcceptanceScenarios(t *testing.T) {
	// US5-AC1: Single package version bump
	t.Run("US5-AC1: single package bump", func(t *testing.T) {
		t.Skip("Implementation pending")
	})

	// US5-AC2: Multiple consignments for same package
	t.Run("US5-AC2: multiple consignments highest bump", func(t *testing.T) {
		t.Skip("Implementation pending")
	})

	// US5-AC3: Dependency propagation
	t.Run("US5-AC3: dependency propagation", func(t *testing.T) {
		t.Skip("Implementation pending")
	})

	// US5-AC4: Custom template
	t.Run("US5-AC4: custom changelog template", func(t *testing.T) {
		t.Skip("Implementation pending")
	})

	// US5-AC5: Preview mode
	t.Run("US5-AC5: preview mode", func(t *testing.T) {
		t.Skip("Implementation pending")
	})

	// US5-AC6: History archival
	t.Run("US5-AC6: history archival", func(t *testing.T) {
		t.Skip("Implementation pending")
	})
}

// Helper functions for test setup
func createTestRepo(t *testing.T) string {
	dir, err := os.MkdirTemp("", "shipyard-test-*")
	require.NoError(t, err)
	t.Cleanup(func() {
		os.RemoveAll(dir)
	})
	return dir
}

func createTestConsignment(t *testing.T, dir, pkg, changeType, summary string) string {
	consignmentDir := filepath.Join(dir, ".shipyard", "consignments")
	require.NoError(t, os.MkdirAll(consignmentDir, 0755))

	// Implementation would create actual consignment file
	return filepath.Join(consignmentDir, "test.md")
}

func assertVersionFileUpdated(t *testing.T, path, expectedVersion string) {
	content, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Contains(t, string(content), expectedVersion)
}

func assertGitTagExists(t *testing.T, repoPath, tag string) {
	// Implementation would check git tags
	t.Skip("Git tag assertion pending")
}

// runVersionInDir runs the version command in a specific directory
func runVersionInDir(dir string, opts *VersionCommandOptions) error {
	return runVersionWithDir(dir, opts)
}

// setupVersionTestRepo creates a fully initialized test repo with config and version files
func setupVersionTestRepo(t *testing.T) string {
	t.Helper()
	tempDir := t.TempDir()

	// Create .shipyard structure
	shipyardDir := filepath.Join(tempDir, ".shipyard")
	require.NoError(t, os.MkdirAll(shipyardDir, 0755))

	consignmentsDir := filepath.Join(shipyardDir, "consignments")
	require.NoError(t, os.MkdirAll(consignmentsDir, 0755))

	// Create config file
	configContent := `packages:
  - name: test-package
    path: ./test-package
    ecosystem: go
templates:
  changelog:
    source: "builtin:default"
consignments:
  path: ".shipyard/consignments"
history:
  path: ".shipyard/history.json"
`
	configPath := filepath.Join(shipyardDir, "shipyard.yaml")
	require.NoError(t, os.WriteFile(configPath, []byte(configContent), 0644))

	// Create package directory and version file
	pkgDir := filepath.Join(tempDir, "test-package")
	require.NoError(t, os.MkdirAll(pkgDir, 0755))

	versionFile := filepath.Join(pkgDir, "version.go")
	versionContent := `package testpackage

const Version = "1.0.0"
`
	require.NoError(t, os.WriteFile(versionFile, []byte(versionContent), 0644))

	// Create history file
	historyPath := filepath.Join(shipyardDir, "history.json")
	require.NoError(t, os.WriteFile(historyPath, []byte("[]"), 0644))

	return tempDir
}

// createTestConsignmentForVersion creates a consignment file for version tests
func createTestConsignmentForVersion(t *testing.T, consignmentsDir, id string, packages []string, changeType, summary string) {
	t.Helper()

	content := fmt.Sprintf(`---
id: %s
packages:
  - %s
changeType: %s
summary: %s
timestamp: %s
---
# Change

%s
`, id, packages[0], changeType, summary, time.Now().UTC().Format(time.RFC3339), summary)

	consignmentPath := filepath.Join(consignmentsDir, id+".md")
	require.NoError(t, os.WriteFile(consignmentPath, []byte(content), 0644))
}

// TestVersionCommand_ChangelogIncludesCurrentVersion verifies that the changelog
// generated by the version command includes the consignment being versioned,
// not just previous versions. This tests the fix for the critical bug where
// changelogs were generated BEFORE archiving consignments.
func TestVersionCommand_ChangelogIncludesCurrentVersion(t *testing.T) {
	// Setup: Create a git repo with shipyard initialized
	tempDir := t.TempDir()
	initGitRepo(t, tempDir)

	// Create a Go version file BEFORE initializing shipyard
	versionFile := filepath.Join(tempDir, "version.go")
	initialVersion := `package main

const Version = "1.0.0"
`
	require.NoError(t, os.WriteFile(versionFile, []byte(initialVersion), 0644))

	// Create go.mod so shipyard auto-detects this as a Go package
	goModFile := filepath.Join(tempDir, "go.mod")
	goModContent := `module example.com/testchangelog

go 1.21
`
	require.NoError(t, os.WriteFile(goModFile, []byte(goModContent), 0644))

	// Initialize shipyard with auto-detection
	err := runInit(tempDir, InitOptions{Yes: true})
	require.NoError(t, err)

	// Create and add consignment for version 1.1.0
	consignmentsDir := filepath.Join(tempDir, ".shipyard", "consignments")
	createTestConsignmentForVersion(t, consignmentsDir, "test-feature-1", []string{"testchangelog"}, "minor", "Add new feature")

	// Run version command to create 1.1.0
	opts := &VersionCommandOptions{
		Preview:  false,
		NoCommit: true,
		NoTag:    true,
	}
	err = runVersionInDir(tempDir, opts)
	require.NoError(t, err)

	// Verify CHANGELOG.md was created and contains the current version's changes
	changelogPath := filepath.Join(tempDir, "CHANGELOG.md")
	require.FileExists(t, changelogPath, "CHANGELOG.md should be created")

	changelogContent, err := os.ReadFile(changelogPath)
	require.NoError(t, err)
	changelogStr := string(changelogContent)

	// The critical assertion: The changelog MUST contain the feature we just versioned
	assert.Contains(t, changelogStr, "Add new feature", "Changelog must include the current version's changes")
	assert.Contains(t, changelogStr, "1.1.0", "Changelog must include the current version number")

	// Create another consignment for version 1.1.1
	createTestConsignmentForVersion(t, consignmentsDir, "test-bugfix-1", []string{"testchangelog"}, "patch", "Fix critical bug")

	// Run version command again to create 1.1.1
	err = runVersionInDir(tempDir, opts)
	require.NoError(t, err)

	// Re-read changelog
	changelogContent, err = os.ReadFile(changelogPath)
	require.NoError(t, err)
	changelogStr = string(changelogContent)

	// Verify the changelog now contains BOTH versions
	assert.Contains(t, changelogStr, "Add new feature", "Changelog must include first version's changes")
	assert.Contains(t, changelogStr, "Fix critical bug", "Changelog must include second version's changes")
	assert.Contains(t, changelogStr, "1.1.0", "Changelog must include first version number")
	assert.Contains(t, changelogStr, "1.1.1", "Changelog must include second version number")
}
