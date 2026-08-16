// Package version exposes build metadata injected at compile time via
// `go build -ldflags "-X vidhya-service/src/version.BuildNumber=... -X vidhya-service/src/version.BuildCommitSHA=..."`
// (see Dockerfile). Both vars default to dev-friendly values when built
// without ldflags, e.g. via `go run ./src`.
package version

var (
	// BuildNumber is the release/build version, e.g. "1.4.0" or a CI build number.
	BuildNumber = "0.1.0-dev"
	// BuildCommitSHA is the git commit the binary was built from.
	BuildCommitSHA = "unknown"
)
