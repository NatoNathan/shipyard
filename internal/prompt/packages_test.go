package prompt

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPromptForPackages_SingleSelection tests selecting one package
func TestPromptForPackages_SingleSelection(t *testing.T) {
	// Setup: Available packages
	available := []string{"core", "api", "web"}

	// Mock: Simulate user selecting "core"
	mockInput := func() ([]string, error) {
		return []string{"core"}, nil
	}

	// Test: Prompt for packages
	selected, err := promptForPackagesWithInput(available, mockInput)

	// Verify: Should return selected package
	require.NoError(t, err)
	assert.Equal(t, []string{"core"}, selected)
}

// TestPromptForPackages_MultipleSelection tests selecting multiple packages
func TestPromptForPackages_MultipleSelection(t *testing.T) {
	// Setup: Available packages
	available := []string{"core", "api", "web"}

	// Mock: Simulate user selecting multiple packages
	mockInput := func() ([]string, error) {
		return []string{"core", "api"}, nil
	}

	// Test: Prompt for packages
	selected, err := promptForPackagesWithInput(available, mockInput)

	// Verify: Should return all selected packages
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"core", "api"}, selected)
}

// TestPromptForPackages_EmptySelection tests when no packages are selected
func TestPromptForPackages_EmptySelection(t *testing.T) {
	// Setup: Available packages
	available := []string{"core", "api", "web"}

	// Mock: Simulate user selecting nothing
	mockInput := func() ([]string, error) {
		return []string{}, nil
	}

	// Test: Prompt for packages
	selected, err := promptForPackagesWithInput(available, mockInput)

	// Verify: Should return error for empty selection
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must select at least one package")
	assert.Nil(t, selected)
}

// TestPromptForPackages_NoPackagesAvailable tests when no packages exist
func TestPromptForPackages_NoPackagesAvailable(t *testing.T) {
	// Setup: No available packages
	available := []string{}

	// Mock: Not used
	mockInput := func() ([]string, error) {
		return []string{}, nil
	}

	// Test: Prompt for packages
	selected, err := promptForPackagesWithInput(available, mockInput)

	// Verify: Should return error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no packages available")
	assert.Nil(t, selected)
}

// TestPromptForPackages_AllPackages tests selecting all available packages
func TestPromptForPackages_AllPackages(t *testing.T) {
	// Setup: Available packages
	available := []string{"core", "api", "web", "cli"}

	// Mock: Simulate user selecting all packages
	mockInput := func() ([]string, error) {
		return []string{"core", "api", "web", "cli"}, nil
	}

	// Test: Prompt for packages
	selected, err := promptForPackagesWithInput(available, mockInput)

	// Verify: Should return all packages
	require.NoError(t, err)
	assert.ElementsMatch(t, available, selected)
}

// promptForPackagesWithInput is a helper that allows testing with mocked input
func promptForPackagesWithInput(available []string, inputFunc func() ([]string, error)) ([]string, error) {
	// This will be implemented to use the inputFunc for testing
	return PromptForPackagesFunc(available, inputFunc)
}
