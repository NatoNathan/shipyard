package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/NatoNathan/shipyard/internal/config"
	"github.com/NatoNathan/shipyard/internal/ecosystem"
	"github.com/NatoNathan/shipyard/pkg/semver"
)

// GetEcosystemHandler returns the appropriate ecosystem handler for a package
func GetEcosystemHandler(pkg config.Package, pkgPath string) (ecosystem.Handler, error) {
	switch pkg.Ecosystem {
	case config.EcosystemGo:
		if pkg.IsTagOnly() {
			return ecosystem.NewGoEcosystemWithOptions(pkgPath, &ecosystem.GoEcosystemOptions{TagOnly: true}), nil
		}
		return ecosystem.NewGoEcosystem(pkgPath), nil
	case config.EcosystemNPM:
		return ecosystem.NewNPMEcosystem(pkgPath), nil
	case config.EcosystemPython:
		return ecosystem.NewPythonEcosystem(pkgPath), nil
	case config.EcosystemHelm:
		return ecosystem.NewHelmEcosystem(pkgPath), nil
	case config.EcosystemCargo:
		return ecosystem.NewCargoEcosystem(pkgPath), nil
	case config.EcosystemDeno:
		return ecosystem.NewDenoEcosystem(pkgPath), nil
	default:
		return nil, fmt.Errorf("unsupported ecosystem: %s", pkg.Ecosystem)
	}
}

// ReadAllCurrentVersions reads current versions for all configured packages
func ReadAllCurrentVersions(projectPath string, cfg *config.Config) (map[string]semver.Version, error) {
	versions := make(map[string]semver.Version)
	for _, pkg := range cfg.Packages {
		pkgPath := filepath.Join(projectPath, pkg.Path)
		handler, err := GetEcosystemHandler(pkg, pkgPath)
		if err != nil {
			return nil, err
		}
		ver, err := handler.ReadVersion()
		if err != nil {
			return nil, fmt.Errorf("failed to read version for %s: %w", pkg.Name, err)
		}
		versions[pkg.Name] = ver
	}
	return versions, nil
}

// CollectVersionFiles collects all version files that should be staged for the given packages
func CollectVersionFiles(projectPath string, cfg *config.Config, packageNames map[string]bool) ([]string, error) {
	var files []string
	for _, pkg := range cfg.Packages {
		if !packageNames[pkg.Name] {
			continue
		}
		pkgPath := filepath.Join(projectPath, pkg.Path)
		handler, err := GetEcosystemHandler(pkg, pkgPath)
		if err != nil {
			return nil, err
		}
		for _, vf := range handler.GetVersionFiles() {
			files = append(files, filepath.Join(pkgPath, vf))
		}
		// Add changelog if it exists
		changelogPath := filepath.Join(pkgPath, "CHANGELOG.md")
		if _, err := os.Stat(changelogPath); err == nil {
			files = append(files, changelogPath)
		}
	}
	return files, nil
}
