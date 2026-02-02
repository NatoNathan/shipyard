package main

import (
	"context"
	"crypto/sha256"
	"dagger/shipyard/internal/dagger"
	"fmt"
	"io"
)

// Package creates distribution archives and checksums
func (m *Shipyard) Package(
	ctx context.Context,
	// Build artifacts directory
	buildArtifacts *dagger.Directory,
	// Version string (e.g., "v1.2.3")
	version string,
) *dagger.Directory {
	output := dag.Directory()
	checksums := ""

	// Create archive for each platform
	for _, platform := range SupportedPlatforms {
		archive := m.createArchive(ctx, buildArtifacts, platform, version)
		filename := archiveFilename(platform, version)

		// Add archive to output
		output = output.WithFile(filename, archive)

		// Calculate checksum
		checksum, _ := calculateChecksum(ctx, archive, filename)
		checksums += checksum + "\n"
	}

	// Add checksums file
	output = output.WithNewFile("checksums.txt", checksums)

	return output
}

// createArchive creates a tar.gz or zip archive for a platform
func (m *Shipyard) createArchive(
	ctx context.Context,
	buildArtifacts *dagger.Directory,
	platform Platform,
	version string,
) *dagger.File {
	dirname := fmt.Sprintf("%s_%s", platform.OS, platform.Arch)
	binaryName := "shipyard"
	if platform.OS == "windows" {
		binaryName += ".exe"
	}

	// Get the binary file
	binaryPath := fmt.Sprintf("%s/%s", dirname, binaryName)

	// Use alpine container with tar/gzip/zip
	archiver := dag.Container().
		From("alpine:latest").
		WithExec([]string{"apk", "add", "--no-cache", "tar", "gzip", "zip"}).
		WithMountedDirectory("/artifacts", buildArtifacts).
		WithWorkdir("/work")

	filename := archiveFilename(platform, version)

	if platform.OS == "windows" {
		// Create zip for Windows
		archiver = archiver.
			WithExec([]string{
				"zip", "-j", filename,
				fmt.Sprintf("/artifacts/%s", binaryPath),
			})
	} else {
		// Create tar.gz for Unix/macOS
		archiver = archiver.
			WithExec([]string{
				"tar", "-czf", filename,
				"-C", fmt.Sprintf("/artifacts/%s", dirname),
				binaryName,
			})
	}

	return archiver.File(filename)
}

// archiveFilename generates the archive filename for a platform
func archiveFilename(platform Platform, version string) string {
	ext := "tar.gz"
	if platform.OS == "windows" {
		ext = "zip"
	}
	return fmt.Sprintf("shipyard_%s_%s_%s.%s", version, platform.OS, platform.Arch, ext)
}

// calculateChecksum computes SHA256 checksum for a file
func calculateChecksum(ctx context.Context, file *dagger.File, filename string) (string, error) {
	// Read file contents
	contents, err := file.Contents(ctx)
	if err != nil {
		return "", err
	}

	// Calculate SHA256
	hash := sha256.New()
	if _, err := io.WriteString(hash, contents); err != nil {
		return "", err
	}

	// Format as "checksum  filename"
	return fmt.Sprintf("%x  %s", hash.Sum(nil), filename), nil
}

// PackageOnly is a convenience function for testing package stage in isolation
func (m *Shipyard) PackageOnly(
	ctx context.Context,
	// Source code directory
	source *dagger.Directory,
	// Version string (e.g., "v1.2.3")
	version string,
) *dagger.Directory {
	buildArtifacts := m.BuildOnly(ctx, source, version)
	return m.Package(ctx, buildArtifacts, version)
}
