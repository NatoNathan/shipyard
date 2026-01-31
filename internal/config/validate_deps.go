package config

import (
	"fmt"
	"strings"
)

// ValidateDependencies checks that all package dependencies reference existing packages.
// Returns an error if any dependency references a non-existent package.
//
// Validation rules:
//  - All dependency package names must exist in the config
//  - Self-dependencies are allowed (package depending on itself)
//  - Circular dependencies are allowed (will be handled by cycle detection)
//  - Empty dependency lists are valid
func ValidateDependencies(cfg *Config) error {
	if cfg == nil {
		return fmt.Errorf("cannot validate dependencies: config is nil")
	}

	// Build a map of existing package names for O(1) lookup
	existingPackages := make(map[string]bool)
	for _, pkg := range cfg.Packages {
		existingPackages[pkg.Name] = true
	}

	// Collect all validation errors
	var errors []string

	// Check each package's dependencies
	for _, pkg := range cfg.Packages {
		for _, dep := range pkg.Dependencies {
			// Check if dependency package exists
			if !existingPackages[dep.Package] {
				errors = append(errors, fmt.Sprintf(
					"package %q depends on non-existent package %q",
					pkg.Name,
					dep.Package,
				))
			}
		}
	}

	// If errors found, return them all
	if len(errors) > 0 {
		return fmt.Errorf("dependency validation failed:\n  - %s", strings.Join(errors, "\n  - "))
	}

	return nil
}
