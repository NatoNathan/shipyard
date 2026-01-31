package version

// Version information for the application
// These values are typically set at build time via ldflags
var (
	// Version is the semantic version of the application
	Version = "0.1.0"
	
	// Commit is the git commit hash
	Commit = "none"
	
	// Date is the build date
	Date = "unknown"
)
