package main

import (
	"context"
	"dagger/shipyard/internal/dagger"
	"fmt"
	"strings"
)

// PublishGitHub creates and publishes a GitHub release
func (m *Shipyard) PublishGitHub(
	ctx context.Context,
	// Package artifacts directory
	artifacts *dagger.Directory,
	// Version string (e.g., "v1.2.3")
	version string,
	// Git commit SHA
	commit string,
	// GitHub token
	githubToken *dagger.Secret,
) error {
	repo := "NatoNathan/shipyard"

	// Container with GitHub CLI (installed from Alpine packages to avoid ghcr.io pull issues)
	gh := dag.Container().
		From("alpine:latest").
		WithExec([]string{"apk", "add", "--no-cache", "--repository=https://dl-cdn.alpinelinux.org/alpine/edge/community", "github-cli"}).
		WithSecretVariable("GITHUB_TOKEN", githubToken).
		WithMountedDirectory("/artifacts", artifacts).
		WithWorkdir("/artifacts")

	// Delete existing release if present (idempotent for re-runs)
	fmt.Printf("Cleaning up any existing release %s...\n", version)
	gh = gh.WithExec([]string{
		"sh", "-c",
		"gh release delete " + version + " --repo " + repo + " --yes 2>/dev/null || true",
	})

	// Create draft release with all artifacts
	fmt.Printf("Creating GitHub release %s...\n", version)
	_, err := gh.
		WithExec([]string{
			"gh", "release", "create", version,
			"--repo", repo,
			"--draft",
			"--title", version,
			"--notes", "Release notes will be generated...",
		}).
		WithExec([]string{
			"sh", "-c",
			"gh release upload " + version + " * --repo " + repo,
		}).
		Sync(ctx)

	if err != nil {
		return fmt.Errorf("failed to create draft release: %w", err)
	}

	// Generate release notes using Shipyard
	// TODO: In the future, import shipyard as Go module instead of building
	fmt.Printf("Generating release notes...\n")
	notes, err := m.generateReleaseNotes(ctx, artifacts, version, commit)
	if err != nil {
		fmt.Printf("Warning: failed to generate release notes: %v\n", err)
		notes = fmt.Sprintf("Release %s\n\nSee commit %s for changes.", version, commit)
	}

	// Publish the release with notes
	fmt.Printf("Publishing release...\n")
	_, err = gh.
		WithExec([]string{
			"gh", "release", "edit", version,
			"--repo", repo,
			"--draft=false",
			"--notes", notes,
		}).
		Sync(ctx)

	if err != nil {
		return fmt.Errorf("failed to publish release: %w", err)
	}

	fmt.Printf("✓ GitHub release published: %s\n", version)
	return nil
}

// generateReleaseNotes generates release notes using Shipyard's release-notes command
func (m *Shipyard) generateReleaseNotes(
	ctx context.Context,
	artifacts *dagger.Directory,
	version string,
	commit string,
) (string, error) {
	// Extract Linux AMD64 binary to run release-notes command
	tarball := fmt.Sprintf("shipyard_%s_linux_amd64.tar.gz", version)

	notes, err := dag.Container().
		From("alpine:latest").
		WithMountedDirectory("/artifacts", artifacts).
		WithWorkdir("/work").
		WithExec([]string{"tar", "-xzf", fmt.Sprintf("/artifacts/%s", tarball)}).
		WithExec([]string{"./shipyard", "release-notes", "--version", version}).
		Stdout(ctx)

	if err != nil {
		return "", err
	}

	return notes, nil
}

