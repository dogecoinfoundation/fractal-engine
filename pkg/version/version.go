package version

var (
	// These are set at build time via -ldflags
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)

func String() string {
	return Version + " (" + Commit + ") " + Date
}
