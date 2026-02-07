package prerelease

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadState_NonExistent(t *testing.T) {
	state, err := ReadState("/nonexistent/path/prerelease.yml")
	require.NoError(t, err)
	assert.NotNil(t, state.Packages)
	assert.Empty(t, state.Packages)
}

func TestWriteAndReadState(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "prerelease.yml")

	state := &State{
		Packages: map[string]PackageState{
			"my-api": {
				Stage:         "alpha",
				Counter:       3,
				TargetVersion: "1.2.0",
			},
			"shared-lib": {
				Stage:         "beta",
				Counter:       1,
				TargetVersion: "0.5.0",
			},
		},
	}

	err := WriteState(path, state)
	require.NoError(t, err)

	// Read back
	got, err := ReadState(path)
	require.NoError(t, err)
	assert.Equal(t, state.Packages["my-api"], got.Packages["my-api"])
	assert.Equal(t, state.Packages["shared-lib"], got.Packages["shared-lib"])
}

func TestDeleteState(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "prerelease.yml")

	// Write file first
	state := &State{Packages: map[string]PackageState{"test": {Stage: "alpha", Counter: 1, TargetVersion: "1.0.0"}}}
	err := WriteState(path, state)
	require.NoError(t, err)
	assert.True(t, Exists(path))

	// Delete it
	err = DeleteState(path)
	require.NoError(t, err)
	assert.False(t, Exists(path))
}

func TestDeleteState_NonExistent(t *testing.T) {
	err := DeleteState("/nonexistent/path/prerelease.yml")
	assert.NoError(t, err)
}

func TestExists(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "prerelease.yml")

	assert.False(t, Exists(path))

	err := os.WriteFile(path, []byte("packages: {}"), 0644)
	require.NoError(t, err)

	assert.True(t, Exists(path))
}

func TestReadState_EmptyPackages(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "prerelease.yml")

	err := os.WriteFile(path, []byte("packages:\n"), 0644)
	require.NoError(t, err)

	state, err := ReadState(path)
	require.NoError(t, err)
	assert.NotNil(t, state.Packages)
}

func TestConcurrentReadWrite(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "prerelease.yml")

	// Write initial state
	initial := &State{
		Packages: map[string]PackageState{
			"pkg": {Stage: "alpha", Counter: 1, TargetVersion: "1.0.0"},
		},
	}
	err := WriteState(path, initial)
	require.NoError(t, err)

	// Run concurrent reads and writes
	done := make(chan bool, 20)
	for i := 0; i < 10; i++ {
		// Concurrent readers
		go func() {
			defer func() { done <- true }()
			state, err := ReadState(path)
			assert.NoError(t, err)
			assert.NotNil(t, state)
			assert.NotNil(t, state.Packages)
		}()

		// Concurrent writers
		go func(counter int) {
			defer func() { done <- true }()
			state := &State{
				Packages: map[string]PackageState{
					"pkg": {Stage: "alpha", Counter: counter, TargetVersion: "1.0.0"},
				},
			}
			err := WriteState(path, state)
			assert.NoError(t, err)
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 20; i++ {
		<-done
	}

	// Final read should succeed and return valid state
	final, err := ReadState(path)
	require.NoError(t, err)
	assert.NotNil(t, final.Packages)
	assert.Contains(t, final.Packages, "pkg")
}
