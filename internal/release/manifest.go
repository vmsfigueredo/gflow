package release

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Masterminds/semver/v3"
)

// BumpKind identifies the semver bump type.
type BumpKind string

const (
	BumpPatch BumpKind = "patch"
	BumpMinor BumpKind = "minor"
	BumpMajor BumpKind = "major"
)

// BumpManifest detects the version manifest in dir and bumps it.
// Returns the new version string.
func BumpManifest(dir string, kind BumpKind, currentVersion string) (string, error) {
	v, err := semver.NewVersion(currentVersion)
	if err != nil {
		return "", fmt.Errorf("invalid current version %q: %w", currentVersion, err)
	}

	var next semver.Version
	switch kind {
	case BumpMajor:
		next = v.IncMajor()
	case BumpMinor:
		next = v.IncMinor()
	default:
		next = v.IncPatch()
	}
	nextStr := next.Original()

	// Try each manifest in priority order.
	if err := bumpPackageJSON(dir, currentVersion, nextStr); err == nil {
		return nextStr, nil
	}
	if err := bumpGoMod(dir, currentVersion, nextStr); err == nil {
		return nextStr, nil
	}
	if err := bumpVersionFile(dir, currentVersion, nextStr); err == nil {
		return nextStr, nil
	}
	if err := bumpCargoToml(dir, currentVersion, nextStr); err == nil {
		return nextStr, nil
	}
	if err := bumpPyproject(dir, currentVersion, nextStr); err == nil {
		return nextStr, nil
	}

	return nextStr, nil // no manifest found — version still computed
}

// DetectVersion reads the current version from the first manifest found in dir.
func DetectVersion(dir string) (string, error) {
	if v, err := readPackageJSONVersion(dir); err == nil {
		return v, nil
	}
	if v, err := readVersionFile(dir); err == nil {
		return v, nil
	}
	if v, err := readCargoVersion(dir); err == nil {
		return v, nil
	}
	if v, err := readPyprojectVersion(dir); err == nil {
		return v, nil
	}
	return "", fmt.Errorf("no version manifest found in %s", dir)
}

// ── package.json ──────────────────────────────────────────────────────────────

func readPackageJSONVersion(dir string) (string, error) {
	data, err := os.ReadFile(filepath.Join(dir, "package.json"))
	if err != nil {
		return "", err
	}
	var m map[string]json.RawMessage
	if err := json.Unmarshal(data, &m); err != nil {
		return "", err
	}
	var v string
	if err := json.Unmarshal(m["version"], &v); err != nil {
		return "", err
	}
	return v, nil
}

func bumpPackageJSON(dir, from, to string) error {
	path := filepath.Join(dir, "package.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	updated := strings.Replace(string(data), `"version": "`+from+`"`, `"version": "`+to+`"`, 1)
	return os.WriteFile(path, []byte(updated), 0o644)
}

// ── go.mod ────────────────────────────────────────────────────────────────────

func bumpGoMod(dir, from, to string) error {
	// go.mod doesn't contain module version — bump VERSION file instead.
	return bumpVersionFile(dir, from, to)
}

// ── VERSION file ──────────────────────────────────────────────────────────────

func readVersionFile(dir string) (string, error) {
	for _, name := range []string{"VERSION", "version.txt", ".version"} {
		data, err := os.ReadFile(filepath.Join(dir, name))
		if err == nil {
			v := strings.TrimSpace(string(data))
			if v != "" {
				return v, nil
			}
		}
	}
	return "", fmt.Errorf("no VERSION file")
}

func bumpVersionFile(dir, from, to string) error {
	for _, name := range []string{"VERSION", "version.txt", ".version"} {
		path := filepath.Join(dir, name)
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		updated := strings.ReplaceAll(string(data), from, to)
		return os.WriteFile(path, []byte(updated), 0o644)
	}
	return fmt.Errorf("no VERSION file found")
}

// ── Cargo.toml ────────────────────────────────────────────────────────────────

var cargoVersionRe = regexp.MustCompile(`(?m)^version\s*=\s*"([^"]+)"`)

func readCargoVersion(dir string) (string, error) {
	data, err := os.ReadFile(filepath.Join(dir, "Cargo.toml"))
	if err != nil {
		return "", err
	}
	m := cargoVersionRe.FindSubmatch(data)
	if m == nil {
		return "", fmt.Errorf("no version in Cargo.toml")
	}
	return string(m[1]), nil
}

func bumpCargoToml(dir, from, to string) error {
	path := filepath.Join(dir, "Cargo.toml")
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	updated := cargoVersionRe.ReplaceAllStringFunc(string(data), func(match string) string {
		if strings.Contains(match, from) {
			return strings.Replace(match, from, to, 1)
		}
		return match
	})
	return os.WriteFile(path, []byte(updated), 0o644)
}

// ── pyproject.toml ────────────────────────────────────────────────────────────

var pyprojectVersionRe = regexp.MustCompile(`(?m)^version\s*=\s*"([^"]+)"`)

func readPyprojectVersion(dir string) (string, error) {
	data, err := os.ReadFile(filepath.Join(dir, "pyproject.toml"))
	if err != nil {
		return "", err
	}
	m := pyprojectVersionRe.FindSubmatch(data)
	if m == nil {
		return "", fmt.Errorf("no version in pyproject.toml")
	}
	return string(m[1]), nil
}

func bumpPyproject(dir, from, to string) error {
	path := filepath.Join(dir, "pyproject.toml")
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	updated := strings.Replace(string(data), `version = "`+from+`"`, `version = "`+to+`"`, 1)
	return os.WriteFile(path, []byte(updated), 0o644)
}
