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
	return GetEcosystemHandlerWithContext(pkg, pkgPath, nil)
}

// GetEcosystemHandlerWithContext returns the appropriate ecosystem handler with optional context
func GetEcosystemHandlerWithContext(pkg config.Package, pkgPath string, ctx *ecosystem.HandlerContext) (ecosystem.Handler, error) {
	var handler ecosystem.Handler

	switch pkg.Ecosystem {
	case config.EcosystemGo:
		if pkg.IsTagOnly() {
			handler = ecosystem.NewGoEcosystemWithOptions(pkgPath, &ecosystem.GoEcosystemOptions{TagOnly: true})
		} else {
			handler = ecosystem.NewGoEcosystem(pkgPath)
		}
	case config.EcosystemNPM:
		handler = ecosystem.NewNPMEcosystem(pkgPath)
	case config.EcosystemPython:
		handler = ecosystem.NewPythonEcosystem(pkgPath)
	case config.EcosystemHelm:
		handler = ecosystem.NewHelmEcosystem(pkgPath)
	case config.EcosystemCargo:
		handler = ecosystem.NewCargoEcosystem(pkgPath)
	case config.EcosystemDeno:
		handler = ecosystem.NewDenoEcosystem(pkgPath)
	default:
		return nil, fmt.Errorf("unsupported ecosystem: %s", pkg.Ecosystem)
	}

	// Set context if handler supports it and context is provided
	if ctx != nil {
		if hwc, ok := handler.(ecosystem.HandlerWithContext); ok {
			hwc.SetContext(ctx)
		}
	}

	return handler, nil
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
