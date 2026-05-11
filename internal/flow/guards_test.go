package flow

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vmsfigueredo/gflow/internal/module"
)

func initRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	cmds := [][]string{
		{"git", "init"},
		{"git", "config", "user.email", "test@test.com"},
		{"git", "config", "user.name", "Test"},
	}
	for _, c := range cmds {
		cmd := exec.Command(c[0], c[1:]...)
		cmd.Dir = dir
		require.NoError(t, cmd.Run())
	}
	return dir
}

func commitFile(t *testing.T, dir, name string) {
	t.Helper()
	require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte("x"), 0o644))
	run := func(args ...string) {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		require.NoError(t, cmd.Run())
	}
	run("add", name)
	run("commit", "-m", "add "+name)
}

func mod(path string) *module.Module {
	return &module.Module{Name: "test", Path: path}
}

// CleanTreeGuard

func TestCleanTreeGuard_CleanRepo(t *testing.T) {
	dir := initRepo(t)
	commitFile(t, dir, "readme.txt")
	g := CleanTreeGuard{}
	assert.NoError(t, g.Check(context.Background(), mod(dir)))
}

func TestCleanTreeGuard_DirtyRepo(t *testing.T) {
	dir := initRepo(t)
	commitFile(t, dir, "readme.txt")
	require.NoError(t, os.WriteFile(filepath.Join(dir, "dirty.txt"), []byte("y"), 0o644))
	g := CleanTreeGuard{}
	assert.Error(t, g.Check(context.Background(), mod(dir)))
}

func TestCleanTreeGuard_ForceBypass(t *testing.T) {
	dir := initRepo(t)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "dirty.txt"), []byte("y"), 0o644))
	g := CleanTreeGuard{Force: true}
	assert.NoError(t, g.Check(context.Background(), mod(dir)))
}

// OnExpectedBranchGuard

func TestOnExpectedBranchGuard_Match(t *testing.T) {
	dir := initRepo(t)
	commitFile(t, dir, "f.txt")
	// get current branch name
	cmd := exec.Command("git", "symbolic-ref", "--short", "HEAD")
	cmd.Dir = dir
	out, err := cmd.Output()
	require.NoError(t, err)
	branch := string(out[:len(out)-1]) // trim newline

	g := OnExpectedBranchGuard{Expected: branch}
	assert.NoError(t, g.Check(context.Background(), mod(dir)))
}

func TestOnExpectedBranchGuard_Mismatch(t *testing.T) {
	dir := initRepo(t)
	commitFile(t, dir, "f.txt")
	g := OnExpectedBranchGuard{Expected: "develop"}
	assert.Error(t, g.Check(context.Background(), mod(dir)))
}

// SemverGuard

func TestSemverGuard_Valid(t *testing.T) {
	g := SemverGuard{Name: "1.2.3"}
	assert.NoError(t, g.Check(context.Background(), mod(t.TempDir())))
}

func TestSemverGuard_Invalid(t *testing.T) {
	g := SemverGuard{Name: "not-a-version"}
	assert.Error(t, g.Check(context.Background(), mod(t.TempDir())))
}

func TestSemverGuard_WithVPrefix(t *testing.T) {
	g := SemverGuard{Name: "v2.0.0"}
	assert.NoError(t, g.Check(context.Background(), mod(t.TempDir())))
}

// RunGuards

func TestRunGuards_StopsOnFirstError(t *testing.T) {
	dir := initRepo(t)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "d.txt"), []byte("y"), 0o644))

	guards := []Guard{
		CleanTreeGuard{},                          // will fail — dirty repo
		OnExpectedBranchGuard{Expected: "develop"}, // would also fail, but never reached
	}
	err := RunGuards(context.Background(), mod(dir), guards)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "uncommitted changes")
}
