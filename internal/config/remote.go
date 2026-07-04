package config

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
	githttp "github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/storage/memory"
	"gopkg.in/yaml.v3"
)

const (
	// maxExtendsDepth bounds recursive extends chains (remote configs that
	// themselves extend other configs).
	maxExtendsDepth = 10

	// remoteFetchTimeout bounds a single HTTP fetch of a remote config.
	remoteFetchTimeout = 30 * time.Second

	// maxRemoteConfigSize bounds the size of a fetched remote config.
	maxRemoteConfigSize = 10 << 20 // 10 MiB
)

// ResolveExtends fetches every remote configuration referenced by the
// config's extends list and merges them beneath the local config. Sources are
// merged in order (first has lowest precedence), with the local config
// applied last. Remote configs may themselves extend other configs; cycles
// and over-deep chains are rejected, and remote-supplied extends must stay on
// the same host as the source that supplied them.
func ResolveExtends(cfg *Config) (*Config, error) {
	r := &extendsResolver{
		httpClient: &http.Client{Timeout: remoteFetchTimeout},
		stack:      make(map[string]bool),
	}
	return r.resolve(cfg, 0, "")
}

type extendsResolver struct {
	httpClient *http.Client
	stack      map[string]bool // sources currently being resolved, for cycle detection
}

// resolve merges cfg's extends sources beneath cfg. pinnedHost is empty for
// the local config (its extends may point anywhere); for fetched configs it is
// the host they were fetched from, and their extends must stay on that host so
// an untrusted upstream cannot pivot fetches to other targets.
func (r *extendsResolver) resolve(cfg *Config, depth int, pinnedHost string) (*Config, error) {
	if len(cfg.Extends) == 0 {
		return cfg, nil
	}
	if depth >= maxExtendsDepth {
		return nil, fmt.Errorf("extends chain exceeds maximum depth of %d", maxExtendsDepth)
	}

	var base *Config
	for _, rc := range cfg.Extends {
		resolved, err := r.resolveSource(rc, depth, pinnedHost)
		if err != nil {
			return nil, err
		}
		if base == nil {
			base = resolved
		} else {
			base = base.Merge(resolved)
		}
	}

	return base.Merge(cfg), nil
}

func (r *extendsResolver) resolveSource(rc RemoteConfig, depth int, pinnedHost string) (*Config, error) {
	key := rc.sourceKey()
	if r.stack[key] {
		return nil, fmt.Errorf("circular dependency in remote configs: %s", key)
	}
	r.stack[key] = true
	defer delete(r.stack, key)

	host, hostErr := sourceHost(rc)
	if depth > 0 {
		if hostErr != nil {
			return nil, fmt.Errorf("refusing extends supplied by a remote config: %w", hostErr)
		}
		if pinnedHost == "" || host != pinnedHost {
			return nil, fmt.Errorf("remote config extends %s on host %q; extends supplied by a remote config must stay on its own host %q", key, host, pinnedHost)
		}
	}

	data, err := r.fetch(rc)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch remote config %s: %w", key, err)
	}

	var remote Config
	if err := yaml.Unmarshal(data, &remote); err != nil {
		return nil, fmt.Errorf("failed to parse remote config %s: %w", key, err)
	}

	return r.resolve(&remote, depth+1, host)
}

// sourceHost extracts the host a remote config source points at, handling
// HTTP(S)/ssh URLs and scp-like git addresses (git@host:org/repo.git).
func sourceHost(rc RemoteConfig) (string, error) {
	raw := rc.URL
	if raw == "" {
		raw = rc.Git
	}
	if rest, ok := strings.CutPrefix(raw, "git@"); ok {
		host, _, found := strings.Cut(rest, ":")
		if found && host != "" {
			return host, nil
		}
		return "", fmt.Errorf("cannot determine host of remote config source %s", rc.sourceKey())
	}
	u, err := url.Parse(raw)
	if err != nil || u.Hostname() == "" {
		return "", fmt.Errorf("cannot determine host of remote config source %s", rc.sourceKey())
	}
	return u.Hostname(), nil
}

