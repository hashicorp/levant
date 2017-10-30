package version

import "fmt"

// Version is the main version number that is being run at the moment.
const Version = "0.0.1"

// VersionPrerelease is a pre-release marker for the version. If this is ""
// (empty string) then it means that it is a final release. Otherwise, this is
// a pre-release such as "dev" (in development), "beta", "rc1", etc.
const VersionPrerelease = ""

// Get returns a human readable version of Levant.
func Get() string {

	if VersionPrerelease != "" {
		return fmt.Sprintf("%s-%s", Version, VersionPrerelease)
	}

	return Version
}
