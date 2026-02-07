package commands

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	shipyarderrors "github.com/NatoNathan/shipyard/internal/errors"
	"github.com/NatoNathan/shipyard/internal/prerelease"
)

func TestSnapshot_Preview(t *testing.T) {
	dir := setupPrereleaseTestProject(t)

	now := time.Date(2026, 2, 4, 15, 30, 45, 0, time.UTC)
	opts := &SnapshotCommandOptions{
		Preview: true,
	}
	err := runSnapshotWithDir(dir, opts, now)
	assert.NoError(t, err)

	// Version file should NOT change in preview
	content, err := os.ReadFile(filepath.Join(dir, "version.go"))
	require.NoError(t, err)
	assert.Contains(t, string(content), `"1.1.5"`)
}

func TestSnapshot_CreateSnapshot(t *testing.T) {
	dir := setupPrereleaseTestProject(t)

	now := time.Date(2026, 2, 4, 15, 30, 45, 0, time.UTC)
	opts := &SnapshotCommandOptions{
		NoCommit: true,
	}
	err := runSnapshotWithDir(dir, opts, now)
	assert.NoError(t, err)

	// Check version file was updated with snapshot version
	content, err := os.ReadFile(filepath.Join(dir, "version.go"))
	require.NoError(t, err)
	assert.Contains(t, string(content), `"1.2.0-snapshot.20260204-153045"`)
}

func TestSnapshot_DoesNotAffectState(t *testing.T) {
	dir := setupPrereleaseTestProject(t)

	// Set up pre-existing state
	statePath := filepath.Join(dir, ".shipyard", "prerelease.yml")
	existingState := &prerelease.State{
		Packages: map[string]prerelease.PackageState{
			"my-api": {Stage: "alpha", Counter: 3, TargetVersion: "1.2.0"},
		},
	}
	require.NoError(t, prerelease.WriteState(statePath, existingState))

	now := time.Date(2026, 2, 4, 15, 30, 45, 0, time.UTC)
	opts := &SnapshotCommandOptions{
		NoCommit: true,
	}
	err := runSnapshotWithDir(dir, opts, now)
	assert.NoError(t, err)

	// State file should be unchanged
	state, err := prerelease.ReadState(statePath)
	require.NoError(t, err)
	pkgState := state.Packages["my-api"]
	assert.Equal(t, "alpha", pkgState.Stage)
	assert.Equal(t, 3, pkgState.Counter)
	assert.Equal(t, "1.2.0", pkgState.TargetVersion)
}

func TestSnapshot_NoStateFileCreated(t *testing.T) {
	dir := setupPrereleaseTestProject(t)

	now := time.Date(2026, 2, 4, 15, 30, 45, 0, time.UTC)
	opts := &SnapshotCommandOptions{
		NoCommit: true,
	}
	err := runSnapshotWithDir(dir, opts, now)
	assert.NoError(t, err)

	// State file should NOT be created by snapshot
	statePath := filepath.Join(dir, ".shipyard", "prerelease.yml")
	assert.False(t, prerelease.Exists(statePath))
}

func TestSnapshot_NoConsignments(t *testing.T) {
	dir := setupPrereleaseTestProject(t)

	// Remove consignments
	os.RemoveAll(filepath.Join(dir, ".shipyard", "consignments"))
	require.NoError(t, os.MkdirAll(filepath.Join(dir, ".shipyard", "consignments"), 0755))

	now := time.Date(2026, 2, 4, 15, 30, 45, 0, time.UTC)
	opts := &SnapshotCommandOptions{}
	err := runSnapshotWithDir(dir, opts, now)
	require.Error(t, err)

	var exitErr *shipyarderrors.ExitCodeError
	require.True(t, errors.As(err, &exitErr))
	assert.Equal(t, 2, exitErr.Code)
}

func TestSnapshot_WithExistingPreRelease(t *testing.T) {
	dir := setupPrereleaseTestProject(t)

	// version.go stays at original 1.1.5
	// Snapshot uses base version for propagation: minor from 1.1.5 → 1.2.0

	now := time.Date(2026, 2, 4, 15, 30, 45, 0, time.UTC)
	opts := &SnapshotCommandOptions{
		NoCommit: true,
	}
	err := runSnapshotWithDir(dir, opts, now)
	assert.NoError(t, err)

	// Snapshot should use target version 1.2.0 from propagation
	content, err := os.ReadFile(filepath.Join(dir, "version.go"))
	require.NoError(t, err)
	assert.Contains(t, string(content), `"1.2.0-snapshot.20260204-153045"`)
}
