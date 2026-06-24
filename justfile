# Shipyard CLI Development Commands
# https://github.com/casey/just

GOLANGCI_LINT_VERSION := "v2.12.2"
GOSEC_VERSION := "v2.27.1"
GOVULNCHECK_VERSION := "v1.3.0"

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

# Run tests via Dagger
test-dagger:
    dagger call test --source=. --show-output

# Run tests with race detection via Dagger (slower, uses CGO)
test-dagger-race:
    dagger call test --source=. --show-output --race

# Run unit tests only
test-unit:
    go test -v -race -cover -short ./...

# Run integration tests only
test-integration:
    go test -v -race -cover -run Integration ./test/integration/...

# Run contract tests only
test-contract:
    go test -v -race -cover -tags contract ./test/contract/...

# Run linters
lint:
    go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@{{GOLANGCI_LINT_VERSION}} run

# Run linters via Dagger
lint-dagger:
    dagger call lint --source=.

# Format code
fmt:
    go fmt ./...
    gofmt -s -w .

# Run pinned security scanners
security:
    go run github.com/securego/gosec/v2/cmd/gosec@{{GOSEC_VERSION}} ./...
    go run golang.org/x/vuln/cmd/govulncheck@{{GOVULNCHECK_VERSION}} ./...

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

# Run all CI checks (lint, test, verify, security)
ci: lint test verify security
    @echo "All CI checks passed!"

# Run all CI checks via Dagger
ci-dagger:
    dagger call ci --source=.

# Run tests and export coverage file via Dagger
coverage-dagger:
    dagger call coverage --source=. export --path=./coverage.out

# Run coverage report via Dagger
coverage-report-dagger:
    dagger call coverage-report --source=.

# Run coverage with threshold check via Dagger
coverage-check-dagger threshold="80":
    dagger call coverage-report --source=. --threshold={{ threshold }}

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
    go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@{{GOLANGCI_LINT_VERSION}}
    go install github.com/securego/gosec/v2/cmd/gosec@{{GOSEC_VERSION}}
    go install golang.org/x/vuln/cmd/govulncheck@{{GOVULNCHECK_VERSION}}
    @echo "Development environment ready!"

# Release build with version information
release VERSION:
    @echo "Building release {{ VERSION }}..."
    go build -ldflags "-X main.version={{ VERSION }} -X main.commit=$(git rev-parse HEAD) -X main.date=$(date -u +%Y-%m-%dT%H:%M:%SZ)" -o bin/shipyard ./cmd/shipyard
    @echo "Release {{ VERSION }} built successfully!"

# Test Dagger build stage
dagger-build:
    dagger call build-only --source=. --version=v0.0.0-dev

# Test Dagger package stage (exports to ./dist)
dagger-package:
    dagger call package-only --source=. --version=v0.0.0-dev export --path=./dist

# Test full Dagger release pipeline (requires tokens)
dagger-test-release:
    dagger call release \
      --source=. \
      --version=v0.0.0-test \
      --github-token=env:GITHUB_TOKEN \
      --npm-token=env:NPM_TOKEN \
      --docker-registry=ghcr.io/natonathan/shipyard \
      --docker-token=env:GITHUB_TOKEN