// sourceKey identifies a remote source for cycle detection and error messages.
func (rc RemoteConfig) sourceKey() string {
	if rc.URL != "" {
		return rc.URL
	}
	return fmt.Sprintf("%s#%s@%s", rc.Git, rc.Path, rc.Ref)
}

func (r *extendsResolver) fetch(rc RemoteConfig) ([]byte, error) {
	token, err := resolveAuthToken(rc.Auth)
	if err != nil {
		return nil, err
	}

	switch {
	case rc.URL != "":
		return r.fetchHTTP(rc.URL, token)
	case rc.Git != "":
		return fetchGitFile(rc, token)
	default:
		return nil, fmt.Errorf("remote config must specify a url or git source")
	}
}

// resolveAuthToken resolves the auth field. The documented format is
// "env:VAR_NAME"; any other non-empty value is treated as a literal token.
func resolveAuthToken(auth string) (string, error) {
	if auth == "" {
		return "", nil
	}
	if envVar, ok := strings.CutPrefix(auth, "env:"); ok {
		val := os.Getenv(envVar)
		if val == "" {
			return "", fmt.Errorf("auth environment variable %s is not set", envVar)
		}
		return val, nil
	}
	return auth, nil
}

func (r *extendsResolver) fetchHTTP(url, token string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %s", resp.Status)
	}

	return io.ReadAll(io.LimitReader(resp.Body, maxRemoteConfigSize))
}

// fetchGitFile clones the repository in memory at the requested ref and reads
// the configured path. The ref may be a branch, tag, or full commit hash.
func fetchGitFile(rc RemoteConfig, token string) ([]byte, error) {
	var auth transport.AuthMethod
	if token != "" && (strings.HasPrefix(rc.Git, "http://") || strings.HasPrefix(rc.Git, "https://")) {
		auth = &githttp.BasicAuth{Username: "git", Password: token}
	}

	shallowClone := func(ref plumbing.ReferenceName) (billy.Filesystem, error) {
		ctx, cancel := context.WithTimeout(context.Background(), remoteFetchTimeout)
		defer cancel()
		fs := memfs.New()
		_, err := gogit.CloneContext(ctx, memory.NewStorage(), fs, &gogit.CloneOptions{
			URL:           rc.Git,
			Auth:          auth,
			ReferenceName: ref,
			SingleBranch:  true,
			Depth:         1,
		})
		return fs, err
	}

	var fs billy.Filesystem
	if plumbing.IsHash(rc.Ref) {
		var err error
		fs, err = cloneAtCommit(rc.Git, auth, plumbing.NewHash(rc.Ref))
		if err != nil {
			return nil, fmt.Errorf("failed to clone %s at ref %s: %w", rc.Git, rc.Ref, err)
		}
	} else {
		var branchErr error
		fs, branchErr = shallowClone(plumbing.NewBranchReferenceName(rc.Ref))
		if branchErr != nil {
			var tagErr error
			fs, tagErr = shallowClone(plumbing.NewTagReferenceName(rc.Ref))
			if tagErr != nil {
				return nil, fmt.Errorf("failed to clone %s at ref %s: %w", rc.Git, rc.Ref, branchErr)
			}
		}
	}

	f, err := fs.Open(rc.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s from %s: %w", rc.Path, rc.Git, err)
	}
	defer func() { _ = f.Close() }()

	return io.ReadAll(io.LimitReader(f, maxRemoteConfigSize))
}

func cloneAtCommit(url string, auth transport.AuthMethod, hash plumbing.Hash) (billy.Filesystem, error) {
	ctx, cancel := context.WithTimeout(context.Background(), remoteFetchTimeout)
	defer cancel()
	fs := memfs.New()
	repo, err := gogit.CloneContext(ctx, memory.NewStorage(), fs, &gogit.CloneOptions{
		URL:  url,
		Auth: auth,
	})
	if err != nil {
		return nil, err
	}
	wt, err := repo.Worktree()
	if err != nil {
		return nil, err
	}
	if err := wt.Checkout(&gogit.CheckoutOptions{Hash: hash}); err != nil {
		return nil, err
	}
	return fs, nil
}
