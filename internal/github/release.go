package github

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/NatoNathan/shipyard/internal/config"
	"github.com/NatoNathan/shipyard/internal/upgrade"
	"github.com/NatoNathan/shipyard/pkg/semver"
)

// ReleasePublisher handles publishing releases to GitHub
type ReleasePublisher struct {
	client   *upgrade.GitHubClient
	repoPath string
	config   *config.Config
}

// NewReleasePublisher creates a new release publisher
func NewReleasePublisher(repoPath string, cfg *config.Config) *ReleasePublisher {
	return &ReleasePublisher{
		client:   upgrade.NewGitHubClient(),
		repoPath: repoPath,
		config:   cfg,
	}
}

// PublishRelease creates a GitHub release for a package version
func (p *ReleasePublisher) PublishRelease(
	ctx context.Context,
	packageName string,
	version semver.Version,
	tagName string,
	releaseNotes string,
	draft bool,
) error {
	// Validate GitHub config
	if p.config.GitHub.Owner == "" || p.config.GitHub.Repo == "" {
		return fmt.Errorf("GitHub not configured in .shipyard.yaml (owner and repo required)")
	}

	// Verify tag exists locally
	if err := p.verifyTagExists(tagName); err != nil {
		return fmt.Errorf("run `shipyard version` first to create version tag %s: %w", tagName, err)
	}

	// Verify tag pushed to remote
	if err := p.verifyTagPushed(tagName); err != nil {
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
		Prerelease: false, // TODO: Make this configurable
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
	cmd := exec.Command("git", "tag", "-l", tagName)
	cmd.Dir = p.repoPath
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to check local tags: %w", err)
	}

	if strings.TrimSpace(string(output)) == "" {
		return fmt.Errorf("tag %s does not exist locally", tagName)
	}

	return nil
}

// verifyTagPushed checks if the tag has been pushed to the remote
func (p *ReleasePublisher) verifyTagPushed(tagName string) error {
	cmd := exec.Command("git", "ls-remote", "--tags", "origin", tagName)
	cmd.Dir = p.repoPath
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to check remote tags: %w", err)
	}

	if strings.TrimSpace(string(output)) == "" {
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

	// Strip markdown heading syntax if present
	title = strings.TrimPrefix(title, "# ")
	title = strings.TrimSpace(title)

	return title
}
