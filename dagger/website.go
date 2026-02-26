package main

import (
	"context"
	"dagger/shipyard/internal/dagger"
)

// BuildWebsite copies the ./docs directory into the website's content tree,
// installs dependencies with bun, and runs the Astro static build.
//
// Returns the built dist/ directory, ready for deployment to Cloudflare Pages
// or any static host.
//
// The ./docs content is merged in-memory before the build container starts,
// so no filesystem mutations occur on the host.
func (m *Shipyard) BuildWebsite(
	ctx context.Context,
	// Source directory (repository root)
	source *dagger.Directory,
) *dagger.Directory {
	// Pull docs and website source from the repo
	docs := source.Directory("docs")
	websiteSrc := source.Directory("website")

	// Overlay docs into the content directory.
	// This mirrors ./docs → ./website/src/content/docs at build time
	// without requiring them to be committed there.
	websiteWithDocs := websiteSrc.WithDirectory("src/content/docs", docs)

	return dag.Container().
		From("oven/bun:1-alpine").
		// git is needed by some Astro plugins that read git metadata
		WithExec([]string{"apk", "add", "--no-cache", "git"}).
		WithMountedDirectory("/website", websiteWithDocs).
		WithWorkdir("/website").
		// Cache bun's install cache across runs for faster subsequent builds
		WithMountedCache("/root/.bun/install/cache", dag.CacheVolume("bun-install-cache")).
		WithExec([]string{"bun", "install", "--frozen-lockfile"}).
		WithExec([]string{"bun", "run", "build"}).
		Directory("/website/dist")
}
