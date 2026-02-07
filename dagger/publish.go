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
	// Source code directory
	source *dagger.Directory,
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

	// Container with GitHub CLI
	gh := dag.Container().
		From("alpine:latest").
		WithExec([]string{"apk", "add", "--no-cache", "github-cli"}).
		WithSecretVariable("GITHUB_TOKEN", githubToken).
		WithMountedDirectory("/artifacts", artifacts).
		WithWorkdir("/artifacts")

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
	notes, err := m.generateReleaseNotes(ctx, source, artifacts, version, commit)
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
	source *dagger.Directory,
	artifacts *dagger.Directory,
	version string,
	commit string,
) (string, error) {
	// Extract Linux AMD64 binary to run release-notes command
	tarball := fmt.Sprintf("shipyard_%s_linux_amd64.tar.gz", version)

	// Strip 'v' prefix from version for history lookup
	// History stores versions without 'v' prefix (e.g., "0.4.0" not "v0.4.0")
	versionWithoutPrefix := version
	if len(version) > 0 && version[0] == 'v' {
		versionWithoutPrefix = version[1:]
	}

	notes, err := dag.Container().
		From("alpine:latest").
		WithMountedDirectory("/artifacts", artifacts).
		WithMountedDirectory("/work", source).
		WithWorkdir("/work").
		WithExec([]string{"tar", "-xzf", fmt.Sprintf("/artifacts/%s", tarball)}).
		WithExec([]string{"./shipyard", "release-notes", "--version", versionWithoutPrefix}).
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
			"git clone https://x-access-token:${GITHUB_TOKEN}@github.com/" + repo + ".git .",
		}).
		WithNewFile("Formula/shipyard.rb", formula).
		WithExec([]string{"git", "config", "user.name", "github-actions[bot]"}).
		WithExec([]string{"git", "config", "user.email", "github-actions[bot]@users.noreply.github.com"}).
		WithExec([]string{"git", "add", "Formula/shipyard.rb"}).
		WithExec([]string{"git", "commit", "-m", fmt.Sprintf("Update shipyard to %s", version)}).
		WithExec([]string{
			"sh", "-c",
			"git push https://x-access-token:${GITHUB_TOKEN}@github.com/" + repo + ".git main",
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
    "tar": "^7.5.0",
    "adm-zip": "^0.5.10"
  },
  "engines": {
    "node": ">=18"
  }
}`, versionNum)

	installScript := m.generateInstallScript(version)
	readme := m.generateNPMReadme(version)

	binWrapper := `#!/usr/bin/env node
const { spawnSync } = require('child_process');
const path = require('path');
const fs = require('fs');

const binDir = path.join(__dirname, '..', 'bin');
const binName = process.platform === 'win32' ? 'shipyard.exe' : 'shipyard';
const binPath = path.join(binDir, binName);
const realBinary = path.join(binDir, '.shipyard-binary');

// Check if real binary exists
if (!fs.existsSync(realBinary)) {
  console.error('Shipyard binary not found. Installing...');
  const installScript = path.join(__dirname, '..', 'install.js');
  const result = spawnSync('node', [installScript], { stdio: 'inherit' });
  if (result.status !== 0) {
    console.error('Failed to install shipyard binary');
    process.exit(1);
  }
}

// Execute the real binary
const result = spawnSync(realBinary, process.argv.slice(2), { stdio: 'inherit' });
process.exit(result.status || 0);
`

	return dag.Directory().
		WithNewFile("package.json", packageJSON).
		WithNewFile("install.js", installScript).
		WithNewFile("README.md", readme).
		WithNewFile("bin/shipyard", binWrapper)
}

// generateInstallScript creates the npm postinstall script
func (m *Shipyard) generateInstallScript(version string) string {
	return fmt.Sprintf(`const fs = require('fs');
const path = require('path');
const tar = require('tar');
const AdmZip = require('adm-zip');

const version = '%s';
const binDir = path.join(__dirname, 'bin');
const binName = '.shipyard-binary' + (process.platform === 'win32' ? '.exe' : '');
const binPath = path.join(binDir, binName);

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

