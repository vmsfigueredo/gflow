package git

import (
	"context"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// makeCommit creates an initial commit in dir so HEAD exists.
func makeCommit(t *testing.T, dir string) {
	t.Helper()
	cmds := [][]string{
		{"git", "commit", "--allow-empty", "-m", "init"},
	}
	for _, c := range cmds {
		cmd := exec.Command(c[0], c[1:]...)
		cmd.Dir = dir
		require.NoError(t, cmd.Run())
	}
}

func TestDetectCurrentFlowBranch_OnFlowBranch(t *testing.T) {
	dir := initRepo(t)
	makeCommit(t, dir)

	// Create and checkout feature/billing.
	cmd := exec.Command("git", "checkout", "-b", "feature/billing")
	cmd.Dir = dir
	require.NoError(t, cmd.Run())

	name, ok, err := DetectCurrentFlowBranch(context.Background(), dir, "feature/")
	require.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, "billing", name)
}

func TestDetectCurrentFlowBranch_NotOnFlowBranch(t *testing.T) {
	dir := initRepo(t)
	makeCommit(t, dir)

	name, ok, err := DetectCurrentFlowBranch(context.Background(), dir, "feature/")
	require.NoError(t, err)
	assert.False(t, ok)
	assert.Equal(t, "", name)
}

func TestDetectCurrentFlowBranch_HotfixPrefix(t *testing.T) {
	dir := initRepo(t)
	makeCommit(t, dir)

	cmd := exec.Command("git", "checkout", "-b", "hotfix/1.2.3")
	cmd.Dir = dir
	require.NoError(t, cmd.Run())

	name, ok, err := DetectCurrentFlowBranch(context.Background(), dir, "hotfix/")
	require.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, "1.2.3", name)
}
