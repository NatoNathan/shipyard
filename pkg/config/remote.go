// Package config provides remote configuration fetching capabilities.
// This allows teams to extend a shared remote configuration from HTTP, Git, or GitHub.
package config

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	nethttp "net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/NatoNathan/shipyard/internal/logger"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/spf13/viper"
)

// RemoteConfigCache represents a cached remote config entry
type RemoteConfigCache struct {
	URL         string    `json:"url" yaml:"url"`
	Hash        string    `json:"hash" yaml:"hash"`
	LastFetched time.Time `json:"last_fetched" yaml:"last_fetched"`
	Content     string    `json:"content" yaml:"content"`
	CachePath   string    `json:"cache_path" yaml:"cache_path"`
	TTL         int       `json:"ttl" yaml:"ttl"` // Time to live in minutes
}

// RemoteConfigFetcher provides functionality to fetch remote configurations
type RemoteConfigFetcher struct {
	cacheDir string
	client   *nethttp.Client
}

// NewRemoteConfigFetcher creates a new remote config fetcher
func NewRemoteConfigFetcher(cacheDir string) *RemoteConfigFetcher {
	if cacheDir == "" {
		// Use system cache directory instead of local .shipyard/cache
		if systemCacheDir, err := os.UserCacheDir(); err == nil {
			cacheDir = filepath.Join(systemCacheDir, "shipyard", "remote-configs")
		} else {
			// Fallback to local cache if system cache dir is not available
			cacheDir = filepath.Join(".shipyard", "cache", "remote-configs")
		}
	}

	return &RemoteConfigFetcher{
		cacheDir: cacheDir,
		client: &nethttp.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// FetchRemoteConfig fetches a remote configuration and returns a viper instance
func (f *RemoteConfigFetcher) FetchRemoteConfig(extendsURL string, forceFresh bool) (*viper.Viper, error) {
	logger.Info("FetchRemoteConfig called", "url", extendsURL, "force_fresh", forceFresh)

	if extendsURL == "" {
		return nil, fmt.Errorf("remote config URL cannot be empty")
	}

	logger.Info("Determining fetch method for URL", "url", extendsURL)

	// Determine the fetch method based on URL prefix
	switch {
	case strings.HasPrefix(extendsURL, "github:"):
		logger.Info("Using GitHub fetcher")
		return f.fetchFromGitHub(extendsURL, forceFresh)
	case strings.HasPrefix(extendsURL, "git+"):
		logger.Info("Using Git fetcher")
		return f.fetchFromGit(extendsURL, forceFresh)
	case strings.HasPrefix(extendsURL, "https://") || strings.HasPrefix(extendsURL, "http://"):
		logger.Info("Using HTTP fetcher")
		return f.fetchFromHTTP(extendsURL, forceFresh)
	default:
		logger.Error("Unsupported URL format", "url", extendsURL)
		return nil, fmt.Errorf("unsupported remote config URL format: %s", extendsURL)
	}
}

// FetchRemoteTemplate fetches a remote template file and returns its content
func (f *RemoteConfigFetcher) FetchRemoteTemplate(templateURL string, forceFresh bool) (string, error) {
	if templateURL == "" {
		return "", fmt.Errorf("remote template URL cannot be empty")
	}

	// Use the same logic as config fetching but return content directly
	switch {
	case strings.HasPrefix(templateURL, "github:"):
		return f.fetchTemplateFromGitHub(templateURL, forceFresh)
	case strings.HasPrefix(templateURL, "git+"):
		return f.fetchTemplateFromGit(templateURL, forceFresh)
	case strings.HasPrefix(templateURL, "https://") || strings.HasPrefix(templateURL, "http://"):
		return f.fetchTemplateFromHTTP(templateURL, forceFresh)
	default:
		return "", fmt.Errorf("unsupported remote template URL format: %s", templateURL)
	}
}

// fetchTemplateFromHTTP fetches template content from an HTTP URL
func (f *RemoteConfigFetcher) fetchTemplateFromHTTP(httpURL string, forceFresh bool) (string, error) {
	// Check cache first unless forcing fresh fetch
	if !forceFresh {
		if content, err := f.loadTemplateFromCache(httpURL); err == nil {
			return content, nil
		}
	}

	// Fetch from HTTP
	resp, err := f.client.Get(httpURL)
	if err != nil {
		return "", fmt.Errorf("failed to fetch template from %s: %w", httpURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != nethttp.StatusOK {
		return "", fmt.Errorf("failed to fetch template from %s: HTTP %d", httpURL, resp.StatusCode)
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read template response body: %w", err)
	}

	templateContent := string(content)

	// Cache the content
	if err := f.saveTemplateToCache(httpURL, templateContent, 60); err != nil {
		// Log warning but don't fail
		fmt.Printf("Warning: failed to cache remote template: %v\n", err)
	}

	return templateContent, nil
}

// fetchTemplateFromGitHub fetches template content from GitHub using Git clone
func (f *RemoteConfigFetcher) fetchTemplateFromGitHub(githubURL string, forceFresh bool) (string, error) {
	logger.Info("Fetching template from GitHub", "url", githubURL, "force_fresh", forceFresh)

	// Parse GitHub URL
	parts := strings.TrimPrefix(githubURL, "github:")
	refParts := strings.Split(parts, "@")

	var pathParts, ref string
	if len(refParts) == 2 {
		pathParts = refParts[0]
		ref = refParts[1]
	} else {
		pathParts = parts
		ref = "main" // default branch
	}

	pathComponents := strings.Split(pathParts, "/")
	if len(pathComponents) < 3 {
		return "", fmt.Errorf("invalid GitHub URL format, expected github:owner/repo/path/to/template[@ref]")
	}

	owner := pathComponents[0]
	repo := pathComponents[1]
	templatePath := strings.Join(pathComponents[2:], "/")

	logger.Info("Parsed GitHub template URL", "owner", owner, "repo", repo, "path", templatePath, "ref", ref)

	// Try SSH first, fallback to HTTPS
	repoURLs := []string{
		fmt.Sprintf("git@github.com:%s/%s.git", owner, repo),
		fmt.Sprintf("https://github.com/%s/%s.git", owner, repo),
	}

	logger.Info("Template repository URLs to try", "urls", repoURLs)

	return f.fetchTemplateFromGitRepository(githubURL, repoURLs, templatePath, ref, forceFresh)
}

// fetchTemplateFromGit fetches template content from a Git repository with SSH/HTTPS support
func (f *RemoteConfigFetcher) fetchTemplateFromGit(gitURL string, forceFresh bool) (string, error) {
	// Parse git URL similar to fetchFromGit
	urlWithoutPrefix := strings.TrimPrefix(gitURL, "git+")

	var repoURL, filePath, ref string

	if strings.HasPrefix(urlWithoutPrefix, "git@") {
		// SSH format
		parts := strings.SplitN(urlWithoutPrefix, "/", 2)
		if len(parts) < 2 {
			return "", fmt.Errorf("invalid SSH git URL format")
		}

		repoURL = parts[0] + ".git"
		filePath = parts[1]

		if strings.Contains(repoURL, ".git/") {
			gitParts := strings.Split(urlWithoutPrefix, ".git/")
			repoURL = gitParts[0] + ".git"
			filePath = gitParts[1]
		}
	} else {
		// HTTPS format
		parts := strings.Split(urlWithoutPrefix, "/")
		if len(parts) < 4 {
			return "", fmt.Errorf("invalid git URL format")
		}

		for i := 3; i < len(parts); i++ {
			if strings.HasSuffix(parts[i], ".git") {
				repoURL = strings.Join(parts[:i+1], "/")
				if i+1 < len(parts) {
					filePath = strings.Join(parts[i+1:], "/")
				}
				break
			}
		}
	}

	if repoURL == "" {
		return "", fmt.Errorf("could not parse repository URL from: %s", gitURL)
	}

	// Check for ref specification
	if strings.Contains(filePath, "@") {
		pathParts := strings.Split(filePath, "@")
		filePath = pathParts[0]
		ref = pathParts[1]
	} else {
		ref = "main" // default branch
	}

	return f.fetchTemplateFromGitRepository(gitURL, []string{repoURL}, filePath, ref, forceFresh)
}

// fetchTemplateFromGitRepository is a common method to fetch templates from Git repositories
func (f *RemoteConfigFetcher) fetchTemplateFromGitRepository(cacheKey string, repoURLs []string, filePath, ref string, forceFresh bool) (string, error) {
	logger.Info("Starting Git repository template fetch", "cache_key", cacheKey, "repo_urls", repoURLs, "file_path", filePath, "ref", ref, "force_fresh", forceFresh)

	// Check cache first unless forcing fresh fetch
	if !forceFresh {
		logger.Info("Checking cache for template")
		if content, err := f.loadTemplateFromCache(cacheKey); err == nil {
			logger.Info("Found valid cached template, using cached version")
			return content, nil
		} else {
			logger.Info("No valid template cache found, will fetch from repository", "cache_error", err)
		}
	} else {
		logger.Info("Forcing fresh template fetch, skipping cache check")
	}

	var lastErr error

	// Try each repository URL (SSH first, then HTTPS)
	for i, repoURL := range repoURLs {
		logger.Info("Attempting to fetch template from repository", "attempt", i+1, "total_urls", len(repoURLs), "repo_url", repoURL)

		content, err := f.cloneAndReadFile(repoURL, filePath, ref)
		if err != nil {
			logger.Warn("Failed to fetch template from repository URL", "repo_url", repoURL, "error", err)
			lastErr = err
			continue // Try next URL
		}

		logger.Info("Successfully fetched template content from repository", "repo_url", repoURL, "content_length", len(content))

		// Cache the content
		if err := f.saveTemplateToCache(cacheKey, content, 60); err != nil {
			// Log warning but don't fail
			logger.Warn("Failed to cache remote template", "error", err)
			fmt.Printf("Warning: failed to cache remote template: %v\n", err)
		} else {
			logger.Info("Successfully cached fetched template")
		}

		return content, nil
	}

	logger.Error("Failed to fetch template from any repository URL", "total_attempts", len(repoURLs), "last_error", lastErr)
	return "", fmt.Errorf("failed to fetch template from any repository URL, last error: %w", lastErr)
}
func (f *RemoteConfigFetcher) fetchFromHTTP(httpURL string, forceFresh bool) (*viper.Viper, error) {
	// Check cache first unless forcing fresh fetch
	if !forceFresh {
		if cachedConfig, err := f.loadFromCache(httpURL); err == nil {
			return cachedConfig, nil
		}
	}

	// Fetch from HTTP
	resp, err := f.client.Get(httpURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch config from %s: %w", httpURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != nethttp.StatusOK {
		return nil, fmt.Errorf("failed to fetch config from %s: HTTP %d", httpURL, resp.StatusCode)
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Cache the content
	if err := f.saveToCache(httpURL, string(content), 60); err != nil {
		// Log warning but don't fail
		fmt.Printf("Warning: failed to cache remote config: %v\n", err)
	}

	return f.parseConfigContent(string(content), httpURL)
}

// fetchFromGitHub fetches configuration from GitHub using Git clone
// Format: github:owner/repo/path/to/config.yaml[@ref]
func (f *RemoteConfigFetcher) fetchFromGitHub(githubURL string, forceFresh bool) (*viper.Viper, error) {
	logger.Info("Fetching from GitHub", "url", githubURL, "force_fresh", forceFresh)

	// Parse GitHub URL
	parts := strings.TrimPrefix(githubURL, "github:")
	refParts := strings.Split(parts, "@")

	var pathParts, ref string
	if len(refParts) == 2 {
		pathParts = refParts[0]
		ref = refParts[1]
	} else {
		pathParts = parts
		ref = "main" // default branch
	}

	pathComponents := strings.Split(pathParts, "/")
	if len(pathComponents) < 3 {
		return nil, fmt.Errorf("invalid GitHub URL format, expected github:owner/repo/path/to/config.yaml[@ref]")
	}

	owner := pathComponents[0]
	repo := pathComponents[1]
	configPath := strings.Join(pathComponents[2:], "/")

	logger.Info("Parsed GitHub URL", "owner", owner, "repo", repo, "path", configPath, "ref", ref)

	// Try SSH first, fallback to HTTPS
	repoURLs := []string{
		fmt.Sprintf("git@github.com:%s/%s.git", owner, repo),
		fmt.Sprintf("https://github.com/%s/%s.git", owner, repo),
	}

	logger.Info("Repository URLs to try", "urls", repoURLs)

	return f.fetchFromGitRepository(githubURL, repoURLs, configPath, ref, forceFresh)
}

// fetchFromGit fetches configuration from a Git repository with SSH/HTTPS support
// Format: git+https://github.com/owner/repo.git/path/to/config.yaml[@ref] or git+git@github.com:owner/repo.git/path/to/config.yaml[@ref]
func (f *RemoteConfigFetcher) fetchFromGit(gitURL string, forceFresh bool) (*viper.Viper, error) {
	logger.Info("Fetching from Git repository", "url", gitURL, "force_fresh", forceFresh)

	// Parse git URL
	urlWithoutPrefix := strings.TrimPrefix(gitURL, "git+")

	// Split URL and path - handle both SSH and HTTPS formats
	var repoURL, filePath, ref string

	if strings.HasPrefix(urlWithoutPrefix, "git@") {
		logger.Info("Detected SSH format URL", "url", urlWithoutPrefix)
		// SSH format: git@github.com:owner/repo.git/path/to/file[@ref]
		parts := strings.SplitN(urlWithoutPrefix, "/", 2)
		if len(parts) < 2 {
			return nil, fmt.Errorf("invalid SSH git URL format")
		}

		repoURL = parts[0] + ".git"
		filePath = parts[1]

		// Extract .git boundary more precisely for SSH
		if strings.Contains(repoURL, ".git/") {
			gitParts := strings.Split(urlWithoutPrefix, ".git/")
			repoURL = gitParts[0] + ".git"
			filePath = gitParts[1]
		}
	} else {
		logger.Info("Detected HTTPS format URL", "url", urlWithoutPrefix)
		// HTTPS format: https://github.com/owner/repo.git/path/to/file[@ref]
		parts := strings.Split(urlWithoutPrefix, "/")
		if len(parts) < 4 {
			return nil, fmt.Errorf("invalid git URL format")
		}

		// Find where the repository URL ends and the file path begins
		for i := 3; i < len(parts); i++ {
			if strings.HasSuffix(parts[i], ".git") {
				repoURL = strings.Join(parts[:i+1], "/")
				if i+1 < len(parts) {
					filePath = strings.Join(parts[i+1:], "/")
				}
				break
			}
		}
	}

	if repoURL == "" {
		return nil, fmt.Errorf("could not parse repository URL from: %s", gitURL)
	}

	// Check for ref specification in file path
	if strings.Contains(filePath, "@") {
		pathParts := strings.Split(filePath, "@")
		filePath = pathParts[0]
		ref = pathParts[1]
	} else {
		ref = "main" // default branch
	}

	logger.Info("Parsed Git URL", "repo_url", repoURL, "file_path", filePath, "ref", ref)

	return f.fetchFromGitRepository(gitURL, []string{repoURL}, filePath, ref, forceFresh)
}

// fetchFromGitRepository is a common method to fetch from Git repositories with auth support
func (f *RemoteConfigFetcher) fetchFromGitRepository(cacheKey string, repoURLs []string, filePath, ref string, forceFresh bool) (*viper.Viper, error) {
	logger.Info("Starting Git repository fetch", "cache_key", cacheKey, "repo_urls", repoURLs, "file_path", filePath, "ref", ref, "force_fresh", forceFresh)

	// Check cache first unless forcing fresh fetch
	if !forceFresh {
		logger.Info("Checking cache for configuration")
		if cachedConfig, err := f.loadFromCache(cacheKey); err == nil {
			logger.Info("Found valid cached configuration, using cached version")
			return cachedConfig, nil
		} else {
			logger.Info("No valid cache found, will fetch from repository", "cache_error", err)
		}
	} else {
		logger.Info("Forcing fresh fetch, skipping cache check")
	}

	var lastErr error

	// Try each repository URL (SSH first, then HTTPS)
	for i, repoURL := range repoURLs {
		logger.Info("Attempting to fetch from repository", "attempt", i+1, "total_urls", len(repoURLs), "repo_url", repoURL)

		content, err := f.cloneAndReadFile(repoURL, filePath, ref)
		if err != nil {
			logger.Warn("Failed to fetch from repository URL", "repo_url", repoURL, "error", err)
			lastErr = err
			continue // Try next URL
		}

		logger.Info("Successfully fetched content from repository", "repo_url", repoURL, "content_length", len(content))

		// Cache the content
		if err := f.saveToCache(cacheKey, content, 60); err != nil {
			// Log warning but don't fail
			logger.Warn("Failed to cache remote config", "error", err)
			fmt.Printf("Warning: failed to cache remote config: %v\n", err)
		} else {
			logger.Info("Successfully cached fetched content")
		}

		logger.Info("Parsing configuration content")
		config, err := f.parseConfigContent(content, cacheKey)
		if err != nil {
			logger.Error("Failed to parse configuration content", "error", err)
			return nil, fmt.Errorf("failed to parse config content: %w", err)
		}

		logger.Info("Successfully parsed configuration content")
		return config, nil
	}

	logger.Error("Failed to fetch from any repository URL", "total_attempts", len(repoURLs), "last_error", lastErr)
	return nil, fmt.Errorf("failed to fetch from any repository URL, last error: %w", lastErr)
}

// cloneAndReadFile clones a repository and reads a specific file using sparse checkout
func (f *RemoteConfigFetcher) cloneAndReadFile(repoURL, filePath, ref string) (string, error) {
	logger.Info("Starting Git sparse clone operation", "repo_url", repoURL, "file_path", filePath, "ref", ref)

	// Clone repository to memory storage with single branch and shallow depth
	cloneOptions := &git.CloneOptions{
		URL:           repoURL,
		ReferenceName: plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", ref)),
		SingleBranch:  true,
		Depth:         1,   // Shallow clone - only latest commit
		Auth:          nil, // Let go-git handle authentication (SSH agent, keys, tokens, etc.)
	}

	logger.Info("Attempting sparse clone to memory", "clone_options", map[string]interface{}{
		"url":            cloneOptions.URL,
		"reference_name": cloneOptions.ReferenceName,
		"single_branch":  cloneOptions.SingleBranch,
		"depth":          cloneOptions.Depth,
	})

	// Clone to memory storage for efficiency
	repo, err := git.Clone(memory.NewStorage(), nil, cloneOptions)
	if err != nil {
		logger.Error("Failed to clone repository to memory", "repo_url", repoURL, "error", err)
		return "", fmt.Errorf("failed to clone repository %s: %w", repoURL, err)
	}

	logger.Info("Successfully cloned repository to memory", "repo_url", repoURL)

	// Get the HEAD reference to find the commit
	head, err := repo.Head()
	if err != nil {
		logger.Error("Failed to get HEAD reference", "error", err)
		return "", fmt.Errorf("failed to get HEAD reference: %w", err)
	}

	logger.Info("Got HEAD reference", "commit_hash", head.Hash())

	// Get the commit object
	commit, err := repo.CommitObject(head.Hash())
	if err != nil {
		logger.Error("Failed to get commit object", "commit_hash", head.Hash(), "error", err)
		return "", fmt.Errorf("failed to get commit object: %w", err)
	}

	logger.Info("Got commit object", "commit_hash", commit.Hash)

	// Get the tree from the commit
	tree, err := commit.Tree()
	if err != nil {
		logger.Error("Failed to get commit tree", "error", err)
		return "", fmt.Errorf("failed to get commit tree: %w", err)
	}

	logger.Info("Got commit tree, looking for file", "file_path", filePath)

	// Get the file from the tree
	file, err := tree.File(filePath)
	if err != nil {
		logger.Error("Failed to find file in tree", "file_path", filePath, "error", err)
		return "", fmt.Errorf("failed to find file %s in repository: %w", filePath, err)
	}

	logger.Info("Found file in tree, reading content", "file_path", filePath, "file_hash", file.Hash)

	// Read the file content
	content, err := file.Contents()
	if err != nil {
		logger.Error("Failed to read file contents", "file_path", filePath, "error", err)
		return "", fmt.Errorf("failed to read file contents: %w", err)
	}

	logger.Info("Successfully read file content", "content_length", len(content))
	return content, nil
}

// parseConfigContent parses configuration content and returns a viper instance
func (f *RemoteConfigFetcher) parseConfigContent(content, sourceURL string) (*viper.Viper, error) {
	v := viper.New()

	// Determine format from URL or content
	format := f.detectFormat(sourceURL, content)

	v.SetConfigType(format)
	if err := v.ReadConfig(strings.NewReader(content)); err != nil {
		return nil, fmt.Errorf("failed to parse config content: %w", err)
	}

	return v, nil
}

// detectFormat detects the configuration file format
func (f *RemoteConfigFetcher) detectFormat(sourceURL, content string) string {
	// Try to detect from URL extension
	if strings.HasSuffix(sourceURL, ".yaml") || strings.HasSuffix(sourceURL, ".yml") {
		return "yaml"
	}
	if strings.HasSuffix(sourceURL, ".json") {
		return "json"
	}
	if strings.HasSuffix(sourceURL, ".toml") {
		return "toml"
	}

	// Try to detect from content
	content = strings.TrimSpace(content)
	if strings.HasPrefix(content, "{") {
		return "json"
	}

	// Default to YAML
	return "yaml"
}

// saveTemplateToCache saves remote template content to cache
func (f *RemoteConfigFetcher) saveTemplateToCache(url, content string, ttlMinutes int) error {
	// Templates use the same cache structure as configs but with a different prefix
	return f.saveToCache("template:"+url, content, ttlMinutes)
}

// loadTemplateFromCache loads template from cache if valid
func (f *RemoteConfigFetcher) loadTemplateFromCache(url string) (string, error) {
	// Check cache using template prefix
	hash := sha256.Sum256([]byte("template:" + url))
	cacheFile := filepath.Join(f.cacheDir, fmt.Sprintf("%x.json", hash))

	// Check if cache file exists
	if _, err := os.Stat(cacheFile); os.IsNotExist(err) {
		return "", fmt.Errorf("cache file does not exist")
	}

	// Read cache file
	data, err := os.ReadFile(cacheFile)
	if err != nil {
		return "", fmt.Errorf("failed to read cache file: %w", err)
	}

	var cache RemoteConfigCache
	if err := json.Unmarshal(data, &cache); err != nil {
		return "", fmt.Errorf("failed to unmarshal cache data: %w", err)
	}

	// Check if cache is still valid
	if cache.TTL > 0 {
		expireTime := cache.LastFetched.Add(time.Duration(cache.TTL) * time.Minute)
		if time.Now().After(expireTime) {
			return "", fmt.Errorf("cache expired")
		}
	}

	return cache.Content, nil
}

// ResolveTemplatePath resolves a template path relative to a config URL
func ResolveTemplatePath(templatePath, configURL string) string {
	// If template path is already a remote URL, return as-is
	if IsValidRemoteURL(templatePath) {
		return templatePath
	}

	// If it's not a relative path, return as-is (local file)
	if filepath.IsAbs(templatePath) || !strings.Contains(templatePath, "/") {
		return templatePath
	}

	// Resolve relative to config URL
	if IsValidRemoteURL(configURL) {
		return resolveRelativeRemoteURL(templatePath, configURL)
	}

	// For local configs, resolve relative to config file directory
	return filepath.Join(filepath.Dir(configURL), templatePath)
}

// resolveRelativeRemoteURL resolves a relative path against a remote URL
func resolveRelativeRemoteURL(relativePath, baseURL string) string {
	switch {
	case strings.HasPrefix(baseURL, "github:"):
		return resolveGitHubRelativePath(relativePath, baseURL)
	case strings.HasPrefix(baseURL, "git+"):
		return resolveGitRelativePath(relativePath, baseURL)
	case strings.HasPrefix(baseURL, "https://") || strings.HasPrefix(baseURL, "http://"):
		return resolveHTTPRelativePath(relativePath, baseURL)
	default:
		return relativePath
	}
}

// resolveGitHubRelativePath resolves relative path for GitHub URLs
func resolveGitHubRelativePath(relativePath, githubURL string) string {
	// Parse the GitHub URL to get the base path
	parts := strings.TrimPrefix(githubURL, "github:")
	refParts := strings.Split(parts, "@")

	var pathParts, ref string
	if len(refParts) == 2 {
		pathParts = refParts[0]
		ref = "@" + refParts[1]
	} else {
		pathParts = parts
		ref = ""
	}

	pathComponents := strings.Split(pathParts, "/")
	if len(pathComponents) < 3 {
		return relativePath
	}

	// Get base directory
	baseDir := strings.Join(pathComponents[:len(pathComponents)-1], "/")
	return "github:" + baseDir + "/" + relativePath + ref
}

// resolveGitRelativePath resolves relative path for Git URLs
func resolveGitRelativePath(relativePath, gitURL string) string {
	// Find the last slash before .git and append the relative path
	urlWithoutPrefix := strings.TrimPrefix(gitURL, "git+")
	lastSlash := strings.LastIndex(urlWithoutPrefix, "/")
	if lastSlash == -1 {
		return relativePath
	}
	baseURL := urlWithoutPrefix[:lastSlash]
	return "git+" + baseURL + "/" + relativePath
}

// resolveHTTPRelativePath resolves relative path for HTTP URLs
func resolveHTTPRelativePath(relativePath, httpURL string) string {
	// Use standard URL resolution
	if baseURL, err := url.Parse(httpURL); err == nil {
		if resolvedURL, err := baseURL.Parse(relativePath); err == nil {
			return resolvedURL.String()
		}
	}
	return relativePath
}
func (f *RemoteConfigFetcher) saveToCache(url, content string, ttlMinutes int) error {
	// Create cache directory
	if err := os.MkdirAll(f.cacheDir, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	// Generate cache file name from URL hash
	hash := sha256.Sum256([]byte(url))
	cacheFile := filepath.Join(f.cacheDir, fmt.Sprintf("%x.json", hash))

	cache := RemoteConfigCache{
		URL:         url,
		Hash:        fmt.Sprintf("%x", sha256.Sum256([]byte(content))),
		LastFetched: time.Now(),
		Content:     content,
		CachePath:   cacheFile,
		TTL:         ttlMinutes,
	}

	data, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal cache data: %w", err)
	}

	if err := os.WriteFile(cacheFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write cache file: %w", err)
	}

	return nil
}

// loadFromCache loads config from cache if valid
func (f *RemoteConfigFetcher) loadFromCache(url string) (*viper.Viper, error) {
	// Generate cache file name from URL hash
	hash := sha256.Sum256([]byte(url))
	cacheFile := filepath.Join(f.cacheDir, fmt.Sprintf("%x.json", hash))

	// Check if cache file exists
	if _, err := os.Stat(cacheFile); os.IsNotExist(err) {
		return nil, fmt.Errorf("cache file does not exist")
	}

	// Read cache file
	data, err := os.ReadFile(cacheFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read cache file: %w", err)
	}

	var cache RemoteConfigCache
	if err := json.Unmarshal(data, &cache); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cache data: %w", err)
	}

	// Check if cache is still valid
	if cache.TTL > 0 {
		expireTime := cache.LastFetched.Add(time.Duration(cache.TTL) * time.Minute)
		if time.Now().After(expireTime) {
			return nil, fmt.Errorf("cache expired")
		}
	}

	// Parse cached content
	return f.parseConfigContent(cache.Content, url)
}

// ClearCache clears the remote config cache
func (f *RemoteConfigFetcher) ClearCache() error {
	if _, err := os.Stat(f.cacheDir); os.IsNotExist(err) {
		return nil // Cache directory doesn't exist, nothing to clear
	}

	return os.RemoveAll(f.cacheDir)
}

// ListCachedConfigs returns a list of cached remote configurations
func (f *RemoteConfigFetcher) ListCachedConfigs() ([]RemoteConfigCache, error) {
	if _, err := os.Stat(f.cacheDir); os.IsNotExist(err) {
		return []RemoteConfigCache{}, nil
	}

	entries, err := os.ReadDir(f.cacheDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read cache directory: %w", err)
	}

	var configs []RemoteConfigCache
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") {
			cacheFile := filepath.Join(f.cacheDir, entry.Name())
			data, err := os.ReadFile(cacheFile)
			if err != nil {
				continue // Skip invalid cache files
			}

			var cache RemoteConfigCache
			if err := json.Unmarshal(data, &cache); err != nil {
				continue // Skip invalid cache files
			}

			configs = append(configs, cache)
		}
	}

	return configs, nil
}

// IsValidRemoteURL checks if a URL is a valid remote config URL
func IsValidRemoteURL(url string) bool {
	return strings.HasPrefix(url, "github:") ||
		strings.HasPrefix(url, "git+") ||
		strings.HasPrefix(url, "https://") ||
		strings.HasPrefix(url, "http://")
}

// ParseRemoteURL parses and validates a remote config URL
func ParseRemoteURL(rawURL string) (*url.URL, error) {
	if !IsValidRemoteURL(rawURL) {
		return nil, fmt.Errorf("invalid remote URL format: %s", rawURL)
	}

	// Handle special formats
	switch {
	case strings.HasPrefix(rawURL, "github:"):
		// Convert github: format to a parseable URL
		parts := strings.TrimPrefix(rawURL, "github:")
		return url.Parse("https://github.com/" + parts)
	case strings.HasPrefix(rawURL, "git+"):
		// Remove git+ prefix and parse
		return url.Parse(strings.TrimPrefix(rawURL, "git+"))
	default:
		return url.Parse(rawURL)
	}
}
