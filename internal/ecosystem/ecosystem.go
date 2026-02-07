package ecosystem

import "github.com/NatoNathan/shipyard/pkg/semver"

// Handler provides a unified interface for ecosystem version operations
type Handler interface {
	ReadVersion() (semver.Version, error)
	UpdateVersion(version semver.Version) error
	GetVersionFiles() []string
}
