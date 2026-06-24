package github

import (
	"context"
	"testing"

	"github.com/NatoNathan/shipyard/internal/config"
	"github.com/NatoNathan/shipyard/internal/upgrade"
	"github.com/NatoNathan/shipyard/pkg/semver"
	"github.com/stretchr/testify/assert"
)

func TestExtractTitleFromNotes(t *testing.T) {
	version := semver.Version{Major: 1, Minor: 2, Patch: 3}

	tests := []struct {
		name     string
		notes    string
		pkg      string
		expected string
	}{
		{
			name:     "empty string",
			notes:    "",
			pkg:      "myapp",
			expected: "myapp v1.2.3",
		},
		{
			name:     "markdown heading prefix",
			notes:    "# Release Notes\nSome details",
			pkg:      "myapp",
			expected: "myapp v1.2.3",
		},
		{
			name:     "normal title string",
			notes:    "New release with bugfixes\nMore details here",
			pkg:      "myapp",
			expected: "New release with bugfixes",
		},
		{
			name:     "whitespace only first line",
			notes:    "   \nActual content",
			pkg:      "myapp",
			expected: "myapp v1.2.3",
		},
		{
			name:     "multiple heading levels",
			notes:    "## Minor heading\nContent",
			pkg:      "myapp",
			expected: "myapp v1.2.3",
		},
		{
			name:     "single line notes",
			notes:    "Quick bugfix release",
			pkg:      "core",
			expected: "Quick bugfix release",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractTitleFromNotes(tt.notes, tt.pkg, version)
			assert.Equal(t, tt.expected, result)
		})
	}
}

type fakeReleaseClient struct {
	owner   string
	repo    string
	request *upgrade.CreateReleaseRequest
	err     error
}

func (c *fakeReleaseClient) CreateRelease(ctx context.Context, owner, repo string, release *upgrade.CreateReleaseRequest) (*upgrade.ReleaseInfo, error) {
	c.owner = owner
	c.repo = repo
	c.request = release
	if c.err != nil {
		return nil, c.err
	}
	return &upgrade.ReleaseInfo{TagName: release.TagName, Name: release.Name, Body: release.Body, Prerelease: release.Prerelease}, nil
}

func TestReleasePublisher_PublishReleaseUsesInjectedClient(t *testing.T) {
	cfg := &config.Config{
		GitHub: config.GitHubConfig{
			Owner: "octo",
			Repo:  "shipyard",
		},
	}
	client := &fakeReleaseClient{}
	publisher := NewReleasePublisherWithClient(t.TempDir(), cfg, client)
	publisher.tagExists = func(tagName string) error { return nil }
	publisher.tagPushed = func(tagName string) error { return nil }

	version := semver.Version{Major: 1, Minor: 2, Patch: 3}
	err := publisher.PublishRelease(context.Background(), "core", version, "core/v1.2.3", "Ship it\nDetails", true, true)

	assert.NoError(t, err)
	assert.Equal(t, "octo", client.owner)
	assert.Equal(t, "shipyard", client.repo)
	assert.NotNil(t, client.request)
	assert.Equal(t, "core/v1.2.3", client.request.TagName)
	assert.Equal(t, "Ship it", client.request.Name)
	assert.Equal(t, "Ship it\nDetails", client.request.Body)
	assert.True(t, client.request.Draft)
	assert.True(t, client.request.Prerelease)
}
