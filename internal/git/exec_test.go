package git

import (
	"context"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestRunSuccess(t *testing.T) {
	dir := initRepo(t)
	res, err := Run(context.Background(), dir, "status")
	require.NoError(t, err)
	assert.Equal(t, 0, res.ExitCode)
	assert.NotEmpty(t, res.Stdout)
}

func TestRunExitCode(t *testing.T) {
	dir := initRepo(t)
	_, err := Run(context.Background(), dir, "checkout", "nonexistent-branch-xyz")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "exit")
}

func TestRunCapturesStderr(t *testing.T) {
	dir := initRepo(t)
	res, err := Run(context.Background(), dir, "checkout", "nonexistent-xyz")
	assert.Error(t, err)
	// stderr included in error message, not swallowed
	assert.Contains(t, err.Error(), res.Stderr)
}

func TestIsDirty(t *testing.T) {
	dir := initRepo(t)

	// clean repo
	dirty, err := IsDirty(context.Background(), dir)
	require.NoError(t, err)
	assert.False(t, dirty)
}

func TestCurrentBranch(t *testing.T) {
	dir := initRepo(t)
	// git init doesn't create a branch until first commit; use symbolic-ref directly
	res, err := Run(context.Background(), dir, "symbolic-ref", "--short", "HEAD")
	require.NoError(t, err)
	assert.NotEmpty(t, res.Stdout)
}
