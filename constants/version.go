package constants

// these are replaced by ldflags during compile time on release
var (
	Version   = "development" // `MAJOR.MINOR.PATCH`, semantic versioning
	Commit    = "unknown commit"
	BuildDate = "unknown date" // `YYYY-MM-DD at HH:MM:SS`
)
