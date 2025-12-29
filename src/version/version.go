package version

var (
	// Version is the current version of the application
	// Injected via ldflags during build
	Version = "dev"

	// Commit is the git commit hash
	// Injected via ldflags during build
	Commit = "unknown"

	// Date is the build date
	// Injected via ldflags during build
	Date = "unknown"
)

// GetVersion returns the formatted version string
func GetVersion() string {
	return Version
}

// GetFullVersion returns the full version info including commit and date
func GetFullVersion() string {
	return Version + " (commit: " + Commit + ", built: " + Date + ")"
}
