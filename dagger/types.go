package main

// Platform represents a target OS and architecture for building
type Platform struct {
	OS   string
	Arch string
}

// SupportedPlatforms lists all platforms to build for
var SupportedPlatforms = []Platform{
	{OS: "linux", Arch: "amd64"},
	{OS: "linux", Arch: "arm64"},
	{OS: "darwin", Arch: "amd64"},
	{OS: "darwin", Arch: "arm64"},
	{OS: "windows", Arch: "amd64"},
}

// BuildInfo contains version metadata to embed in binaries
type BuildInfo struct {
	Version string
	Commit  string
	Date    string
}

const (
	GoImage             = "golang:1.25.11-alpine"
	GoImageCGO          = "golang:1.25.11"
	GolangCILintVersion = "v2.12.2"
	GosecVersion        = "v2.27.1"
	GovulncheckVersion  = "v1.3.0"
)
