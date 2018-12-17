package version

import (
	"fmt"

	"github.com/Masterminds/semver"
)

// Git SHA1 commit hash of the release (set via linker flags - ldflags)
var gitCommit = "master"

// CentrifugeNodeVersion is the current version of the app
const CentrifugeNodeVersion = "0.0.1-alpha"

// GetVersion returns current cent node version in semvar format.
func GetVersion() *semver.Version {
	v, err := semver.NewVersion(fmt.Sprintf("%s+%s", CentrifugeNodeVersion, gitCommit))
	if err != nil {
		log.Panicf("Invalid CentrifugeNodeVersion specified: %s", CentrifugeNodeVersion)
	}

	return v
}
