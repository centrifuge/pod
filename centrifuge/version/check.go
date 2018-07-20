package version

import (
	"github.com/Masterminds/semver"
	logging "github.com/ipfs/go-log"
)

var log = logging.Logger("centrifuge-version")

// CheckMajorCompatibility ensures that a version string matches the major version of
// the app.
func CheckMajorCompatibility(versionString string) (match bool, err error) {
	v, err := semver.NewVersion(versionString)
	if err != nil {
		return false, err
	}
	return v.Major() == GetVersion().Major(), nil
}
