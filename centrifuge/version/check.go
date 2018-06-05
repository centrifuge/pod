package version

import (
	"github.com/Masterminds/semver"
	logging "github.com/ipfs/go-log"
)

var log = logging.Logger("centrifuge-version")

func SemverNodeVersion() *semver.Version {
	v, err := semver.NewVersion(CENTRIFUGE_NODE_VERSION)

	if err != nil {
		log.Panicf("Invalid CENTRIFUGE_NODE_VERSION specified: %s", CENTRIFUGE_NODE_VERSION)
	}
	return v

}

func CheckMajorCompatibility(versionString string) (match bool, err error) {
	v, err := semver.NewVersion(versionString)
	if err != nil {
		return false, err
	}
	return v.Major() == SemverNodeVersion().Major(), nil
}
