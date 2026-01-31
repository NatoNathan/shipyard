package prompt

import (
	"testing"

	"github.com/NatoNathan/shipyard/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPromptReviewPackages_AllSelected tests selecting all packages
func TestPromptReviewPackages_AllSelected(t *testing.T) {
	// Setup: Available packages
	packages := []config.Package{
		{Name: "core", Path: "./core", Ecosystem: config.EcosystemGo},
		{Name: "api", Path: "./api", Ecosystem: config.EcosystemGo},
	}

	// Mock: Simulate user accepting all
	mockInput := func() ([]config.Package, error) {
		return packages, nil
	}

	// Test: Review packages
	selected, err := PromptReviewPackagesFunc(packages, mockInput)

	// Verify: Should return all packages
	require.NoError(t, err)
	assert.Len(t, selected, 2)
	assert.Equal(t, packages, selected)
}

// TestPromptReviewPackages_PartialSelection tests selecting some packages
func TestPromptReviewPackages_PartialSelection(t *testing.T) {
	// Setup: Available packages
	packages := []config.Package{
		{Name: "core", Path: "./core", Ecosystem: config.EcosystemGo},
		{Name: "api", Path: "./api", Ecosystem: config.EcosystemGo},
		{Name: "web", Path: "./web", Ecosystem: config.EcosystemNPM},
	}

	// Mock: Simulate user selecting only first two
	mockInput := func() ([]config.Package, error) {
		return []config.Package{packages[0], packages[1]}, nil
	}

	// Test: Review packages
	selected, err := PromptReviewPackagesFunc(packages, mockInput)

	// Verify: Should return only selected packages
	require.NoError(t, err)
	assert.Len(t, selected, 2)
	assert.Contains(t, selected, packages[0])
	assert.Contains(t, selected, packages[1])
}

// TestPromptReviewPackages_SinglePackage tests single package scenario
func TestPromptReviewPackages_SinglePackage(t *testing.T) {
	// Setup: Single package
	packages := []config.Package{
		{Name: "app", Path: "./", Ecosystem: config.EcosystemGo},
	}

	// Mock: Simulate user accepting it
	mockInput := func() ([]config.Package, error) {
		return packages, nil
	}

	// Test: Review packages
	selected, err := PromptReviewPackagesFunc(packages, mockInput)

	// Verify: Should return the package
	require.NoError(t, err)
	assert.Len(t, selected, 1)
	assert.Equal(t, packages[0], selected[0])
}

// TestPromptReviewPackages_EmptyPackages tests empty package list
func TestPromptReviewPackages_EmptyPackages(t *testing.T) {
	// Setup: No packages
	packages := []config.Package{}

	// Mock: Not used
	mockInput := func() ([]config.Package, error) {
		return nil, nil
	}

	// Test: Review packages
	selected, err := PromptReviewPackagesFunc(packages, mockInput)

	// Verify: Should return error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no packages to review")
	assert.Nil(t, selected)
}
