package handlers

import (
	"testing"

	"github.com/NatoNathan/shipyard/pkg/config"
)

func TestEcosystemHandlers(t *testing.T) {
	// Test that all expected handlers are registered
	ecosystems := GetRegisteredEcosystems()
	expectedEcosystems := []config.PackageEcosystem{config.EcosystemNPM, config.EcosystemGo, config.EcosystemHelm}

	if len(ecosystems) != len(expectedEcosystems) {
		t.Errorf("Expected %d ecosystems, got %d", len(expectedEcosystems), len(ecosystems))
	}

	for _, expected := range expectedEcosystems {
		found := false
		for _, actual := range ecosystems {
			if actual == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected ecosystem %s not found in registered ecosystems", expected)
		}
	}
}

func TestLoadPackageInvalidEcosystem(t *testing.T) {
	_, err := LoadPackage("invalid", "/some/path")
	if err == nil {
		t.Error("Expected error for invalid ecosystem")
	}
}

func TestLoadPackageEmptyEcosystem(t *testing.T) {
	_, err := LoadPackage("", "/some/path")
	if err == nil {
		t.Error("Expected error for empty ecosystem")
	}
}

func TestHandlerInterfaces(t *testing.T) {
	// Test that all handlers implement the interface correctly
	handlers := []EcosystemHandler{
		&NPMHandler{},
		&GoHandler{},
		&HelmHandler{},
	}

	for _, handler := range handlers {
		if handler.GetEcosystem() == "" {
			t.Errorf("Handler %T returned empty ecosystem", handler)
		}
		if handler.GetManifestFile() == "" {
			t.Errorf("Handler %T returned empty manifest file", handler)
		}
	}
}

func TestGetHandler(t *testing.T) {
	// Test getting a valid handler
	handler, ok := GetHandler(config.EcosystemNPM)
	if !ok {
		t.Error("Expected to find NPM handler")
	}
	if handler.GetEcosystem() != config.EcosystemNPM {
		t.Errorf("Expected NPM ecosystem, got %s", handler.GetEcosystem())
	}

	// Test getting an invalid handler
	_, ok = GetHandler("invalid")
	if ok {
		t.Error("Expected not to find invalid handler")
	}
}
