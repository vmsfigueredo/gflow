package gitflow

import (
	"context"
	"fmt"

	"github.com/vmsfigueredo/gflow/internal/config"
	"github.com/vmsfigueredo/gflow/internal/executor"
	"github.com/vmsfigueredo/gflow/internal/git"
	"github.com/vmsfigueredo/gflow/internal/module"
)

// Init runs git flow init and submodule bootstrap for all modules.
func Init(ctx context.Context, cfg *config.Config, mods []*module.Module, dryRun, debug bool) ([]executor.Result, error) {
	runner := executor.New(cfg, false, false, dryRun, debug)
	results := runner.Run(ctx, mods, func(_ interface{}, m *module.Module) executor.Result {
		return initModule(ctx, cfg, m, dryRun)
	})
	return results, nil
}

func initModule(ctx context.Context, cfg *config.Config, m *module.Module, dryRun bool) executor.Result {
	if dryRun {
		return executor.Result{Module: m, Action: "init", Status: executor.StatusDryRun,
			Output: "git flow init -d (with prefixes from config)"}
	}

	// init git-flow with real prefixes from config
	args := []string{
		"init", "-d",
		"--feature", cfg.Gitflow.FeaturePrefix,
		"--hotfix", cfg.Gitflow.HotfixPrefix,
		"--bugfix", cfg.Gitflow.BugfixPrefix,
		"--release", cfg.Gitflow.ReleasePrefix,
		"--support", cfg.Gitflow.SupportPrefix,
		"--tag", cfg.Gitflow.VersionTagPrefix,
	}

	out, err := runGitFlow(ctx, m.Path, args...)
	if err != nil {
		return executor.Result{Module: m, Action: "init", Status: executor.StatusError,
			Err: fmt.Errorf("git flow init in %s: %w", m.Name, err)}
	}

	// submodule init/update for the root module
	if m.Root {
		if err := git.UpdateInit(ctx, m.Path); err != nil {
			return executor.Result{Module: m, Action: "init", Status: executor.StatusError,
				Err: fmt.Errorf("submodule update in %s: %w", m.Name, err)}
		}
	}

	return executor.Result{Module: m, Action: "init", Status: executor.StatusOK, Output: out}
}
