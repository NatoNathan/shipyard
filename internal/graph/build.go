package graph

import (
	"fmt"

	"github.com/NatoNathan/shipyard/internal/config"
)

// BuildGraph constructs a dependency graph from package configurations
// Returns an error if any dependency references a non-existent package
func BuildGraph(cfg *config.Config) (*DependencyGraph, error) {
	g := NewGraph()

	// First pass: Add all package nodes
	for _, pkg := range cfg.Packages {
		if err := g.AddNode(pkg); err != nil {
			return nil, fmt.Errorf("failed to add package node %s: %w", pkg.Name, err)
		}
	}

	// Second pass: Add dependency edges
	for _, pkg := range cfg.Packages {
		for _, dep := range pkg.Dependencies {
			// Verify dependency target exists
			if _, exists := g.GetNode(dep.Package); !exists {
				return nil, fmt.Errorf("package %s depends on non-existent package %s", pkg.Name, dep.Package)
			}

			// Add edge from dependent package to dependency
			// (Note: edge direction is from dependent TO dependency,
			// representing "pkg depends on dep.Package")
			strategy := dep.Strategy
			if strategy == "" {
				strategy = "linked" // Default strategy
			}

			if err := g.AddEdge(pkg.Name, dep.Package, strategy, dep.BumpMapping); err != nil {
				return nil, fmt.Errorf("failed to add dependency edge from %s to %s: %w", pkg.Name, dep.Package, err)
			}
		}
	}

	return g, nil
}

// BuildGraphFromPackages is a convenience function for building a graph from a slice of packages
// It creates a temporary config and calls BuildGraph
func BuildGraphFromPackages(packages []config.Package) (*DependencyGraph, error) {
	cfg := &config.Config{
		Packages: packages,
	}
	return BuildGraph(cfg)
}
