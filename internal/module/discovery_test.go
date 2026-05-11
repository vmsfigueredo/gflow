package module

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vmsfigueredo/gflow/internal/config"
)

func cfg(modules []string, aliases map[string][]string) *config.Config {
	return &config.Config{Modules: modules, Aliases: aliases}
}

func TestResolveExplicitModules(t *testing.T) {
	dir := t.TempDir()
	// create module dirs so path resolution is meaningful
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "api"), 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "web"), 0o755))

	mods, err := Resolve(cfg([]string{"api", "web"}, nil), dir, "", false)
	require.NoError(t, err)

	names := names(mods)
	assert.Equal(t, []string{".", "api", "web"}, names)
}

func TestResolveNoRoot(t *testing.T) {
	dir := t.TempDir()
	mods, err := Resolve(cfg([]string{"api"}, nil), dir, "", true)
	require.NoError(t, err)
	assert.NotContains(t, names(mods), ".")
}

func TestResolveProjectAlias(t *testing.T) {
	dir := t.TempDir()
	aliases := map[string][]string{"backend": {"api", "services/auth"}}
	mods, err := Resolve(cfg([]string{"api", "web", "services/auth"}, aliases), dir, "backend", false)
	require.NoError(t, err)

	n := names(mods)
	assert.Contains(t, n, "api")
	assert.Contains(t, n, "auth")
	assert.NotContains(t, n, "web")
}

func TestResolveFromGitmodules(t *testing.T) {
	dir := t.TempDir()
	gitmodules := `[submodule "api"]
	path = api
	url = ../api.git
[submodule "web"]
	path = web
	url = ../web.git
`
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".gitmodules"), []byte(gitmodules), 0o644))

	mods, err := Resolve(cfg(nil, nil), dir, "", false)
	require.NoError(t, err)

	n := names(mods)
	assert.Contains(t, n, "api")
	assert.Contains(t, n, "web")
}

func TestResolveNoConfigNoGitmodules(t *testing.T) {
	dir := t.TempDir()
	mods, err := Resolve(cfg(nil, nil), dir, "", false)
	require.NoError(t, err)
	// only root
	assert.Equal(t, []string{"."}, names(mods))
}

func names(mods []*Module) []string {
	out := make([]string, len(mods))
	for i, m := range mods {
		out[i] = m.Name
	}
	return out
}
