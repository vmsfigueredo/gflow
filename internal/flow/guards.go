package flow

import (
	"context"
	"fmt"

	"github.com/Masterminds/semver/v3"
	"github.com/vmsfigueredo/gflow/internal/git"
	"github.com/vmsfigueredo/gflow/internal/module"
)

// Guard checks preconditions before an operation runs on a module.
type Guard interface {
	Check(ctx context.Context, m *module.Module) error
}

// CleanTreeGuard refuses if working tree is dirty.
type CleanTreeGuard struct {
	Force bool // bypass check entirely
	Stash bool // auto-stash before, pop after (handled by orchestrator)
}

func (g CleanTreeGuard) Check(ctx context.Context, m *module.Module) error {
	if g.Force {
		return nil
	}
	dirty, err := git.IsDirty(ctx, m.Path)
	if err != nil {
		return err
	}
	if dirty {
		return fmt.Errorf("module %s has uncommitted changes (use --force to bypass or --stash to auto-stash)", m.Display)
	}
	return nil
}

// OnExpectedBranchGuard verifies HEAD is on the expected base branch.
type OnExpectedBranchGuard struct {
	Expected string
}

func (g OnExpectedBranchGuard) Check(ctx context.Context, m *module.Module) error {
	current, err := git.CurrentBranch(ctx, m.Path)
	if err != nil {
		return err
	}
	if current != g.Expected {
		return fmt.Errorf("module %s: expected branch %q, got %q", m.Display, g.Expected, current)
	}
	return nil
}

// RemoteSyncGuard verifies the local branch is in sync with remote before finish.
type RemoteSyncGuard struct {
	Remote string
	Branch string
}

func (g RemoteSyncGuard) Check(ctx context.Context, m *module.Module) error {
	synced, err := git.IsRemoteSynced(ctx, m.Path, g.Remote, g.Branch)
	if err != nil {
		return err
	}
	if !synced {
		return fmt.Errorf("module %s: branch %q is not synced with %s/%s — pull first",
			m.Display, g.Branch, g.Remote, g.Branch)
	}
	return nil
}

// SemverGuard validates that name is a valid semver string.
type SemverGuard struct {
	Name string
}

func (g SemverGuard) Check(_ context.Context, m *module.Module) error {
	if _, err := semver.NewVersion(g.Name); err != nil {
		return fmt.Errorf("module %s: %q is not a valid semver version", m.Display, g.Name)
	}
	return nil
}

// BranchExistsPolicy controls what to do when a branch already exists.
type BranchExistsPolicy string

const (
	PolicySkip     BranchExistsPolicy = "skip"
	PolicyCheckout BranchExistsPolicy = "checkout"
	PolicyError    BranchExistsPolicy = "error"
)

// RunGuards executes all guards in order; returns first error.
func RunGuards(ctx context.Context, m *module.Module, guards []Guard) error {
	for _, g := range guards {
		if err := g.Check(ctx, m); err != nil {
			return err
		}
	}
	return nil
}
