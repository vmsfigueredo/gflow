package flow

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vmsfigueredo/gflow/internal/config"
	"github.com/vmsfigueredo/gflow/internal/executor"
	"github.com/vmsfigueredo/gflow/internal/gitflow"
	"github.com/vmsfigueredo/gflow/internal/module"
)

// makeModules builds a slice of test modules backed by temp git repos.
func makeModules(t *testing.T, names []string) ([]*module.Module, *config.Config) {
	t.Helper()
	cfg := config.Config{}
	cfg.Gitflow.FeaturePrefix = "feature/"
	cfg.Gitflow.HotfixPrefix = "hotfix/"
	cfg.Gitflow.BugfixPrefix = "bugfix/"
	cfg.Gitflow.ReleasePrefix = "release/"
	cfg.Submod.AutoCommit = false

	mods := make([]*module.Module, len(names))
	for i, n := range names {
		dir := initRepo(t)
		mods[i] = &module.Module{Name: n, Path: dir, Root: i == 0}
	}
	return mods, &cfg
}

func TestRunNamesByModule_SkipsAbsentModules(t *testing.T) {
	mods, cfg := makeModules(t, []string{"core", "api", "web"})

	// Inject a stub that records calls without running real git flow.
	var called []string
	orig := gitflow.RunOpFn
	t.Cleanup(func() { gitflow.RunOpFn = orig })
	gitflow.RunOpFn = func(_ context.Context, _ gitflow.Variant, _ *config.Config, m *module.Module, _, _, name, _ string, _ bool) (string, error) {
		called = append(called, m.Name+"="+name)
		return "", nil
	}

	// Use "publish" op — no guards fire (no branch check, no clean-tree, no remote sync).
	opts := Options{
		BranchType: "feature",
		Op:         "publish",
		NamesByModule: map[string]string{
			"api": "billing",
		},
	}

	results, err := Run(context.Background(), cfg, mods, opts)
	require.NoError(t, err)

	// Only "api" should have been called.
	assert.Equal(t, []string{"api=billing"}, called)

	// Result statuses: core=skip, api=ok, web=skip
	statusByName := map[string]executor.ResultStatus{}
	for _, r := range results {
		statusByName[r.Module.Name] = r.Status
	}
	assert.Equal(t, executor.StatusSkip, statusByName["core"])
	assert.Equal(t, executor.StatusOK, statusByName["api"])
	assert.Equal(t, executor.StatusSkip, statusByName["web"])
}

func TestRunSingleName_AllModulesRun(t *testing.T) {
	mods, cfg := makeModules(t, []string{"core", "api"})

	var called []string
	orig := gitflow.RunOpFn
	t.Cleanup(func() { gitflow.RunOpFn = orig })
	gitflow.RunOpFn = func(_ context.Context, _ gitflow.Variant, _ *config.Config, m *module.Module, _, _, name, _ string, _ bool) (string, error) {
		called = append(called, m.Name)
		return "", nil
	}

	// "publish" has no guards — clean for unit testing.
	opts := Options{
		BranchType: "feature",
		Op:         "publish",
		Name:       "billing",
	}

	results, err := Run(context.Background(), cfg, mods, opts)
	require.NoError(t, err)

	assert.Len(t, called, 2)

	for _, r := range results {
		assert.Equal(t, executor.StatusOK, r.Status, "module %s", r.Module.Name)
	}
}
