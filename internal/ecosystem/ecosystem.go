package ecosystem

import (
	"github.com/NatoNathan/shipyard/internal/config"
	"github.com/NatoNathan/shipyard/pkg/semver"
)

// Handler provides a unified interface for ecosystem version operations
type Handler interface {
	ReadVersion() (semver.Version, error)
	UpdateVersion(version semver.Version) error
	GetVersionFiles() []string
}

// HandlerContext provides additional context for handlers that need it
type HandlerContext struct {
	AllVersions   map[string]semver.Version // All package versions (new versions after bumps)
	PackageConfig *config.Package           // Full package configuration
}

// HandlerWithContext is an optional interface for handlers that need additional context
type HandlerWithContext interface {
	Handler
	SetContext(ctx *HandlerContext)
}
