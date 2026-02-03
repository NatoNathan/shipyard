package main

import (
	"context"
	"dagger/shipyard/internal/dagger"
	"fmt"
	"strings"
)

// Test runs all tests with coverage
func (m *Shipyard) Test(
	ctx context.Context,
	// Source code directory
	source *dagger.Directory,
	// Show verbose test output
	// +optional
	// +default=false
	showOutput bool,
	// Enable race detection (requires CGO, uses larger image)
	// +optional
	// +default=false
	race bool,
) (string, error) {
	args := []string{"go", "test", "-coverprofile=coverage.out"}

	if race {
		args = append(args, "-race")
	}

	if showOutput {
		args = append(args, "-v")
	}

	args = append(args, "./...")

	container := m.goContainerWithCGO(source, race).
		WithExec(args)

	output, err := container.Stdout(ctx)
	if err != nil {
		return "", fmt.Errorf("tests failed: %w", err)
	}

	return output, nil
}

// Coverage runs tests and returns the coverage file
func (m *Shipyard) Coverage(
	ctx context.Context,
	// Source code directory
	source *dagger.Directory,
) *dagger.File {
	container := m.goContainer(source).
		WithExec([]string{"go", "test", "-coverprofile=coverage.out", "-covermode=atomic", "./..."})

	return container.File("/src/coverage.out")
}

// CoverageReport runs tests and returns the coverage percentage
func (m *Shipyard) CoverageReport(
	ctx context.Context,
	// Source code directory
	source *dagger.Directory,
	// Minimum coverage threshold (0-100)
	// +optional
	// +default=0
	threshold int,
) (string, error) {
	container := m.goContainer(source).
		WithExec([]string{"go", "test", "-coverprofile=coverage.out", "-covermode=atomic", "./..."}).
		WithExec([]string{"go", "tool", "cover", "-func=coverage.out"})

	output, err := container.Stdout(ctx)
	if err != nil {
		return "", fmt.Errorf("coverage failed: %w", err)
	}

	// Check threshold if specified
	if threshold > 0 {
		// Parse the total coverage from the last line
		lines := strings.Split(strings.TrimSpace(output), "\n")
		if len(lines) > 0 {
			lastLine := lines[len(lines)-1]
			// Format: "total:    (statements)    XX.X%"
			fields := strings.Fields(lastLine)
			if len(fields) >= 3 {
				percentStr := strings.TrimSuffix(fields[len(fields)-1], "%")
				var percent float64
				fmt.Sscanf(percentStr, "%f", &percent)
				if percent < float64(threshold) {
					return output, fmt.Errorf("coverage %.1f%% is below %d%% threshold", percent, threshold)
				}
			}
		}
	}

	return output, nil
}

// Lint runs golangci-lint on the source code
func (m *Shipyard) Lint(
	ctx context.Context,
	// Source code directory
	source *dagger.Directory,
	// Timeout for linting (e.g., "5m")
	// +optional
	// +default="5m"
	timeout string,
) (string, error) {
	container := m.goContainer(source).
		// Install golangci-lint v2
		WithExec([]string{"go", "install", "github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest"}).
		// Run linter
		WithExec([]string{"golangci-lint", "run", "--timeout=" + timeout})

	output, err := container.Stdout(ctx)
	if err != nil {
		return "", fmt.Errorf("linting failed: %w", err)
	}

	return output, nil
}

// CI runs the full CI pipeline: test, lint, and build
func (m *Shipyard) CI(
	ctx context.Context,
	// Source code directory
	source *dagger.Directory,
	// Enable race detection in tests
	// +optional
	// +default=false
	race bool,
) error {
	fmt.Println("🧪 Running tests...")
	testOutput, err := m.Test(ctx, source, true, race)
	if err != nil {
		return fmt.Errorf("test stage failed: %w", err)
	}
	fmt.Println(testOutput)

	fmt.Println("🔍 Running linter...")
	lintOutput, err := m.Lint(ctx, source, "5m")
	if err != nil {
		return fmt.Errorf("lint stage failed: %w", err)
	}
	fmt.Println(lintOutput)

	fmt.Println("🔨 Building...")
	_ = m.BuildOnly(ctx, source, "dev")

	fmt.Println("✅ CI passed!")
	return nil
}

// goContainer returns a Go container with source mounted and dependencies downloaded
func (m *Shipyard) goContainer(source *dagger.Directory) *dagger.Container {
	return dag.Container().
		From("golang:1.25-alpine").
		// Install git (needed for some go tools)
		WithExec([]string{"apk", "add", "--no-cache", "git"}).
		WithMountedDirectory("/src", source).
		WithWorkdir("/src").
		// Cache go modules
		WithMountedCache("/go/pkg/mod", dag.CacheVolume("go-mod")).
		WithMountedCache("/root/.cache/go-build", dag.CacheVolume("go-build")).
		// Download dependencies
		WithExec([]string{"go", "mod", "download"})
}

// goContainerWithCGO returns a Go container with optional CGO support for race detection
func (m *Shipyard) goContainerWithCGO(source *dagger.Directory, enableCGO bool) *dagger.Container {
	if !enableCGO {
		return m.goContainer(source)
	}

	// Use full golang image (not alpine) for CGO support
	return dag.Container().
		From("golang:1.25").
		WithMountedDirectory("/src", source).
		WithWorkdir("/src").
		// Enable CGO for race detection
		WithEnvVariable("CGO_ENABLED", "1").
		// Cache go modules
		WithMountedCache("/go/pkg/mod", dag.CacheVolume("go-mod")).
		WithMountedCache("/root/.cache/go-build", dag.CacheVolume("go-build")).
		// Download dependencies
		WithExec([]string{"go", "mod", "download"})
}
