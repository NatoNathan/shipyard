// Shipyard Dagger Module
//
// This module handles building, packaging, and releasing the Shipyard CLI tool
// across multiple platforms and distribution channels.

package main

import (
	"context"
	"dagger/shipyard/internal/dagger"
	"fmt"
	"sync"
)

type Shipyard struct{}

// Release builds, packages, and publishes Shipyard to all distribution channels
func (m *Shipyard) Release(
	ctx context.Context,
	// Source code directory
	source *dagger.Directory,
	// Version string (e.g., "v1.2.3")
	version string,
	// GitHub token for releases and Docker registry
	githubToken *dagger.Secret,
	// npm registry token
	npmToken *dagger.Secret,
	// Docker registry (e.g., "ghcr.io/natonathan/shipyard")
	// +default="ghcr.io/natonathan/shipyard"
	dockerRegistry string,
	// Docker registry username (GitHub actor for GHCR)
	dockerUsername string,
	// Docker registry token (usually same as GitHub token for ghcr.io)
	// +optional
	dockerToken *dagger.Secret,
) error {
	// Use GitHub token for Docker if not provided
	if dockerToken == nil {
		dockerToken = githubToken
	}

	// Get commit SHA from git
	commit, err := dag.Container().
		From("alpine/git:latest").
		WithMountedDirectory("/src", source).
		WithWorkdir("/src").
		WithExec([]string{"git", "rev-parse", "HEAD"}).
		Stdout(ctx)

	if err != nil {
		return fmt.Errorf("failed to get git commit: %w", err)
	}

	fmt.Printf("🚢 Building Shipyard %s (commit: %.7s)\n", version, commit)

	// Stage 1: Build
	fmt.Printf("\n📦 Stage 1: Building binaries...\n")
	buildArtifacts := m.Build(ctx, source, version, commit)

	// Stage 2: Package
	fmt.Printf("\n📦 Stage 2: Creating distribution packages...\n")
	packageArtifacts := m.Package(ctx, buildArtifacts, version)

	// Stage 3: Publish (all in parallel)
	fmt.Printf("\n🚀 Stage 3: Publishing to distribution channels...\n")

	var wg sync.WaitGroup
	errors := make(chan error, 4)

	// Publish to GitHub
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := m.PublishGitHub(ctx, packageArtifacts, version, commit, githubToken); err != nil {
			errors <- fmt.Errorf("GitHub: %w", err)
		}
	}()

	// Publish to Homebrew
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := m.PublishHomebrew(ctx, packageArtifacts, version, githubToken); err != nil {
			errors <- fmt.Errorf("Homebrew: %w", err)
		}
	}()

	// Publish to npm
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := m.PublishNPM(ctx, packageArtifacts, version, npmToken); err != nil {
			errors <- fmt.Errorf("npm: %w", err)
		}
	}()

	// Publish to Docker
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := m.PublishDocker(ctx, buildArtifacts, version, dockerRegistry, dockerUsername, dockerToken); err != nil {
			errors <- fmt.Errorf("Docker: %w", err)
		}
	}()

	// Wait for all publishers
	wg.Wait()
	close(errors)

	// Collect errors
	var publishErrors []error
	for err := range errors {
		publishErrors = append(publishErrors, err)
	}

	if len(publishErrors) > 0 {
		fmt.Printf("\n❌ Release failed with %d error(s):\n", len(publishErrors))
		for _, err := range publishErrors {
			fmt.Printf("  - %v\n", err)
		}
		return fmt.Errorf("release failed")
	}

	fmt.Printf("\n✅ Release %s completed successfully!\n", version)
	return nil
}
