package update

import (
	"github.com/Masterminds/semver/v3"
)

// IsNewer reports whether latest is strictly newer than current.
// "dev" is always treated as outdated so local builds can be force-updated.
func IsNewer(latest, current string) (bool, error) {
	if current == "dev" {
		return true, nil
	}
	l, err := semver.NewVersion(latest)
	if err != nil {
		return false, err
	}
	c, err := semver.NewVersion(current)
	if err != nil {
		return false, err
	}
	return l.GreaterThan(c), nil
}
