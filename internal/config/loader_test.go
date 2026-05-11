package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadYAML(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "gflow.yaml", `
modules:
  - api
  - web
parallel: true
remote: upstream
gitflow:
  finish_options: "--keep-branch"
  feature_prefix: "feat/"
submodule:
  auto_commit: false
aliases:
  backend:
    - api
    - services/auth
branch_exists_policy: skip
`)
	cfg, err := Load(dir)
	require.NoError(t, err)

	assert.Equal(t, []string{"api", "web"}, cfg.Modules)
	assert.True(t, cfg.Parallel)
	assert.Equal(t, "upstream", cfg.Remote)
	assert.Equal(t, "--keep-branch", cfg.Gitflow.FinishOptions)
	assert.Equal(t, "feat/", cfg.Gitflow.FeaturePrefix)
	assert.False(t, cfg.Submod.AutoCommit)
	assert.Equal(t, []string{"api", "services/auth"}, cfg.Aliases["backend"])
	assert.Equal(t, "skip", cfg.BranchExistsPolicy)
}

func TestLoadBashCompat(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, ".gflow.conf", `
# gflow config
MODULES=(api web services/auth)
REMOTE_NAME=upstream
PARALLEL_EXECUTION=true
DEBUG=false
add_path_alias "backend" "api services/auth"
add_path_alias "frontend" "web"
`)
	cfg, err := Load(dir)
	require.NoError(t, err)

	assert.Equal(t, []string{"api", "web", "services/auth"}, cfg.Modules)
	assert.Equal(t, "upstream", cfg.Remote)
	assert.True(t, cfg.Parallel)
	assert.Equal(t, []string{"api", "services/auth"}, cfg.Aliases["backend"])
	assert.Equal(t, []string{"web"}, cfg.Aliases["frontend"])
}

func TestLoadBashMultilineArray(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, ".gflow.conf", `
# gflow configuration file
MODULES=(
    "central/api"
    "hagape/api"
    "hagape/web"
)
REMOTE_NAME=upstream
add_path_alias "central" "central/api"
add_path_alias "api" "hagape/api"
`)
	cfg, err := Load(dir)
	require.NoError(t, err)

	assert.Equal(t, []string{"central/api", "hagape/api", "hagape/web"}, cfg.Modules)
	assert.Equal(t, "upstream", cfg.Remote)
	assert.Equal(t, []string{"central/api"}, cfg.Aliases["central"])
}

func TestLoadBashPhantomVarsIgnored(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, ".gflow.conf", `
MODULES=(api)
BRANCH_TYPES=(feature hotfix)
GITFLOW_VARIANT=avh
FINISH_OPTIONS=--keep-branch
GIT_COMMANDS=(git)
`)
	cfg, err := Load(dir)
	require.NoError(t, err)
	assert.Equal(t, []string{"api"}, cfg.Modules)
}

func TestLoadBashUnsupportedShape(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, ".gflow.conf", `
MODULES=(api)
SOME_VAR=$(echo hi)
`)
	_, err := Load(dir)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported .gflow.conf shape")
	assert.Contains(t, err.Error(), "gflow config migrate")
}

func TestLoadDefaults(t *testing.T) {
	dir := t.TempDir() // no config file
	cfg, err := Load(dir)
	require.NoError(t, err)

	assert.Equal(t, "origin", cfg.Remote)
	assert.Equal(t, "feature/", cfg.Gitflow.FeaturePrefix)
	assert.True(t, cfg.Submod.AutoCommit)
	assert.Equal(t, "checkout", cfg.BranchExistsPolicy)
}

func TestYAMLTakesPrecedenceOverBash(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "gflow.yaml", `modules: [yaml-only]`)
	write(t, dir, ".gflow.conf", `MODULES=(bash-only)`)

	cfg, err := Load(dir)
	require.NoError(t, err)
	assert.Equal(t, []string{"yaml-only"}, cfg.Modules)
}

func TestMigrate(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, ".gflow.conf", `
MODULES=(api web)
REMOTE_NAME=origin
add_path_alias "be" "api"
`)
	require.NoError(t, Migrate(dir))

	cfg, err := Load(dir) // should now load gflow.yaml
	require.NoError(t, err)
	assert.Equal(t, []string{"api", "web"}, cfg.Modules)
	assert.Equal(t, []string{"api"}, cfg.Aliases["be"])
}

func TestValidate(t *testing.T) {
	cfg := defaults()
	assert.NoError(t, Validate(&cfg))

	cfg.BranchExistsPolicy = "invalid"
	assert.Error(t, Validate(&cfg))

	cfg.BranchExistsPolicy = "skip"
	cfg.Workers = -1
	assert.Error(t, Validate(&cfg))
}

func write(t *testing.T, dir, name, content string) {
	t.Helper()
	require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644))
}