// PublishHomebrew updates the Homebrew tap with new formula
func (m *Shipyard) PublishHomebrew(
	ctx context.Context,
	// Package artifacts directory
	artifacts *dagger.Directory,
	// Version string (e.g., "v1.2.3")
	version string,
	// GitHub token
	githubToken *dagger.Secret,
) error {
	repo := "natonathan/homebrew-tap"

	// Extract checksums
	checksums, err := m.extractChecksums(ctx, artifacts)
	if err != nil {
		return fmt.Errorf("failed to extract checksums: %w", err)
	}

	// Generate Homebrew formula
	formula := m.generateFormula(version, checksums)

	// Update tap repository
	fmt.Printf("Updating Homebrew tap...\n")
	git := dag.Container().
		From("alpine/git:latest").
		WithExec([]string{"apk", "add", "--no-cache", "ruby"}).
		WithSecretVariable("GITHUB_TOKEN", githubToken).
		WithWorkdir("/work").
		WithExec([]string{
			"sh", "-c",
			fmt.Sprintf("git clone https://x-access-token:$GITHUB_TOKEN@github.com/%s.git .", repo),
		}).
		WithNewFile("Formula/shipyard.rb", formula).
		WithExec([]string{"git", "config", "user.name", "github-actions[bot]"}).
		WithExec([]string{"git", "config", "user.email", "github-actions[bot]@users.noreply.github.com"}).
		WithExec([]string{"git", "add", "Formula/shipyard.rb"}).
		WithExec([]string{"git", "commit", "-m", fmt.Sprintf("Update shipyard to %s", version)}).
		WithExec([]string{
			"sh", "-c",
			"git push https://x-access-token:$GITHUB_TOKEN@github.com/" + repo + ".git main",
		})

	_, err = git.Sync(ctx)
	if err != nil {
		return fmt.Errorf("failed to update Homebrew tap: %w", err)
	}

	fmt.Printf("✓ Homebrew tap updated: %s\n", version)
	return nil
}

// extractChecksums reads the checksums.txt file and parses it
func (m *Shipyard) extractChecksums(ctx context.Context, artifacts *dagger.Directory) (map[string]string, error) {
	checksumFile := artifacts.File("checksums.txt")
	content, err := checksumFile.Contents(ctx)
	if err != nil {
		return nil, err
	}

	checksums := make(map[string]string)
	lines := strings.Split(strings.TrimSpace(content), "\n")
	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) == 2 {
			checksums[parts[1]] = parts[0]
		}
	}

	return checksums, nil
}

// generateFormula creates the Homebrew formula Ruby file
func (m *Shipyard) generateFormula(version string, checksums map[string]string) string {
	versionNum := strings.TrimPrefix(version, "v")
	baseURL := fmt.Sprintf("https://github.com/NatoNathan/shipyard/releases/download/%s", version)

	darwinAMD64 := fmt.Sprintf("shipyard_%s_darwin_amd64.tar.gz", version)
	darwinARM64 := fmt.Sprintf("shipyard_%s_darwin_arm64.tar.gz", version)
	linuxAMD64 := fmt.Sprintf("shipyard_%s_linux_amd64.tar.gz", version)
	linuxARM64 := fmt.Sprintf("shipyard_%s_linux_arm64.tar.gz", version)

	return fmt.Sprintf(`class Shipyard < Formula
  desc "CLI tool for managing project workflows with a nautical theme"
  homepage "https://github.com/NatoNathan/shipyard"
  version "%s"

  if OS.mac? && Hardware::CPU.intel?
    url "%s/%s"
    sha256 "%s"
  elsif OS.mac? && Hardware::CPU.arm?
    url "%s/%s"
    sha256 "%s"
  elsif OS.linux? && Hardware::CPU.intel?
    url "%s/%s"
    sha256 "%s"
  elsif OS.linux? && Hardware::CPU.arm?
    url "%s/%s"
    sha256 "%s"
  end

  def install
    bin.install "shipyard"
    generate_completions_from_executable(bin/"shipyard", "completion")
  end

  test do
    assert_match "shipyard version %s", shell_output("#{bin}/shipyard version")
  end
end
`,
		versionNum,
		baseURL, darwinAMD64, checksums[darwinAMD64],
		baseURL, darwinARM64, checksums[darwinARM64],
		baseURL, linuxAMD64, checksums[linuxAMD64],
		baseURL, linuxARM64, checksums[linuxARM64],
		versionNum,
	)
}

