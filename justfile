# Shipyard CLI Development Commands
# https://github.com/casey/just

# Default recipe to display help
default:
    @just --list

# Build the CLI binary
build:
    go build -o bin/shipyard ./cmd/shipyard

# Install the CLI to $GOPATH/bin
install:
    go install ./cmd/shipyard

# Run all tests
test:
    go test -v -race -cover ./...

# Run unit tests only
test-unit:
    go test -v -race -cover -short ./...

# Run integration tests only
test-integration:
    go test -v -race -cover -run Integration ./tests/integration/...

# Run contract tests only
test-contract:
    go test -v -race -cover -tags contract ./tests/contract/...

# Run linters
lint:
    golangci-lint run

# Format code
fmt:
    go fmt ./...
    gofmt -s -w .

# Run security scanner
security:
    gosec ./...

# Clean build artifacts
clean:
    rm -rf bin/
    go clean

# Run benchmarks
bench:
    go test -bench=. -benchmem ./...

# Generate code coverage report
coverage:
    go test -coverprofile=coverage.out ./...
    go tool cover -html=coverage.out -o coverage.html
    @echo "Coverage report generated: coverage.html"

# Check for outdated dependencies
check-deps:
    go list -u -m all

# Update dependencies
update-deps:
    go get -u ./...
    go mod tidy

# Verify go.mod is tidy
verify:
    go mod verify
    go mod tidy
    git diff --exit-code go.mod go.sum

# Run all CI checks (lint, test, verify)
ci: lint test verify
    @echo "All CI checks passed!"

# Build for multiple platforms
build-all:
    GOOS=linux GOARCH=amd64 go build -o bin/shipyard-linux-amd64 ./cmd/shipyard
    GOOS=linux GOARCH=arm64 go build -o bin/shipyard-linux-arm64 ./cmd/shipyard
    GOOS=darwin GOARCH=amd64 go build -o bin/shipyard-darwin-amd64 ./cmd/shipyard
    GOOS=darwin GOARCH=arm64 go build -o bin/shipyard-darwin-arm64 ./cmd/shipyard
    GOOS=windows GOARCH=amd64 go build -o bin/shipyard-windows-amd64.exe ./cmd/shipyard
    @echo "Built for all platforms in bin/"

# Generate mocks for testing (if using mockgen)
mocks:
    @echo "Generating mocks..."
    go generate ./...

# Run the CLI with example config
run *ARGS: build
    go run ./cmd/shipyard {{ ARGS }}

# Watch for changes and run tests
watch:
    @echo "Watching for changes..."
    find . -name '*.go' | entr -c just test-unit

# Initialize development environment
dev-setup:
    go mod download
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
    go install github.com/securego/gosec/v2/cmd/gosec@latest
    @echo "Development environment ready!"

# Release build with version information
release VERSION:
    @echo "Building release {{ VERSION }}..."
    go build -ldflags "-X main.version={{ VERSION }} -X main.commit=$(git rev-parse HEAD) -X main.date=$(date -u +%Y-%m-%dT%H:%M:%SZ)" -o bin/shipyard ./cmd/shipyard
    @echo "Release {{ VERSION }} built successfully!"

# Test Dagger build stage
dagger-build:
    ./bin/dagger call -m ./dagger build-only --source=. --version=v0.0.0-dev

# Test Dagger package stage (exports to ./dist)
dagger-package:
    ./bin/dagger call -m ./dagger package-only --source=. --version=v0.0.0-dev export --path=./dist

# Test full Dagger release pipeline (requires tokens)
dagger-test-release:
    ./bin/dagger call -m ./dagger release \
      --source=. \
      --version=v0.0.0-test \
      --github-token=env:GITHUB_TOKEN \
      --npm-token=env:NPM_TOKEN \
      --docker-registry=ghcr.io/natonathan/shipyard \
      --docker-token=env:GITHUB_TOKEN
