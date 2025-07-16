// Package handlers provides ecosystem-specific package management functionality.
//
// This package implements the ecosystem handler pattern, where each supported
// package ecosystem (NPM, Go, Helm, etc.) has its own handler that implements
// the EcosystemHandler interface.
//
// The handlers are responsible for:
//   - Loading package information from manifest files
//   - Updating package versions in manifest files
//   - Providing ecosystem-specific metadata
//
// Usage:
//
//	handler, ok := handlers.GetHandler(config.EcosystemNPM)
//	if !ok {
//	    return fmt.Errorf("unsupported ecosystem")
//	}
//	pkg, err := handler.LoadPackage("/path/to/package")
//
// Supported ecosystems:
//   - NPM (package.json)
//   - Go (go.mod)
//   - Helm (Chart.yaml)
//
// New ecosystems can be added by implementing the EcosystemHandler interface
// and registering with RegisterEcosystemHandler.
package handlers
