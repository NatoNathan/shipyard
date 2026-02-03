package upgrade

import "time"

// InstallMethod represents how shipyard was installed
type InstallMethod int

const (
	InstallMethodUnknown InstallMethod = iota
	InstallMethodHomebrew
	InstallMethodNPM
	InstallMethodGo
	InstallMethodScript
	InstallMethodDocker
)

// String returns the human-readable name of the install method
func (m InstallMethod) String() string {
	switch m {
	case InstallMethodHomebrew:
		return "Homebrew"
	case InstallMethodNPM:
		return "npm"
	case InstallMethodGo:
		return "Go install"
	case InstallMethodScript:
		return "Script install"
	case InstallMethodDocker:
		return "Docker"
	default:
		return "Unknown"
	}
}

// InstallInfo contains information about the current shipyard installation
type InstallInfo struct {
	Method     InstallMethod
	BinaryPath string
	Version    string
	Commit     string
	Date       string
	CanUpgrade bool
	Reason     string // Why it can't upgrade
}

// ReleaseInfo contains information about a GitHub release
type ReleaseInfo struct {
	TagName     string
	Name        string
	Body        string
	PublishedAt time.Time
	Assets      []ReleaseAsset
	Prerelease  bool
}

// ReleaseAsset represents a downloadable asset from a GitHub release
type ReleaseAsset struct {
	Name        string
	DownloadURL string
	Size        int64
}