// PublishNPM publishes the npm wrapper package
func (m *Shipyard) PublishNPM(
	ctx context.Context,
	// Package artifacts directory
	artifacts *dagger.Directory,
	// Version string (e.g., "v1.2.3")
	version string,
	// npm auth token
	npmToken *dagger.Secret,
) error {
	versionNum := strings.TrimPrefix(version, "v")

	// Create npm package structure
	fmt.Printf("Creating npm package...\n")
	npmPackage := m.createNPMPackage(ctx, version)

	// Publish to npm
	fmt.Printf("Publishing to npm...\n")
	publisher := dag.Container().
		From("node:20-alpine").
		WithSecretVariable("NPM_TOKEN", npmToken).
		WithMountedDirectory("/package", npmPackage).
		WithWorkdir("/package").
		WithExec([]string{
			"sh", "-c",
			"echo '//registry.npmjs.org/:_authToken=${NPM_TOKEN}' > .npmrc",
		}).
		WithExec([]string{"npm", "publish", "--access", "public"})

	_, err := publisher.Sync(ctx)
	if err != nil {
		return fmt.Errorf("failed to publish npm package: %w", err)
	}

	fmt.Printf("✓ npm package published: shipyard-cli@%s\n", versionNum)
	return nil
}

// createNPMPackage creates the npm package directory structure
func (m *Shipyard) createNPMPackage(ctx context.Context, version string) *dagger.Directory {
	versionNum := strings.TrimPrefix(version, "v")

	packageJSON := fmt.Sprintf(`{
  "name": "shipyard-cli",
  "version": "%s",
  "description": "CLI tool for managing project workflows with a nautical theme",
  "bin": {
    "shipyard": "bin/shipyard"
  },
  "scripts": {
    "postinstall": "node install.js"
  },
  "keywords": ["shipyard", "cli", "workflow", "project-management"],
  "author": "NatoNathan",
  "license": "MIT",
  "repository": {
    "type": "git",
    "url": "https://github.com/NatoNathan/shipyard.git"
  },
  "homepage": "https://github.com/NatoNathan/shipyard",
  "dependencies": {
    "tar": "^6.2.0",
    "adm-zip": "^0.5.10"
  }
}`, versionNum)

	installScript := m.generateInstallScript(version)
	readme := m.generateNPMReadme(version)

	return dag.Directory().
		WithNewFile("package.json", packageJSON).
		WithNewFile("install.js", installScript).
		WithNewFile("README.md", readme).
		WithNewFile("bin/shipyard", "#!/bin/sh\n# Placeholder - real binary installed by postinstall\n")
}

// generateInstallScript creates the npm postinstall script
func (m *Shipyard) generateInstallScript(version string) string {
	return fmt.Sprintf(`const fs = require('fs');
const path = require('path');
const https = require('https');
const tar = require('tar');
const AdmZip = require('adm-zip');

const version = '%s';
const binDir = path.join(__dirname, 'bin');
const binPath = path.join(binDir, process.platform === 'win32' ? 'shipyard.exe' : 'shipyard');

// Platform mapping
const platformMap = {
  'darwin': { 'x64': 'darwin_amd64', 'arm64': 'darwin_arm64' },
  'linux': { 'x64': 'linux_amd64', 'arm64': 'linux_arm64' },
  'win32': { 'x64': 'windows_amd64', 'arm64': 'windows_amd64' }
};

const platform = platformMap[process.platform]?.[process.arch];
if (!platform) {
  console.error('Unsupported platform:', process.platform, process.arch);
  process.exit(1);
}

const ext = process.platform === 'win32' ? 'zip' : 'tar.gz';
const filename = 'shipyard_' + version + '_' + platform.replace('_', '_') + '.' + ext;
const url = 'https://github.com/NatoNathan/shipyard/releases/download/' + version + '/' + filename;

console.log('Downloading shipyard binary for', platform);

https.get(url, (res) => {
  if (res.statusCode !== 200) {
    console.error('Failed to download:', res.statusCode);
    process.exit(1);
  }

  const chunks = [];
  res.on('data', (chunk) => chunks.push(chunk));
  res.on('end', () => {
    const buffer = Buffer.concat(chunks);

    // Ensure bin directory exists
    if (!fs.existsSync(binDir)) {
      fs.mkdirSync(binDir, { recursive: true });
    }

    // Extract based on format
    if (ext === 'zip') {
      const zip = new AdmZip(buffer);
      zip.extractAllTo(binDir, true);
    } else {
      const tmpPath = path.join(binDir, filename);
      fs.writeFileSync(tmpPath, buffer);
      tar.x({ file: tmpPath, cwd: binDir, sync: true });
      fs.unlinkSync(tmpPath);
    }

    // Make executable
    if (process.platform !== 'win32') {
      fs.chmodSync(binPath, 0o755);
    }

    console.log('✓ Shipyard installed successfully');
  });
}).on('error', (err) => {
  console.error('Download failed:', err.message);
  process.exit(1);
});
`, version)
}

