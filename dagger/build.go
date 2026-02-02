package main

import (
	"context"
	"dagger/shipyard/internal/dagger"
	"fmt"
	"time"
)

// Build compiles the Shipyard binary for all supported platforms
func (m *Shipyard) Build(
	ctx context.Context,
	// Source code directory
	source *dagger.Directory,
	// Version string (e.g., "v1.2.3")
	version string,
	// Git commit SHA
	commit string,
) *dagger.Directory {
	buildInfo := BuildInfo{
		Version: version,
		Commit:  commit,
		Date:    time.Now().Format(time.RFC3339),
	}

	// Create output directory
	output := dag.Directory()

	// Build for each platform (Dagger parallelizes automatically)
	for _, platform := range SupportedPlatforms {
		binary := m.buildPlatform(ctx, source, platform, buildInfo)

		// Place binary in platform-specific subdirectory
		dirname := fmt.Sprintf("%s_%s", platform.OS, platform.Arch)
		filename := "shipyard"
		if platform.OS == "windows" {
			filename += ".exe"
		}

		output = output.WithFile(fmt.Sprintf("%s/%s", dirname, filename), binary)
	}

	return output
}

// buildPlatform compiles a single platform binary
func (m *Shipyard) buildPlatform(
	ctx context.Context,
	source *dagger.Directory,
	platform Platform,
	buildInfo BuildInfo,
) *dagger.File {
	// Use Go 1.25 alpine image for building
	builder := dag.Container().
		From("golang:1.25-alpine").
		WithMountedDirectory("/src", source).
		WithWorkdir("/src").
		// Set build environment
		WithEnvVariable("GOOS", platform.OS).
		WithEnvVariable("GOARCH", platform.Arch).
		WithEnvVariable("CGO_ENABLED", "0")

	// Build ldflags for version info
	ldflags := fmt.Sprintf(
		"-s -w -X main.version=%s -X main.commit=%s -X main.date=%s",
		buildInfo.Version,
		buildInfo.Commit,
		buildInfo.Date,
	)

	// Output binary name
	outputName := "shipyard"
	if platform.OS == "windows" {
		outputName += ".exe"
	}

	// Execute build
	builder = builder.WithExec([]string{
		"go", "build",
		"-ldflags", ldflags,
		"-o", outputName,
		"./cmd/shipyard",
	})

	return builder.File(outputName)
}

// BuildOnly is a convenience function for testing build stage in isolation
func (m *Shipyard) BuildOnly(
	ctx context.Context,
	// Source code directory
	source *dagger.Directory,
	// Version string (e.g., "v1.2.3")
	version string,
) *dagger.Directory {
	return m.Build(ctx, source, version, "dev")
}
