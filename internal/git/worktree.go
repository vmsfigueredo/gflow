package git

import (
	"context"
	"strings"
)

// WorktreeEntry represents one git worktree.
type WorktreeEntry struct {
	Path   string
	Branch string
	HEAD   string
	Bare   bool
}

// WorktreeAdd creates a new worktree for branch at targetPath.
// If branch does not exist locally it is created tracking remote/branch.
func WorktreeAdd(ctx context.Context, repoDir, targetPath, branch string) error {
	// Try to add with existing branch first.
	_, err := Run(ctx, repoDir, "worktree", "add", targetPath, branch)
	if err != nil {
		// Branch may not exist locally — create it.
		_, err = Run(ctx, repoDir, "worktree", "add", "-b", branch, targetPath)
	}
	return err
}

// WorktreeList returns all worktrees for repoDir.
func WorktreeList(ctx context.Context, repoDir string) ([]WorktreeEntry, error) {
	res, err := Run(ctx, repoDir, "worktree", "list", "--porcelain")
	if err != nil {
		return nil, err
	}
	return parseWorktreePorcelain(res.Stdout), nil
}

// WorktreeRemove removes worktree at targetPath and prunes stale entries.
func WorktreeRemove(ctx context.Context, repoDir, targetPath string) error {
	if _, err := Run(ctx, repoDir, "worktree", "remove", "--force", targetPath); err != nil {
		return err
	}
	_, _ = Run(ctx, repoDir, "worktree", "prune")
	return nil
}

func parseWorktreePorcelain(output string) []WorktreeEntry {
	var entries []WorktreeEntry
	var cur WorktreeEntry

	for _, line := range strings.Split(output, "\n") {
		switch {
		case strings.HasPrefix(line, "worktree "):
			if cur.Path != "" {
				entries = append(entries, cur)
			}
			cur = WorktreeEntry{Path: strings.TrimPrefix(line, "worktree ")}
		case strings.HasPrefix(line, "HEAD "):
			cur.HEAD = strings.TrimPrefix(line, "HEAD ")
		case strings.HasPrefix(line, "branch "):
			cur.Branch = strings.TrimPrefix(line, "branch ")
			cur.Branch = strings.TrimPrefix(cur.Branch, "refs/heads/")
		case line == "bare":
			cur.Bare = true
		}
	}
	if cur.Path != "" {
		entries = append(entries, cur)
	}
	return entries
}
