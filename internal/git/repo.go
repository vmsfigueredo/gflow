package git

import (
	"context"
	"strings"

	"github.com/vmsfigueredo/gflow/internal/executor"
	"github.com/vmsfigueredo/gflow/internal/module"
)

// IsDirty returns true if the working tree has uncommitted changes.
func IsDirty(ctx context.Context, dir string) (bool, error) {
	res, err := Run(ctx, dir, "status", "--porcelain")
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(res.Stdout) != "", nil
}

// CurrentBranch returns HEAD branch name.
func CurrentBranch(ctx context.Context, dir string) (string, error) {
	res, err := Run(ctx, dir, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", err
	}
	return res.Stdout, nil
}

// BranchExists returns true if the named branch exists locally.
func BranchExists(ctx context.Context, dir, branch string) (bool, error) {
	_, err := Run(ctx, dir, "rev-parse", "--verify", branch)
	if err != nil {
		return false, nil //nolint:nilerr
	}
	return true, nil
}

// IsGitflowInitialized checks for [gitflow "branch"] section in .git/config.
func IsGitflowInitialized(ctx context.Context, dir string) bool {
	res, _ := Run(ctx, dir, "config", "--local", "--get-regexp", `^gitflow\.branch\.`)
	return strings.TrimSpace(res.Stdout) != ""
}

// CurrentSHA returns the current HEAD commit SHA.
func CurrentSHA(ctx context.Context, dir string) (string, error) {
	res, err := Run(ctx, dir, "rev-parse", "HEAD")
	if err != nil {
		return "", err
	}
	return res.Stdout, nil
}

// Status runs git status for a module and returns an executor.Result.
// Output includes branch name and clean/modified state, followed by the
// short diff when the tree is dirty.
func Status(m *module.Module, dryRun bool) executor.Result {
	if dryRun {
		return executor.Result{Module: m, Action: "status", Status: executor.StatusDryRun, Output: "git status --short"}
	}
	ctx := context.Background()

	branch, err := CurrentBranch(ctx, m.Path)
	if err != nil {
		return executor.Result{Module: m, Action: "status", Status: executor.StatusError, Err: err}
	}

	short, err := Run(ctx, m.Path, "status", "--short")
	if err != nil {
		return executor.Result{Module: m, Action: "status", Status: executor.StatusError, Err: err}
	}

	dirty := strings.TrimSpace(short.Stdout) != ""
	state := "Clean"
	if dirty {
		state = "Modified"
	}

	out := "Branch: " + strings.TrimSpace(branch) + "\nStatus: " + state
	if dirty {
		out += "\n" + strings.TrimRight(short.Stdout, "\n")
	}

	return executor.Result{Module: m, Action: "status", Status: executor.StatusOK, Output: out}
}

// Checkout runs git checkout <branch> in module dir.
func Checkout(m *module.Module, branch string, dryRun bool) executor.Result {
	if dryRun {
		return executor.Result{Module: m, Action: "checkout", Status: executor.StatusDryRun, Output: "git checkout " + branch}
	}
	_, err := Run(context.Background(), m.Path, "checkout", branch)
	if err != nil {
		return executor.Result{Module: m, Action: "checkout", Status: executor.StatusError, Err: err}
	}
	return executor.Result{Module: m, Action: "checkout", Status: executor.StatusOK}
}

// Pull runs git pull in module dir.
func Pull(m *module.Module, remote, branch string, dryRun bool) executor.Result {
	args := []string{"pull", remote}
	if branch != "" {
		args = append(args, branch)
	}
	if dryRun {
		return executor.Result{Module: m, Action: "pull", Status: executor.StatusDryRun, Output: "git " + strings.Join(args, " ")}
	}
	_, err := Run(context.Background(), m.Path, args...)
	if err != nil {
		return executor.Result{Module: m, Action: "pull", Status: executor.StatusError, Err: err}
	}
	return executor.Result{Module: m, Action: "pull", Status: executor.StatusOK}
}

// Push runs git push in module dir.
func Push(m *module.Module, remote, branch string, dryRun bool) executor.Result {
	args := []string{"push", remote}
	if branch != "" {
		args = append(args, branch)
	}
	if dryRun {
		return executor.Result{Module: m, Action: "push", Status: executor.StatusDryRun, Output: "git " + strings.Join(args, " ")}
	}
	_, err := Run(context.Background(), m.Path, args...)
	if err != nil && strings.Contains(err.Error(), "no upstream branch") {
		trackBranch := branch
		if trackBranch == "" {
			trackBranch, _ = CurrentBranch(context.Background(), m.Path)
		}
		_, err = Run(context.Background(), m.Path, "push", "--set-upstream", remote, trackBranch)
	}
	if err != nil {
		return executor.Result{Module: m, Action: "push", Status: executor.StatusError, Err: err}
	}
	return executor.Result{Module: m, Action: "push", Status: executor.StatusOK}
}

// LatestTag returns the most recent semver-like tag reachable from HEAD.
// Returns empty string when no tag exists or git fails.
func LatestTag(ctx context.Context, dir string) string {
	res, err := Run(ctx, dir, "describe", "--tags", "--abbrev=0")
	if err != nil {
		return ""
	}
	return strings.TrimSpace(res.Stdout)
}

// Commit runs git commit with argv passthrough in module dir.
func Commit(m *module.Module, args []string, dryRun bool) executor.Result {
	full := append([]string{"commit"}, args...)
	if dryRun {
		return executor.Result{Module: m, Action: "commit", Status: executor.StatusDryRun, Output: "git " + strings.Join(full, " ")}
	}
	_, err := Run(context.Background(), m.Path, full...)
	if err != nil {
		return executor.Result{Module: m, Action: "commit", Status: executor.StatusError, Err: err}
	}
	return executor.Result{Module: m, Action: "commit", Status: executor.StatusOK}
}
