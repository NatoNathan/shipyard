
run +OPTIONS:
	go run ./cmd/shipyard/main.go {{OPTIONS}}

clean:
    rm -rf dist

gitCommit := `git rev-parse --short HEAD 2>/dev/null || echo "unknown"`
buildDate := `date -u +%Y-%m-%dT%H:%M:%SZ`
devVersion := `git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0-dev"`

# Build with version information
[group("build")]
build VERSION=devVersion:
    echo "Building with version: {{VERSION}}, git commit: {{gitCommit}}, build date: {{buildDate}}"
    mkdir -p dist
    go build \
        -ldflags "-X 'github.com/NatoNathan/shipyard/internal/cli.Version={{VERSION}}' \
                  -X 'github.com/NatoNathan/shipyard/internal/cli.GitCommit={{gitCommit}}' \
                  -X 'github.com/NatoNathan/shipyard/internal/cli.BuildDate={{buildDate}}'" \
        -o dist/shipyard \
        ./cmd/shipyard/main.go
# Run tests
[group("test")]
test TEST="./...":
    go test {{TEST}} 
