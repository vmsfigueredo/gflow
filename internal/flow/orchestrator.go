package flow

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"time"

	"github.com/vmsfigueredo/gflow/internal/config"
	"github.com/vmsfigueredo/gflow/internal/executor"
	"github.com/vmsfigueredo/gflow/internal/git"
	"github.com/vmsfigueredo/gflow/internal/gitflow"
	"github.com/vmsfigueredo/gflow/internal/journal"
	"github.com/vmsfigueredo/gflow/internal/module"
)

// Options configures a flow operation.
type Options struct {
	BranchType   string
	Op           string // start | finish | publish | track | delete
	Name         string
	// NamesByModule maps module.Name to per-module branch suffix.
	// When non-nil, modules absent from the map are skipped and Name is ignored.
	NamesByModule map[string]string
	Remote        string
	Parallel      bool
	FailFast      bool
	DryRun        bool
	Debug         bool
	Force         bool
	Stash         bool
	NoAutoCommit  bool
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

	// Snapshot SHAs before op for journal.
	refsBefore := snapshotRefs(ctx, mods)

	runner := executor.New(cfg, opts.Parallel, opts.FailFast, opts.DryRun, opts.Debug)

	results := runner.Run(ctx, mods, func(_ interface{}, m *module.Module) executor.Result {
		start := time.Now()

		// Resolve name: per-module map takes priority over global Name.
		name := opts.Name
		if opts.NamesByModule != nil {
			n, ok := opts.NamesByModule[m.Name]
			if !ok {
				return executor.Result{Module: m, Action: opts.Op, Status: executor.StatusSkip, Duration: time.Since(start)}
			}
			name = n
		}

		// Build and run guards with the resolved per-module name.
		guards := buildGuardsWithName(cfg, m, opts, name)
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

		// hotfix start: auto-stash dirty tree, sync main and develop before branching
		if opts.BranchType == "hotfix" && opts.Op == "start" && !opts.DryRun {
			dirty, _ := git.IsDirty(ctx, m.Path)
			if dirty && !stashed {
				if _, err := git.Run(ctx, m.Path, "stash", "push", "-m", "gflow-hotfix-auto-stash"); err == nil {
					stashed = true
				}
			}
			mainBranch, _ := git.DetectMainBranch(ctx, m.Path, opts.Remote)
			_, _ = git.Run(ctx, m.Path, "checkout", mainBranch)
			_, _ = git.Run(ctx, m.Path, "pull", opts.Remote, mainBranch)
			_, _ = git.Run(ctx, m.Path, "checkout", "develop")
			_, _ = git.Run(ctx, m.Path, "pull", opts.Remote, "develop")
			_, _ = git.Run(ctx, m.Path, "checkout", mainBranch)
		}

		out, err := gitflow.RunOpFn(ctx, variant, cfg, m, opts.BranchType, opts.Op, name, opts.DryRun)

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

	// Write journal entry (best-effort — never fail the op on journal error).
	refsAfter := snapshotRefs(ctx, mods)
	writeJournal(ctx, rootMod, opts, mods, refsBefore, refsAfter, results)

	return results, nil
}

func buildGuardsWithName(cfg *config.Config, m *module.Module, opts Options, name string) []Guard {
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
		// hotfix: no branch guard — orchestrator handles checkout+pull automatically
		}
	}

	// remote sync guard before finish
	if opts.Op == "finish" {
		guards = append(guards, RemoteSyncGuard{Remote: opts.Remote, Branch: "develop"})
	}

	// semver guard for release/hotfix start using per-module name
	if opts.Op == "start" && (opts.BranchType == "release" || opts.BranchType == "hotfix") {
		guards = append(guards, SemverGuard{Name: name})
	}

	return guards
}

func onExpectedMainBranch(cfg *config.Config, m *module.Module, remote string) Guard {
	main, _ := git.DetectMainBranch(context.Background(), m.Path, remote)
	return OnExpectedBranchGuard{Expected: main}
}

func snapshotRefs(ctx context.Context, mods []*module.Module) map[string]string {
	refs := make(map[string]string, len(mods))
	for _, m := range mods {
		sha, _ := git.CurrentSHA(ctx, m.Path)
		refs[m.Name] = sha
	}
	return refs
}

func writeJournal(ctx context.Context, rootMod *module.Module, opts Options, mods []*module.Module,
	before, after map[string]string, results []executor.Result) {
	if opts.DryRun {
		return
	}

	repoRoot := "."
	if rootMod != nil {
		repoRoot = rootMod.Path
	} else if len(mods) > 0 {
		repoRoot = mods[0].Path
	}

	j, err := journal.Open(repoRoot)
	if err != nil {
		return
	}

	modNames := make([]string, len(mods))
	for i, m := range mods {
		modNames[i] = m.Name
	}

	jResults := make([]journal.ModuleResult, 0, len(results))
	for _, r := range results {
		jr := journal.ModuleResult{Module: r.Module.Name, Status: string(r.Status)}
		if r.Err != nil {
			jr.Error = r.Err.Error()
		}
		jResults = append(jResults, jr)
	}

	_ = j.Append(ctx, journal.Entry{
		ID:         newID(),
		Timestamp:  time.Now(),
		Op:         opts.BranchType + " " + opts.Op,
		Modules:    modNames,
		RefsBefore: before,
		RefsAfter:  after,
		Results:    jResults,
	})
}

func newID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
