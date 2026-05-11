package module

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	"github.com/vmsfigueredo/gflow/internal/config"
)

// Resolve returns the ordered list of modules to operate on.
// Priority: explicit MODULES list in config > .gitmodules discovery.
// --project/-P resolves an alias and limits to matching paths (repeatable, union).
// --no-root excludes the superproject root. Root is also excluded when -P is used.
// --except/-E excludes named modules (by Name or Display) from the final list.
func Resolve(cfg *config.Config, root string, projects []string, noRoot bool, except []string) ([]*Module, error) {
	if root == "" {
		var err error
		root, err = os.Getwd()
		if err != nil {
			return nil, err
		}
	}

	var paths []string

	if len(projects) > 0 {
		seen := map[string]bool{}
		for _, project := range projects {
			var resolved []string
			if alias, ok := cfg.Aliases[project]; ok {
				resolved = alias
			} else {
				resolved = []string{project}
			}
			for _, p := range resolved {
				if !seen[p] {
					seen[p] = true
					paths = append(paths, p)
				}
			}
		}
		noRoot = true
	} else if len(cfg.Modules) > 0 {
		paths = cfg.Modules
	} else {
		// auto-discover from .gitmodules
		discovered, err := fromGitmodules(root)
		if err != nil {
			return nil, err
		}
		paths = discovered
	}

	var mods []*Module
	noRootEffective := noRoot || cfg.NoRoot

	aliasIdx := buildAliasIndex(cfg.Aliases)

	// include root unless excluded
	if !noRootEffective {
		mods = append(mods, &Module{Name: ".", Display: ".", Path: root, Root: true})
	}

	for _, p := range paths {
		abs := p
		if !filepath.IsAbs(p) {
			abs = filepath.Join(root, p)
		}
		name := filepath.Base(p)
		mods = append(mods, &Module{Name: name, Display: displayFor(p, aliasIdx), Path: abs})
	}

	if len(except) > 0 {
		excluded := make(map[string]bool, len(except))
		for _, e := range except {
			excluded[e] = true
		}
		filtered := mods[:0]
		for _, m := range mods {
			if !excluded[m.Name] && !excluded[m.Display] {
				filtered = append(filtered, m)
			}
		}
		mods = filtered
	}

	return mods, nil
}

// buildAliasIndex builds a path→alias reverse map. Only paths that are the
// sole entry in a single-element alias list are eligible (avoids ambiguity
// when an alias groups multiple paths).
func buildAliasIndex(aliases map[string][]string) map[string]string {
	idx := map[string]string{}
	for alias, paths := range aliases {
		if len(paths) == 1 {
			idx[paths[0]] = alias
		}
	}
	return idx
}

// displayFor returns the human-facing label for a submodule path.
func displayFor(p string, aliasIdx map[string]string) string {
	if a, ok := aliasIdx[p]; ok {
		return a
	}
	dir := filepath.Dir(p)
	if dir == "." || dir == "" || dir == "/" {
		return filepath.Base(p)
	}
	return filepath.Base(dir) + "/" + filepath.Base(p)
}

// fromGitmodules parses .gitmodules for submodule paths.
func fromGitmodules(root string) ([]string, error) {
	path := filepath.Join(root, ".gitmodules")
	f, err := os.Open(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var paths []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "path") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				paths = append(paths, strings.TrimSpace(parts[1]))
			}
		}
	}
	return paths, scanner.Err()
}
