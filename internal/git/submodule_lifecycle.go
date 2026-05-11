package git

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// SubmoduleAdd adds a new submodule at subPath from url, optionally pinned to branch.
// Runs: git submodule add [-b branch] <url> <subPath>
func SubmoduleAdd(ctx context.Context, repoDir, url, subPath, branch string) error {
	args := []string{"submodule", "add"}
	if branch != "" {
		args = append(args, "-b", branch)
	}
	args = append(args, url, subPath)
	_, err := Run(ctx, repoDir, args...)
	return err
}

// SubmoduleRemove fully removes a submodule: deinit, git rm, remove .git/modules/<path>.
func SubmoduleRemove(ctx context.Context, repoDir, subPath string) error {
	// deinit
	if _, err := Run(ctx, repoDir, "submodule", "deinit", "-f", subPath); err != nil {
		return fmt.Errorf("submodule deinit: %w", err)
	}
	// git rm
	if _, err := Run(ctx, repoDir, "rm", "-f", subPath); err != nil {
		return fmt.Errorf("git rm submodule: %w", err)
	}
	// remove .git/modules/<subPath>
	modulesDir := filepath.Join(repoDir, ".git", "modules", subPath)
	_ = os.RemoveAll(modulesDir) // best-effort
	return nil
}

// SubmoduleMove moves a submodule from oldPath to newPath using git mv.
func SubmoduleMove(ctx context.Context, repoDir, oldPath, newPath string) error {
	if _, err := Run(ctx, repoDir, "mv", oldPath, newPath); err != nil {
		return fmt.Errorf("git mv submodule: %w", err)
	}
	// sync .gitmodules paths
	if _, err := Run(ctx, repoDir, "submodule", "sync"); err != nil {
		return fmt.Errorf("submodule sync after move: %w", err)
	}
	return nil
}

// SubmoduleSync runs git submodule sync --recursive to refresh URLs from .gitmodules.
func SubmoduleSync(ctx context.Context, repoDir string) error {
	_, err := Run(ctx, repoDir, "submodule", "sync", "--recursive")
	return err
}

// DriftEntry describes a submodule whose registered pointer differs from its HEAD.
type DriftEntry struct {
	Name      string
	Path      string
	Registered string // SHA stored in parent tree
	Actual    string // SHA at submodule HEAD
	Detached  bool
}

// SubmoduleDrift returns submodules where HEAD ≠ parent-registered pointer.
func SubmoduleDrift(ctx context.Context, repoDir string) ([]DriftEntry, error) {
	// git submodule status: prefixes ' '=ok, '-'=not init, '+'=drift, 'U'=conflict
	res, err := Run(ctx, repoDir, "submodule", "status", "--recursive")
	if err != nil {
		return nil, err
	}

	var drifts []DriftEntry
	for _, line := range strings.Split(res.Stdout, "\n") {
		if len(line) < 42 {
			continue
		}
		prefix := string(line[0])
		if prefix == " " || prefix == "-" {
			continue // ok or uninit — not drift
		}
		// '+' or 'U'
		rest := strings.TrimSpace(line[1:])
		parts := strings.Fields(rest)
		if len(parts) < 2 {
			continue
		}
		sha := parts[0]
		subPath := parts[1]
		absPath := filepath.Join(repoDir, subPath)
		actual, _ := CurrentSHA(ctx, absPath)

		// registered SHA = SHA in parent index (ls-files --stage)
		regRes, _ := Run(ctx, repoDir, "ls-files", "--stage", subPath)
		regSHA := ""
		if f := strings.Fields(regRes.Stdout); len(f) >= 2 {
			regSHA = f[1]
		}

		drifts = append(drifts, DriftEntry{
			Name:       subPath,
			Path:       absPath,
			Registered: regSHA,
			Actual:     actual,
			Detached:   sha != actual,
		})
	}
	return drifts, nil
}