// generateNPMReadme creates the npm package README
func (m *Shipyard) generateNPMReadme(version string) string {
	return fmt.Sprintf(`# shipyard-cli

CLI tool for managing project workflows with a nautical theme.

Version: %s

## Installation

'''bash
npm install -g shipyard-cli
'''

Or use with npx:

'''bash
npx shipyard-cli [command]
'''

## Usage

'''bash
shipyard --help
'''

## Links

- [GitHub Repository](https://github.com/NatoNathan/shipyard)
- [Documentation](https://github.com/NatoNathan/shipyard#readme)

## License

MIT
`, strings.TrimPrefix(version, "v"))
}

// PublishDocker builds and pushes multi-arch Docker images
func (m *Shipyard) PublishDocker(
	ctx context.Context,
	// Build artifacts directory (not packaged)
	buildArtifacts *dagger.Directory,
	// Version string (e.g., "v1.2.3")
	version string,
	// Docker registry (e.g., "ghcr.io/natonathan/shipyard")
	dockerRegistry string,
	// Docker registry token
	dockerToken *dagger.Secret,
) error {
	versionNum := strings.TrimPrefix(version, "v")
	tags := []string{"latest", version, versionNum}

	// Build multi-arch image
	fmt.Printf("Building Docker images...\n")

	// Build for linux/amd64
	amd64Image := m.buildDockerImage(ctx, buildArtifacts, version, "linux", "amd64")

	// Build for linux/arm64
	arm64Image := m.buildDockerImage(ctx, buildArtifacts, version, "linux", "arm64")

	// Push all tags
	registry := strings.Split(dockerRegistry, "/")[0]
	username := strings.Split(dockerRegistry, "/")[1]

	for _, tag := range tags {
		imageRef := fmt.Sprintf("%s:%s", dockerRegistry, tag)
		fmt.Printf("Pushing %s...\n", imageRef)

		// Push AMD64 variant
		_, err := amd64Image.
			WithRegistryAuth(registry, username, dockerToken).
			Publish(ctx, imageRef+"-amd64")
		if err != nil {
			return fmt.Errorf("failed to push amd64 image: %w", err)
		}

		// Push ARM64 variant
		_, err = arm64Image.
			WithRegistryAuth(registry, username, dockerToken).
			Publish(ctx, imageRef+"-arm64")
		if err != nil {
			return fmt.Errorf("failed to push arm64 image: %w", err)
		}

		// Create and push manifest list
		// Note: Dagger handles manifest creation automatically
	}

	fmt.Printf("✓ Docker images published to %s\n", dockerRegistry)
	return nil
}

// buildDockerImage creates a Docker image for a specific platform
func (m *Shipyard) buildDockerImage(
	ctx context.Context,
	buildArtifacts *dagger.Directory,
	version string,
	os string,
	arch string,
) *dagger.Container {
	dirname := fmt.Sprintf("%s_%s", os, arch)
	binaryPath := fmt.Sprintf("%s/shipyard", dirname)

	versionNum := strings.TrimPrefix(version, "v")

	return dag.Container().
		From("alpine:latest").
		WithFile("/usr/local/bin/shipyard", buildArtifacts.File(binaryPath)).
		WithExec([]string{"chmod", "+x", "/usr/local/bin/shipyard"}).
		WithEntrypoint([]string{"shipyard"}).
		WithLabel("org.opencontainers.image.title", "Shipyard").
		WithLabel("org.opencontainers.image.description", "CLI tool for managing project workflows").
		WithLabel("org.opencontainers.image.version", versionNum).
		WithLabel("org.opencontainers.image.source", "https://github.com/NatoNathan/shipyard").
		WithLabel("org.opencontainers.image.licenses", "MIT")
}
