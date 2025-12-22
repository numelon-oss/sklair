package constants

var (
	Version   = "0.1.0"   // to be replaced by ldflags during compile time in a GitHub runner
	Commit    = "unknown" // likewise
	BuildDate = "unknown" // likewise, ISO-8601 UTC
)

/*
VERSION=v0.0.0
COMMIT=$(git rev-parse --short HEAD)
DATE=$(date -u +%Y-%m-%dT%H:%M:%SZ)

go build \
  -ldflags "\
    -X sklair/version.Version=$VERSION \
    -X sklair/version.Commit=$COMMIT \
    -X sklair/version.BuildDate=$DATE \
  "
*/
