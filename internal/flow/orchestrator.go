package flow

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/vmsfigueredo/gflow/internal/config"
	"github.com/vmsfigueredo/gflow/internal/executor"
	"github.com/vmsfigueredo/gflow/internal/git"
	"github.com/vmsfigueredo/gflow/internal/gitflow"
	"github.com/vmsfigueredo/gflow/internal/module"
)

// Options configures a flow operation.
type Options struct {
	BranchType   string
	Op           string // start | finish | publish | track | delete
	Name         string
	Remote       string
	Parallel     bool
	FailFast     bool
	DryRun       bool
	Debug        bool
	Force        bool
	Stash        bool
	NoAutoCommit bool
}

// Run orchestrates guards → git-flow op → SubmodulePointerSync for each module.
func Run(ctx context.Context, cfg *config.Config, mods []*module.Module, opts Options) ([]executor.Result, error) {
	variant, err := gitflow.DetectVariant(ctx)
	if err != nil {
		return nil, fmt.Errorf("detect git-flow variant: %w", err)
	}

	// find root module (superproject) for pointer sync
	var rootMod *module.Module
	for _, m := range mods {
		if m.Root {
			rootMod = m
			break
		}
	}

	runner := executor.New(cfg, opts.Parallel, opts.FailFast, opts.DryRun, opts.Debug)

	results := runner.Run(ctx, mods, func(_ interface{}, m *module.Module) executor.Result {
		start := time.Now()

		guards := buildGuards(cfg, m, opts)
		if err := RunGuards(ctx, m, guards); err != nil {
			return executor.Result{Module: m, Action: opts.Op, Status: executor.StatusError, Err: err, Duration: time.Since(start)}
		}

		// stash if requested and dirty
		stashed := false
		if opts.Stash {
			dirty, _ := git.IsDirty(ctx, m.Path)
			if dirty {
				if _, err := git.Run(ctx, m.Path, "stash", "push", "-m", "gflow-auto-stash"); err == nil {
					stashed = true
				}
			}
		}

		out, err := gitflow.RunOp(ctx, variant, cfg, m, opts.BranchType, opts.Op, opts.Name, opts.DryRun)

		if stashed {
			_, _ = git.Run(ctx, m.Path, "stash", "pop")
		}

		if err != nil {
			return executor.Result{Module: m, Action: opts.Op, Status: executor.StatusError, Err: err, Duration: time.Since(start)}
		}

		// SubmodulePointerSync: after finish, auto-commit submodule pointer in superproject
		if opts.Op == "finish" && !opts.NoAutoCommit && cfg.Submod.AutoCommit && rootMod != nil && !m.Root {
			subPath := filepath.Join(rootMod.Path, m.Name)
			_ = git.SubmoduleAutoCommit(ctx, rootMod.Path, subPath, m.Name, cfg.Submod.CommitMessage)
		}

		return executor.Result{Module: m, Action: opts.Op, Status: executor.StatusOK, Output: out, Duration: time.Since(start)}
	})

	return results, nil
}

func buildGuards(cfg *config.Config, m *module.Module, opts Options) []Guard {
	var guards []Guard

	// CleanTree guard for destructive ops
	if opts.Op == "start" || opts.Op == "finish" || opts.Op == "checkout" {
		guards = append(guards, CleanTreeGuard{Force: opts.Force, Stash: opts.Stash})
	}

	// base branch guard for start
	if opts.Op == "start" {
		switch opts.BranchType {
		case "feature", "bugfix", "release":
			guards = append(guards, OnExpectedBranchGuard{Expected: "develop"})
		case "hotfix":
			guards = append(guards, onExpectedMainBranch(cfg, m, opts.Remote))
		}
	}

	// remote sync guard before finish
	if opts.Op == "finish" {
		guards = append(guards, RemoteSyncGuard{Remote: opts.Remote, Branch: "develop"})
	}

	// semver guard for release/hotfix start
	if opts.Op == "start" && (opts.BranchType == "release" || opts.BranchType == "hotfix") {
		guards = append(guards, SemverGuard{Name: opts.Name})
	}

	return guards
}

func onExpectedMainBranch(cfg *config.Config, m *module.Module, remote string) Guard {
	main, _ := git.DetectMainBranch(context.Background(), m.Path, remote)
	return OnExpectedBranchGuard{Expected: main}
}
