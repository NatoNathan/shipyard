package commands

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	"github.com/NatoNathan/shipyard/internal/config"
	shipyarderrors "github.com/NatoNathan/shipyard/internal/errors"
	"github.com/NatoNathan/shipyard/internal/prerelease"
)

// setupPrereleaseTestProject creates a test project with Go ecosystem, config with prerelease stages, and a git repo
func setupPrereleaseTestProject(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	// Create .shipyard directory structure
	require.NoError(t, os.MkdirAll(filepath.Join(dir, ".shipyard", "consignments"), 0755))

	// Create config
	cfg := config.Config{
		Packages: []config.Package{
			{
				Name:      "my-api",
				Path:      ".",
				Ecosystem: "go",
			},
		},
		PreRelease: config.PreReleaseConfig{
			Stages: []config.StageConfig{
				{Name: "alpha", Order: 1, TagTemplate: "v{{.Version}}-alpha.{{.Counter}}"},
				{Name: "beta", Order: 2, TagTemplate: "v{{.Version}}-beta.{{.Counter}}"},
				{Name: "rc", Order: 3, TagTemplate: "v{{.Version}}-rc.{{.Counter}}"},
			},
			SnapshotTagTemplate: "v{{.Version}}-snapshot.{{.Timestamp}}",
		},
	}
	cfgData, err := yaml.Marshal(&cfg)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".shipyard", "shipyard.yaml"), cfgData, 0644))

	// Create version.go with current version
	require.NoError(t, os.WriteFile(filepath.Join(dir, "version.go"), []byte(`package main

const Version = "1.1.5"
`), 0644))

	// Create a consignment
	consignmentContent := `---
id: 20240130-120000-abc123
timestamp: 2024-01-30T12:00:00Z
packages:
  - my-api
changeType: minor
---

Add new API endpoint
`
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".shipyard", "consignments", "20240130-120000-abc123.md"), []byte(consignmentContent), 0644))

	// Init git repo and make initial commit
	repo, err := gogit.PlainInit(dir, false)
	require.NoError(t, err)
	wt, err := repo.Worktree()
	require.NoError(t, err)
	_, err = wt.Add(".")
	require.NoError(t, err)
	_, err = wt.Commit("initial commit", &gogit.CommitOptions{
		Author: &object.Signature{Name: "Test", Email: "test@example.com"},
	})
	require.NoError(t, err)

	return dir
}

func TestPrerelease_Preview(t *testing.T) {
	dir := setupPrereleaseTestProject(t)

	opts := &PrereleaseCommandOptions{
		Preview: true,
	}
	err := runPrereleaseWithDir(dir, opts)
	assert.NoError(t, err)

	// State file should NOT be created in preview mode
	assert.False(t, prerelease.Exists(filepath.Join(dir, ".shipyard", "prerelease.yml")))
}

func TestPrerelease_FirstPrerelease(t *testing.T) {
	dir := setupPrereleaseTestProject(t)

	opts := &PrereleaseCommandOptions{
		NoCommit: true,
	}
	err := runPrereleaseWithDir(dir, opts)
	assert.NoError(t, err)

	// Check state was created
	statePath := filepath.Join(dir, ".shipyard", "prerelease.yml")
	assert.True(t, prerelease.Exists(statePath))

	state, err := prerelease.ReadState(statePath)
	require.NoError(t, err)

	pkgState, ok := state.Packages["my-api"]
	require.True(t, ok)
	assert.Equal(t, "alpha", pkgState.Stage)
	assert.Equal(t, 1, pkgState.Counter)
	assert.Equal(t, "1.2.0", pkgState.TargetVersion)

	// Check version file was updated
	content, err := os.ReadFile(filepath.Join(dir, "version.go"))
	require.NoError(t, err)
	assert.Contains(t, string(content), `"1.2.0-alpha.1"`)
}

func TestPrerelease_IncrementCounter(t *testing.T) {
	dir := setupPrereleaseTestProject(t)

	// Set up existing state — target matches what consignments produce
	statePath := filepath.Join(dir, ".shipyard", "prerelease.yml")
	existingState := &prerelease.State{
		Packages: map[string]prerelease.PackageState{
			"my-api": {Stage: "alpha", Counter: 2, TargetVersion: "1.2.0"},
		},
	}
	require.NoError(t, prerelease.WriteState(statePath, existingState))

	// version.go keeps the original stable version (1.1.5)
	// The propagator calculates target from base version: minor bump from 1.1.5 → 1.2.0
	// which matches state target, so counter increments

	opts := &PrereleaseCommandOptions{
		NoCommit: true,
	}
	err := runPrereleaseWithDir(dir, opts)
	assert.NoError(t, err)

	state, err := prerelease.ReadState(statePath)
	require.NoError(t, err)

	pkgState := state.Packages["my-api"]
	assert.Equal(t, "alpha", pkgState.Stage)
	assert.Equal(t, 3, pkgState.Counter)

	// Check version file was updated
	content, err := os.ReadFile(filepath.Join(dir, "version.go"))
	require.NoError(t, err)
	assert.Contains(t, string(content), `"1.2.0-alpha.3"`)
}

func TestPrerelease_NoConsignments(t *testing.T) {
	dir := setupPrereleaseTestProject(t)

	// Remove consignments
	os.RemoveAll(filepath.Join(dir, ".shipyard", "consignments"))
	require.NoError(t, os.MkdirAll(filepath.Join(dir, ".shipyard", "consignments"), 0755))

	opts := &PrereleaseCommandOptions{}
	err := runPrereleaseWithDir(dir, opts)
	require.Error(t, err)

	var exitErr *shipyarderrors.ExitCodeError
	require.True(t, errors.As(err, &exitErr))
	assert.Equal(t, 2, exitErr.Code)
}

func TestPrerelease_TargetVersionChanged(t *testing.T) {
	dir := setupPrereleaseTestProject(t)

	// State says target was 1.1.0 (a patch bump), but consignment says minor → 1.2.0
	statePath := filepath.Join(dir, ".shipyard", "prerelease.yml")
	existingState := &prerelease.State{
		Packages: map[string]prerelease.PackageState{
			"my-api": {Stage: "beta", Counter: 5, TargetVersion: "1.1.0"},
		},
	}
	require.NoError(t, prerelease.WriteState(statePath, existingState))

	// version.go at original stable version (1.1.5)
	// Propagator: minor bump from 1.1.5 → 1.2.0 (doesn't match state target 1.1.0)

	opts := &PrereleaseCommandOptions{
		NoCommit: true,
	}
	err := runPrereleaseWithDir(dir, opts)
	assert.NoError(t, err)

	state, err := prerelease.ReadState(statePath)
	require.NoError(t, err)

	pkgState := state.Packages["my-api"]
	assert.Equal(t, "beta", pkgState.Stage)           // Stage unchanged
	assert.Equal(t, 1, pkgState.Counter)               // Counter reset due to target change
	assert.Equal(t, "1.2.0", pkgState.TargetVersion)   // New target
}
