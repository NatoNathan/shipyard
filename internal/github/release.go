package github

import (
	"context"
	"fmt"
	"strings"

	"github.com/NatoNathan/shipyard/internal/config"
	"github.com/NatoNathan/shipyard/internal/git"
	"github.com/NatoNathan/shipyard/internal/upgrade"
	"github.com/NatoNathan/shipyard/pkg/semver"
)

// ReleaseClient creates releases in a remote release service.
type ReleaseClient interface {
	CreateRelease(ctx context.Context, owner, repo string, release *upgrade.CreateReleaseRequest) (*upgrade.ReleaseInfo, error)
}

// ReleasePublisher handles publishing releases to GitHub
type ReleasePublisher struct {
	client    ReleaseClient
	repoPath  string
	config    *config.Config
	tagExists func(tagName string) error
	tagPushed func(tagName string) error
}

// NewReleasePublisher creates a new release publisher
func NewReleasePublisher(repoPath string, cfg *config.Config) *ReleasePublisher {
	return NewReleasePublisherWithClient(repoPath, cfg, upgrade.NewGitHubClient())
}

// NewReleasePublisherWithClient creates a release publisher with an injected client.
func NewReleasePublisherWithClient(repoPath string, cfg *config.Config, client ReleaseClient) *ReleasePublisher {
	publisher := &ReleasePublisher{
		client:   client,
		repoPath: repoPath,
		config:   cfg,
	}
	publisher.tagExists = publisher.verifyTagExists
	publisher.tagPushed = publisher.verifyTagPushed
	return publisher
}

// PublishRelease creates a GitHub release for a package version
func (p *ReleasePublisher) PublishRelease(
	ctx context.Context,
	packageName string,
	version semver.Version,
	tagName string,
	releaseNotes string,
	draft bool,
	prerelease bool,
) error {
	// Validate GitHub config
	if p.config.GitHub.Owner == "" || p.config.GitHub.Repo == "" {
		return fmt.Errorf("GitHub not configured in .shipyard.yaml (owner and repo required)")
	}

	// Verify tag exists locally
	if err := p.tagExists(tagName); err != nil {
		return fmt.Errorf("run `shipyard version` first to create version tag %s: %w", tagName, err)
	}

	// Verify tag pushed to remote
	if err := p.tagPushed(tagName); err != nil {
		return fmt.Errorf("push tags first: `git push --tags`: %w", err)
	}

	// Extract release title from first line of release notes
	title := extractTitleFromNotes(releaseNotes, packageName, version)

	// Create release request
	releaseReq := &upgrade.CreateReleaseRequest{
		TagName:    tagName,
		Name:       title,
		Body:       releaseNotes,
		Draft:      draft,
		Prerelease: prerelease,
	}

	// Call GitHub API
	_, err := p.client.CreateRelease(ctx, p.config.GitHub.Owner, p.config.GitHub.Repo, releaseReq)
	if err != nil {
		return fmt.Errorf("failed to create GitHub release: %w", err)
	}

	return nil
}

// verifyTagExists checks if the tag exists locally
func (p *ReleasePublisher) verifyTagExists(tagName string) error {
	exists, err := git.VerifyTagExists(p.repoPath, tagName)
	if err != nil {
		return fmt.Errorf("failed to check local tags: %w", err)
	}

	if !exists {
		return fmt.Errorf("tag %s does not exist locally", tagName)
	}

	return nil
}

// verifyTagPushed checks if the tag has been pushed to the remote
func (p *ReleasePublisher) verifyTagPushed(tagName string) error {
	pushed, err := git.VerifyTagPushedToRemote(p.repoPath, "origin", tagName)
	if err != nil {
		return fmt.Errorf("failed to check remote tags: %w", err)
	}

	if !pushed {
		return fmt.Errorf("tag %s not found on remote", tagName)
	}

	return nil
}

// extractTitleFromNotes extracts the first line from release notes as title
// Falls back to "{package} v{version}" if empty or starts with markdown heading
func extractTitleFromNotes(releaseNotes string, packageName string, version semver.Version) string {
	lines := strings.Split(releaseNotes, "\n")
	if len(lines) == 0 {
		return fmt.Sprintf("%s v%s", packageName, version)
	}

	title := strings.TrimSpace(lines[0])

	// Fall back if empty or is a markdown heading
	if title == "" || strings.HasPrefix(title, "#") {
		return fmt.Sprintf("%s v%s", packageName, version)
	}

	return title
}
