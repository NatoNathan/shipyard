package commands

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	shipyarderrors "github.com/NatoNathan/shipyard/internal/errors"
	"github.com/NatoNathan/shipyard/internal/prerelease"
)

func TestPromote_Preview(t *testing.T) {
	dir := setupPrereleaseTestProject(t)

	// Set up existing state at alpha
	statePath := filepath.Join(dir, ".shipyard", "prerelease.yml")
	existingState := &prerelease.State{
		Packages: map[string]prerelease.PackageState{
			"my-api": {Stage: "alpha", Counter: 5, TargetVersion: "1.2.0"},
		},
	}
	require.NoError(t, prerelease.WriteState(statePath, existingState))

	// version.go stays at original stable version
	// (promote reads base version and recalculates target)

	opts := &PromoteCommandOptions{
		Preview: true,
	}
	err := runPromoteWithDir(dir, opts)
	assert.NoError(t, err)

	// State should NOT change in preview mode
	state, err := prerelease.ReadState(statePath)
	require.NoError(t, err)
	assert.Equal(t, "alpha", state.Packages["my-api"].Stage)
	assert.Equal(t, 5, state.Packages["my-api"].Counter)
}

func TestPromote_AlphaToBeta(t *testing.T) {
	dir := setupPrereleaseTestProject(t)

	statePath := filepath.Join(dir, ".shipyard", "prerelease.yml")
	existingState := &prerelease.State{
		Packages: map[string]prerelease.PackageState{
			"my-api": {Stage: "alpha", Counter: 5, TargetVersion: "1.2.0"},
		},
	}
	require.NoError(t, prerelease.WriteState(statePath, existingState))

	// version.go at original stable version (1.1.5)
	// Propagator: minor bump from 1.1.5 → 1.2.0, matches state target

	opts := &PromoteCommandOptions{
		NoCommit: true,
	}
	err := runPromoteWithDir(dir, opts)
	assert.NoError(t, err)

	state, err := prerelease.ReadState(statePath)
	require.NoError(t, err)

	pkgState := state.Packages["my-api"]
	assert.Equal(t, "beta", pkgState.Stage)
	assert.Equal(t, 1, pkgState.Counter) // Counter reset to 1
	assert.Equal(t, "1.2.0", pkgState.TargetVersion)

	// Check version file
	content, err := os.ReadFile(filepath.Join(dir, "version.go"))
	require.NoError(t, err)
	assert.Contains(t, string(content), `"1.2.0-beta.1"`)
}

func TestPromote_BetaToRC(t *testing.T) {
	dir := setupPrereleaseTestProject(t)

	statePath := filepath.Join(dir, ".shipyard", "prerelease.yml")
	existingState := &prerelease.State{
		Packages: map[string]prerelease.PackageState{
			"my-api": {Stage: "beta", Counter: 3, TargetVersion: "1.2.0"},
		},
	}
	require.NoError(t, prerelease.WriteState(statePath, existingState))

	// version.go at original stable version

	opts := &PromoteCommandOptions{
		NoCommit: true,
	}
	err := runPromoteWithDir(dir, opts)
	assert.NoError(t, err)

	state, err := prerelease.ReadState(statePath)
	require.NoError(t, err)

	pkgState := state.Packages["my-api"]
	assert.Equal(t, "rc", pkgState.Stage)
	assert.Equal(t, 1, pkgState.Counter)
}

func TestPromote_NoStateFile(t *testing.T) {
	dir := setupPrereleaseTestProject(t)

	// No state file exists — promote should return ExitCodeError with code 3
	opts := &PromoteCommandOptions{}
	err := runPromoteWithDir(dir, opts)
	require.Error(t, err)

	var exitErr *shipyarderrors.ExitCodeError
	require.True(t, errors.As(err, &exitErr))
	assert.Equal(t, 3, exitErr.Code)
}

func TestPromote_HighestStage(t *testing.T) {
	dir := setupPrereleaseTestProject(t)

	// Set state at highest stage (rc, order=3)
	statePath := filepath.Join(dir, ".shipyard", "prerelease.yml")
	existingState := &prerelease.State{
		Packages: map[string]prerelease.PackageState{
			"my-api": {Stage: "rc", Counter: 2, TargetVersion: "1.2.0"},
		},
	}
	require.NoError(t, prerelease.WriteState(statePath, existingState))

	opts := &PromoteCommandOptions{NoCommit: true}
	err := runPromoteWithDir(dir, opts)
	require.Error(t, err)

	var exitErr *shipyarderrors.ExitCodeError
	require.True(t, errors.As(err, &exitErr))
	assert.Equal(t, 2, exitErr.Code)
}

func TestPromote_TargetVersionChanged(t *testing.T) {
	dir := setupPrereleaseTestProject(t)

	// State target was 1.1.0, but consignments produce 1.2.0
	statePath := filepath.Join(dir, ".shipyard", "prerelease.yml")
	existingState := &prerelease.State{
		Packages: map[string]prerelease.PackageState{
			"my-api": {Stage: "alpha", Counter: 2, TargetVersion: "1.1.0"},
		},
	}
	require.NoError(t, prerelease.WriteState(statePath, existingState))

	// version.go at original stable version (1.1.5)
	// Propagator: minor bump from 1.1.5 → 1.2.0 (doesn't match state target 1.1.0)

	opts := &PromoteCommandOptions{
		NoCommit: true,
	}
	err := runPromoteWithDir(dir, opts)
	assert.NoError(t, err)

	state, err := prerelease.ReadState(statePath)
	require.NoError(t, err)

	pkgState := state.Packages["my-api"]
	assert.Equal(t, "beta", pkgState.Stage)
	assert.Equal(t, 1, pkgState.Counter)
	assert.Equal(t, "1.2.0", pkgState.TargetVersion) // Updated to new target
}
