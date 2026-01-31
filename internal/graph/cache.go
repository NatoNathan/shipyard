package graph

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/NatoNathan/shipyard/internal/config"
)

// GraphCache provides caching for dependency graphs to avoid rebuilding
// on every operation. Thread-safe for concurrent access.
type GraphCache struct {
	mu    sync.RWMutex
	cache map[string]*DependencyGraph
}

// NewGraphCache creates a new graph cache.
func NewGraphCache() *GraphCache {
	return &GraphCache{
		cache: make(map[string]*DependencyGraph),
	}
}

// GetOrBuild returns a cached graph if available, otherwise builds and caches it.
// Errors are not cached - failed builds will retry on next call.
func (gc *GraphCache) GetOrBuild(cfg *config.Config) (*DependencyGraph, error) {
	if cfg == nil {
		return nil, fmt.Errorf("cannot build graph: config is nil")
	}

	// Compute cache key from config
	key, err := gc.cacheKey(cfg)
	if err != nil {
		// If we can't compute cache key, just build without caching
		return BuildGraph(cfg)
	}

	// Check cache (read lock)
	gc.mu.RLock()
	if cached, ok := gc.cache[key]; ok {
		gc.mu.RUnlock()
		return cached, nil
	}
	gc.mu.RUnlock()

	// Build graph (no lock during build to allow concurrent builds)
	g, err := BuildGraph(cfg)
	if err != nil {
		// Don't cache errors
		return nil, err
	}

	// Store in cache (write lock)
	gc.mu.Lock()
	gc.cache[key] = g
	gc.mu.Unlock()

	return g, nil
}

// Clear removes all entries from the cache.
func (gc *GraphCache) Clear() {
	gc.mu.Lock()
	defer gc.mu.Unlock()

	gc.cache = make(map[string]*DependencyGraph)
}

// Invalidate removes the cache entry for a specific config.
func (gc *GraphCache) Invalidate(cfg *config.Config) {
	if cfg == nil {
		return
	}

	key, err := gc.cacheKey(cfg)
	if err != nil {
		return
	}

	gc.mu.Lock()
	defer gc.mu.Unlock()

	delete(gc.cache, key)
}

// cacheKey computes a deterministic cache key from a config.
// Uses SHA-256 hash of the JSON representation.
func (gc *GraphCache) cacheKey(cfg *config.Config) (string, error) {
	// Serialize config to JSON for hashing
	data, err := json.Marshal(cfg)
	if err != nil {
		return "", fmt.Errorf("failed to marshal config for cache key: %w", err)
	}

	// Compute SHA-256 hash
	hash := sha256.Sum256(data)
	return fmt.Sprintf("%x", hash), nil
}
