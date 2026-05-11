package update

import (
	"os"
	"path/filepath"
	"strings"
)

type Source int

const (
	SourceUnknown  Source = iota
	SourceBinary          // self-managed binary in standard path
	SourceHomebrew        // managed by Homebrew
)

func (s Source) String() string {
	switch s {
	case SourceHomebrew:
		return "homebrew"
	case SourceBinary:
		return "binary"
	default:
		return "unknown"
	}
}

var homebrewPrefixes = []string{
	"/opt/homebrew/",
	"/usr/local/Cellar/",
	"/home/linuxbrew/.linuxbrew/",
}

var binaryPrefixes = []string{
	"/usr/local/bin/",
	"/usr/bin/",
}

// classifyPath maps a resolved binary path to a Source.
func classifyPath(real, home string) Source {
	for _, prefix := range homebrewPrefixes {
		if strings.HasPrefix(real, prefix) {
			return SourceHomebrew
		}
	}

	userPrefixes := []string{
		filepath.Join(home, ".local", "bin") + string(filepath.Separator),
		filepath.Join(home, "bin") + string(filepath.Separator),
	}
	for _, prefix := range append(binaryPrefixes, userPrefixes...) {
		if strings.HasPrefix(real, prefix) {
			return SourceBinary
		}
	}
	return SourceUnknown
}

// Detect resolves the real path of the running executable and classifies its install source.
func Detect() (Source, string, error) {
	exe, err := os.Executable()
	if err != nil {
		return SourceUnknown, "", err
	}

	// Resolve symlinks (Homebrew installs symlinks in /opt/homebrew/bin → Cellar).
	real, err := filepath.EvalSymlinks(exe)
	if err != nil {
		real = exe
	}

	home, _ := os.UserHomeDir()
	return classifyPath(real, home), real, nil
}