(async () => {
  const res = await fetch(url);
  if (!res.ok) {
    console.error('Failed to download:', res.status, res.statusText);
    process.exit(1);
  }

  const buffer = Buffer.from(await res.arrayBuffer());

  // Ensure bin directory exists
  if (!fs.existsSync(binDir)) {
    fs.mkdirSync(binDir, { recursive: true });
  }

  // Extract based on format
  const extractedName = process.platform === 'win32' ? 'shipyard.exe' : 'shipyard';
  const extractedPath = path.join(binDir, extractedName);

  if (ext === 'zip') {
    const zip = new AdmZip(buffer);
    zip.extractAllTo(binDir, true);
  } else {
    const tmpPath = path.join(binDir, filename);
    fs.writeFileSync(tmpPath, buffer);
    tar.x({ file: tmpPath, cwd: binDir, sync: true });
    fs.unlinkSync(tmpPath);
  }

  // Rename extracted binary to hidden name
  if (fs.existsSync(extractedPath)) {
    if (fs.existsSync(binPath)) {
      fs.unlinkSync(binPath);
    }
    fs.renameSync(extractedPath, binPath);
  }

  // Make executable
  if (process.platform !== 'win32') {
    fs.chmodSync(binPath, 0o755);
  }

  console.log('✓ Shipyard installed successfully');
})().catch((err) => {
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

// generateDockerTags creates all version tags for a given version string
// For stable versions (e.g., "v1.2.3"), generates:
//   - latest, v1.2.3, 1.2.3, v1.2, 1.2, v1, 1
//
// For pre-releases (e.g., "v1.2.3-beta.1"), generates:
//   - v1.2.3-beta.1, 1.2.3-beta.1
func generateDockerTags(version string) []string {
	versionNum := strings.TrimPrefix(version, "v")

	// Check for pre-release suffix (contains hyphen)
	// Examples: "1.2.3-beta.1", "1.2.3-rc.2", "1.2.3-alpha"
	isPreRelease := strings.Contains(versionNum, "-")

	if isPreRelease {
		// Pre-release: only exact version tags
		return []string{
			version,    // "v1.2.3-beta.1"
			versionNum, // "1.2.3-beta.1"
		}
	}

	// Stable release: full tag set
	parts := strings.Split(versionNum, ".")
	if len(parts) < 2 {
		// Invalid version, return basic tags
		return []string{"latest", version, versionNum}
	}

	major := parts[0]
	minor := parts[1]

	return []string{
		"latest",
		version,                             // v1.2.3
		versionNum,                          // 1.2.3
		fmt.Sprintf("v%s.%s", major, minor), // v1.2
		fmt.Sprintf("%s.%s", major, minor),  // 1.2
		fmt.Sprintf("v%s", major),           // v1
		major,                               // 1
	}
}

// buildDockerImageWithPlatform creates a Docker image for a specific platform
// with proper platform metadata for multi-arch manifest creation
func (m *Shipyard) buildDockerImageWithPlatform(
	buildArtifacts *dagger.Directory,
	version string,
	platform dagger.Platform, // "linux/amd64" or "linux/arm64"
) *dagger.Container {
	// Convert platform to os_arch format for binary path
	// "linux/amd64" -> "linux_amd64"
	dirname := strings.Replace(string(platform), "/", "_", 1)
	binaryPath := fmt.Sprintf("%s/shipyard", dirname)

	versionNum := strings.TrimPrefix(version, "v")

	return dag.Container(dagger.ContainerOpts{Platform: platform}).
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

// buildPlatformVariants creates platform-specific container images for multi-arch publishing
func (m *Shipyard) buildPlatformVariants(
	buildArtifacts *dagger.Directory,
	version string,
) []*dagger.Container {
	platforms := []dagger.Platform{
		"linux/amd64",
		"linux/arm64",
	}

	variants := make([]*dagger.Container, len(platforms))
	for i, platform := range platforms {
		variants[i] = m.buildDockerImageWithPlatform(buildArtifacts, version, platform)
	}

	return variants
}

// PublishDocker builds and pushes multi-arch Docker images to registry
func (m *Shipyard) PublishDocker(
	ctx context.Context,
	// Build artifacts directory (not packaged)
	buildArtifacts *dagger.Directory,
	// Version string (e.g., "v1.2.3")
	version string,
	// Docker registry (e.g., "ghcr.io/natonathan/shipyard")
	dockerRegistry string,
	// Docker registry username (GitHub actor for GHCR)
	dockerUsername string,
	// Docker registry token
	dockerToken *dagger.Secret,
) error {
	fmt.Printf("Building multi-arch Docker images...\n")

	// Build platform variants once (reused for all tags)
	platformVariants := m.buildPlatformVariants(buildArtifacts, version)

	// Generate all version tags
	tags := generateDockerTags(version)

	// Extract registry domain for authentication
	registry := strings.Split(dockerRegistry, "/")[0]

	// Publish each tag with platform variants
	fmt.Printf("Publishing %d tags to %s...\n", len(tags), dockerRegistry)

	for _, tag := range tags {
		imageRef := fmt.Sprintf("%s:%s", dockerRegistry, tag)
		fmt.Printf("  → %s (linux/amd64, linux/arm64)\n", imageRef)

		// Publish multi-arch manifest using platform variants
		_, err := platformVariants[0].
			WithRegistryAuth(registry, dockerUsername, dockerToken).
			Publish(ctx, imageRef, dagger.ContainerPublishOpts{
				PlatformVariants: platformVariants,
			})

		if err != nil {
			return fmt.Errorf("failed to publish tag %s: %w", tag, err)
		}
	}

	fmt.Printf("✓ Docker images published to %s\n", dockerRegistry)
	fmt.Printf("  Tags: %s\n", strings.Join(tags, ", "))
	return nil
}
