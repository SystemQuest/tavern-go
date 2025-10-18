package version

// Version information. These variables are set at build time via ldflags.
var (
	// Version is the current version of tavern-go (e.g., "v0.1.2")
	Version = "dev"

	// GitCommit is the git commit hash
	GitCommit = "unknown"

	// BuildDate is the date when the binary was built
	BuildDate = "unknown"
)

// GetVersionInfo returns a formatted version string with all build info
func GetVersionInfo() string {
	return Version + " (commit: " + GitCommit + ", built: " + BuildDate + ")"
}
