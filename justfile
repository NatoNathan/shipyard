# Shipyard Build System
# This file contains common tasks for building, testing, and managing the Shipyard project.

# Variables
gitCommit := `git rev-parse --short HEAD 2>/dev/null || echo "unknown"`
buildDate := `date -u +%Y-%m-%dT%H:%M:%SZ`
devVersion := `git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0-dev"`

# Default recipe - show help
default:
    @just --list

# Development recipes
[group("dev")]
run +OPTIONS:
    go run ./cmd/shipyard/main.go {{OPTIONS}}

[group("dev")]
fmt:
    go fmt ./...

[group("dev")]
vet:
    go vet ./...

[group("dev")]
mod-tidy:
    go mod tidy

[group("dev")]
dev-setup: mod-tidy fmt vet
    @echo "Development setup complete"

# Build recipes
[group("build")]
clean:
    rm -rf dist

[group("build")]
build VERSION=devVersion:
    @echo "Building Shipyard..."
    @echo "  Version: {{VERSION}}"
    @echo "  Git Commit: {{gitCommit}}"
    @echo "  Build Date: {{buildDate}}"
    @mkdir -p dist
    go build \
        -ldflags "-X 'github.com/NatoNathan/shipyard/internal/cli.Version={{VERSION}}' \
                  -X 'github.com/NatoNathan/shipyard/internal/cli.GitCommit={{gitCommit}}' \
                  -X 'github.com/NatoNathan/shipyard/internal/cli.BuildDate={{buildDate}}'" \
        -o dist/shipyard \
        ./cmd/shipyard/main.go
    @echo "Build complete: dist/shipyard"

[group("build")]
build-all: clean build
    @echo "Full build complete"

# Test recipes
[group("test")]
test TEST="./...":
    go test {{TEST}}

[group("test")]
test-verbose TEST="./...":
    go test -v {{TEST}}

[group("test")]
test-coverage TEST="./...":
    go test -cover {{TEST}}

[group("test")]
test-race TEST="./...":
    go test -race {{TEST}}

[group("test")]
test-all: test-race test-coverage
    @echo "All tests complete"

# Quality assurance recipes
[group("qa")]
lint:
    golangci-lint run

[group("qa")]
security:
    gosec ./...

[group("qa")]
check-deps:
    go mod tidy
    go mod verify

[group("qa")]
qa-all: fmt vet lint security check-deps test-all
    @echo "Quality assurance complete"

# Release recipes
[group("release")]
tag VERSION:
    git tag {{VERSION}}
    git push origin {{VERSION}}
    @echo "Tagged version {{VERSION}}"

[group("release")]
release-prep VERSION: (build VERSION) test-all
    @echo "Release {{VERSION}} prepared"

# Future implementation for shipment history
[group("future")]
git-tags:
    @echo "Creating git tags for shipment history..."
    @echo "This will be implemented in the future to create annotated git tags"
    @echo "with shipyard-history prefix for each shipment instead of JSON files"
