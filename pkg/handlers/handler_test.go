package handlers

import (
	"testing"

	"github.com/NatoNathan/shipyard/pkg/config"
)

// TestEcosystemHandlers verifies that all expected ecosystem handlers are registered
func TestEcosystemHandlers(t *testing.T) {
	ecosystems := GetRegisteredEcosystems()
	expectedEcosystems := []config.PackageEcosystem{
		config.EcosystemNPM,
		config.EcosystemGo,
		config.EcosystemHelm,
	}

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

// TestLoadPackageInvalidEcosystem tests error handling for invalid ecosystem types
func TestLoadPackageInvalidEcosystem(t *testing.T) {
	_, err := LoadPackage("invalid", "/some/path")
	if err == nil {
		t.Error("Expected error for invalid ecosystem")
	}
}

// TestLoadPackageEmptyEcosystem tests error handling for empty ecosystem
func TestLoadPackageEmptyEcosystem(t *testing.T) {
	_, err := LoadPackage("", "/some/path")
	if err == nil {
		t.Error("Expected error for empty ecosystem")
	}
}

// TestHandlerInterfaces verifies that all handlers implement the interface correctly
func TestHandlerInterfaces(t *testing.T) {
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

// TestGetHandler verifies handler retrieval functionality
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

// TestRegisterEcosystemHandler tests dynamic handler registration
func TestRegisterEcosystemHandler(t *testing.T) {
	// Create a mock handler for testing
	mockHandler := &MockHandler{}

	// Register it
	RegisterEcosystemHandler(mockHandler)

	// Verify it was registered
	handler, ok := GetHandler(mockHandler.GetEcosystem())
	if !ok {
		t.Error("Expected to find mock handler after registration")
	}
	if handler != mockHandler {
		t.Error("Expected registered handler to be the same instance")
	}
}

// MockHandler is a test implementation of EcosystemHandler
type MockHandler struct{}

func (m *MockHandler) GetEcosystem() config.PackageEcosystem { return "mock" }
func (m *MockHandler) GetManifestFile() string               { return "mock.json" }
func (m *MockHandler) LoadPackage(path string) (*EcosystemPackage, error) {
	return &EcosystemPackage{Name: "mock", Path: path, Ecosystem: "mock"}, nil
}
func (m *MockHandler) UpdateVersion(path string, version string) error { return nil }
