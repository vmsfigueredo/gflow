package gitflow

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// runGitFlow executes `git flow <args>` in dir and returns stdout.
func runGitFlow(ctx context.Context, dir string, args ...string) (string, error) {
	full := append([]string{"flow"}, args...)
	cmd := exec.CommandContext(ctx, "git", full...)
	if dir != "" {
		cmd.Dir = dir
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if exit, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("git flow %s: exit %d: %s",
				strings.Join(args, " "), exit.ExitCode(), strings.TrimSpace(stderr.String()))
		}
		return "", fmt.Errorf("git flow %s: %w", strings.Join(args, " "), err)
	}
	return strings.TrimRight(stdout.String(), "\n"), nil
}
