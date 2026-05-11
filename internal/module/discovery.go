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
// --project/-P resolves an alias and limits to matching paths.
// --no-root excludes the superproject root.
func Resolve(cfg *config.Config, root, project string, noRoot bool) ([]*Module, error) {
	if root == "" {
		var err error
		root, err = os.Getwd()
		if err != nil {
			return nil, err
		}
	}

	var paths []string

	if project != "" {
		if alias, ok := cfg.Aliases[project]; ok {
			paths = alias
		} else {
			paths = []string{project}
		}
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

	// include root unless excluded
	if !noRootEffective {
		mods = append(mods, &Module{Name: ".", Path: root, Root: true})
	}

	for _, p := range paths {
		abs := p
		if !filepath.IsAbs(p) {
			abs = filepath.Join(root, p)
		}
		name := filepath.Base(p)
		mods = append(mods, &Module{Name: name, Path: abs})
	}

	return mods, nil
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
